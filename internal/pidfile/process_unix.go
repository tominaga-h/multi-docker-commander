//go:build !windows

package pidfile

import (
	"os"
	"syscall"
	"time"
)

func IsRunning(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}

// GracefulKill sends SIGTERM, polls at 100ms intervals until the process
// exits or the timeout elapses, then falls back to SIGKILL.
func GracefulKill(pid int, timeout time.Duration) error {
	if !IsRunning(pid) {
		return nil
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}

	if err := p.Signal(syscall.SIGTERM); err != nil {
		return nil
	}

	deadline := time.After(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !IsRunning(pid) {
				return nil
			}
		case <-deadline:
			_ = p.Kill()
			_, _ = p.Wait()
			return nil
		}
	}
}
