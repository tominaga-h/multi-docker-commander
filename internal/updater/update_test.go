package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mdc/internal/version"
)

func TestUpdate(t *testing.T) {
	tag := "v9.9.9"
	binaryName := assetBinaryName(tag)
	binaryData := []byte("updated-binary")
	hash := sha256.Sum256(binaryData)
	checksumContent := fmt.Sprintf("%s  %s\n", hex.EncodeToString(hash[:]), binaryName)

	makeServer := func(rel release) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == latestReleaseAPIPath():
				_ = json.NewEncoder(w).Encode(rel)
			case strings.HasSuffix(r.URL.Path, "/binary"):
				_, _ = w.Write(binaryData)
			case strings.HasSuffix(r.URL.Path, ".sha256"):
				_, _ = w.Write([]byte(checksumContent))
			default:
				http.NotFound(w, r)
			}
		}))
	}

	t.Run("skips when already up to date", func(t *testing.T) {
		origVersion := version.Version
		version.Version = "v2.0.3"
		t.Cleanup(func() { version.Version = origVersion })

		rel := release{
			TagName: "v2.0.3",
			Assets: []asset{
				{Name: assetBinaryName("v2.0.3"), BrowserDownloadURL: "/binary"},
				{Name: assetChecksumName(assetBinaryName("v2.0.3")), BrowserDownloadURL: "/checksum.sha256"},
			},
		}
		server := makeServer(rel)
		defer server.Close()

		result, err := Update(Options{
			HTTPClient:    server.Client(),
			APIBaseURL:    server.URL,
			CheckPlatform: func() error { return nil },
		})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if !result.Skipped {
			t.Fatal("expected skipped result")
		}
		if result.LocalNewer {
			t.Fatal("expected LocalNewer false when versions are equal")
		}
	})

	t.Run("skips and reports local newer when ahead of latest", func(t *testing.T) {
		origVersion := version.Version
		version.Version = "v2.1.0"
		t.Cleanup(func() { version.Version = origVersion })

		rel := release{
			TagName: "v2.0.3",
			Assets: []asset{
				{Name: assetBinaryName("v2.0.3"), BrowserDownloadURL: "/binary"},
				{Name: assetChecksumName(assetBinaryName("v2.0.3")), BrowserDownloadURL: "/checksum.sha256"},
			},
		}
		server := makeServer(rel)
		defer server.Close()

		result, err := Update(Options{
			HTTPClient:    server.Client(),
			APIBaseURL:    server.URL,
			CheckPlatform: func() error { return nil },
		})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if !result.Skipped {
			t.Fatal("expected skipped result")
		}
		if !result.LocalNewer {
			t.Fatal("expected LocalNewer true when local version is ahead of latest")
		}
	})

	t.Run("updates binary on newer release", func(t *testing.T) {
		origVersion := version.Version
		version.Version = "2.0.0"
		t.Cleanup(func() { version.Version = origVersion })

		dir := t.TempDir()
		target := filepath.Join(dir, "mdc")
		if err := os.WriteFile(target, []byte("old"), 0755); err != nil {
			t.Fatal(err)
		}

		rel := release{
			TagName: tag,
			Assets: []asset{
				{Name: binaryName, BrowserDownloadURL: "/binary"},
				{Name: assetChecksumName(binaryName), BrowserDownloadURL: "/checksum.sha256"},
			},
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == latestReleaseAPIPath():
				_ = json.NewEncoder(w).Encode(rel)
			case r.URL.Path == "/binary":
				_, _ = w.Write(binaryData)
			case r.URL.Path == "/checksum.sha256":
				_, _ = w.Write([]byte(checksumContent))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		rel.Assets[0].BrowserDownloadURL = server.URL + "/binary"
		rel.Assets[1].BrowserDownloadURL = server.URL + "/checksum.sha256"

		result, err := Update(Options{
			HTTPClient:     server.Client(),
			APIBaseURL:     server.URL,
			ExecutablePath: target,
			CheckPlatform:  func() error { return nil },
		})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if result.Skipped {
			t.Fatal("expected update, got skipped")
		}
		if result.Version != tag {
			t.Fatalf("Version = %q, want %q", result.Version, tag)
		}

		got, err := os.ReadFile(target)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(binaryData) {
			t.Fatalf("binary not replaced")
		}
	})

	t.Run("returns error on unsupported platform", func(t *testing.T) {
		_, err := Update(Options{
			CheckPlatform: func() error {
				return fmt.Errorf("mdc update is only supported on linux/amd64")
			},
		})
		if err == nil {
			t.Fatal("Update() expected error, got nil")
		}
	})

	t.Run("returns error when API fails", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "fail", http.StatusInternalServerError)
		}))
		defer server.Close()

		_, err := Update(Options{
			HTTPClient:    server.Client(),
			APIBaseURL:    server.URL,
			CheckPlatform: func() error { return nil },
		})
		if err == nil {
			t.Fatal("Update() expected error, got nil")
		}
	})
}
