package cmd

import (
	"fmt"
	"os"

	"mdc/internal/config"
	"mdc/internal/runner"
	"mdc/internal/version"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "mdc",
	Version: version.Version,
	Short:   "Multi-Docker-Compose — manage multiple repos with one command",
	Long: `mdc is a CLI tool that manages Docker environments across multiple
repositories. Define your projects in a YAML config file and run
"mdc up" or "mdc down" to start/stop them all at once.`,
}

func init() {
	rootCmd.InitDefaultVersionFlag()
	rootCmd.Flags().Lookup("version").Shorthand = "v"
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadAndRun(configName, action string) {
	cfg, err := config.Load(configName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := runner.Run(cfg, action, configName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
