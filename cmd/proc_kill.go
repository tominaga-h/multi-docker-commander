package cmd

import (
	"fmt"
	"os"
	"time"

	"mdc/internal/logger"
	"mdc/internal/pidfile"

	"github.com/spf13/cobra"
)

var (
	procKillConfigName string
	procKillPID        int
	procKillAll        bool
	procKillDead       bool
)

var procKillCmd = &cobra.Command{
	Use:   "kill",
	Short: "Kill background processes by config name, PID, or all configs",
	Long: `Kill background processes tracked by mdc.

Use -c to kill all processes belonging to a config, -p to kill a single process by PID,
or --all to kill all tracked processes across all configs.
Use --dead to remove tracked entries whose process is no longer running without
killing any live processes; it may be combined with -c (single config) or used
alone/with --all (all configs).
This command can be used in YAML down commands; the runner automatically adds -c <config-name>.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		hasConfig := cmd.Flags().Changed("config")
		hasPID := cmd.Flags().Changed("pid")
		hasAll := cmd.Flags().Changed("all")
		hasDead := cmd.Flags().Changed("dead")

		if err := validateKillFlags(hasConfig, hasPID, hasAll, hasDead); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if hasDead {
			removeDead(hasConfig, procKillConfigName)
			return
		}

		if hasPID {
			killByPID(procKillPID)
		} else if hasAll {
			killAllConfigs()
		} else {
			killByConfig(procKillConfigName)
		}
	},
}

// validateKillFlags validates the combination of kill flags.
//
// In non-dead mode, exactly one of -c / -p / --all must be set and they
// cannot be combined. In dead mode (--dead), -c may be combined (clean a
// single config) and --all may be combined (clean all configs), but -p is
// rejected since removing dead entries does not target a single PID.
func validateKillFlags(hasConfig, hasPID, hasAll, hasDead bool) error {
	if hasDead {
		if hasPID {
			return fmt.Errorf("error: --dead cannot be combined with -p")
		}
		return nil
	}

	selected := 0
	if hasConfig {
		selected++
	}
	if hasPID {
		selected++
	}
	if hasAll {
		selected++
	}

	if selected == 0 {
		return fmt.Errorf("error: specify one of -c <config-name>, -p <PID>, or --all")
	}
	if selected > 1 {
		return fmt.Errorf("error: -c, -p, and --all cannot be used together")
	}
	return nil
}

func removeDead(hasConfig bool, configName string) {
	var (
		removed int
		err     error
	)
	if hasConfig {
		removed, err = pidfile.RemoveDead(configName)
	} else {
		removed, err = pidfile.RemoveDeadAllConfigs()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: failed to clean up dead processes: %v\n", err)
	}
	logger.DeadCleaned(removed)
}

func killByPID(pid int) {
	configName, projectName, entry, err := pidfile.FindByPID(pid)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	logger.Stop(projectName, entry.Command, pid)

	_ = pidfile.GracefulKill(pid, 10*time.Second)

	if err := pidfile.RemoveEntry(configName, projectName, pid); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: failed to remove PID entry: %v\n", err)
	}

	logger.Stopped(projectName)
}

func killByConfig(configName string) {
	if err := pidfile.KillAllWithCallback(configName, logger.Stop); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: failed to kill processes: %v\n", err)
	}
}

func killAllConfigs() {
	if err := pidfile.KillAllConfigsWithCallback(logger.Stop); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: failed to kill processes: %v\n", err)
	}
}

func init() {
	procKillCmd.Flags().StringVarP(&procKillConfigName, "config", "c", "", "Config name to kill all processes for")
	procKillCmd.Flags().IntVarP(&procKillPID, "pid", "p", 0, "PID of the process to kill")
	procKillCmd.Flags().BoolVar(&procKillAll, "all", false, "Kill all tracked processes across all configs")
	procKillCmd.Flags().BoolVar(&procKillDead, "dead", false, "Remove tracked entries whose process is no longer running (does not kill live processes)")
	procCmd.AddCommand(procKillCmd)
}
