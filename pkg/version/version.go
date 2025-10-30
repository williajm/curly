// Package version provides version information for the curly application.
package version

import (
	"fmt"
	"runtime"
)

// Version information. These variables are set via ldflags during build.
var (
	// Version is the semantic version of the application.
	Version = "dev"
	// Commit is the git commit hash.
	Commit = "unknown"
	// BuildDate is the date the binary was built.
	BuildDate = "unknown"
)

// Info holds the version information.
type Info struct {
	Version   string
	Commit    string
	BuildDate string
	GoVersion string
	Platform  string
}

// Get returns the version information.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string.
func (i Info) String() string {
	return fmt.Sprintf("curly %s (commit: %s, built: %s, go: %s, platform: %s)",
		i.Version, i.Commit, i.BuildDate, i.GoVersion, i.Platform)
}
