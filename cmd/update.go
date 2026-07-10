package cmd

import (
	"fmt"
	"os"

	"mdc/internal/logger"
	"mdc/internal/updater"
	"mdc/internal/version"

	"github.com/spf13/cobra"
)

var updateFunc = updater.Update

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update mdc to the latest release from GitHub",
	Long:  "Download the latest Linux amd64 binary from GitHub Releases and replace the running executable.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runUpdate(); err != nil {
			logger.Error("update", "mdc update", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

// runUpdate performs the update and prints the outcome. It returns any error
// so the caller (or tests) can handle it; user-facing output goes through
// logger.
func runUpdate() error {
	result, err := updateFunc(updater.Options{})
	if err != nil {
		return err
	}

	switch {
	case !result.Skipped:
		logger.Info("update", fmt.Sprintf("updated to %s", result.Version))
	case result.LocalNewer:
		logger.Info("update", fmt.Sprintf("local version v%s is ahead of the latest release %s",
			updater.DisplayVersion(version.Version), result.Version))
	default:
		logger.Info("update", fmt.Sprintf("already up to date (%s)", result.Version))
	}
	return nil
}

// setUpdateFunc replaces the update implementation for tests.
func setUpdateFunc(fn func(updater.Options) (updater.Result, error)) {
	updateFunc = fn
}

// resetUpdateFunc restores the default update implementation.
func resetUpdateFunc() {
	updateFunc = updater.Update
}
