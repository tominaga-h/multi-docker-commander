package cmd

import (
	"fmt"
	"os"

	"mdc/internal/logger"
	"mdc/internal/pidfile"

	"github.com/spf13/cobra"
)

var downDryRun bool

var downCmd = &cobra.Command{
	Use:   "down [config-name]",
	Short: "Stop all projects defined in a config",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configName := args[0]
		loadAndRun(configName, "down", downDryRun)

		if downDryRun {
			printDryRunStopEntries(configName)
			return
		}

		if err := pidfile.KillAllWithCallback(configName, logger.Stop); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: failed to clean up background processes: %v\n", err)
		}
	},
}

func printDryRunStopEntries(configName string) {
	projects, err := pidfile.LoadAll(configName)
	if err != nil || len(projects) == 0 {
		return
	}
	logger.DryRunStopHeader()
	for projectName, entries := range projects {
		for _, e := range entries {
			if pidfile.IsRunning(e.PID) {
				logger.DryRunStopEntry(projectName, e.Command, e.PID)
			}
		}
	}
}

func init() {
	downCmd.Flags().BoolVar(&downDryRun, "dry-run", false, "Print execution plan without running commands")
	rootCmd.AddCommand(downCmd)
}
