package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/loy-go/loy/internal/version"
	"golang.org/x/mod/semver"
)

const (
	DefaultGitHubRepo = "loy-go/loy"
)

// ReleaseAsset represents an asset in a GitHub release.
type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// GitHubRelease represents release metadata returned by GitHub API.
type GitHubRelease struct {
	TagName     string         `json:"tag_name"`
	Name        string         `json:"name"`
	PublishedAt string         `json:"published_at"`
	Assets      []ReleaseAsset `json:"assets"`
}

// ReleaseInfo models resolved release assets.
type ReleaseInfo struct {
	TagName        string
	Version        string
	ArchiveName    string
	ArchiveURL     string
	ChecksumsURL   string
	PublishedAt    string
}

// CheckResult represents the outcome of an update check.
type CheckResult struct {
	CurrentVersion   string `json:"current_version"`
	TargetVersion    string `json:"target_version"`
	UpdateAvailable  bool   `json:"update_available"`
	IsManagedPackage bool   `json:"is_managed_package"`
	PackageManager   string `json:"package_manager,omitempty"`
}

// ApplyResult represents the outcome of an applied update.
type ApplyResult struct {
	PreviousVersion string `json:"previous_version"`
	UpdatedVersion  string `json:"updated_version"`
	ExecutablePath  string `json:"executable_path"`
}

// HTTPClient abstracts HTTP calls for testing.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Updater manages checking, downloading, verifying, and atomically applying binary updates.
type Updater struct {
	repo           string
	httpClient     HTTPClient
	currentVersion string
	executablePath string
	goos           string
	goarch         string
}

// NewUpdater constructs an Updater.
func NewUpdater(repo string, client HTTPClient) (*Updater, error) {
	if repo == "" {
		repo = DefaultGitHubRepo
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolving current executable path: %w", err)
	}

	// Resolve symlinks to find the real binary path
	if evalPath, err := filepath.EvalSymlinks(execPath); err == nil {
		execPath = evalPath
	}

	curVer := version.Get().Version
	if !strings.HasPrefix(curVer, "v") && curVer != "dev" {
		curVer = "v" + curVer
	}

	return &Updater{
		repo:           repo,
		httpClient:     client,
		currentVersion: curVer,
		executablePath: execPath,
		goos:           runtime.GOOS,
		goarch:         runtime.GOARCH,
	}, nil
}

// DetectPackageManager checks if the binary is installed via Homebrew or Nix.
func (u *Updater) DetectPackageManager() (isManaged bool, name string) {
	p := filepath.ToSlash(u.executablePath)
	if strings.Contains(p, "/Cellar/") || strings.Contains(p, "/homebrew/") || strings.Contains(p, "/opt/homebrew/") {
		return true, "homebrew"
	}
	if strings.Contains(p, "/nix/store/") {
		return true, "nix"
	}
	return false, ""
}

// Check queries GitHub Releases and determines if an update is available.
func (u *Updater) Check(ctx context.Context, targetTag string) (*CheckResult, *ReleaseInfo, error) {
	rel, err := u.fetchRelease(ctx, targetTag)
	if err != nil {
		return nil, nil, err
	}

	info, err := u.resolveAssets(rel)
	if err != nil {
		return nil, nil, err
	}

	isManaged, pkgMgr := u.DetectPackageManager()

	updateAvailable := false
	if targetTag != "" {
		updateAvailable = info.Version != u.currentVersion
	} else if u.currentVersion == "dev" {
		updateAvailable = true
	} else {
		// Semver comparison
		curV := u.currentVersion
		if !strings.HasPrefix(curV, "v") {
			curV = "v" + curV
		}
		targetV := info.Version
		if !strings.HasPrefix(targetV, "v") {
			targetV = "v" + targetV
		}
		if semver.IsValid(curV) && semver.IsValid(targetV) {
			updateAvailable = semver.Compare(targetV, curV) > 0
		} else {
			updateAvailable = targetV != curV
		}
	}

	return &CheckResult{
		CurrentVersion:   u.currentVersion,
		TargetVersion:    info.Version,
		UpdateAvailable:  updateAvailable,
		IsManagedPackage: isManaged,
		PackageManager:   pkgMgr,
	}, info, nil
}

// Apply downloads, cryptographically verifies, and replaces the current binary atomically.
func (u *Updater) Apply(ctx context.Context, targetTag string) (*ApplyResult, error) {
	isManaged, pkgMgr := u.DetectPackageManager()
	if isManaged {
		switch pkgMgr {
		case "homebrew":
			return nil, fmt.Errorf("loy was installed via Homebrew; please update via: brew upgrade loy")
		case "nix":
			return nil, fmt.Errorf("loy was installed via Nix; please update via nix profile or flake")
		default:
			return nil, fmt.Errorf("loy is managed by %s; please update via your package manager", pkgMgr)
		}
	}

	_, info, err := u.Check(ctx, targetTag)
	if err != nil {
		return nil, fmt.Errorf("checking release: %w", err)
	}

	if info.ArchiveURL == "" {
		return nil, fmt.Errorf("no matching archive asset found for %s/%s in release %s", u.goos, u.goarch, info.TagName)
	}

	// 1. Download checksums.txt
	checksumsData, err := u.downloadBytes(ctx, info.ChecksumsURL)
	if err != nil {
		return nil, fmt.Errorf("downloading release checksums: %w", err)
	}

	// 2. Download release archive
	archiveData, err := u.downloadBytes(ctx, info.ArchiveURL)
	if err != nil {
		return nil, fmt.Errorf("downloading release archive: %w", err)
	}

	// 3. Verify SHA256 Checksum
	expectedHash, err := extractExpectedChecksum(checksumsData, info.ArchiveName)
	if err != nil {
		return nil, fmt.Errorf("verifying checksum: %w", err)
	}

	hasher := sha256.New()
	hasher.Write(archiveData)
	actualHash := hex.EncodeToString(hasher.Sum(nil))

	if !strings.EqualFold(actualHash, expectedHash) {
		return nil, fmt.Errorf("checksum mismatch for %s: expected %s, got %s", info.ArchiveName, expectedHash, actualHash)
	}

	// 4. Extract executable from archive
	newBinaryBytes, err := extractBinary(archiveData, info.ArchiveName, u.goos)
	if err != nil {
		return nil, fmt.Errorf("extracting binary from archive: %w", err)
	}

	// 5. Atomically replace executable
	if err := u.replaceExecutable(newBinaryBytes); err != nil {
		return nil, fmt.Errorf("replacing binary: %w", err)
	}

	return &ApplyResult{
		PreviousVersion: u.currentVersion,
		UpdatedVersion:  info.Version,
		ExecutablePath:  u.executablePath,
	}, nil
}

func (u *Updater) fetchRelease(ctx context.Context, tag string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", u.repo)
	if tag != "" {
		if !strings.HasPrefix(tag, "v") && !strings.Contains(tag, "/") {
			tag = "v" + tag
		}
		url = fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", u.repo, tag)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "loy-updater/"+u.currentVersion)

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching release metadata: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release %s not found on GitHub", tag)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var rel GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("parsing release metadata: %w", err)
	}

	return &rel, nil
}

func (u *Updater) resolveAssets(rel *GitHubRelease) (*ReleaseInfo, error) {
	info := &ReleaseInfo{
		TagName:     rel.TagName,
		Version:     strings.TrimPrefix(rel.TagName, "v"),
		PublishedAt: rel.PublishedAt,
	}

	// Archive name pattern: loy_<version>_<os>_<arch>.tar.gz (or .zip on windows)
	// Example: loy_0.3.0_linux_amd64.tar.gz or loy_0.3.0_darwin_arm64.tar.gz
	ext := ".tar.gz"
	if u.goos == "windows" {
		ext = ".zip"
	}

	targetSuffix := fmt.Sprintf("_%s_%s%s", u.goos, u.goarch, ext)

	for _, asset := range rel.Assets {
		if asset.Name == "checksums.txt" {
			info.ChecksumsURL = asset.BrowserDownloadURL
		}
		if strings.HasSuffix(asset.Name, targetSuffix) {
			info.ArchiveName = asset.Name
			info.ArchiveURL = asset.BrowserDownloadURL
		}
	}

	if info.ChecksumsURL == "" {
		return nil, fmt.Errorf("release %s missing checksums.txt", rel.TagName)
	}
	if info.ArchiveURL == "" {
		return nil, fmt.Errorf("release %s has no matching asset for %s/%s", rel.TagName, u.goos, u.goarch)
	}

	return info, nil
}

func (u *Updater) downloadBytes(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "loy-updater/"+u.currentVersion)

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with HTTP %d for %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}

func extractExpectedChecksum(checksumsData []byte, assetName string) (string, error) {
	lines := strings.Split(string(checksumsData), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == assetName {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("asset %s not found in checksums.txt", assetName)
}

func extractBinary(archiveData []byte, archiveName, targetOS string) ([]byte, error) {
	binName := "loy"
	if targetOS == "windows" {
		binName = "loy.exe"
	}

	if strings.HasSuffix(archiveName, ".tar.gz") || strings.HasSuffix(archiveName, ".tgz") {
		gzr, err := gzip.NewReader(bytes.NewReader(archiveData))
		if err != nil {
			return nil, fmt.Errorf("opening gzip reader: %w", err)
		}
		defer func() { _ = gzr.Close() }()

		tr := tar.NewReader(gzr)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("reading tar header: %w", err)
			}
			if filepath.Base(hdr.Name) == binName && hdr.Typeflag == tar.TypeReg {
				return io.ReadAll(tr)
			}
		}
		return nil, fmt.Errorf("binary %s not found in tar.gz archive", binName)
	}

	if strings.HasSuffix(archiveName, ".zip") {
		zr, err := zip.NewReader(bytes.NewReader(archiveData), int64(len(archiveData)))
		if err != nil {
			return nil, fmt.Errorf("opening zip reader: %w", err)
		}

		for _, f := range zr.File {
			if filepath.Base(f.Name) == binName {
				rc, err := f.Open()
				if err != nil {
					return nil, fmt.Errorf("opening zip file %s: %w", f.Name, err)
				}
				defer func() { _ = rc.Close() }()
				return io.ReadAll(rc)
			}
		}
		return nil, fmt.Errorf("binary %s not found in zip archive", binName)
	}

	return nil, fmt.Errorf("unsupported archive format: %s", archiveName)
}

func (u *Updater) replaceExecutable(newBinary []byte) error {
	dir := filepath.Dir(u.executablePath)

	// Write to temporary file in the same directory (guarantees same filesystem for atomic rename)
	tmpFile, err := os.CreateTemp(dir, ".loy-update-*")
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied writing to %s; please run with appropriate administrative privileges (e.g. sudo): %w", dir, err)
		}
		return fmt.Errorf("creating temporary binary file: %w", err)
	}
	tmpName := tmpFile.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmpFile.Write(newBinary); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("writing temporary binary: %w", err)
	}
	if err := tmpFile.Chmod(0755); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("setting binary executable permissions: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temporary binary: %w", err)
	}

	// On Windows or active filesystems where rename over active binary might fail,
	// rename existing to .old first then rename tmp to target
	oldBackup := u.executablePath + ".old"
	_ = os.Remove(oldBackup)
	if err := os.Rename(u.executablePath, oldBackup); err == nil {
		defer func() { _ = os.Remove(oldBackup) }()
	}

	if err := os.Rename(tmpName, u.executablePath); err != nil {
		// Attempt to restore backup if rename failed
		if _, statErr := os.Stat(oldBackup); statErr == nil {
			_ = os.Rename(oldBackup, u.executablePath)
		}
		return fmt.Errorf("atomic rename failed: %w", err)
	}

	return nil
}
