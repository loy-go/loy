package filesystem

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type memFile struct {
	data    []byte
	perm    os.FileMode
	isDir   bool
	modTime time.Time
}

// MemFileSystem implements an in-memory FileSystem for fast, isolated tests.
type MemFileSystem struct {
	mu    sync.RWMutex
	files map[string]*memFile
}

// NewMemFileSystem initializes an empty in-memory filesystem.
func NewMemFileSystem() *MemFileSystem {
	m := &MemFileSystem{
		files: make(map[string]*memFile),
	}
	m.files["/"] = &memFile{isDir: true, perm: 0755, modTime: time.Now()}
	return m
}

func cleanPath(p string) string {
	p = filepath.ToSlash(p)
	if vol := filepath.VolumeName(p); vol != "" {
		p = strings.TrimPrefix(p, vol)
	}
	cleaned := path.Clean(p)
	if !strings.HasPrefix(cleaned, "/") {
		cleaned = "/" + cleaned
	}
	return path.Clean(cleaned)
}

func (m *MemFileSystem) ReadFile(name string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p := cleanPath(name)
	f, ok := m.files[p]
	if !ok || f.isDir {
		return nil, os.ErrNotExist
	}
	res := make([]byte, len(f.data))
	copy(res, f.data)
	return res, nil
}

func (m *MemFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p := cleanPath(name)
	if existing, ok := m.files[p]; ok && existing.isDir {
		return errors.New("cannot overwrite existing directory with file: " + p)
	}

	dir := path.Dir(p)

	// Ensure parent directories exist
	if err := m.mkdirAllLocked(dir, 0755); err != nil {
		return err
	}

	buf := make([]byte, len(data))
	copy(buf, data)
	m.files[p] = &memFile{
		data:    buf,
		perm:    perm,
		isDir:   false,
		modTime: time.Now(),
	}
	return nil
}

func (m *MemFileSystem) Exists(name string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p := cleanPath(name)
	_, ok := m.files[p]
	return ok, nil
}

func (m *MemFileSystem) MkdirAll(path string, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.mkdirAllLocked(cleanPath(path), perm)
}

func (m *MemFileSystem) mkdirAllLocked(p string, perm os.FileMode) error {
	if p == "/" || p == "." {
		return nil
	}
	parts := strings.Split(strings.Trim(p, "/"), "/")
	curr := ""
	for _, part := range parts {
		curr += "/" + part
		if f, exists := m.files[curr]; exists {
			if !f.isDir {
				return errors.New("path component is not a directory: " + curr)
			}
		} else {
			m.files[curr] = &memFile{
				perm:    perm,
				isDir:   true,
				modTime: time.Now(),
			}
		}
	}
	return nil
}

func (m *MemFileSystem) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p := cleanPath(name)
	if _, ok := m.files[p]; !ok {
		return os.ErrNotExist
	}
	delete(m.files, p)
	return nil
}

func (m *MemFileSystem) RemoveAll(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p := cleanPath(path)
	prefix := p
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	for k := range m.files {
		if k == p || strings.HasPrefix(k, prefix) {
			delete(m.files, k)
		}
	}
	return nil
}

type memFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (fi memFileInfo) Name() string       { return fi.name }
func (fi memFileInfo) Size() int64        { return fi.size }
func (fi memFileInfo) Mode() os.FileMode {
	if fi.isDir {
		return fi.mode | os.ModeDir
	}
	return fi.mode
}
func (fi memFileInfo) ModTime() time.Time { return fi.modTime }
func (fi memFileInfo) IsDir() bool        { return fi.isDir }
func (fi memFileInfo) Sys() any           { return nil }

func (m *MemFileSystem) Stat(name string) (os.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p := cleanPath(name)
	f, ok := m.files[p]
	if !ok {
		return nil, os.ErrNotExist
	}
	return memFileInfo{
		name:    path.Base(p),
		size:    int64(len(f.data)),
		mode:    f.perm,
		modTime: f.modTime,
		isDir:   f.isDir,
	}, nil
}

type memDirEntry struct {
	fi os.FileInfo
}

func (e memDirEntry) Name() string               { return e.fi.Name() }
func (e memDirEntry) IsDir() bool                { return e.fi.IsDir() }
func (e memDirEntry) Type() fs.FileMode          { return e.fi.Mode().Type() }
func (e memDirEntry) Info() (fs.FileInfo, error) { return e.fi, nil }

func (m *MemFileSystem) Walk(root string, fn fs.WalkDirFunc) error {
	m.mu.RLock()
	cleanRoot := cleanPath(root)
	prefix := cleanRoot
	if !strings.HasSuffix(prefix, "/") && prefix != "/" {
		prefix += "/"
	}

	var matchedPaths []string
	for p := range m.files {
		if p == cleanRoot || strings.HasPrefix(p, prefix) {
			matchedPaths = append(matchedPaths, p)
		}
	}
	m.mu.RUnlock()

	sort.Strings(matchedPaths)
	skipPrefix := ""
	vol := filepath.VolumeName(root)
	for _, p := range matchedPaths {
		if skipPrefix != "" && strings.HasPrefix(p, skipPrefix) {
			continue
		}
		skipPrefix = ""

		st, err := m.Stat(p)
		if err != nil {
			return err
		}
		callPath := p
		if vol != "" {
			callPath = filepath.FromSlash(vol + p)
		}
		if err := fn(callPath, memDirEntry{fi: st}, nil); err != nil {
			if errors.Is(err, fs.SkipDir) {
				if st.IsDir() {
					skipPrefix = p
					if !strings.HasSuffix(skipPrefix, "/") {
						skipPrefix += "/"
					}
				}
				continue
			}
			return err
		}
	}
	return nil
}
