package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"mdc/internal/logger"
	"mdc/internal/pidfile"

	"github.com/spf13/cobra"
)

var procStopCmd = &cobra.Command{
	Use:   "stop <PID>",
	Short: "Stop a background process by PID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pid, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid PID: %s\n", args[0])
			os.Exit(1)
		}

		configName, projectName, entry, err := pidfile.FindByPID(pid)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		logger.Stop(projectName, entry.Command, pid)

		_ = pidfile.GracefulKill(pid, 10*time.Second)

		if err := pidfile.RemoveEntry(configName, projectName, pid); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: failed to remove PID entry: %v\n", err)
		}

		logger.Stopped(projectName)
	},
}

func init() {
	procCmd.AddCommand(procStopCmd)
}
