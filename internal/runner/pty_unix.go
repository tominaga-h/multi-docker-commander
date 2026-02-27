//go:build !windows

package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty/v2"
)

func hasPTYSupport() bool { return true }

func isTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func execWithPTY(cmd *exec.Cmd, buffered bool) (output string, err error) {
	ptmx, tty, err := pty.Open()
	if err != nil {
		return "", fmt.Errorf("pty open: %w", err)
	}
	defer func() { _ = ptmx.Close() }()

	if sz, err := pty.GetsizeFull(os.Stdout); err == nil {
		_ = pty.Setsize(ptmx, sz)
	}

	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty

	if err := cmd.Start(); err != nil {
		_ = tty.Close()
		return "", fmt.Errorf("start: %w", err)
	}
	_ = tty.Close()

	if buffered {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, ptmx)
		waitErr := cmd.Wait()
		return buf.String(), waitErr
	}

	go func() {
		_, _ = io.Copy(ptmx, os.Stdin)
	}()
	_, _ = io.Copy(os.Stdout, ptmx)

	waitErr := cmd.Wait()
	return "", waitErr
}
