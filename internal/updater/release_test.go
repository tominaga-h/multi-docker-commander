package updater

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchLatestRelease(t *testing.T) {
	t.Run("returns release on success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != latestReleaseAPIPath() {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			_ = json.NewEncoder(w).Encode(release{
				TagName: "v2.0.4",
				Assets:  []asset{{Name: assetBinaryName("v2.0.4"), BrowserDownloadURL: "http://example/binary"}},
			})
		}))
		defer server.Close()

		got, err := fetchLatestRelease(server.Client(), server.URL)
		if err != nil {
			t.Fatalf("fetchLatestRelease() error = %v", err)
		}
		if got.TagName != "v2.0.4" {
			t.Fatalf("TagName = %q, want v2.0.4", got.TagName)
		}
	})

	t.Run("returns error on API failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "not found", http.StatusNotFound)
		}))
		defer server.Close()

		_, err := fetchLatestRelease(server.Client(), server.URL)
		if err == nil {
			t.Fatal("fetchLatestRelease() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "404") {
			t.Fatalf("error = %v, want status 404 mention", err)
		}
	})
}

func TestResolveAssetURL(t *testing.T) {
	tag := "v2.0.3"
	binaryName := assetBinaryName(tag)
	checksumName := assetChecksumName(binaryName)

	t.Run("resolves binary and checksum URLs", func(t *testing.T) {
		rel := &release{
			TagName: tag,
			Assets: []asset{
				{Name: binaryName, BrowserDownloadURL: "http://example/binary"},
				{Name: checksumName, BrowserDownloadURL: "http://example/checksum"},
			},
		}

		urls, err := resolveAssetURL(rel)
		if err != nil {
			t.Fatalf("resolveAssetURL() error = %v", err)
		}
		if urls.BinaryURL != "http://example/binary" {
			t.Fatalf("BinaryURL = %q", urls.BinaryURL)
		}
		if urls.ChecksumURL != "http://example/checksum" {
			t.Fatalf("ChecksumURL = %q", urls.ChecksumURL)
		}
	})

	t.Run("returns error when binary asset is missing", func(t *testing.T) {
		rel := &release{
			TagName: tag,
			Assets: []asset{
				{Name: checksumName, BrowserDownloadURL: "http://example/checksum"},
			},
		}

		_, err := resolveAssetURL(rel)
		if err == nil {
			t.Fatal("resolveAssetURL() expected error, got nil")
		}
		if !strings.Contains(err.Error(), binaryName) {
			t.Fatalf("error = %v, want binary name mention", err)
		}
	})

	t.Run("returns error when checksum asset is missing", func(t *testing.T) {
		rel := &release{
			TagName: tag,
			Assets: []asset{
				{Name: binaryName, BrowserDownloadURL: "http://example/binary"},
			},
		}

		_, err := resolveAssetURL(rel)
		if err == nil {
			t.Fatal("resolveAssetURL() expected error, got nil")
		}
		if !strings.Contains(err.Error(), checksumName) {
			t.Fatalf("error = %v, want checksum name mention", err)
		}
	})
}
