package runner

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"mdc/internal/config"
	"mdc/internal/logger"
	"mdc/internal/pidfile"
)

func Run(cfg *config.Config, action string, configName string) error {
	commands, err := commandsForAction(cfg, action)
	if err != nil {
		return err
	}

	switch cfg.ExecutionMode {
	case "sequential":
		return runSequential(cfg.Projects, commands, configName)
	case "parallel":
		return runParallel(cfg.Projects, commands, configName)
	default:
		return fmt.Errorf("unknown execution_mode: %q", cfg.ExecutionMode)
	}
}

func commandsForAction(cfg *config.Config, action string) ([][]config.CommandItem, error) {
	result := make([][]config.CommandItem, len(cfg.Projects))
	for i, p := range cfg.Projects {
		switch action {
		case "up":
			result[i] = p.Commands.Up
		case "down":
			result[i] = p.Commands.Down
		default:
			return nil, fmt.Errorf("unknown action: %q", action)
		}
		if len(result[i]) == 0 {
			return nil, fmt.Errorf("project %q: no commands defined for %q", p.Name, action)
		}
	}
	return result, nil
}

func runSequential(projects []config.Project, commands [][]config.CommandItem, configName string) error {
	for i, p := range projects {
		if err := validateProjectPath(p); err != nil {
			return err
		}
		for _, item := range commands[i] {
			if err := execCommand(p, item, configName, false); err != nil {
				return err
			}
		}
		logger.ProjectDone(p.Name)
	}
	return nil
}

func runParallel(projects []config.Project, commands [][]config.CommandItem, configName string) error {
	var wg sync.WaitGroup
	errs := make([]error, len(projects))

	for _, p := range projects {
		if err := validateProjectPath(p); err != nil {
			return err
		}
	}

	for i, p := range projects {
		wg.Add(1)
		go func(idx int, proj config.Project, cmds []config.CommandItem) {
			defer wg.Done()
			errs[idx] = runProjectBuffered(proj, cmds, configName)
		}(i, p, commands[i])
	}

	wg.Wait()

	var failures []string
	for _, err := range errs {
		if err != nil {
			failures = append(failures, err.Error())
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("some projects failed:\n  %s", strings.Join(failures, "\n  "))
	}
	return nil
}

func runProjectBuffered(p config.Project, cmds []config.CommandItem, configName string) error {
	for _, item := range cmds {
		if err := execCommand(p, item, configName, true); err != nil {
			return err
		}
	}
	logger.ProjectDone(p.Name)
	return nil
}

func execCommand(p config.Project, item config.CommandItem, configName string, buffered bool) error {
	logger.Start(p.Name, item.Command)

	cmd := newShellCommand(item.Command, p.Path)

	if item.Background {
		setSysProcAttr(cmd)
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil

		if err := cmd.Start(); err != nil {
			logger.Error(p.Name, item.Command, err)
			return fmt.Errorf("project %q: background command %q failed to start: %w", p.Name, item.Command, err)
		}

		if err := pidfile.Append(configName, p.Name, pidfile.Entry{
			PID:     cmd.Process.Pid,
			Command: item.Command,
			Dir:     p.Path,
		}); err != nil {
			return fmt.Errorf("project %q: failed to save PID: %w", p.Name, err)
		}
		logger.Background(p.Name, item.Command, cmd.Process.Pid)
		return nil
	}

	if hasPTYSupport() && isTerminal(os.Stdout) {
		if !buffered {
			logger.Border()
		}
		output, err := execWithPTY(cmd, buffered)
		if !buffered {
			logger.Border()
		}
		if err != nil {
			logger.Error(p.Name, item.Command, err)
			if buffered && output != "" {
				logger.Output(p.Name, output)
			}
			return fmt.Errorf("project %q: command %q failed: %w", p.Name, item.Command, err)
		}
		logger.Success(p.Name, item.Command)
		return nil
	}

	if buffered {
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf

		if err := cmd.Run(); err != nil {
			logger.Error(p.Name, item.Command, err)
			combined := stderrBuf.String() + stdoutBuf.String()
			logger.Output(p.Name, combined)
			return fmt.Errorf("project %q: command %q failed: %w", p.Name, item.Command, err)
		}
	} else {
		logger.Border()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		logger.Border()

		if err != nil {
			logger.Error(p.Name, item.Command, err)
			return fmt.Errorf("project %q: command %q failed: %w", p.Name, item.Command, err)
		}
	}

	logger.Success(p.Name, item.Command)
	return nil
}

func newShellCommand(cmdStr, dir string) *exec.Cmd {
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	return cmd
}

func validateProjectPath(p config.Project) error {
	info, err := os.Stat(p.Path)
	if err != nil {
		return fmt.Errorf("project %q: path %q does not exist: %w", p.Name, p.Path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("project %q: path %q is not a directory", p.Name, p.Path)
	}
	return nil
}
