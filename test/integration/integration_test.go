package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mdc/internal/config"
	"mdc/internal/logger"
	"mdc/internal/pidfile"
	"mdc/internal/runner"
)

func TestMain(m *testing.M) {
	logger.SetOutput(os.Stderr)
	os.Exit(m.Run())
}

func writeConfig(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestSequentialUpDown(t *testing.T) {
	configDir := t.TempDir()
	projectDir := t.TempDir()

	yaml := fmt.Sprintf(`execution_mode: sequential
projects:
  - name: svc
    path: %s
    commands:
      up:
        - "touch up.txt"
      down:
        - "rm up.txt"
`, projectDir)

	writeConfig(t, configDir, "test.yml", yaml)

	cfg, err := config.LoadFromDir(configDir, "test")
	if err != nil {
		t.Fatalf("LoadFromDir() error: %v", err)
	}

	if err := runner.Run(cfg, "up", "test"); err != nil {
		t.Fatalf("Run(up) error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, "up.txt")); err != nil {
		t.Fatalf("up.txt should exist after 'up': %v", err)
	}

	if err := runner.Run(cfg, "down", "test"); err != nil {
		t.Fatalf("Run(down) error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, "up.txt")); !os.IsNotExist(err) {
		t.Fatal("up.txt should not exist after 'down'")
	}
}

func TestParallelMultipleProjects(t *testing.T) {
	configDir := t.TempDir()
	proj1 := t.TempDir()
	proj2 := t.TempDir()
	proj3 := t.TempDir()

	yaml := fmt.Sprintf(`execution_mode: parallel
projects:
  - name: api
    path: %s
    commands:
      up:
        - "touch api_started.txt"
  - name: web
    path: %s
    commands:
      up:
        - "touch web_started.txt"
  - name: worker
    path: %s
    commands:
      up:
        - "touch worker_started.txt"
`, proj1, proj2, proj3)

	writeConfig(t, configDir, "multi.yml", yaml)

	cfg, err := config.LoadFromDir(configDir, "multi")
	if err != nil {
		t.Fatalf("LoadFromDir() error: %v", err)
	}

	if err := runner.Run(cfg, "up", "multi"); err != nil {
		t.Fatalf("Run(up) error: %v", err)
	}

	checks := []struct {
		dir  string
		file string
	}{
		{proj1, "api_started.txt"},
		{proj2, "web_started.txt"},
		{proj3, "worker_started.txt"},
	}
	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(c.dir, c.file)); err != nil {
			t.Errorf("%s should exist: %v", c.file, err)
		}
	}
}

func TestSequentialMultipleCommands(t *testing.T) {
	configDir := t.TempDir()
	projectDir := t.TempDir()

	yaml := fmt.Sprintf(`execution_mode: sequential
projects:
  - name: svc
    path: %s
    commands:
      up:
        - "echo step1 > log.txt"
        - "echo step2 >> log.txt"
        - "echo step3 >> log.txt"
`, projectDir)

	writeConfig(t, configDir, "steps.yml", yaml)

	cfg, err := config.LoadFromDir(configDir, "steps")
	if err != nil {
		t.Fatalf("LoadFromDir() error: %v", err)
	}

	if err := runner.Run(cfg, "up", "steps"); err != nil {
		t.Fatalf("Run(up) error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(projectDir, "log.txt"))
	if err != nil {
		t.Fatalf("reading log.txt: %v", err)
	}
	content := string(data)
	for _, step := range []string{"step1", "step2", "step3"} {
		if !strings.Contains(content, step) {
			t.Errorf("log.txt should contain %q, got %q", step, content)
		}
	}
}

func TestCommandFailureStopsSequential(t *testing.T) {
	configDir := t.TempDir()
	projectDir := t.TempDir()

	yaml := fmt.Sprintf(`execution_mode: sequential
projects:
  - name: svc
    path: %s
    commands:
      up:
        - "touch before_fail.txt"
        - "false"
        - "touch after_fail.txt"
`, projectDir)

	writeConfig(t, configDir, "fail.yml", yaml)

	cfg, err := config.LoadFromDir(configDir, "fail")
	if err != nil {
		t.Fatalf("LoadFromDir() error: %v", err)
	}

	err = runner.Run(cfg, "up", "fail")
	if err == nil {
		t.Fatal("Run(up) expected error, got nil")
	}

	if _, err := os.Stat(filepath.Join(projectDir, "before_fail.txt")); err != nil {
		t.Error("before_fail.txt should exist")
	}
	if _, err := os.Stat(filepath.Join(projectDir, "after_fail.txt")); !os.IsNotExist(err) {
		t.Error("after_fail.txt should NOT exist (command after failure should not run)")
	}
}

func TestInvalidConfigPath(t *testing.T) {
	configDir := t.TempDir()

	yaml := `execution_mode: sequential
projects:
  - name: bad
    path: /nonexistent/path/abc123
    commands:
      up:
        - "echo hello"
`
	writeConfig(t, configDir, "bad.yml", yaml)

	cfg, err := config.LoadFromDir(configDir, "bad")
	if err != nil {
		t.Fatalf("LoadFromDir() error: %v", err)
	}

	err = runner.Run(cfg, "up", "bad")
	if err == nil {
		t.Fatal("Run(up) expected error for nonexistent path, got nil")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("error = %q, want containing 'does not exist'", err.Error())
	}
}

func TestInvalidExecutionMode(t *testing.T) {
	configDir := t.TempDir()

	yaml := `execution_mode: invalid_mode
projects:
  - name: svc
    path: /tmp
`
	writeConfig(t, configDir, "invalid.yml", yaml)

	_, err := config.LoadFromDir(configDir, "invalid")
	if err == nil {
		t.Fatal("LoadFromDir() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "execution_mode") {
		t.Errorf("error = %q, want containing 'execution_mode'", err.Error())
	}
}

func TestBackgroundCommandIntegration(t *testing.T) {
	configDir := t.TempDir()
	projectDir := t.TempDir()
	pidDir := t.TempDir()
	oldBaseDir := pidfile.BaseDir
	pidfile.BaseDir = pidDir
	defer func() { pidfile.BaseDir = oldBaseDir }()

	yaml := fmt.Sprintf(`execution_mode: sequential
projects:
  - name: daemon-svc
    path: %s
    commands:
      up:
        - "touch setup.txt"
        - command: "sleep 60"
          background: true
`, projectDir)

	writeConfig(t, configDir, "bg.yml", yaml)

	cfg, err := config.LoadFromDir(configDir, "bg")
	if err != nil {
		t.Fatalf("LoadFromDir() error: %v", err)
	}

	if err := runner.Run(cfg, "up", "bg"); err != nil {
		t.Fatalf("Run(up) error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(projectDir, "setup.txt")); err != nil {
		t.Error("setup.txt should exist (foreground command ran)")
	}

	entries, err := pidfile.Load("bg", "daemon-svc")
	if err != nil {
		t.Fatalf("pidfile.Load() error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 PID entry, got %d", len(entries))
	}
	if !pidfile.IsRunning(entries[0].PID) {
		t.Error("background process should be running")
	}

	// Simulate mdc down: kill background processes
	if err := pidfile.KillAll("bg"); err != nil {
		t.Fatalf("KillAll() error: %v", err)
	}

	// Verify process is stopped (give it a moment)
	if pidfile.IsRunning(entries[0].PID) {
		t.Error("background process should be stopped after KillAll")
	}
}
