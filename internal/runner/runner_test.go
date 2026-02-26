package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mdc/internal/config"
	"mdc/internal/logger"
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
					Up:   []string{"echo up-a"},
					Down: []string{"echo down-a"},
				},
			},
			{
				Name: "svc-b",
				Path: "/tmp",
				Commands: config.Commands{
					Up:   []string{"echo up-b1", "echo up-b2"},
					Down: []string{"echo down-b"},
				},
			},
		},
	}

	t.Run("action up", func(t *testing.T) {
		cmds, err := commandsForAction(cfg, "up")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cmds) != 2 {
			t.Fatalf("len = %d, want 2", len(cmds))
		}
		if cmds[0][0] != "echo up-a" {
			t.Errorf("cmds[0][0] = %q, want %q", cmds[0][0], "echo up-a")
		}
		if len(cmds[1]) != 2 {
			t.Fatalf("len(cmds[1]) = %d, want 2", len(cmds[1]))
		}
	})

	t.Run("action down", func(t *testing.T) {
		cmds, err := commandsForAction(cfg, "down")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cmds[0][0] != "echo down-a" {
			t.Errorf("cmds[0][0] = %q, want %q", cmds[0][0], "echo down-a")
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
					Up: []string{
						"touch first.txt",
						"touch second.txt",
					},
				},
			},
		},
	}

	if err := Run(cfg, "up"); err != nil {
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
					Up: []string{"touch parallel_a.txt"},
				},
			},
			{
				Name: "proj-b",
				Path: dir2,
				Commands: config.Commands{
					Up: []string{"touch parallel_b.txt"},
				},
			},
		},
	}

	if err := Run(cfg, "up"); err != nil {
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
					Commands: config.Commands{Up: []string{"false"}},
				},
			},
		}
		err := Run(cfg, "up")
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
					Commands: config.Commands{Up: []string{"false"}},
				},
			},
		}
		err := Run(cfg, "up")
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
				Commands: config.Commands{Up: []string{"echo hi"}},
			},
		},
	}

	err := Run(cfg, "up")
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
				Commands: config.Commands{Up: []string{"echo hi"}},
			},
		},
	}

	err := Run(cfg, "up")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unknown execution_mode") {
		t.Errorf("error = %q, want containing 'unknown execution_mode'", err.Error())
	}
}
