package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type mockHTTPClient struct {
	responses map[string]*http.Response
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	url := req.URL.String()
	if resp, ok := m.responses[url]; ok {
		return resp, nil
	}
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader("not found")),
	}, nil
}

func createTestTarGz(t *testing.T, binaryName, content string) []byte {
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)

	hdr := &tar.Header{
		Name: binaryName,
		Mode: 0755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestUpdater_Check(t *testing.T) {
	releaseJSON := fmt.Sprintf(`{
  "tag_name": "v1.5.0",
  "name": "v1.5.0",
  "published_at": "2026-09-15T00:00:00Z",
  "assets": [
    {
      "name": "checksums.txt",
      "browser_download_url": "https://example.com/checksums.txt"
    },
    {
      "name": "loy_1.5.0_%s_%s.tar.gz",
      "browser_download_url": "https://example.com/loy.tar.gz"
    }
  ]
}`, runtime.GOOS, runtime.GOARCH)

	client := &mockHTTPClient{
		responses: map[string]*http.Response{
			"https://api.github.com/repos/loy-go/loy/releases/latest": {
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(releaseJSON)),
			},
		},
	}

	u, err := NewUpdater("loy-go/loy", client)
	if err != nil {
		t.Fatal(err)
	}
	u.currentVersion = "v1.0.0"

	ctx := context.Background()
	res, info, err := u.Check(ctx, "")
	if err != nil {
		t.Fatalf("check failed: %v", err)
	}

	if !res.UpdateAvailable {
		t.Errorf("expected update to be available: v1.0.0 -> v1.5.0")
	}
	if res.TargetVersion != "1.5.0" {
		t.Errorf("expected target version 1.5.0, got: %s", res.TargetVersion)
	}
	if info.ArchiveName != fmt.Sprintf("loy_1.5.0_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH) {
		t.Errorf("unexpected archive name: %s", info.ArchiveName)
	}
}

func TestUpdater_Apply_Success(t *testing.T) {
	tempDir := t.TempDir()
	binName := "loy"
	if runtime.GOOS == "windows" {
		binName = "loy.exe"
	}
	fakeExecPath := filepath.Join(tempDir, binName)
	if err := os.WriteFile(fakeExecPath, []byte("old-binary-v1.0.0"), 0755); err != nil {
		t.Fatal(err)
	}

	archiveName := fmt.Sprintf("loy_1.5.0_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	archiveBytes := createTestTarGz(t, binName, "new-binary-v1.5.0")

	hasher := sha256.New()
	hasher.Write(archiveBytes)
	checksumHex := hex.EncodeToString(hasher.Sum(nil))

	checksumsContent := fmt.Sprintf("%s  %s\n", checksumHex, archiveName)

	releaseJSON := fmt.Sprintf(`{
  "tag_name": "v1.5.0",
  "assets": [
    {
      "name": "checksums.txt",
      "browser_download_url": "https://example.com/checksums.txt"
    },
    {
      "name": "%s",
      "browser_download_url": "https://example.com/%s"
    }
  ]
}`, archiveName, archiveName)

	client := &mockHTTPClient{
		responses: map[string]*http.Response{
			"https://api.github.com/repos/loy-go/loy/releases/latest": {
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(releaseJSON)),
			},
			"https://example.com/checksums.txt": {
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(checksumsContent)),
			},
			"https://example.com/" + archiveName: {
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(archiveBytes)),
			},
		},
	}

	u, err := NewUpdater("loy-go/loy", client)
	if err != nil {
		t.Fatal(err)
	}
	u.executablePath = fakeExecPath
	u.currentVersion = "v1.0.0"

	ctx := context.Background()
	applyRes, err := u.Apply(ctx, "")
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	if applyRes.UpdatedVersion != "1.5.0" {
		t.Errorf("expected updated version 1.5.0, got: %s", applyRes.UpdatedVersion)
	}

	// Verify updated binary on disk
	updatedBytes, err := os.ReadFile(fakeExecPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(updatedBytes) != "new-binary-v1.5.0" {
		t.Errorf("expected 'new-binary-v1.5.0', got: %s", string(updatedBytes))
	}
}

func TestUpdater_Apply_ChecksumMismatchAbort(t *testing.T) {
	tempDir := t.TempDir()
	binName := "loy"
	if runtime.GOOS == "windows" {
		binName = "loy.exe"
	}
	fakeExecPath := filepath.Join(tempDir, binName)
	_ = os.WriteFile(fakeExecPath, []byte("old-binary"), 0755)

	archiveName := fmt.Sprintf("loy_1.5.0_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	archiveBytes := createTestTarGz(t, binName, "tampered-content")

	// Provide incorrect checksum in checksums.txt
	checksumsContent := fmt.Sprintf("0000000000000000000000000000000000000000000000000000000000000000  %s\n", archiveName)

	releaseJSON := fmt.Sprintf(`{
  "tag_name": "v1.5.0",
  "assets": [
    {
      "name": "checksums.txt",
      "browser_download_url": "https://example.com/checksums.txt"
    },
    {
      "name": "%s",
      "browser_download_url": "https://example.com/%s"
    }
  ]
}`, archiveName, archiveName)

	client := &mockHTTPClient{
		responses: map[string]*http.Response{
			"https://api.github.com/repos/loy-go/loy/releases/latest": {
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(releaseJSON)),
			},
			"https://example.com/checksums.txt": {
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(checksumsContent)),
			},
			"https://example.com/" + archiveName: {
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(archiveBytes)),
			},
		},
	}

	u, _ := NewUpdater("loy-go/loy", client)
	u.executablePath = fakeExecPath

	_, err := u.Apply(context.Background(), "")
	if err == nil {
		t.Fatalf("expected checksum mismatch error")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Errorf("expected checksum mismatch message, got: %v", err)
	}

	// Verify original binary was preserved
	content, _ := os.ReadFile(fakeExecPath)
	if string(content) != "old-binary" {
		t.Errorf("original binary should not be overwritten on checksum failure")
	}
}

func TestUpdater_DetectPackageManager(t *testing.T) {
	u := &Updater{executablePath: "/opt/homebrew/bin/loy"}
	isManaged, name := u.DetectPackageManager()
	if !isManaged || name != "homebrew" {
		t.Errorf("expected homebrew detection for /opt/homebrew/bin/loy, got: %v, %s", isManaged, name)
	}

	u = &Updater{executablePath: "/nix/store/abc-loy/bin/loy"}
	isManaged, name = u.DetectPackageManager()
	if !isManaged || name != "nix" {
		t.Errorf("expected nix detection, got: %v, %s", isManaged, name)
	}

	u = &Updater{executablePath: "/home/user/bin/loy"}
	isManaged, _ = u.DetectPackageManager()
	if isManaged {
		t.Errorf("expected not managed for /home/user/bin/loy")
	}
}
