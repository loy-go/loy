package filesystem

import (
	"io/fs"
	"os"
)

// FileSystem represents a unified abstraction for file operations.
type FileSystem interface {
	// ReadFile reads the named file and returns the contents.
	ReadFile(name string) ([]byte, error)

	// WriteFile writes data to the named file, creating it if necessary.
	WriteFile(name string, data []byte, perm os.FileMode) error

	// Exists reports whether the named file or directory exists.
	Exists(name string) (bool, error)

	// MkdirAll creates a directory named path, along with any necessary parents.
	MkdirAll(path string, perm os.FileMode) error

	// Remove removes the named file or directory.
	Remove(name string) error

	// RemoveAll removes path and any children it contains.
	RemoveAll(path string) error

	// Stat returns a FileInfo describing the named file.
	Stat(name string) (os.FileInfo, error)

	// Walk walks the file tree rooted at root, calling fn for each file or directory.
	Walk(root string, fn fs.WalkDirFunc) error
}
