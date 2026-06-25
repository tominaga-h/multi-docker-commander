package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

func downloadAndVerify(client *http.Client, urls *assetURLs) ([]byte, error) {
	if client == nil {
		return nil, fmt.Errorf("http client is required")
	}
	if urls == nil {
		return nil, fmt.Errorf("asset URLs are required")
	}

	// The binary and checksum are independent GETs; fetch them concurrently
	// so the small checksum round-trip overlaps the large binary download.
	var (
		binaryData, checksumData []byte
		binaryErr, checksumErr   error
		wg                       sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		binaryData, binaryErr = downloadBytes(client, urls.BinaryURL)
	}()
	go func() {
		defer wg.Done()
		checksumData, checksumErr = downloadBytes(client, urls.ChecksumURL)
	}()
	wg.Wait()

	if binaryErr != nil {
		return nil, fmt.Errorf("download binary: %w", binaryErr)
	}
	if checksumErr != nil {
		return nil, fmt.Errorf("download checksum: %w", checksumErr)
	}

	expectedHash, err := parseChecksumFile(string(checksumData), urls.BinaryName)
	if err != nil {
		return nil, fmt.Errorf("parse checksum file: %w", err)
	}

	actualHash := sha256.Sum256(binaryData)
	actualHex := hex.EncodeToString(actualHash[:])
	if !strings.EqualFold(actualHex, expectedHash) {
		return nil, fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHex)
	}

	return binaryData, nil
}

func downloadBytes(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("status %d: read body: %w", resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func parseChecksumFile(content, binaryName string) (string, error) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := strings.TrimPrefix(fields[1], "*")
		if name == binaryName {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("checksum for %q not found in checksum file", binaryName)
}
