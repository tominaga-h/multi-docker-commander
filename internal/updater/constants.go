package updater

import "fmt"

const (
	githubOwner = "tominaga-h"
	githubRepo  = "multi-docker-commander"

	defaultAPIBaseURL = "https://api.github.com"

	platformOS   = "linux"
	platformArch = "amd64"
)

func latestReleaseAPIPath() string {
	return fmt.Sprintf("/repos/%s/%s/releases/latest", githubOwner, githubRepo)
}

// assetBinaryName / assetChecksumName must stay in lockstep with the release
// asset naming in .github/workflows/release.yml (BIN_NAME = mdc-<ver>-<goos>-<goarch>,
// checksum = <BIN_NAME>.sha256), which is the source of truth. If that workflow
// changes the pattern, update these or `mdc update` will fail to match any asset.
func assetBinaryName(tag string) string {
	return fmt.Sprintf("mdc-%s-%s-%s", tag, platformOS, platformArch)
}

func assetChecksumName(binaryName string) string {
	return binaryName + ".sha256"
}
