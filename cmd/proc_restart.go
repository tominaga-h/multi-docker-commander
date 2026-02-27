package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"mdc/internal/logger"
	"mdc/internal/pidfile"
	"mdc/internal/runner"

	"github.com/spf13/cobra"
)

var procRestartCmd = &cobra.Command{
	Use:   "restart <PID>",
	Short: "Restart a background process by PID",
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

		tmpLog, _ := pidfile.ProcLogTmpPath(configName, projectName)
		newPID, err := runner.StartBackgroundProcess(entry.Command, entry.Dir, tmpLog)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ [%s] Failed to restart %q: %v\n", projectName, entry.Command, err)
			_ = pidfile.RemoveEntry(configName, projectName, pid)
			os.Exit(1)
		}
		if _, err := pidfile.RenameProcLog(tmpLog, newPID); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: log rename failed: %v\n", err)
		}

		if err := pidfile.RemoveEntry(configName, projectName, pid); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: failed to remove old PID entry: %v\n", err)
		}

		if err := pidfile.Append(configName, projectName, pidfile.Entry{
			PID:     newPID,
			Command: entry.Command,
			Dir:     entry.Dir,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: failed to save new PID entry: %v\n", err)
		}

		logger.Background(projectName, entry.Command, newPID)
	},
}

func init() {
	procCmd.AddCommand(procRestartCmd)
}
