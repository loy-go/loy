package version

import (
	"runtime"
	"runtime/debug"
)

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

// Get returns the populated version Info, falling back to runtime/debug.ReadBuildInfo()
// when build-time ldflags are absent (e.g. installed via 'go install').
func Get() Info {
	v := Version
	c := Commit
	d := Date

	if (v == "dev" || c == "none" || d == "unknown") {
		if bi, ok := debug.ReadBuildInfo(); ok {
			if v == "dev" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
				v = bi.Main.Version
			}
			var rev string
			var modified bool
			for _, s := range bi.Settings {
				switch s.Key {
				case "vcs.revision":
					rev = s.Value
				case "vcs.time":
					if d == "unknown" && s.Value != "" {
						d = s.Value
					}
				case "vcs.modified":
					modified = s.Value == "true"
				}
			}
			if c == "none" && rev != "" {
				c = rev
				if modified {
					c += "-dirty"
				}
			}
		}
	}

	return Info{
		Version:   v,
		Commit:    c,
		BuildDate: d,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
}
