package logger

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	SetOutput(&buf)
	t.Cleanup(func() { SetOutput(&buf) })
	fn()
	return buf.String()
}

func TestPrefix(t *testing.T) {
	got := prefix("my-app")
	want := "[my-app]"
	if got != want {
		t.Errorf("prefix(%q) = %q, want %q", "my-app", got, want)
	}
}

func TestStart(t *testing.T) {
	out := captureOutput(t, func() {
		Start("api", "docker compose up -d")
	})
	if !strings.Contains(out, "[api]") {
		t.Errorf("output missing project prefix: %q", out)
	}
	if !strings.Contains(out, "Executing: docker compose up -d") {
		t.Errorf("output missing command: %q", out)
	}
	if !strings.Contains(out, "🚀") {
		t.Errorf("output missing emoji: %q", out)
	}
}

func TestSuccess(t *testing.T) {
	out := captureOutput(t, func() {
		Success("api", "docker compose up -d")
	})
	if !strings.Contains(out, "[api]") {
		t.Errorf("output missing project prefix: %q", out)
	}
	if !strings.Contains(out, "Completed: docker compose up -d") {
		t.Errorf("output missing command: %q", out)
	}
	if !strings.Contains(out, "✅") {
		t.Errorf("output missing emoji: %q", out)
	}
}

func TestError(t *testing.T) {
	out := captureOutput(t, func() {
		Error("api", "docker compose up -d", errors.New("exit code 1"))
	})
	if !strings.Contains(out, "[api]") {
		t.Errorf("output missing project prefix: %q", out)
	}
	if !strings.Contains(out, "Failed: docker compose up -d") {
		t.Errorf("output missing command: %q", out)
	}
	if !strings.Contains(out, "exit code 1") {
		t.Errorf("output missing error message: %q", out)
	}
	if !strings.Contains(out, "❌") {
		t.Errorf("output missing emoji: %q", out)
	}
}

func TestProjectDone(t *testing.T) {
	out := captureOutput(t, func() {
		ProjectDone("api")
	})
	if !strings.Contains(out, "[api]") {
		t.Errorf("output missing project prefix: %q", out)
	}
	if !strings.Contains(out, "All commands completed") {
		t.Errorf("output missing message: %q", out)
	}
}

func TestProjectFailed(t *testing.T) {
	out := captureOutput(t, func() {
		ProjectFailed("api", errors.New("something went wrong"))
	})
	if !strings.Contains(out, "[api]") {
		t.Errorf("output missing project prefix: %q", out)
	}
	if !strings.Contains(out, "Aborted") {
		t.Errorf("output missing message: %q", out)
	}
	if !strings.Contains(out, "something went wrong") {
		t.Errorf("output missing error: %q", out)
	}
}

func TestOutput(t *testing.T) {
	t.Run("single line", func(t *testing.T) {
		out := captureOutput(t, func() {
			Output("api", "hello world")
		})
		if !strings.Contains(out, "[api]") {
			t.Errorf("output missing project prefix: %q", out)
		}
		if !strings.Contains(out, "hello world") {
			t.Errorf("output missing content: %q", out)
		}
	})

	t.Run("multi line", func(t *testing.T) {
		out := captureOutput(t, func() {
			Output("api", "line1\nline2\nline3")
		})
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		if len(lines) != 3 {
			t.Errorf("expected 3 lines, got %d: %q", len(lines), out)
		}
		for _, line := range lines {
			if !strings.Contains(line, "[api]") {
				t.Errorf("line missing prefix: %q", line)
			}
		}
	})

	t.Run("empty output", func(t *testing.T) {
		out := captureOutput(t, func() {
			Output("api", "")
		})
		if out != "" {
			t.Errorf("expected empty output, got %q", out)
		}
	})

	t.Run("only newlines", func(t *testing.T) {
		out := captureOutput(t, func() {
			Output("api", "\n\n\n")
		})
		if out != "" {
			t.Errorf("expected empty output, got %q", out)
		}
	})

	t.Run("trailing newlines trimmed", func(t *testing.T) {
		out := captureOutput(t, func() {
			Output("api", "hello\n\n")
		})
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		if len(lines) != 1 {
			t.Errorf("expected 1 line, got %d: %q", len(lines), out)
		}
	})
}
