package cmd

import (
	"fmt"
	"os"

	"mdc/internal/config"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init <config-name>",
	Short: "Create a new YAML config template in ~/.config/mdc/",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path, err := config.CreateDefaultConfig(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("Created %s\n", path)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
