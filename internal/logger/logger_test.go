package logger

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/jedib0t/go-pretty/v6/text"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	text.EnableColors()
	var buf bytes.Buffer
	SetOutput(&buf)
	ResetColors()
	t.Cleanup(func() {
		SetOutput(&buf)
		ResetColors()
	})
	fn()
	return buf.String()
}

func TestPrefix(t *testing.T) {
	text.EnableColors()
	ResetColors()
	raw := prefix("my-app")
	got := stripANSI(raw)
	want := "my-app"
	if got != want {
		t.Errorf("prefix(%q) = %q, want %q", "my-app", got, want)
	}
	if raw == got {
		t.Error("prefix should contain ANSI color codes")
	}
}

func TestStart(t *testing.T) {
	out := captureOutput(t, func() {
		Start("api", "docker compose up -d")
	})
	plain := stripANSI(out)
	if !strings.Contains(plain, "[api]") {
		t.Errorf("output missing project prefix: %q", plain)
	}
	if !strings.Contains(plain, "Executing: docker compose up -d") {
		t.Errorf("output missing command: %q", plain)
	}
	if !strings.Contains(out, "🚀") {
		t.Errorf("output missing emoji: %q", out)
	}
}

func TestSuccess(t *testing.T) {
	out := captureOutput(t, func() {
		Success("api", "docker compose up -d")
	})
	plain := stripANSI(out)
	if !strings.Contains(plain, "[api]") {
		t.Errorf("output missing project prefix: %q", plain)
	}
	if !strings.Contains(plain, "Completed: docker compose up -d") {
		t.Errorf("output missing command: %q", plain)
	}
	if !strings.Contains(out, "✅") {
		t.Errorf("output missing emoji: %q", out)
	}
}

func TestError(t *testing.T) {
	out := captureOutput(t, func() {
		Error("api", "docker compose up -d", errors.New("exit code 1"))
	})
	plain := stripANSI(out)
	if !strings.Contains(plain, "[api]") {
		t.Errorf("output missing project prefix: %q", plain)
	}
	if !strings.Contains(plain, "Failed: docker compose up -d") {
		t.Errorf("output missing command: %q", plain)
	}
	if !strings.Contains(plain, "exit code 1") {
		t.Errorf("output missing error message: %q", plain)
	}
	if !strings.Contains(out, "❌") {
		t.Errorf("output missing emoji: %q", out)
	}
}

func TestBackground(t *testing.T) {
	out := captureOutput(t, func() {
		Background("api", "make run", 12345)
	})
	plain := stripANSI(out)
	if !strings.Contains(plain, "[api]") {
		t.Errorf("output missing project prefix: %q", plain)
	}
	if !strings.Contains(plain, "Background: make run") {
		t.Errorf("output missing command: %q", plain)
	}
	if !strings.Contains(plain, "PID: 12345") {
		t.Errorf("output missing PID: %q", plain)
	}
	if !strings.Contains(out, "🔄") {
		t.Errorf("output missing emoji: %q", out)
	}
}

func TestStop(t *testing.T) {
	out := captureOutput(t, func() {
		Stop("api", "make run", 12345)
	})
	plain := stripANSI(out)
	if !strings.Contains(plain, "[api]") {
		t.Errorf("output missing project prefix: %q", plain)
	}
	if !strings.Contains(plain, "Stopping: make run") {
		t.Errorf("output missing command: %q", plain)
	}
	if !strings.Contains(plain, "PID: 12345") {
		t.Errorf("output missing PID: %q", plain)
	}
	if !strings.Contains(out, "🛑") {
		t.Errorf("output missing emoji: %q", out)
	}
}

func TestProjectDone(t *testing.T) {
	out := captureOutput(t, func() {
		ProjectDone("api")
	})
	plain := stripANSI(out)
	if !strings.Contains(plain, "[api]") {
		t.Errorf("output missing project prefix: %q", plain)
	}
	if !strings.Contains(plain, "All commands completed") {
		t.Errorf("output missing message: %q", plain)
	}
}

func TestProjectFailed(t *testing.T) {
	out := captureOutput(t, func() {
		ProjectFailed("api", errors.New("something went wrong"))
	})
	plain := stripANSI(out)
	if !strings.Contains(plain, "[api]") {
		t.Errorf("output missing project prefix: %q", plain)
	}
	if !strings.Contains(plain, "Aborted") {
		t.Errorf("output missing message: %q", plain)
	}
	if !strings.Contains(plain, "something went wrong") {
		t.Errorf("output missing error: %q", plain)
	}
}

func TestOutput(t *testing.T) {
	t.Run("single line with borders", func(t *testing.T) {
		out := captureOutput(t, func() {
			Output("api", "hello world")
		})
		plain := stripANSI(out)
		if !strings.Contains(plain, "[api]") {
			t.Errorf("output missing project prefix: %q", plain)
		}
		if !strings.Contains(plain, "hello world") {
			t.Errorf("output missing content: %q", plain)
		}
		lines := strings.Split(strings.TrimRight(plain, "\n"), "\n")
		if len(lines) != 3 {
			t.Errorf("expected 3 lines (border + content + border), got %d: %q", len(lines), plain)
		}
		border := strings.Repeat("=", defaultBorderWidth)
		if !strings.Contains(lines[0], border) {
			t.Errorf("first line missing border: %q", lines[0])
		}
		if !strings.Contains(lines[2], border) {
			t.Errorf("last line missing border: %q", lines[2])
		}
	})

	t.Run("multi line with borders", func(t *testing.T) {
		out := captureOutput(t, func() {
			Output("api", "line1\nline2\nline3")
		})
		plain := stripANSI(out)
		lines := strings.Split(strings.TrimRight(plain, "\n"), "\n")
		if len(lines) != 5 {
			t.Errorf("expected 5 lines (border + 3 content + border), got %d: %q", len(lines), plain)
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
		plain := stripANSI(out)
		lines := strings.Split(strings.TrimRight(plain, "\n"), "\n")
		if len(lines) != 3 {
			t.Errorf("expected 3 lines (border + content + border), got %d: %q", len(lines), plain)
		}
	})
}

func TestBorder(t *testing.T) {
	out := captureOutput(t, func() {
		Border()
	})
	plain := stripANSI(out)
	trimmed := strings.TrimRight(plain, "\n")
	if trimmed == "" {
		t.Fatal("expected border output, got empty")
	}
	for _, ch := range trimmed {
		if ch != '=' {
			t.Errorf("expected only '=' characters, got %q in %q", string(ch), trimmed)
			break
		}
	}
	if len(trimmed) < defaultBorderWidth {
		t.Errorf("border width %d < default %d", len(trimmed), defaultBorderWidth)
	}
}

func TestProjectColorConsistency(t *testing.T) {
	text.EnableColors()
	ResetColors()
	first := prefix("my-app")
	second := prefix("my-app")
	if first != second {
		t.Errorf("same project should get same color, got %q and %q", first, second)
	}
}

func TestProjectColorVariety(t *testing.T) {
	text.EnableColors()
	ResetColors()
	a := prefix("project-a")
	b := prefix("project-b")

	ansiA := ansiRe.FindString(a)
	ansiB := ansiRe.FindString(b)
	if ansiA == "" || ansiB == "" {
		t.Fatal("expected ANSI color codes in both prefixes")
	}
	if ansiA == ansiB {
		t.Errorf("different projects should get different colors, both got %q", ansiA)
	}
}
