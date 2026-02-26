package cmd

import (
	"fmt"
	"os"

	"mdc/internal/config"
	"mdc/internal/logger"
	"mdc/internal/pidfile"
	"mdc/internal/runner"

	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down [config-name]",
	Short: "Stop all projects defined in a config",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configName := args[0]
		cfg, err := config.Load(configName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := runner.Run(cfg, "down", configName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		projects, _ := pidfile.LoadAll(configName)
		for projectName, entries := range projects {
			for _, e := range entries {
				logger.Stop(projectName, e.Command, e.PID)
			}
		}
		if err := pidfile.KillAll(configName); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: failed to clean up background processes: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
