//go:build windows

package runner

import (
	"os"
	"os/exec"
)

func hasPTYSupport() bool { return false }

func isTerminal(_ *os.File) bool { return false }

func execWithPTY(_ *exec.Cmd, _ bool) (string, error) {
	panic("execWithPTY called on unsupported platform")
}
