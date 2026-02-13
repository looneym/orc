package version

import "fmt"

// These variables are set at build time via ldflags
var (
	Commit    = "unknown"
	BuildTime = "unknown"
	DevBuild  = "" // Set to "true" by `make dev`; empty for installed builds
)

// String returns the version string (commit-hash based, no semver)
func String() string {
	return fmt.Sprintf("orc dev (commit: %s, built: %s)", shortCommit(), BuildTime)
}

func shortCommit() string {
	if len(Commit) > 7 {
		return Commit[:7]
	}
	return Commit
}
