//go:build windows

package runner

import "os/exec"

func setSysProcAttr(cmd *exec.Cmd) {
	// On Windows, child processes are already independent from the parent
	// process group by default. No additional configuration is needed.
}
