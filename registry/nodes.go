package registry

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/absfs/fscomposer/absfs"
)

func init() {
	// Register built-in node types
	registerOSFS()
	registerMemFS()
	registerCacheFS()
	registerEncryptFS()
	registerRetryFS()
	registerMetricsFS()
}

// ============================================================================
// OSFS - Operating System Filesystem
// ============================================================================

type osFS struct {
	root string
}

func registerOSFS() {
	Register("osfs", newOSFS, NodeSchema{
		Type:        "osfs",
		Description: "Local operating system filesystem",
		Fields: []SchemaField{
			{
				Name:        "root",
				Type:        "string",
				Required:    true,
				Description: "Root directory path",
			},
		},
	})
}

func newOSFS(config map[string]interface{}, _ absfs.FileSystem) (absfs.FileSystem, error) {
	root, ok := config["root"].(string)
	if !ok {
		return nil, fmt.Errorf("osfs requires 'root' config (string path)")
	}

	// Ensure directory exists
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, fmt.Errorf("failed to create root directory: %w", err)
	}

	return &osFS{root: root}, nil
}

func (fs *osFS) Open(name string) (absfs.File, error) {
	return os.Open(filepath.Join(fs.root, name))
}

func (fs *osFS) Create(name string) (absfs.File, error) {
	path := filepath.Join(fs.root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	return os.Create(path)
}

func (fs *osFS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(filepath.Join(fs.root, name))
}

func (fs *osFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(filepath.Join(fs.root, name))
}

func (fs *osFS) Remove(name string) error {
	return os.Remove(filepath.Join(fs.root, name))
}

func (fs *osFS) Mkdir(name string, perm fs.FileMode) error {
	return os.Mkdir(filepath.Join(fs.root, name), perm)
}

// ============================================================================
// MemFS - In-Memory Filesystem (POC implementation)
// ============================================================================

type memFS struct {
	mu    sync.RWMutex
	files map[string]*memFile
}

type memFile struct {
	name    string
	data    *bytes.Buffer
	modTime time.Time
	mode    fs.FileMode
}

func (mf *memFile) Read(p []byte) (int, error) {
	return mf.data.Read(p)
}

func (mf *memFile) Write(p []byte) (int, error) {
	mf.modTime = time.Now()
	return mf.data.Write(p)
}

func (mf *memFile) Close() error {
	return nil
}

func (mf *memFile) Seek(offset int64, whence int) (int64, error) {
	// Simple seek implementation
	return 0, nil
}

func (mf *memFile) Stat() (fs.FileInfo, error) {
	return absfs.NewFileInfo(mf.name, int64(mf.data.Len()), mf.mode, mf.modTime, false), nil
}

func registerMemFS() {
	Register("memfs", newMemFS, NodeSchema{
		Type:        "memfs",
		Description: "In-memory filesystem",
		Fields:      []SchemaField{},
	})
}

func newMemFS(config map[string]interface{}, _ absfs.FileSystem) (absfs.FileSystem, error) {
	return &memFS{
		files: make(map[string]*memFile),
	}, nil
}

func (fs *memFS) Open(name string) (absfs.File, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	f, ok := fs.files[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return f, nil
}

func (fs *memFS) Create(name string) (absfs.File, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	f := &memFile{
		name:    name,
		data:    new(bytes.Buffer),
		modTime: time.Now(),
		mode:    0644,
	}
	fs.files[name] = f
	return f, nil
}

func (fs *memFS) Stat(name string) (fs.FileInfo, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	f, ok := fs.files[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return absfs.NewFileInfo(f.name, int64(f.data.Len()), f.mode, f.modTime, false), nil
}

func (fs *memFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return nil, fmt.Errorf("memfs ReadDir not implemented in POC")
}

func (fs *memFS) Remove(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	delete(fs.files, name)
	return nil
}

func (fs *memFS) Mkdir(name string, perm fs.FileMode) error {
	return nil // No-op for POC
}

// ============================================================================
// CacheFS - Caching Wrapper
// ============================================================================

type cacheFS struct {
	underlying absfs.FileSystem
	cache      map[string][]byte
	mu         sync.RWMutex
	maxSize    int
}

func registerCacheFS() {
	Register("cachefs", newCacheFS, NodeSchema{
		Type:        "cachefs",
		Description: "Caching filesystem wrapper",
		Fields: []SchemaField{
			{
				Name:        "size",
				Type:        "int",
				Required:    false,
				Default:     1073741824, // 1GB
				Description: "Maximum cache size in bytes",
			},
			{
				Name:        "policy",
				Type:        "select",
				Required:    false,
				Default:     "LRU",
				Options:     []string{"LRU", "LFU", "ARC"},
				Description: "Cache eviction policy",
			},
		},
	})
}

func newCacheFS(config map[string]interface{}, underlying absfs.FileSystem) (absfs.FileSystem, error) {
	if underlying == nil {
		return nil, fmt.Errorf("cachefs requires an underlying filesystem")
	}

	size := 1073741824 // Default 1GB
	if s, ok := config["size"]; ok {
		// Handle both int and float64 (YAML unmarshals numbers as float64)
		switch v := s.(type) {
		case int:
			size = v
		case float64:
			size = int(v)
		}
	}

	return &cacheFS{
		underlying: underlying,
		cache:      make(map[string][]byte),
		maxSize:    size,
	}, nil
}

func (fs *cacheFS) Open(name string) (absfs.File, error) {
	// Check cache first
	fs.mu.RLock()
	data, cached := fs.cache[name]
	fs.mu.RUnlock()

	if cached {
		// Return cached data
		return &memFile{
			name:    name,
			data:    bytes.NewBuffer(data),
			modTime: time.Now(),
			mode:    0644,
		}, nil
	}

	// Not cached, read from underlying
	f, err := fs.underlying.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err = io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// Cache the data
	fs.mu.Lock()
	fs.cache[name] = data
	fs.mu.Unlock()

	return &memFile{
		name:    name,
		data:    bytes.NewBuffer(data),
		modTime: time.Now(),
		mode:    0644,
	}, nil
}

func (fs *cacheFS) Create(name string) (absfs.File, error) {
	// Invalidate cache
	fs.mu.Lock()
	delete(fs.cache, name)
	fs.mu.Unlock()

	return fs.underlying.Create(name)
}

func (fs *cacheFS) Stat(name string) (fs.FileInfo, error) {
	return fs.underlying.Stat(name)
}

func (fs *cacheFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return fs.underlying.ReadDir(name)
}

func (fs *cacheFS) Remove(name string) error {
	fs.mu.Lock()
	delete(fs.cache, name)
	fs.mu.Unlock()

	return fs.underlying.Remove(name)
}

func (fs *cacheFS) Mkdir(name string, perm fs.FileMode) error {
	return fs.underlying.Mkdir(name, perm)
}

// ============================================================================
// EncryptFS - Encryption Wrapper (Stub for POC)
// ============================================================================

type encryptFS struct {
	underlying absfs.FileSystem
}

func registerEncryptFS() {
	Register("encryptfs", newEncryptFS, NodeSchema{
		Type:        "encryptfs",
		Description: "Encryption filesystem wrapper",
		Fields: []SchemaField{
			{
				Name:     "algorithm",
				Type:     "select",
				Required: true,
				Options:  []string{"AES-256-GCM", "ChaCha20-Poly1305"},
			},
			{
				Name:     "keySource",
				Type:     "select",
				Required: true,
				Options:  []string{"env", "file"},
			},
		},
	})
}

func newEncryptFS(config map[string]interface{}, underlying absfs.FileSystem) (absfs.FileSystem, error) {
	if underlying == nil {
		return nil, fmt.Errorf("encryptfs requires an underlying filesystem")
	}
	// POC: Pass-through implementation (no actual encryption)
	return &encryptFS{underlying: underlying}, nil
}

func (fs *encryptFS) Open(name string) (absfs.File, error)              { return fs.underlying.Open(name) }
func (fs *encryptFS) Create(name string) (absfs.File, error)            { return fs.underlying.Create(name) }
func (fs *encryptFS) Stat(name string) (fs.FileInfo, error)             { return fs.underlying.Stat(name) }
func (fs *encryptFS) ReadDir(name string) ([]fs.DirEntry, error)        { return fs.underlying.ReadDir(name) }
func (fs *encryptFS) Remove(name string) error                          { return fs.underlying.Remove(name) }
func (fs *encryptFS) Mkdir(name string, perm fs.FileMode) error         { return fs.underlying.Mkdir(name, perm) }

// ============================================================================
// RetryFS - Retry Wrapper (Stub for POC)
// ============================================================================

type retryFS struct {
	underlying absfs.FileSystem
}

func registerRetryFS() {
	Register("retryfs", newRetryFS, NodeSchema{
		Type:        "retryfs",
		Description: "Retry wrapper for resilience",
		Fields:      []SchemaField{},
	})
}

func newRetryFS(config map[string]interface{}, underlying absfs.FileSystem) (absfs.FileSystem, error) {
	if underlying == nil {
		return nil, fmt.Errorf("retryfs requires an underlying filesystem")
	}
	return &retryFS{underlying: underlying}, nil
}

func (fs *retryFS) Open(name string) (absfs.File, error)              { return fs.underlying.Open(name) }
func (fs *retryFS) Create(name string) (absfs.File, error)            { return fs.underlying.Create(name) }
func (fs *retryFS) Stat(name string) (fs.FileInfo, error)             { return fs.underlying.Stat(name) }
func (fs *retryFS) ReadDir(name string) ([]fs.DirEntry, error)        { return fs.underlying.ReadDir(name) }
func (fs *retryFS) Remove(name string) error                          { return fs.underlying.Remove(name) }
func (fs *retryFS) Mkdir(name string, perm fs.FileMode) error         { return fs.underlying.Mkdir(name, perm) }

// ============================================================================
// MetricsFS - Metrics Wrapper (Stub for POC)
// ============================================================================

type metricsFS struct {
	underlying absfs.FileSystem
}

func registerMetricsFS() {
	Register("metricsfs", newMetricsFS, NodeSchema{
		Type:        "metricsfs",
		Description: "Metrics collection wrapper",
		Fields:      []SchemaField{},
	})
}

func newMetricsFS(config map[string]interface{}, underlying absfs.FileSystem) (absfs.FileSystem, error) {
	if underlying == nil {
		return nil, fmt.Errorf("metricsfs requires an underlying filesystem")
	}
	return &metricsFS{underlying: underlying}, nil
}

func (fs *metricsFS) Open(name string) (absfs.File, error)              { return fs.underlying.Open(name) }
func (fs *metricsFS) Create(name string) (absfs.File, error)            { return fs.underlying.Create(name) }
func (fs *metricsFS) Stat(name string) (fs.FileInfo, error)             { return fs.underlying.Stat(name) }
func (fs *metricsFS) ReadDir(name string) ([]fs.DirEntry, error)        { return fs.underlying.ReadDir(name) }
func (fs *metricsFS) Remove(name string) error                          { return fs.underlying.Remove(name) }
func (fs *metricsFS) Mkdir(name string, perm fs.FileMode) error         { return fs.underlying.Mkdir(name, perm) }
