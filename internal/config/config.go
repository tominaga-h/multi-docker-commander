package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type CommandItem struct {
	Command    string `yaml:"command"`
	Background bool   `yaml:"background"`
}

func (c *CommandItem) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		c.Command = value.Value
		return nil
	}
	type raw CommandItem
	var r raw
	if err := value.Decode(&r); err != nil {
		return err
	}
	*c = CommandItem(r)
	return nil
}

type Commands struct {
	Up   []CommandItem `yaml:"up"`
	Down []CommandItem `yaml:"down"`
}

type Project struct {
	Name     string   `yaml:"name"`
	Path     string   `yaml:"path"`
	Commands Commands `yaml:"commands"`
}

type Config struct {
	ExecutionMode string    `yaml:"execution_mode"`
	Projects      []Project `yaml:"projects"`
}

func ExpandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, path[1:]), nil
}

func BaseMDCDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "mdc"), nil
}

func DefaultConfigDir() (string, error) {
	return BaseMDCDir()
}

func Load(name string) (*Config, error) {
	configDir, err := DefaultConfigDir()
	if err != nil {
		return nil, err
	}
	return LoadFromDir(configDir, name)
}

func resolveConfigPath(configDir, name string) (string, error) {
	path := filepath.Join(configDir, name)
	if filepath.Ext(path) != "" {
		return path, nil
	}
	for _, ext := range []string{".yml", ".yaml"} {
		candidate := path + ext
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("config file not found: tried %s.yml and %s.yaml", path, path)
}

func LoadFromDir(configDir, name string) (*Config, error) {
	path, err := resolveConfigPath(configDir, name)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config %q: %w", name, err)
	}

	for i := range cfg.Projects {
		expanded, err := ExpandHome(cfg.Projects[i].Path)
		if err != nil {
			return nil, fmt.Errorf("project %q: %w", cfg.Projects[i].Name, err)
		}
		cfg.Projects[i].Path = expanded
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	switch c.ExecutionMode {
	case "parallel", "sequential":
	default:
		return fmt.Errorf("execution_mode must be \"parallel\" or \"sequential\", got %q", c.ExecutionMode)
	}

	if len(c.Projects) == 0 {
		return fmt.Errorf("at least one project must be defined")
	}

	for i, p := range c.Projects {
		if p.Name == "" {
			return fmt.Errorf("project[%d]: name is required", i)
		}
		if p.Path == "" {
			return fmt.Errorf("project %q: path is required", p.Name)
		}
	}

	return nil
}
