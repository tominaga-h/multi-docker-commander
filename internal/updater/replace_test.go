package updater

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplaceExecutable(t *testing.T) {
	t.Run("replaces executable successfully", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "mdc")
		original := []byte("original-binary")
		if err := os.WriteFile(target, original, 0755); err != nil {
			t.Fatal(err)
		}

		updated := []byte("updated-binary")
		if err := replaceExecutable(updated, target); err != nil {
			t.Fatalf("replaceExecutable() error = %v", err)
		}

		got, err := os.ReadFile(target)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(updated) {
			t.Fatalf("file content = %q, want %q", got, updated)
		}

		info, err := os.Stat(target)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm()&0111 == 0 {
			t.Fatal("executable bit not preserved")
		}
	})

	t.Run("replaces executable through symlink", func(t *testing.T) {
		dir := t.TempDir()
		realTarget := filepath.Join(dir, "mdc-real")
		original := []byte("original-binary")
		if err := os.WriteFile(realTarget, original, 0755); err != nil {
			t.Fatal(err)
		}

		linkPath := filepath.Join(dir, "mdc-link")
		if err := os.Symlink(realTarget, linkPath); err != nil {
			t.Skip("symlinks not supported in this environment")
		}

		updated := []byte("updated-binary")
		if err := replaceExecutable(updated, linkPath); err != nil {
			t.Fatalf("replaceExecutable() error = %v", err)
		}

		got, err := os.ReadFile(realTarget)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(updated) {
			t.Fatalf("real file content = %q, want %q", got, updated)
		}
	})

	t.Run("permission error includes sudo hint", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "mdc")
		if err := os.WriteFile(target, []byte("original"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(dir, 0555); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = os.Chmod(dir, 0755)
		})

		err := replaceExecutable([]byte("updated"), target)
		if err == nil {
			t.Fatal("replaceExecutable() expected error, got nil")
		}
		if !strings.Contains(err.Error(), sudoHint) {
			t.Fatalf("error = %v, want %q hint", err, sudoHint)
		}
	})
}
