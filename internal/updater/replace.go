package updater

import (
	"fmt"
	"os"
	"path/filepath"
)

const sudoHint = "sudo mdc update"

func replaceExecutable(data []byte, executablePath string) error {
	if len(data) == 0 {
		return fmt.Errorf("binary data is empty")
	}

	target, err := resolveExecutablePath(executablePath)
	if err != nil {
		return err
	}

	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("stat executable %s: %w", target, err)
	}
	mode := info.Mode().Perm()

	dir := filepath.Dir(target)
	tmpFile, err := os.CreateTemp(dir, ".mdc-update-*")
	if err != nil {
		return permissionError(target, err)
	}
	tmpPath := tmpFile.Name()

	cleanup := func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
	}
	defer cleanup()

	if _, err := tmpFile.Write(data); err != nil {
		return permissionError(target, err)
	}
	if err := tmpFile.Chmod(mode | 0111); err != nil {
		return permissionError(target, err)
	}
	if err := tmpFile.Close(); err != nil {
		return permissionError(target, err)
	}

	if err := os.Rename(tmpPath, target); err != nil {
		return permissionError(target, err)
	}

	return nil
}

func resolveExecutablePath(executablePath string) (string, error) {
	path := executablePath
	if path == "" {
		var err error
		path, err = os.Executable()
		if err != nil {
			return "", fmt.Errorf("resolve executable path: %w", err)
		}
	}

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("resolve symlinks for %s: %w", path, err)
	}
	return resolved, nil
}

func permissionError(target string, err error) error {
	if os.IsPermission(err) {
		return fmt.Errorf("cannot replace executable at %s: %w (try: %s)", target, err, sudoHint)
	}
	return fmt.Errorf("cannot replace executable at %s: %w", target, err)
}
