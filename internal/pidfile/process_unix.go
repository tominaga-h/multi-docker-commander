//go:build !windows

package pidfile

import (
	"os"
	"syscall"
)

func IsRunning(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}
