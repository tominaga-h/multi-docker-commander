package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name: "absolute path unchanged",
			path: "/usr/local/bin",
			want: "/usr/local/bin",
		},
		{
			name: "relative path unchanged",
			path: "relative/path",
			want: "relative/path",
		},
		{
			name: "empty string unchanged",
			path: "",
			want: "",
		},
		{
			name: "tilde expands to home",
			path: "~/projects",
			want: filepath.Join(home, "projects"),
		},
		{
			name: "tilde only expands to home",
			path: "~",
			want: home,
		},
		{
			name: "tilde with nested path",
			path: "~/a/b/c",
			want: filepath.Join(home, "a", "b", "c"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandHome(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandHome(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExpandHome(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "valid sequential config",
			cfg: Config{
				ExecutionMode: "sequential",
				Projects:      []Project{{Name: "svc", Path: "/tmp"}},
			},
		},
		{
			name: "valid parallel config",
			cfg: Config{
				ExecutionMode: "parallel",
				Projects:      []Project{{Name: "svc", Path: "/tmp"}},
			},
		},
		{
			name: "invalid execution_mode",
			cfg: Config{
				ExecutionMode: "invalid",
				Projects:      []Project{{Name: "svc", Path: "/tmp"}},
			},
			wantErr: "execution_mode must be",
		},
		{
			name: "empty execution_mode",
			cfg: Config{
				ExecutionMode: "",
				Projects:      []Project{{Name: "svc", Path: "/tmp"}},
			},
			wantErr: "execution_mode must be",
		},
		{
			name: "no projects",
			cfg: Config{
				ExecutionMode: "parallel",
				Projects:      []Project{},
			},
			wantErr: "at least one project",
		},
		{
			name: "project missing name",
			cfg: Config{
				ExecutionMode: "parallel",
				Projects:      []Project{{Name: "", Path: "/tmp"}},
			},
			wantErr: "name is required",
		},
		{
			name: "project missing path",
			cfg: Config{
				ExecutionMode: "parallel",
				Projects:      []Project{{Name: "svc", Path: ""}},
			},
			wantErr: "path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("validate() unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Errorf("validate() expected error containing %q, got nil", tt.wantErr)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("validate() error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestLoadFromDir(t *testing.T) {
	t.Run("valid yaml with extension", func(t *testing.T) {
		dir := t.TempDir()
		yaml := `execution_mode: sequential
projects:
  - name: app
    path: /tmp
    commands:
      up: ["echo up"]
      down: ["echo down"]
`
		if err := os.WriteFile(filepath.Join(dir, "test.yml"), []byte(yaml), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadFromDir(dir, "test.yml")
		if err != nil {
			t.Fatalf("LoadFromDir() error: %v", err)
		}
		if cfg.ExecutionMode != "sequential" {
			t.Errorf("ExecutionMode = %q, want %q", cfg.ExecutionMode, "sequential")
		}
		if len(cfg.Projects) != 1 {
			t.Fatalf("len(Projects) = %d, want 1", len(cfg.Projects))
		}
		if cfg.Projects[0].Name != "app" {
			t.Errorf("Projects[0].Name = %q, want %q", cfg.Projects[0].Name, "app")
		}
	})

	t.Run("auto-appends .yml extension", func(t *testing.T) {
		dir := t.TempDir()
		yaml := `execution_mode: parallel
projects:
  - name: svc
    path: /tmp
    commands:
      up: ["echo up"]
`
		if err := os.WriteFile(filepath.Join(dir, "dev.yml"), []byte(yaml), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadFromDir(dir, "dev")
		if err != nil {
			t.Fatalf("LoadFromDir() error: %v", err)
		}
		if cfg.ExecutionMode != "parallel" {
			t.Errorf("ExecutionMode = %q, want %q", cfg.ExecutionMode, "parallel")
		}
	})

	t.Run("tilde path expansion", func(t *testing.T) {
		dir := t.TempDir()
		yaml := `execution_mode: sequential
projects:
  - name: app
    path: ~/myproject
    commands:
      up: ["echo up"]
`
		if err := os.WriteFile(filepath.Join(dir, "test.yml"), []byte(yaml), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadFromDir(dir, "test")
		if err != nil {
			t.Fatalf("LoadFromDir() error: %v", err)
		}

		home, _ := os.UserHomeDir()
		want := filepath.Join(home, "myproject")
		if cfg.Projects[0].Path != want {
			t.Errorf("Projects[0].Path = %q, want %q", cfg.Projects[0].Path, want)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		dir := t.TempDir()
		_, err := LoadFromDir(dir, "nonexistent")
		if err == nil {
			t.Fatal("LoadFromDir() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to read config file") {
			t.Errorf("error = %q, want containing 'failed to read config file'", err.Error())
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "bad.yml"), []byte(":::invalid"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := LoadFromDir(dir, "bad")
		if err == nil {
			t.Fatal("LoadFromDir() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to parse config file") {
			t.Errorf("error = %q, want containing 'failed to parse config file'", err.Error())
		}
	})

	t.Run("validation failure", func(t *testing.T) {
		dir := t.TempDir()
		yaml := `execution_mode: invalid
projects:
  - name: svc
    path: /tmp
`
		if err := os.WriteFile(filepath.Join(dir, "bad.yml"), []byte(yaml), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := LoadFromDir(dir, "bad")
		if err == nil {
			t.Fatal("LoadFromDir() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid config") {
			t.Errorf("error = %q, want containing 'invalid config'", err.Error())
		}
	})

	t.Run("multiple projects", func(t *testing.T) {
		dir := t.TempDir()
		yaml := `execution_mode: parallel
projects:
  - name: api
    path: /tmp
    commands:
      up: ["echo api-up"]
      down: ["echo api-down"]
  - name: web
    path: /tmp
    commands:
      up: ["echo web-up"]
      down: ["echo web-down"]
`
		if err := os.WriteFile(filepath.Join(dir, "multi.yml"), []byte(yaml), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadFromDir(dir, "multi")
		if err != nil {
			t.Fatalf("LoadFromDir() error: %v", err)
		}
		if len(cfg.Projects) != 2 {
			t.Fatalf("len(Projects) = %d, want 2", len(cfg.Projects))
		}
		if cfg.Projects[0].Name != "api" {
			t.Errorf("Projects[0].Name = %q, want %q", cfg.Projects[0].Name, "api")
		}
		if cfg.Projects[1].Name != "web" {
			t.Errorf("Projects[1].Name = %q, want %q", cfg.Projects[1].Name, "web")
		}
	})
}
