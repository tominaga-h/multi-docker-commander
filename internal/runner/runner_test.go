package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mdc/internal/config"
	"mdc/internal/logger"
	"mdc/internal/pidfile"
)

func init() {
	logger.SetOutput(os.Stderr)
}

func TestCommandsForAction(t *testing.T) {
	cfg := &config.Config{
		ExecutionMode: "sequential",
		Projects: []config.Project{
			{
				Name: "svc-a",
				Path: "/tmp",
				Commands: config.Commands{
					Up:   []config.CommandItem{{Command: "echo up-a"}},
					Down: []config.CommandItem{{Command: "echo down-a"}},
				},
			},
			{
				Name: "svc-b",
				Path: "/tmp",
				Commands: config.Commands{
					Up:   []config.CommandItem{{Command: "echo up-b1"}, {Command: "echo up-b2"}},
					Down: []config.CommandItem{{Command: "echo down-b"}},
				},
			},
		},
	}

	t.Run("action up", func(t *testing.T) {
		pcs, err := commandsForAction(cfg, "up")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pcs) != 2 {
			t.Fatalf("len = %d, want 2", len(pcs))
		}
		if pcs[0].Project.Name != "svc-a" {
			t.Errorf("pcs[0].Project.Name = %q, want %q", pcs[0].Project.Name, "svc-a")
		}
		if pcs[0].Commands[0].Command != "echo up-a" {
			t.Errorf("pcs[0].Commands[0].Command = %q, want %q", pcs[0].Commands[0].Command, "echo up-a")
		}
		if len(pcs[1].Commands) != 2 {
			t.Fatalf("len(pcs[1].Commands) = %d, want 2", len(pcs[1].Commands))
		}
	})

	t.Run("action down", func(t *testing.T) {
		pcs, err := commandsForAction(cfg, "down")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pcs[0].Commands[0].Command != "echo down-a" {
			t.Errorf("pcs[0].Commands[0].Command = %q, want %q", pcs[0].Commands[0].Command, "echo down-a")
		}
	})

	t.Run("unknown action", func(t *testing.T) {
		_, err := commandsForAction(cfg, "restart")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "unknown action") {
			t.Errorf("error = %q, want containing 'unknown action'", err.Error())
		}
	})

	t.Run("empty commands", func(t *testing.T) {
		emptyCfg := &config.Config{
			ExecutionMode: "sequential",
			Projects: []config.Project{
				{Name: "svc", Path: "/tmp", Commands: config.Commands{}},
			},
		}
		_, err := commandsForAction(emptyCfg, "up")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "no commands defined") {
			t.Errorf("error = %q, want containing 'no commands defined'", err.Error())
		}
	})
}

func TestValidateProjectPath(t *testing.T) {
	t.Run("valid directory", func(t *testing.T) {
		dir := t.TempDir()
		p := config.Project{Name: "svc", Path: dir}
		if err := validateProjectPath(p); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("nonexistent path", func(t *testing.T) {
		p := config.Project{Name: "svc", Path: "/nonexistent/path/xyz"}
		err := validateProjectPath(p)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("error = %q, want containing 'does not exist'", err.Error())
		}
	})

	t.Run("path is a file not directory", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "file.txt")
		if err := os.WriteFile(f, []byte("hello"), 0644); err != nil {
			t.Fatal(err)
		}
		p := config.Project{Name: "svc", Path: f}
		err := validateProjectPath(p)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "is not a directory") {
			t.Errorf("error = %q, want containing 'is not a directory'", err.Error())
		}
	})
}

func TestNewShellCommand(t *testing.T) {
	cmd := newShellCommand("echo hello", "/tmp")
	if cmd.Dir != "/tmp" {
		t.Errorf("Dir = %q, want %q", cmd.Dir, "/tmp")
	}
	args := cmd.Args
	if len(args) != 3 || args[0] != "sh" || args[1] != "-c" || args[2] != "echo hello" {
		t.Errorf("Args = %v, want [sh -c echo hello]", args)
	}
	if cmd.Stdin != os.Stdin {
		t.Error("Stdin should be os.Stdin so that TTY-dependent commands (e.g. docker compose exec) work")
	}
}

func TestRunSequential(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		ExecutionMode: "sequential",
		Projects: []config.Project{
			{
				Name: "test-proj",
				Path: dir,
				Commands: config.Commands{
					Up: []config.CommandItem{
						{Command: "touch first.txt"},
						{Command: "touch second.txt"},
					},
				},
			},
		},
	}

	if err := Run(cfg, "up", "test-config"); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	for _, name := range []string{"first.txt", "second.txt"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s to exist: %v", name, err)
		}
	}
}

func TestRunParallel(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	cfg := &config.Config{
		ExecutionMode: "parallel",
		Projects: []config.Project{
			{
				Name: "proj-a",
				Path: dir1,
				Commands: config.Commands{
					Up: []config.CommandItem{{Command: "touch parallel_a.txt"}},
				},
			},
			{
				Name: "proj-b",
				Path: dir2,
				Commands: config.Commands{
					Up: []config.CommandItem{{Command: "touch parallel_b.txt"}},
				},
			},
		},
	}

	if err := Run(cfg, "up", "test-config"); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir1, "parallel_a.txt")); err != nil {
		t.Errorf("expected parallel_a.txt in dir1: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir2, "parallel_b.txt")); err != nil {
		t.Errorf("expected parallel_b.txt in dir2: %v", err)
	}
}

func TestRunCommandFailure(t *testing.T) {
	dir := t.TempDir()

	t.Run("sequential failure", func(t *testing.T) {
		cfg := &config.Config{
			ExecutionMode: "sequential",
			Projects: []config.Project{
				{
					Name:     "fail-proj",
					Path:     dir,
					Commands: config.Commands{Up: []config.CommandItem{{Command: "false"}}},
				},
			},
		}
		err := Run(cfg, "up", "test-config")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed") {
			t.Errorf("error = %q, want containing 'failed'", err.Error())
		}
	})

	t.Run("parallel failure", func(t *testing.T) {
		cfg := &config.Config{
			ExecutionMode: "parallel",
			Projects: []config.Project{
				{
					Name:     "fail-proj",
					Path:     dir,
					Commands: config.Commands{Up: []config.CommandItem{{Command: "false"}}},
				},
			},
		}
		err := Run(cfg, "up", "test-config")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed") {
			t.Errorf("error = %q, want containing 'failed'", err.Error())
		}
	})
}

func TestRunInvalidPath(t *testing.T) {
	cfg := &config.Config{
		ExecutionMode: "sequential",
		Projects: []config.Project{
			{
				Name:     "bad-proj",
				Path:     "/nonexistent/path/xyz",
				Commands: config.Commands{Up: []config.CommandItem{{Command: "echo hi"}}},
			},
		},
	}

	err := Run(cfg, "up", "test-config")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("error = %q, want containing 'does not exist'", err.Error())
	}
}

func TestRunUnknownExecutionMode(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		ExecutionMode: "unknown",
		Projects: []config.Project{
			{
				Name:     "proj",
				Path:     dir,
				Commands: config.Commands{Up: []config.CommandItem{{Command: "echo hi"}}},
			},
		},
	}

	err := Run(cfg, "up", "test-config")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unknown execution_mode") {
		t.Errorf("error = %q, want containing 'unknown execution_mode'", err.Error())
	}
}

func TestRunBackgroundCommand(t *testing.T) {
	dir := t.TempDir()
	pidDir := t.TempDir()
	oldBaseDir := pidfile.BaseDir
	pidfile.BaseDir = pidDir
	defer func() { pidfile.BaseDir = oldBaseDir }()

	cfg := &config.Config{
		ExecutionMode: "sequential",
		Projects: []config.Project{
			{
				Name: "bg-proj",
				Path: dir,
				Commands: config.Commands{
					Up: []config.CommandItem{
						{Command: "sleep 60", Background: true},
					},
				},
			},
		},
	}

	if err := Run(cfg, "up", "test-bg"); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	entries, err := pidfile.Load("test-bg", "bg-proj")
	if err != nil {
		t.Fatalf("pidfile.Load() error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 PID entry, got %d", len(entries))
	}
	if entries[0].Command != "sleep 60" {
		t.Errorf("Command = %q, want %q", entries[0].Command, "sleep 60")
	}
	if entries[0].PID <= 0 {
		t.Errorf("PID = %d, want > 0", entries[0].PID)
	}

	if !pidfile.IsRunning(entries[0].PID) {
		t.Error("background process should be running")
	}

	// Clean up
	if p, err := os.FindProcess(entries[0].PID); err == nil {
		p.Kill()
		p.Wait()
	}
}

func TestRunMixedForegroundAndBackground(t *testing.T) {
	dir := t.TempDir()
	pidDir := t.TempDir()
	oldBaseDir := pidfile.BaseDir
	pidfile.BaseDir = pidDir
	defer func() { pidfile.BaseDir = oldBaseDir }()

	cfg := &config.Config{
		ExecutionMode: "sequential",
		Projects: []config.Project{
			{
				Name: "mix-proj",
				Path: dir,
				Commands: config.Commands{
					Up: []config.CommandItem{
						{Command: "touch fg.txt"},
						{Command: "sleep 60", Background: true},
						{Command: "touch fg2.txt"},
					},
				},
			},
		},
	}

	if err := Run(cfg, "up", "test-mix"); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "fg.txt")); err != nil {
		t.Error("fg.txt should exist (foreground command before background)")
	}
	if _, err := os.Stat(filepath.Join(dir, "fg2.txt")); err != nil {
		t.Error("fg2.txt should exist (foreground command after background)")
	}

	entries, err := pidfile.Load("test-mix", "mix-proj")
	if err != nil {
		t.Fatalf("pidfile.Load() error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 background PID, got %d", len(entries))
	}

	// Clean up
	if p, err := os.FindProcess(entries[0].PID); err == nil {
		p.Kill()
		p.Wait()
	}
}
