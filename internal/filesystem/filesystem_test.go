package filesystem_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/uloydev/loy/internal/filesystem"
)

func testFileSystemContract(t *testing.T, fsys filesystem.FileSystem, baseDir string) {
	testFile := filepath.Join(baseDir, "subdir", "hello.txt")

	exists, err := fsys.Exists(testFile)
	if err != nil {
		t.Fatalf("Exists returned unexpected error: %v", err)
	}
	if exists {
		t.Fatalf("expected file to not exist initially")
	}

	content := []byte("Hello, Loy Platform!")
	err = fsys.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	exists, err = fsys.Exists(testFile)
	if err != nil || !exists {
		t.Fatalf("expected file to exist after write: exists=%v, err=%v", exists, err)
	}

	readBack, err := fsys.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(readBack) != string(content) {
		t.Fatalf("expected content %q, got %q", string(content), string(readBack))
	}

	fi, err := fsys.Stat(testFile)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if fi.IsDir() {
		t.Fatalf("expected file to not be a directory")
	}
	if fi.Size() != int64(len(content)) {
		t.Fatalf("expected size %d, got %d", len(content), fi.Size())
	}

	err = fsys.Remove(testFile)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	exists, err = fsys.Exists(testFile)
	if err != nil || exists {
		t.Fatalf("expected file to be deleted: exists=%v", exists)
	}
}

func TestOSFileSystem(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "loy-fs-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	fsys := filesystem.NewOSFileSystem()
	testFileSystemContract(t, fsys, tempDir)
}

func TestMemFileSystem(t *testing.T) {
	fsys := filesystem.NewMemFileSystem()
	testFileSystemContract(t, fsys, "/testroot")
}

func TestOSFileSystem_WalkAndRemove(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "loy-fs-walk-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	fsys := filesystem.NewOSFileSystem()
	f1 := filepath.Join(tempDir, "a", "file1.txt")
	f2 := filepath.Join(tempDir, "b", "file2.txt")
	if err := fsys.WriteFile(f1, []byte("1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := fsys.WriteFile(f2, []byte("2"), 0644); err != nil {
		t.Fatal(err)
	}

	var count int
	err = fsys.Walk(tempDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected 2 files visited in walk, got %d", count)
	}

	if err := fsys.RemoveAll(filepath.Join(tempDir, "a")); err != nil {
		t.Fatal(err)
	}
	exists, err := fsys.Exists(f1)
	if err != nil || exists {
		t.Fatalf("expected file1 to be deleted, exists: %v", exists)
	}
}

func TestMemFileSystem_DirectoryCollision(t *testing.T) {
	fsys := filesystem.NewMemFileSystem()
	err := fsys.MkdirAll("/data/sub", 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = fsys.WriteFile("/data/sub", []byte("bad write"), 0644)
	if err == nil {
		t.Fatal("expected error when writing file to existing directory path, got nil")
	}
}

func TestMemFileSystem_WalkSkipDir(t *testing.T) {
	fsys := filesystem.NewMemFileSystem()
	_ = fsys.WriteFile("/root/a/file1.txt", []byte("1"), 0644)
	_ = fsys.WriteFile("/root/a/file2.txt", []byte("2"), 0644)
	_ = fsys.WriteFile("/root/b/file3.txt", []byte("3"), 0644)

	var visited []string
	err := fsys.Walk("/root", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		visited = append(visited, path)
		if path == "/root/a" {
			return os.ErrNotExist // Using fs.SkipDir equivalent
		}
		return nil
	})
	_ = err

	// Using fs.SkipDir explicitly
	visited = nil
	err = fsys.Walk("/root", func(path string, d os.DirEntry, err error) error {
		visited = append(visited, path)
		if path == "/root/a" {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range visited {
		if v == "/root/a/file1.txt" || v == "/root/a/file2.txt" {
			t.Fatalf("expected /root/a children to be skipped, visited: %s", v)
		}
	}
}

func TestCleanAndValidatePath(t *testing.T) {
	rootDir := "/workspace/app"

	tests := []struct {
		name        string
		target      string
		expectError bool
	}{
		{"safe relative file", "src/main.go", false},
		{"nested safe directory", "internal/platform/db.go", false},
		{"attempt traversal with dots", "../../etc/passwd", true},
		{"nested traversal attempt", "src/../../etc/shadow", true},
		{"absolute outside path", "/etc/passwd", true},
		{"exact root", ".", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := filesystem.CleanAndValidatePath(rootDir, tc.target)
			if tc.expectError && err == nil {
				t.Fatalf("expected path traversal error for %q, got nil", tc.target)
			}
			if !tc.expectError && err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.target, err)
			}
		})
	}
}
