package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSeekToLastNLines(t *testing.T) {
	content := "line1\nline2\nline3\nline4\nline5\n"
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		n        int
		expected string
	}{
		{"last 2 lines", 2, "line4\nline5\n"},
		{"last 5 lines", 5, "line1\nline2\nline3\nline4\nline5\n"},
		{"more than available", 10, "line1\nline2\nline3\nline4\nline5\n"},
		{"last 1 line", 1, "line5\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = f.Close() }()

			seekToLastNLines(f, tt.n)

			data := make([]byte, len(content))
			n, _ := f.Read(data)
			got := string(data[:n])
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSeekToLastNLinesNoTrailingNewline(t *testing.T) {
	content := "line1\nline2\nline3"
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	seekToLastNLines(f, 2)

	data := make([]byte, len(content))
	n, _ := f.Read(data)
	got := string(data[:n])
	if !strings.Contains(got, "line2") || !strings.Contains(got, "line3") {
		t.Errorf("got %q, want containing 'line2' and 'line3'", got)
	}
}

func TestSeekToLastNLinesEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.log")
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	seekToLastNLines(f, 5)

	data := make([]byte, 100)
	n, _ := f.Read(data)
	if n != 0 {
		t.Errorf("expected no data from empty file, got %d bytes", n)
	}
}

func TestSeekToLastNLinesLargeFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "large.log")

	var sb strings.Builder
	for i := 0; i < 10000; i++ {
		sb.WriteString("this is a line of log output\n")
	}
	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	seekToLastNLines(f, 3)

	data := make([]byte, 1024)
	n, _ := f.Read(data)
	got := string(data[:n])
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("got %d lines, want 3", len(lines))
	}
}
