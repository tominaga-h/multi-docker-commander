package pidfile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"mdc/internal/config"
)

// BaseDir overrides the default PID directory for testing.
var BaseDir string

type Entry struct {
	PID     int    `json:"pid"`
	Command string `json:"command"`
	Dir     string `json:"dir"`
}

func baseDir() (string, error) {
	if BaseDir != "" {
		return BaseDir, nil
	}
	base, err := config.BaseMDCDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "pids"), nil
}

func Dir(configName string) (string, error) {
	base, err := baseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, configName), nil
}

func filePath(configName, projectName string) (string, error) {
	dir, err := Dir(configName)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, projectName+".json"), nil
}

func Append(configName, projectName string, entry Entry) error {
	entries, err := Load(configName, projectName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	entries = append(entries, entry)
	return Save(configName, projectName, entries)
}

func Save(configName, projectName string, entries []Entry) error {
	path, err := filePath(configName, projectName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func Load(configName, projectName string) ([]Entry, error) {
	path, err := filePath(configName, projectName)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func LoadAll(configName string) (map[string][]Entry, error) {
	dir, err := Dir(configName)
	if err != nil {
		return nil, err
	}
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	result := make(map[string][]Entry)
	for _, de := range dirEntries {
		if de.IsDir() || filepath.Ext(de.Name()) != ".json" {
			continue
		}
		name := strings.TrimSuffix(de.Name(), ".json")
		data, err := os.ReadFile(filepath.Join(dir, de.Name()))
		if err != nil {
			return nil, err
		}
		var items []Entry
		if err := json.Unmarshal(data, &items); err != nil {
			return nil, err
		}
		result[name] = items
	}
	return result, nil
}

func LoadAllConfigs() (map[string]map[string][]Entry, error) {
	base, err := baseDir()
	if err != nil {
		return nil, err
	}
	dirs, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	result := make(map[string]map[string][]Entry)
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		projects, err := LoadAll(d.Name())
		if err != nil {
			return nil, err
		}
		if len(projects) > 0 {
			result[d.Name()] = projects
		}
	}
	return result, nil
}

// StopFunc is called for each tracked process before it is killed.
type StopFunc func(projectName, command string, pid int)

func KillAll(configName string) error {
	return KillAllWithCallback(configName, nil)
}

func KillAllWithCallback(configName string, onStop StopFunc) error {
	projects, err := LoadAll(configName)
	if err != nil {
		return err
	}
	for projectName, entries := range projects {
		for _, e := range entries {
			if onStop != nil {
				onStop(projectName, e.Command, e.PID)
			}
			if p, err := os.FindProcess(e.PID); err == nil {
				p.Kill()
				p.Wait()
			}
		}
	}
	dir, err := Dir(configName)
	if err != nil {
		return err
	}
	return os.RemoveAll(dir)
}
