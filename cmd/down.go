package cmd

import (
	"fmt"
	"os"

	"mdc/internal/logger"
	"mdc/internal/pidfile"

	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down [config-name]",
	Short: "Stop all projects defined in a config",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configName := args[0]
		loadAndRun(configName, "down")

		if err := pidfile.KillAllWithCallback(configName, logger.Stop); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: failed to clean up background processes: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
