package cmd

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"mdc/internal/logger"
	"mdc/internal/updater"
)

func TestUpdateCommandRegistered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"update"})
	if err != nil {
		t.Fatalf("rootCmd.Find(update) error = %v", err)
	}
	if cmd.Use != "update" {
		t.Fatalf("command Use = %q, want update", cmd.Use)
	}
}

func TestRunUpdateWithStub(t *testing.T) {
	t.Run("returns nil on successful update", func(t *testing.T) {
		setUpdateFunc(func(opts updater.Options) (updater.Result, error) {
			return updater.Result{Skipped: false, Version: "v2.0.4"}, nil
		})
		t.Cleanup(resetUpdateFunc)

		if err := runUpdate(); err != nil {
			t.Fatalf("runUpdate() error = %v", err)
		}
	})

	t.Run("returns nil when already up to date", func(t *testing.T) {
		setUpdateFunc(func(opts updater.Options) (updater.Result, error) {
			return updater.Result{Skipped: true, Version: "v2.0.3"}, nil
		})
		t.Cleanup(resetUpdateFunc)

		if err := runUpdate(); err != nil {
			t.Fatalf("runUpdate() error = %v", err)
		}
	})

	t.Run("returns error from updater", func(t *testing.T) {
		setUpdateFunc(func(opts updater.Options) (updater.Result, error) {
			return updater.Result{}, errors.New("update failed")
		})
		t.Cleanup(resetUpdateFunc)

		err := runUpdate()
		if err == nil {
			t.Fatal("runUpdate() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "update failed") {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestRunUpdateMessages(t *testing.T) {
	cases := []struct {
		name   string
		result updater.Result
		want   string
	}{
		{
			name:   "updated reports the new release tag",
			result: updater.Result{Skipped: false, Version: "v2.0.4"},
			want:   "updated to v2.0.4",
		},
		{
			name:   "up to date prints the latest release tag",
			result: updater.Result{Skipped: true, Version: "v2.0.3"},
			want:   "already up to date (v2.0.3)",
		},
		{
			name:   "local newer reports being ahead of the latest release",
			result: updater.Result{Skipped: true, LocalNewer: true, Version: "v2.0.3"},
			want:   "ahead of the latest release v2.0.3",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setUpdateFunc(func(opts updater.Options) (updater.Result, error) {
				return tc.result, nil
			})
			t.Cleanup(resetUpdateFunc)

			var buf bytes.Buffer
			logger.SetOutput(&buf)
			t.Cleanup(func() { logger.SetOutput(os.Stdout) })

			if err := runUpdate(); err != nil {
				t.Fatalf("runUpdate() error = %v", err)
			}
			if !strings.Contains(buf.String(), tc.want) {
				t.Fatalf("output = %q, want substring %q", buf.String(), tc.want)
			}
		})
	}
}
