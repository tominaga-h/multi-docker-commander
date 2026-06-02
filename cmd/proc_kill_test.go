package cmd

import "testing"

func TestValidateKillFlags(t *testing.T) {
	tests := []struct {
		name                              string
		hasConfig, hasPID, hasAll, hasDead bool
		wantErr                           bool
	}{
		// Non-dead mode: exactly one selector required.
		{name: "none", wantErr: true},
		{name: "config only", hasConfig: true, wantErr: false},
		{name: "pid only", hasPID: true, wantErr: false},
		{name: "all only", hasAll: true, wantErr: false},
		{name: "config and pid", hasConfig: true, hasPID: true, wantErr: true},
		{name: "config and all", hasConfig: true, hasAll: true, wantErr: true},
		{name: "pid and all", hasPID: true, hasAll: true, wantErr: true},
		{name: "all three", hasConfig: true, hasPID: true, hasAll: true, wantErr: true},

		// Dead mode.
		{name: "dead only", hasDead: true, wantErr: false},
		{name: "dead and config", hasDead: true, hasConfig: true, wantErr: false},
		{name: "dead and all", hasDead: true, hasAll: true, wantErr: false},
		{name: "dead config all", hasDead: true, hasConfig: true, hasAll: true, wantErr: false},
		{name: "dead and pid", hasDead: true, hasPID: true, wantErr: true},
		{name: "dead config pid", hasDead: true, hasConfig: true, hasPID: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateKillFlags(tt.hasConfig, tt.hasPID, tt.hasAll, tt.hasDead)
			if tt.wantErr && err == nil {
				t.Errorf("validateKillFlags() = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateKillFlags() = %v, want nil", err)
			}
		})
	}
}
