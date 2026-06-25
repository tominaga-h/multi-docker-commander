package updater

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"

	"mdc/internal/version"
)

// defaultHTTPClient bounds every stage of the update download so a stalled
// network cannot hang `mdc update` indefinitely: connect/TLS/response-header
// timeouts catch a dead connection quickly, and the overall Timeout caps the
// whole request (multi-MB binary included).
func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Minute,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: 30 * time.Second}).DialContext,
			TLSHandshakeTimeout:   30 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
		},
	}
}

func Update(opts Options) (Result, error) {
	client := opts.HTTPClient
	if client == nil {
		client = defaultHTTPClient()
	}

	apiBase := opts.APIBaseURL
	if apiBase == "" {
		apiBase = defaultAPIBaseURL
	}

	checkPlatform := opts.CheckPlatform
	if checkPlatform == nil {
		checkPlatform = defaultCheckPlatform
	}
	if err := checkPlatform(); err != nil {
		return Result{}, err
	}

	release, err := fetchLatestRelease(client, apiBase)
	if err != nil {
		return Result{}, err
	}

	needsUpdate, localNewer := compareVersion(version.Version, release.TagName)
	if !needsUpdate {
		return Result{Skipped: true, Version: release.TagName, LocalNewer: localNewer}, nil
	}

	urls, err := resolveAssetURL(release)
	if err != nil {
		return Result{}, err
	}

	binaryData, err := downloadAndVerify(client, urls)
	if err != nil {
		return Result{}, err
	}

	if err := replaceExecutable(binaryData, opts.ExecutablePath); err != nil {
		return Result{}, err
	}

	return Result{Skipped: false, Version: release.TagName}, nil
}

func defaultCheckPlatform() error {
	if runtime.GOOS != platformOS || runtime.GOARCH != platformArch {
		return fmt.Errorf("mdc update is only supported on %s/%s (current: %s/%s)",
			platformOS, platformArch, runtime.GOOS, runtime.GOARCH)
	}
	return nil
}
