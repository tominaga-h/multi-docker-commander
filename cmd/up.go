package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var upDryRun bool

var upCmd = &cobra.Command{
	Use:   "up [config-name]",
	Short: "Start all projects defined in a config",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := loadAndRun(args[0], "up", upDryRun); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func init() {
	upCmd.Flags().BoolVar(&upDryRun, "dry-run", false, "Print execution plan without running commands")
	rootCmd.AddCommand(upCmd)
}
