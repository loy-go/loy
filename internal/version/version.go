package version

import "runtime"

var (
	// Version is the semver release tag, set at build time via -ldflags.
	Version = "dev"
	// Commit is the git commit hash.
	Commit = "none"
	// Date is the ISO-8601 build timestamp.
	Date = "unknown"
)

// Info holds structured version information for loy.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Compiler  string `json:"compiler"`
	Platform  string `json:"platform"`
}

// Get returns the populated version Info.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: Date,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
}
