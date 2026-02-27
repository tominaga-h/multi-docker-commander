package runner

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"mdc/internal/config"
)

type ContainerInfo struct {
	ID     string `json:"ID"`
	Name   string `json:"Name"`
	State  string `json:"State"`
	Status string `json:"Status"`
	Ports  string `json:"Ports"`
}

type ProjectContainers struct {
	ProjectName string
	Containers  []ContainerInfo
	Err         error
}

// CollectPS runs "docker compose ps --format json" in each project directory
// concurrently and returns the aggregated results.
func CollectPS(cfg *config.Config) []ProjectContainers {
	results := make([]ProjectContainers, len(cfg.Projects))
	var wg sync.WaitGroup

	for i, p := range cfg.Projects {
		wg.Add(1)
		go func(idx int, project config.Project) {
			defer wg.Done()
			results[idx] = collectProjectPS(project)
		}(i, p)
	}

	wg.Wait()
	return results
}

func collectProjectPS(project config.Project) ProjectContainers {
	result := ProjectContainers{ProjectName: project.Name}

	if err := validateProjectPath(project); err != nil {
		result.Err = err
		return result
	}

	stdout, stderr, err := runDockerComposePS(project.Path)
	if err != nil {
		if isNoComposeConfigError(stderr) {
			return result
		}
		result.Err = fmt.Errorf("%s (stderr: %s)", err, strings.TrimSpace(stderr))
		return result
	}

	containers, parseErr := ParseContainers(stdout)
	if parseErr != nil {
		result.Err = fmt.Errorf("failed to parse docker compose ps output: %w", parseErr)
		return result
	}
	result.Containers = containers
	return result
}

func runDockerComposePS(dir string) (stdout, stderr string, err error) {
	stdout, stderr, err = execComposePS(dir, "docker", "compose", "ps", "--format", "json")
	if err != nil && isCommandNotFound(err) {
		stdout, stderr, err = execComposePS(dir, "docker-compose", "ps", "--format", "json")
		if err != nil && isFormatUnsupported(stderr) {
			return "", "", fmt.Errorf("docker-compose (V1) does not support --format json; please upgrade to Docker Compose V2")
		}
	}
	return stdout, stderr, err
}

func execComposePS(dir string, name string, args ...string) (string, string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
}

// ParseContainers handles both JSON array format and NDJSON (one JSON object per line).
func ParseContainers(output string) ([]ContainerInfo, error) {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return nil, nil
	}

	if strings.HasPrefix(trimmed, "[") {
		return parseJSONArray(trimmed)
	}
	return parseNDJSON(trimmed)
}

func parseJSONArray(data string) ([]ContainerInfo, error) {
	var containers []ContainerInfo
	if err := json.Unmarshal([]byte(data), &containers); err != nil {
		return nil, fmt.Errorf("JSON array parse error: %w", err)
	}
	return containers, nil
}

func parseNDJSON(data string) ([]ContainerInfo, error) {
	var containers []ContainerInfo
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var c ContainerInfo
		if err := json.Unmarshal([]byte(line), &c); err != nil {
			return nil, fmt.Errorf("NDJSON parse error on line %q: %w", line, err)
		}
		containers = append(containers, c)
	}
	return containers, nil
}

type dockerPSEntry struct {
	ID     string `json:"ID"`
	Names  string `json:"Names"`
	State  string `json:"State"`
	Status string `json:"Status"`
	Ports  string `json:"Ports"`
}

// CollectDockerPS runs "docker ps --format json" and returns all containers.
func CollectDockerPS() ([]ContainerInfo, error) {
	cmd := exec.Command("docker", "ps", "--format", "json")
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("docker ps failed: %s (stderr: %s)", err, strings.TrimSpace(stderrBuf.String()))
	}
	return ParseDockerPS(stdoutBuf.String())
}

// ParseDockerPS parses the JSON output of "docker ps --format json".
// The output uses "Names" instead of "Name".
func ParseDockerPS(output string) ([]ContainerInfo, error) {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return nil, nil
	}

	var entries []dockerPSEntry
	if strings.HasPrefix(trimmed, "[") {
		if err := json.Unmarshal([]byte(trimmed), &entries); err != nil {
			return nil, fmt.Errorf("JSON array parse error: %w", err)
		}
	} else {
		for _, line := range strings.Split(trimmed, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var e dockerPSEntry
			if err := json.Unmarshal([]byte(line), &e); err != nil {
				return nil, fmt.Errorf("NDJSON parse error on line %q: %w", line, err)
			}
			entries = append(entries, e)
		}
	}

	containers := make([]ContainerInfo, len(entries))
	for i, e := range entries {
		containers[i] = ContainerInfo{
			ID:     e.ID,
			Name:   e.Names,
			State:  e.State,
			Status: e.Status,
			Ports:  e.Ports,
		}
	}
	return containers, nil
}

func isCommandNotFound(err error) bool {
	var exitErr *exec.ExitError
	if ok := errors.As(err, &exitErr); ok {
		return exitErr.ExitCode() == 127
	}
	return errors.Is(err, exec.ErrNotFound)
}

func isFormatUnsupported(stderr string) bool {
	lower := strings.ToLower(stderr)
	return strings.Contains(lower, "unknown flag") ||
		strings.Contains(lower, "unknown shorthand flag")
}

func isNoComposeConfigError(stderr string) bool {
	lower := strings.ToLower(stderr)
	markers := []string{
		"no configuration file",
		"can't find a suitable configuration file",
		"not find a compose file",
		"no compose file",
	}
	for _, m := range markers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}
