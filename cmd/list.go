package cmd

import (
	"fmt"
	"os"

	"mdc/internal/config"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List config YAML files in ~/.config/mdc/",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		files, err := config.ListConfigs()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if len(files) == 0 {
			fmt.Println("No config files found in ~/.config/mdc/")
			return
		}
		fmt.Println("Config YAML Files:")
		for _, f := range files {
			fmt.Printf("  - %s\n", f)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
