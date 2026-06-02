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
	Short:   "Multi-Docker-Commander — manage multiple repos with one command",
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

// loadAndRun loads the config and dispatches to the runner.
// It returns any error encountered instead of exiting, so callers
// (e.g. down) can perform cleanup even when the run fails. Callers are
// responsible for printing the error and exiting.
func loadAndRun(configName, action string, dryRun bool) error {
	cfg, err := config.Load(configName)
	if err != nil {
		return err
	}
	if dryRun {
		return runner.DryRun(cfg, action)
	}
	return runner.Run(cfg, action, configName)
}
