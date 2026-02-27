package runner

import (
	"testing"
)

func TestParseContainers_NDJSON(t *testing.T) {
	input := `{"ID":"abc123def456","Name":"web-1","State":"running","Status":"Up 2 hours","Ports":"0.0.0.0:3000->3000/tcp"}
{"ID":"def789012345","Name":"db-1","State":"running","Status":"Up 2 hours","Ports":"5432/tcp"}`

	containers, err := ParseContainers(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(containers))
	}

	if containers[0].ID != "abc123def456" {
		t.Errorf("expected ID abc123def456, got %s", containers[0].ID)
	}
	if containers[0].Name != "web-1" {
		t.Errorf("expected Name web-1, got %s", containers[0].Name)
	}
	if containers[0].State != "running" {
		t.Errorf("expected State running, got %s", containers[0].State)
	}
	if containers[1].Ports != "5432/tcp" {
		t.Errorf("expected Ports 5432/tcp, got %s", containers[1].Ports)
	}
}

func TestParseContainers_JSONArray(t *testing.T) {
	input := `[{"ID":"abc123","Name":"app-1","State":"running","Status":"Up 5 min","Ports":"8080/tcp"},{"ID":"def456","Name":"redis-1","State":"exited","Status":"Exited (0) 1 min ago","Ports":""}]`

	containers, err := ParseContainers(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(containers))
	}

	if containers[0].Name != "app-1" {
		t.Errorf("expected Name app-1, got %s", containers[0].Name)
	}
	if containers[1].State != "exited" {
		t.Errorf("expected State exited, got %s", containers[1].State)
	}
	if containers[1].Status != "Exited (0) 1 min ago" {
		t.Errorf("expected Status 'Exited (0) 1 min ago', got %s", containers[1].Status)
	}
}

func TestParseContainers_Empty(t *testing.T) {
	containers, err := ParseContainers("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 0 {
		t.Fatalf("expected 0 containers, got %d", len(containers))
	}
}

func TestParseContainers_WhitespaceOnly(t *testing.T) {
	containers, err := ParseContainers("  \n  \n  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 0 {
		t.Fatalf("expected 0 containers, got %d", len(containers))
	}
}

func TestParseContainers_EmptyArray(t *testing.T) {
	containers, err := ParseContainers("[]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 0 {
		t.Fatalf("expected 0 containers, got %d", len(containers))
	}
}

func TestParseContainers_InvalidJSON(t *testing.T) {
	_, err := ParseContainers("{not valid json}")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseContainers_InvalidJSONArray(t *testing.T) {
	_, err := ParseContainers("[{not valid}]")
	if err == nil {
		t.Fatal("expected error for invalid JSON array, got nil")
	}
}

func TestParseContainers_NDJSONWithTrailingNewline(t *testing.T) {
	input := `{"ID":"aaa","Name":"svc-1","State":"running","Status":"Up","Ports":""}
`
	containers, err := ParseContainers(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(containers))
	}
}

func TestParseDockerPS_NDJSON(t *testing.T) {
	input := `{"ID":"abc123def456","Names":"my-web","State":"running","Status":"Up 3 hours","Ports":"0.0.0.0:8080->80/tcp"}
{"ID":"def789012345","Names":"my-db","State":"running","Status":"Up 3 hours","Ports":"3306/tcp"}`

	containers, err := ParseDockerPS(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(containers))
	}

	if containers[0].Name != "my-web" {
		t.Errorf("expected Name my-web, got %s", containers[0].Name)
	}
	if containers[1].Name != "my-db" {
		t.Errorf("expected Name my-db, got %s", containers[1].Name)
	}
	if containers[0].Ports != "0.0.0.0:8080->80/tcp" {
		t.Errorf("expected Ports 0.0.0.0:8080->80/tcp, got %s", containers[0].Ports)
	}
}

func TestParseDockerPS_JSONArray(t *testing.T) {
	input := `[{"ID":"aaa","Names":"svc-a","State":"running","Status":"Up","Ports":""},{"ID":"bbb","Names":"svc-b","State":"exited","Status":"Exited (1)","Ports":""}]`

	containers, err := ParseDockerPS(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(containers))
	}
	if containers[1].State != "exited" {
		t.Errorf("expected State exited, got %s", containers[1].State)
	}
}

func TestParseDockerPS_Empty(t *testing.T) {
	containers, err := ParseDockerPS("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 0 {
		t.Fatalf("expected 0 containers, got %d", len(containers))
	}
}

func TestIsNoComposeConfigError(t *testing.T) {
	tests := []struct {
		stderr string
		want   bool
	}{
		{"no configuration file provided: not found", true},
		{"can't find a suitable configuration file in this directory", true},
		{"no compose file found", true},
		{"error: something else happened", false},
		{"", false},
	}
	for _, tt := range tests {
		got := isNoComposeConfigError(tt.stderr)
		if got != tt.want {
			t.Errorf("isNoComposeConfigError(%q) = %v, want %v", tt.stderr, got, tt.want)
		}
	}
}

func TestIsFormatUnsupported(t *testing.T) {
	tests := []struct {
		stderr string
		want   bool
	}{
		{"unknown flag: --format", true},
		{"Unknown shorthand flag: 'f' in -format", true},
		{"error: container not found", false},
		{"", false},
	}
	for _, tt := range tests {
		got := isFormatUnsupported(tt.stderr)
		if got != tt.want {
			t.Errorf("isFormatUnsupported(%q) = %v, want %v", tt.stderr, got, tt.want)
		}
	}
}
