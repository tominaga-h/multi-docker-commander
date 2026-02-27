package cmd

import (
	"github.com/spf13/cobra"
)

var upDryRun bool

var upCmd = &cobra.Command{
	Use:   "up [config-name]",
	Short: "Start all projects defined in a config",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		loadAndRun(args[0], "up", upDryRun)
	},
}

func init() {
	upCmd.Flags().BoolVar(&upDryRun, "dry-run", false, "Print execution plan without running commands")
	rootCmd.AddCommand(upCmd)
}
