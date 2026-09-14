package doctor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/uloydev/loy/internal/diagnostics"
	"github.com/uloydev/loy/internal/discovery"
	"github.com/uloydev/loy/internal/filesystem"
	"github.com/uloydev/loy/internal/manifest"
	"github.com/uloydev/loy/internal/process"
)

var goVersionRegex = regexp.MustCompile(`go(\d+)\.(\d+)`)

// CheckItem models an individual health check result.
type CheckItem struct {
	Name        string                  `json:"name"`
	Passed      bool                    `json:"passed"`
	Detail      string                  `json:"detail"`
	Diagnostic  *diagnostics.Diagnostic `json:"diagnostic,omitempty"`
}

// Doctor aggregates and executes system and project diagnostics.
type Doctor struct {
	fs     filesystem.FileSystem
	runner process.Runner
}

// NewDoctor creates a new Doctor instance.
func NewDoctor(fs filesystem.FileSystem, runner process.Runner) *Doctor {
	if fs == nil {
		fs = filesystem.NewOSFileSystem()
	}
	if runner == nil {
		runner = process.NewExecRunner()
	}
	return &Doctor{
		fs:     fs,
		runner: runner,
	}
}

// Run executes all health and environment checks on the specified directory.
func (d *Doctor) Run(ctx context.Context, targetDir string) ([]CheckItem, error) {
	var results []CheckItem

	// 1. Go Compiler Check
	results = append(results, d.checkGo(ctx))

	// 2. Git Check
	results = append(results, d.checkTool(ctx, "git", "git --version", diagnostics.CodeDoctorGitMissing, "Install git via your system package manager"))

	// 3. sqlc CLI Check
	results = append(results, d.checkTool(ctx, "sqlc", "sqlc version", diagnostics.CodeDoctorSqlcMissing, "run 'go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest'"))

	// 4. Docker CLI Check
	results = append(results, d.checkDocker(ctx))

	// 5. Project Manifest & Module Check
	results = append(results, d.checkProject(ctx, targetDir)...)

	// 6. Fullstack / Web Tooling Checks (Templ, Node, Package Manager)
	results = append(results, d.checkFullstackTools(ctx, targetDir)...)

	return results, nil
}

func (d *Doctor) checkGo(ctx context.Context) CheckItem {
	res, err := d.runner.Run(ctx, "", "go", "version")
	if err != nil || res.ExitCode != 0 {
		return CheckItem{
			Name:   "Go Compiler",
			Passed: false,
			Detail: "Go compiler not found in PATH",
			Diagnostic: &diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Code:     diagnostics.CodeDoctorGoMissing,
				Message:  "Go toolchain is missing",
				Hint:     "Install Go 1.22+ from https://go.dev/dl/",
			},
		}
	}

	versionStr := strings.TrimSpace(string(res.Stdout))
	// Parse version e.g. "go version go1.23.1 linux/amd64"
	matches := goVersionRegex.FindStringSubmatch(versionStr)
	if len(matches) >= 3 {
		major, _ := strconv.Atoi(matches[1])
		minor, _ := strconv.Atoi(matches[2])
		if major < 1 || (major == 1 && minor < 22) {
			return CheckItem{
				Name:   "Go Compiler",
				Passed: false,
				Detail: fmt.Sprintf("Unsupported version %s (minimum Go 1.22 required)", versionStr),
				Diagnostic: &diagnostics.Diagnostic{
					Severity: diagnostics.SeverityError,
					Code:     diagnostics.CodeDoctorGoMissing,
					Message:  "Go version is older than 1.22",
					Hint:     "Upgrade Go to 1.22 or higher",
				},
			}
		}
	}

	return CheckItem{
		Name:   "Go Compiler",
		Passed: true,
		Detail: versionStr,
	}
}

func (d *Doctor) checkTool(ctx context.Context, name, command, code, hint string) CheckItem {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return CheckItem{Name: name, Passed: false}
	}

	res, err := d.runner.Run(ctx, "", parts[0], parts[1:]...)
	if err != nil || res.ExitCode != 0 {
		return CheckItem{
			Name:   name,
			Passed: false,
			Detail: fmt.Sprintf("%s not found in PATH", name),
			Diagnostic: &diagnostics.Diagnostic{
				Severity: diagnostics.SeverityWarning,
				Code:     code,
				Message:  fmt.Sprintf("%s is not installed or not in PATH", name),
				Hint:     hint,
			},
		}
	}

	out := strings.TrimSpace(string(res.Stdout))
	if out == "" {
		out = "installed"
	} else if len(out) > 50 {
		out = out[:50] + "..."
	}

	return CheckItem{
		Name:   name,
		Passed: true,
		Detail: out,
	}
}

func (d *Doctor) checkDocker(ctx context.Context) CheckItem {
	res, err := d.runner.Run(ctx, "", "docker", "info")
	if err != nil || res.ExitCode != 0 {
		// Distinguish docker not installed vs docker daemon not running
		versionRes, vErr := d.runner.Run(ctx, "", "docker", "--version")
		if vErr != nil || versionRes.ExitCode != 0 {
			return CheckItem{
				Name:   "Docker",
				Passed: false,
				Detail: "Docker CLI not found in PATH",
				Diagnostic: &diagnostics.Diagnostic{
					Severity: diagnostics.SeverityWarning,
					Code:     diagnostics.CodeDoctorDockerMissing,
					Message:  "Docker is not installed",
					Hint:     "Install Docker from https://docs.docker.com/get-docker/",
				},
			}
		}
		return CheckItem{
			Name:   "Docker",
			Passed: false,
			Detail: "Docker daemon is not running",
			Diagnostic: &diagnostics.Diagnostic{
				Severity: diagnostics.SeverityWarning,
				Code:     diagnostics.CodeDoctorDockerMissing,
				Message:  "Docker daemon is not reachable",
				Hint:     "Start the Docker service or daemon",
			},
		}
	}

	return CheckItem{
		Name:   "Docker",
		Passed: true,
		Detail: "Docker daemon is active and running",
	}
}

func (d *Doctor) checkProject(ctx context.Context, targetDir string) []CheckItem {
	if targetDir == "" {
		targetDir, _ = os.Getwd()
	}

	var items []CheckItem

	disc, err := discovery.NewDiscoverer(d.fs, d.runner)
	if err != nil {
		return items
	}

	res, diag := disc.Discover(ctx, targetDir)
	if diag != nil || res == nil || !res.HasGoMod {
		items = append(items, CheckItem{
			Name:   "Go Module",
			Passed: false,
			Detail: "go.mod not found in project or parent directories",
			Diagnostic: &diagnostics.Diagnostic{
				Severity: diagnostics.SeverityWarning,
				Code:     diagnostics.CodeDoctorModuleMissing,
				Message:  "Current directory is not inside a Go module",
				Hint:     "Run 'go mod init <module>' or 'loy new <app>'",
			},
		})
	} else {
		items = append(items, CheckItem{
			Name:   "Go Module",
			Passed: true,
			Detail: fmt.Sprintf("Found go.mod at %s", res.GoModPath),
		})
	}

	if res != nil && res.HasManifest {
		mPath := res.ManifestPath
		data, err := d.fs.ReadFile(mPath)
		if err != nil {
			items = append(items, CheckItem{
				Name:   "Loy Manifest",
				Passed: false,
				Detail: fmt.Sprintf("Failed reading %s: %v", mPath, err),
				Diagnostic: &diagnostics.Diagnostic{
					Severity: diagnostics.SeverityError,
					Code:     diagnostics.CodeDoctorManifestMissing,
					Message:  fmt.Sprintf("Failed reading manifest %s: %v", mPath, err),
					Hint:     "Check file permissions or reinitialize with 'loy init'",
					File:     mPath,
				},
			})
		} else {
			parser := manifest.NewParser()
			man, mDiag := parser.ParseStrict(mPath, data)
			if mDiag != nil {
				items = append(items, CheckItem{
					Name:       "Loy Manifest",
					Passed:     false,
					Detail:     mDiag.Message,
					Diagnostic: mDiag,
				})
			} else {
				items = append(items, CheckItem{
					Name:   "Loy Manifest",
					Passed: true,
					Detail: fmt.Sprintf("Valid manifest (project: %s, version: %d)", man.Project.Name, man.Version),
				})
			}
		}
	} else {
		items = append(items, CheckItem{
			Name:   "Loy Manifest",
			Passed: false,
			Detail: "loy.yaml not found",
			Diagnostic: &diagnostics.Diagnostic{
				Severity: diagnostics.SeverityWarning,
				Code:     diagnostics.CodeDoctorManifestMissing,
				Message:  "Project missing loy.yaml configuration",
				Hint:     "Run 'loy init' to create a standard manifest",
			},
		})
	}

	return items
}

func (d *Doctor) checkFullstackTools(ctx context.Context, targetDir string) []CheckItem {
	if targetDir == "" {
		targetDir, _ = os.Getwd()
	}

	var items []CheckItem

	hasViews, _ := d.fs.Exists(filepath.Join(targetDir, "views"))
	hasPkgJson, _ := d.fs.Exists(filepath.Join(targetDir, "package.json"))
	hasViteConfig, _ := d.fs.Exists(filepath.Join(targetDir, "vite.config.ts"))

	// Also check manifest if available
	mPath := filepath.Join(targetDir, "loy.yaml")
	if data, err := d.fs.ReadFile(mPath); err == nil {
		strData := string(data)
		if strings.Contains(strData, "template: templ") {
			hasViews = true
		}
		if strings.Contains(strData, "assets: vite") {
			hasPkgJson = true
		}
	}

	if hasViews {
		items = append(items, d.checkTool(ctx, "templ CLI", "templ version", diagnostics.CodeDoctorTemplMissing, "run 'go install github.com/a-h/templ/cmd/templ@latest'"))
	}

	if hasPkgJson || hasViteConfig {
		items = append(items, d.checkTool(ctx, "Node.js", "node --version", diagnostics.CodeDoctorNodeMissing, "Install Node.js 18+ from https://nodejs.org/"))

		// Detect package manager
		pm := "npm"
		if exists, _ := d.fs.Exists(filepath.Join(targetDir, "pnpm-lock.yaml")); exists {
			pm = "pnpm"
		} else if exists, _ := d.fs.Exists(filepath.Join(targetDir, "bun.lockb")); exists {
			pm = "bun"
		} else if exists, _ := d.fs.Exists(filepath.Join(targetDir, "yarn.lock")); exists {
			pm = "yarn"
		}

		items = append(items, d.checkTool(ctx, pm+" Package Manager", pm+" --version", diagnostics.CodeDoctorPackageMgrMissing, fmt.Sprintf("Install %s via your system package manager", pm)))
	}

	return items
}
