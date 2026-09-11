package filesystem

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// OSFileSystem implements FileSystem backed by the local operating system.
type OSFileSystem struct{}

// NewOSFileSystem creates an OS-backed FileSystem.
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

func (fsys *OSFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (fsys *OSFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(name, data, perm)
}

func (fsys *OSFileSystem) Exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (fsys *OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (fsys *OSFileSystem) Remove(name string) error {
	return os.Remove(name)
}

func (fsys *OSFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (fsys *OSFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (fsys *OSFileSystem) Walk(root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
}
