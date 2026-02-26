//go:build windows

package pidfile

import "os"

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
