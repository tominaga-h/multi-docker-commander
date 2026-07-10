package updater

import (
	"testing"
)

func TestCompareVersion(t *testing.T) {
	tests := []struct {
		name        string
		current     string
		latest      string
		needsUpdate bool
		localNewer  bool
	}{
		{
			name:        "same version without v prefix",
			current:     "2.0.3",
			latest:      "v2.0.3",
			needsUpdate: false,
			localNewer:  false,
		},
		{
			name:        "same version with v prefix",
			current:     "v2.0.3",
			latest:      "2.0.3",
			needsUpdate: false,
			localNewer:  false,
		},
		{
			name:        "current is older",
			current:     "2.0.2",
			latest:      "v2.0.3",
			needsUpdate: true,
			localNewer:  false,
		},
		{
			name:        "current is newer",
			current:     "v2.1.0",
			latest:      "v2.0.3",
			needsUpdate: false,
			localNewer:  true,
		},
		{
			name:        "git-describe dev build past the latest tag is not a downgrade",
			current:     "v2.0.3-5-gabcdef",
			latest:      "v2.0.3",
			needsUpdate: false,
			localNewer:  true,
		},
		{
			name:        "git-describe dev build older than latest still updates",
			current:     "v2.0.3-5-gabcdef",
			latest:      "v2.0.4",
			needsUpdate: true,
			localNewer:  false,
		},
		{
			name:        "dirty build of the latest tag is not a downgrade",
			current:     "v2.0.3-dirty",
			latest:      "v2.0.3",
			needsUpdate: false,
			localNewer:  true,
		},
		{
			name:        "genuine prerelease ranks below its release",
			current:     "v2.0.3-rc1",
			latest:      "v2.0.3",
			needsUpdate: true,
			localNewer:  false,
		},
		{
			name:        "non-semver current forces update attempt",
			current:     "dev",
			latest:      "v2.0.3",
			needsUpdate: true,
			localNewer:  false,
		},
		{
			name:        "non-semver latest forces update attempt",
			current:     "2.0.3",
			latest:      "snapshot",
			needsUpdate: true,
			localNewer:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			needsUpdate, localNewer := compareVersion(tt.current, tt.latest)
			if needsUpdate != tt.needsUpdate || localNewer != tt.localNewer {
				t.Fatalf("compareVersion(%q, %q) = (%v, %v), want (%v, %v)",
					tt.current, tt.latest, needsUpdate, localNewer, tt.needsUpdate, tt.localNewer)
			}
		})
	}
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "2.0.3", want: "v2.0.3"},
		{input: "v2.0.3", want: "v2.0.3"},
		{input: " v2.0.3 ", want: "v2.0.3"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := normalizeVersion(tt.input); got != tt.want {
				t.Fatalf("normalizeVersion(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
