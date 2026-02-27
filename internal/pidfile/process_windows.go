//go:build windows

package pidfile

import (
	"os"
	"time"
)

func IsRunning(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Windows, Signal only supports os.Kill and os.Interrupt.
	// If Kill returns "process already finished", the process is dead.
	err = p.Signal(os.Kill)
	if err == nil {
		return true
	}
	return false
}

// GracefulKill on Windows falls back to Kill since SIGTERM is not supported.
func GracefulKill(pid int, _ time.Duration) error {
	if !IsRunning(pid) {
		return nil
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	_ = p.Kill()
	_, _ = p.Wait()
	return nil
}
