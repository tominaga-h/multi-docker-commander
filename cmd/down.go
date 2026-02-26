package cmd

import (
	"fmt"
	"os"

	"mdc/internal/config"
	"mdc/internal/runner"

	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down [config-name]",
	Short: "Stop all projects defined in a config",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := runner.Run(cfg, "down"); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
