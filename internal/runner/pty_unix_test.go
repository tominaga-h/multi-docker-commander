//go:build !windows

package runner

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestIsTerminal(t *testing.T) {
	t.Run("stdout in test is not a terminal", func(t *testing.T) {
		if isTerminal(os.Stdout) {
			t.Skip("os.Stdout is a terminal in this environment; skipping")
		}
	})

	t.Run("nil file", func(t *testing.T) {
		if isTerminal(nil) {
			t.Error("isTerminal(nil) should be false")
		}
	})

	t.Run("regular file is not a terminal", func(t *testing.T) {
		f, err := os.CreateTemp(t.TempDir(), "test")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = f.Close() }()
		if isTerminal(f) {
			t.Error("regular file should not be a terminal")
		}
	})
}

func TestExecWithPTY_Buffered(t *testing.T) {
	cmd := exec.Command("sh", "-c", "echo hello-pty")
	cmd.Dir = t.TempDir()

	output, err := execWithPTY(cmd, true)
	if err != nil {
		t.Fatalf("execWithPTY() error: %v", err)
	}
	if !strings.Contains(output, "hello-pty") {
		t.Errorf("output = %q, want containing %q", output, "hello-pty")
	}
}

func TestExecWithPTY_BufferedFailure(t *testing.T) {
	cmd := exec.Command("sh", "-c", "echo fail-output && exit 1")
	cmd.Dir = t.TempDir()

	output, err := execWithPTY(cmd, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(output, "fail-output") {
		t.Errorf("output = %q, want containing %q", output, "fail-output")
	}
}

func TestExecWithPTY_Direct(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("sh", "-c", "touch pty-direct.txt")
	cmd.Dir = dir

	_, err := execWithPTY(cmd, false)
	if err != nil {
		t.Fatalf("execWithPTY() error: %v", err)
	}
	if _, err := os.Stat(dir + "/pty-direct.txt"); err != nil {
		t.Errorf("expected pty-direct.txt to exist: %v", err)
	}
}

func TestHasPTYSupport(t *testing.T) {
	if !hasPTYSupport() {
		t.Error("hasPTYSupport() should be true on Unix")
	}
}
