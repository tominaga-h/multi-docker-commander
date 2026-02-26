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
)

func Run(cfg *config.Config, action string) error {
	commands, err := commandsForAction(cfg, action)
	if err != nil {
		return err
	}

	switch cfg.ExecutionMode {
	case "sequential":
		return runSequential(cfg.Projects, commands)
	case "parallel":
		return runParallel(cfg.Projects, commands)
	default:
		return fmt.Errorf("unknown execution_mode: %q", cfg.ExecutionMode)
	}
}

func commandsForAction(cfg *config.Config, action string) ([][]string, error) {
	result := make([][]string, len(cfg.Projects))
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

func runSequential(projects []config.Project, commands [][]string) error {
	for i, p := range projects {
		if err := validateProjectPath(p); err != nil {
			return err
		}
		for _, cmdStr := range commands[i] {
			logger.Start(p.Name, cmdStr)

			cmd := newShellCommand(cmdStr, p.Path)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				logger.Error(p.Name, cmdStr, err)
				return fmt.Errorf("project %q: command %q failed: %w", p.Name, cmdStr, err)
			}
			logger.Success(p.Name, cmdStr)
		}
		logger.ProjectDone(p.Name)
	}
	return nil
}

func runParallel(projects []config.Project, commands [][]string) error {
	var wg sync.WaitGroup
	errs := make([]error, len(projects))

	for _, p := range projects {
		if err := validateProjectPath(p); err != nil {
			return err
		}
	}

	for i, p := range projects {
		wg.Add(1)
		go func(idx int, proj config.Project, cmds []string) {
			defer wg.Done()
			errs[idx] = runProjectBuffered(proj, cmds)
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

func runProjectBuffered(p config.Project, cmds []string) error {
	for _, cmdStr := range cmds {
		logger.Start(p.Name, cmdStr)

		var stdoutBuf, stderrBuf bytes.Buffer
		cmd := newShellCommand(cmdStr, p.Path)
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf

		if err := cmd.Run(); err != nil {
			logger.Error(p.Name, cmdStr, err)
			combined := stderrBuf.String() + stdoutBuf.String()
			logger.Output(p.Name, combined)
			return fmt.Errorf("project %q: command %q failed: %w", p.Name, cmdStr, err)
		}
		logger.Success(p.Name, cmdStr)
	}
	logger.ProjectDone(p.Name)
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
