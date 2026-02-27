//go:build windows

package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// newScriptCommand on Windows falls back to a plain shell command with
// stdout/stderr redirected to the log file, since the `script` utility
// is not available. ANSI color codes will not be preserved.
func newScriptCommand(command, dir, logFile string) *exec.Cmd {
	if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
		cmd := exec.Command("cmd", "/c", command)
		cmd.Dir = dir
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
		return cmd
	}

	redirect := fmt.Sprintf("%s > %q 2>&1", command, logFile)
	cmd := exec.Command("cmd", "/c", redirect)
	cmd.Dir = dir
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd
}
