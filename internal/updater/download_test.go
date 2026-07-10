package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDownloadAndVerify(t *testing.T) {
	binaryName := assetBinaryName("v2.0.3")
	binaryData := []byte("fake-binary-content")
	hash := sha256.Sum256(binaryData)
	checksumContent := fmt.Sprintf("%s  %s\n", hex.EncodeToString(hash[:]), binaryName)

	urls := &assetURLs{
		BinaryURL:   "/binary",
		ChecksumURL: "/checksum",
		BinaryName:  binaryName,
	}

	t.Run("verifies matching checksum", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/binary":
				_, _ = w.Write(binaryData)
			case "/checksum":
				_, _ = w.Write([]byte(checksumContent))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		urlsWithHost := *urls
		urlsWithHost.BinaryURL = server.URL + "/binary"
		urlsWithHost.ChecksumURL = server.URL + "/checksum"

		got, err := downloadAndVerify(server.Client(), &urlsWithHost)
		if err != nil {
			t.Fatalf("downloadAndVerify() error = %v", err)
		}
		if string(got) != string(binaryData) {
			t.Fatalf("binary data mismatch")
		}
	})

	t.Run("returns error when binary download fails", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/binary":
				http.Error(w, "not found", http.StatusNotFound)
			case "/checksum":
				_, _ = w.Write([]byte(checksumContent))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		urlsWithHost := *urls
		urlsWithHost.BinaryURL = server.URL + "/binary"
		urlsWithHost.ChecksumURL = server.URL + "/checksum"

		_, err := downloadAndVerify(server.Client(), &urlsWithHost)
		if err == nil {
			t.Fatal("downloadAndVerify() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "download binary") {
			t.Fatalf("error = %v, want download binary failure", err)
		}
	})

	t.Run("returns error when checksum download fails", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/binary":
				_, _ = w.Write(binaryData)
			case "/checksum":
				http.Error(w, "internal error", http.StatusInternalServerError)
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		urlsWithHost := *urls
		urlsWithHost.BinaryURL = server.URL + "/binary"
		urlsWithHost.ChecksumURL = server.URL + "/checksum"

		_, err := downloadAndVerify(server.Client(), &urlsWithHost)
		if err == nil {
			t.Fatal("downloadAndVerify() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "download checksum") {
			t.Fatalf("error = %v, want download checksum failure", err)
		}
	})

	t.Run("returns error on checksum mismatch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/binary":
				_, _ = w.Write(binaryData)
			case "/checksum":
				_, _ = w.Write([]byte("deadbeef  " + binaryName + "\n"))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		urlsWithHost := *urls
		urlsWithHost.BinaryURL = server.URL + "/binary"
		urlsWithHost.ChecksumURL = server.URL + "/checksum"

		_, err := downloadAndVerify(server.Client(), &urlsWithHost)
		if err == nil {
			t.Fatal("downloadAndVerify() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "checksum mismatch") {
			t.Fatalf("error = %v, want checksum mismatch", err)
		}
	})
}

func TestParseChecksumFile(t *testing.T) {
	binaryName := "mdc-v2.0.3-linux-amd64"
	content := "abc123  " + binaryName + "\n"

	got, err := parseChecksumFile(content, binaryName)
	if err != nil {
		t.Fatalf("parseChecksumFile() error = %v", err)
	}
	if got != "abc123" {
		t.Fatalf("hash = %q, want abc123", got)
	}
}
