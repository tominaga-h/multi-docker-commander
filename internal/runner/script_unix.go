//go:build !windows

package runner

import (
	"os/exec"
	"runtime"
)

// newScriptCommand wraps a shell command with the `script` utility so that
// the child process runs inside a PTY. This preserves ANSI color codes in
// the log file, because the child sees a terminal as its stdout.
//
//   - macOS (BSD script): script -qF <logfile> sh -c "<command>"
//   - Linux (util-linux): script -qf -c "<command>" <logfile>
//     -f flushes after each write for real-time streaming.
func newScriptCommand(command, dir, logFile string) *exec.Cmd {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("script", "-qF", logFile, "sh", "-c", command)
	default:
		cmd = exec.Command("script", "-qf", "-c", command, logFile)
	}
	cmd.Dir = dir
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd
}
