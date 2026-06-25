package updater

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/mod/semver"
)

// gitDescribeSuffix matches the "<commits>-g<hash>" trailer that
// `git describe --tags` appends to a tag when HEAD is past it
// (e.g. "v2.0.3-5-gabcdef" → suffix "-5-gabcdef").
var gitDescribeSuffix = regexp.MustCompile(`-[0-9]+-g[0-9a-f]+$`)

func DisplayVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

// compareVersion reports how the locally-built `current` version relates to
// the `latest` release tag.
//
//   - needsUpdate is true when `latest` is a newer release than `current`
//     (or when the versions cannot be compared, so an explicit `mdc update`
//     still proceeds).
//   - localNewer is true when `current` is ahead of `latest` — e.g. a
//     `make build-v` dev build with commits past the latest tag.
//
// git-describe metadata ("-<commits>-g<hash>" and/or "-dirty") is stripped
// before comparing so a dev build ranks at or above its base tag instead of
// being treated as a prerelease below it (which would cause a silent
// downgrade to the older release binary).
func compareVersion(current, latest string) (needsUpdate, localNewer bool) {
	latestNorm := normalizeVersion(latest)
	if !semver.IsValid(latestNorm) {
		// The published tag is not semver; we cannot reason about it, but the
		// user explicitly asked to update, so attempt it.
		return true, false
	}

	base, ahead := baseReleaseTag(current)
	currentNorm := normalizeVersion(base)
	if !semver.IsValid(currentNorm) {
		// Unidentifiable local build (e.g. "dev" or a bare commit hash);
		// attempt the update the user asked for.
		return true, false
	}

	switch cmp := semver.Compare(currentNorm, latestNorm); {
	case cmp < 0:
		return true, false
	case cmp > 0:
		return false, true
	default:
		// Base tag equals the latest release. A dev build with commits past
		// the tag (or a dirty worktree) is ahead of the published release.
		return false, ahead
	}
}

// baseReleaseTag strips git-describe metadata from a version string,
// returning the underlying release tag and whether the build is ahead of
// that tag (extra commits and/or a dirty worktree). Genuine prerelease tags
// such as "v2.1.0-rc1" are left untouched.
func baseReleaseTag(v string) (base string, ahead bool) {
	s := strings.TrimSpace(v)
	if trimmed := strings.TrimSuffix(s, "-dirty"); trimmed != s {
		s = trimmed
		ahead = true
	}
	if loc := gitDescribeSuffix.FindStringIndex(s); loc != nil {
		s = s[:loc[0]]
		ahead = true
	}
	return s, ahead
}

func normalizeVersion(v string) string {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return ""
	}
	withoutV := strings.TrimPrefix(trimmed, "v")
	return semver.Canonical("v" + withoutV)
}

func validateTagName(tag string) error {
	if strings.TrimSpace(tag) == "" {
		return fmt.Errorf("release tag_name is empty")
	}
	return nil
}
