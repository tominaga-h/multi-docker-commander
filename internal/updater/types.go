package updater

import "net/http"

type Options struct {
	HTTPClient     *http.Client
	APIBaseURL     string
	ExecutablePath string
	CheckPlatform  func() error
}

type Result struct {
	Skipped bool
	Version string
	// LocalNewer is true when the locally-built version is ahead of the
	// latest published release (a dev build), as opposed to being equal.
	LocalNewer bool
}

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type assetURLs struct {
	BinaryURL   string
	ChecksumURL string
	BinaryName  string
}
