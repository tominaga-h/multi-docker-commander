package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func fetchLatestRelease(client *http.Client, apiBaseURL string) (*release, error) {
	if client == nil {
		return nil, fmt.Errorf("http client is required")
	}
	base := strings.TrimRight(apiBaseURL, "/")
	url := base + latestReleaseAPIPath()

	body, err := downloadBytes(client, url)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}

	var release release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("decode release response: %w", err)
	}
	if err := validateTagName(release.TagName); err != nil {
		return nil, err
	}
	return &release, nil
}

func resolveAssetURL(release *release) (*assetURLs, error) {
	if release == nil {
		return nil, fmt.Errorf("release is required")
	}

	binaryName := assetBinaryName(release.TagName)
	checksumName := assetChecksumName(binaryName)

	var binaryURL, checksumURL string
	for _, asset := range release.Assets {
		switch asset.Name {
		case binaryName:
			binaryURL = asset.BrowserDownloadURL
		case checksumName:
			checksumURL = asset.BrowserDownloadURL
		}
	}

	if binaryURL == "" {
		return nil, fmt.Errorf("binary asset %q not found in release %s", binaryName, release.TagName)
	}
	if checksumURL == "" {
		return nil, fmt.Errorf("checksum asset %q not found in release %s", checksumName, release.TagName)
	}

	return &assetURLs{
		BinaryURL:   binaryURL,
		ChecksumURL: checksumURL,
		BinaryName:  binaryName,
	}, nil
}
