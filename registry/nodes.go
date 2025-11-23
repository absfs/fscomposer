package registry

import (
	"fmt"
	"time"

	"github.com/absfs/absfs"
	"github.com/absfs/cachefs"
	"github.com/absfs/encryptfs"
	"github.com/absfs/memfs"
	"github.com/absfs/metricsfs"
	"github.com/absfs/osfs"
)

func init() {
	// Register built-in node types
	registerOSFS()
	registerMemFS()
	registerCacheFS()
	registerEncryptFS()
	registerMetricsFS()
}

// ============================================================================
// OSFS - Operating System Filesystem
// ============================================================================

func registerOSFS() {
	Register("osfs", newOSFS, NodeSchema{
		Type:        "osfs",
		Description: "Local operating system filesystem",
		Fields: []SchemaField{
			{
				Name:        "root",
				Type:        "string",
				Required:    false,
				Default:     ".",
				Description: "Root directory path (default: current directory)",
			},
		},
	})
}

func newOSFS(config map[string]interface{}, _ absfs.FileSystem) (absfs.FileSystem, error) {
	fs, err := osfs.NewFS()
	if err != nil {
		return nil, fmt.Errorf("failed to create osfs: %w", err)
	}

	// Handle root directory if specified
	if root, ok := config["root"].(string); ok && root != "" && root != "." {
		if err := fs.Chdir(root); err != nil {
			return nil, fmt.Errorf("failed to change to root directory %s: %w", root, err)
		}
	}

	return fs, nil
}

// ============================================================================
// MemFS - In-Memory Filesystem
// ============================================================================

func registerMemFS() {
	Register("memfs", newMemFS, NodeSchema{
		Type:        "memfs",
		Description: "In-memory filesystem",
		Fields:      []SchemaField{},
	})
}

func newMemFS(config map[string]interface{}, _ absfs.FileSystem) (absfs.FileSystem, error) {
	fs, err := memfs.NewFS()
	if err != nil {
		return nil, fmt.Errorf("failed to create memfs: %w", err)
	}
	return fs, nil
}

// ============================================================================
// CacheFS - Caching Wrapper
// ============================================================================

func registerCacheFS() {
	Register("cachefs", newCacheFS, NodeSchema{
		Type:        "cachefs",
		Description: "Caching filesystem wrapper",
		Fields: []SchemaField{
			{
				Name:        "maxBytes",
				Type:        "int",
				Required:    false,
				Default:     1073741824, // 1GB
				Description: "Maximum cache size in bytes",
			},
			{
				Name:        "maxEntries",
				Type:        "int",
				Required:    false,
				Description: "Maximum number of cached entries",
			},
			{
				Name:        "policy",
				Type:        "select",
				Required:    false,
				Default:     "LRU",
				Options:     []string{"LRU", "LFU"},
				Description: "Cache eviction policy",
			},
			{
				Name:        "ttl",
				Type:        "int",
				Required:    false,
				Description: "Time-to-live for cache entries in seconds",
			},
			{
				Name:        "metadataCache",
				Type:        "bool",
				Required:    false,
				Default:     true,
				Description: "Enable metadata caching",
			},
		},
	})
}

func newCacheFS(config map[string]interface{}, underlying absfs.FileSystem) (absfs.FileSystem, error) {
	if underlying == nil {
		return nil, fmt.Errorf("cachefs requires an underlying filesystem")
	}

	var opts []cachefs.Option

	// Handle maxBytes
	if mb, ok := config["maxBytes"]; ok {
		var maxBytes uint64
		switch v := mb.(type) {
		case int:
			maxBytes = uint64(v)
		case float64:
			maxBytes = uint64(v)
		}
		opts = append(opts, cachefs.WithMaxBytes(maxBytes))
	}

	// Handle maxEntries
	if me, ok := config["maxEntries"]; ok {
		var maxEntries uint64
		switch v := me.(type) {
		case int:
			maxEntries = uint64(v)
		case float64:
			maxEntries = uint64(v)
		}
		opts = append(opts, cachefs.WithMaxEntries(maxEntries))
	}

	// Handle policy
	if policy, ok := config["policy"].(string); ok {
		switch policy {
		case "LRU":
			opts = append(opts, cachefs.WithEvictionPolicy(cachefs.EvictionLRU))
		case "LFU":
			opts = append(opts, cachefs.WithEvictionPolicy(cachefs.EvictionLFU))
		}
	}

	// Handle TTL
	if ttl, ok := config["ttl"]; ok {
		var ttlSeconds int
		switch v := ttl.(type) {
		case int:
			ttlSeconds = v
		case float64:
			ttlSeconds = int(v)
		}
		opts = append(opts, cachefs.WithTTL(time.Duration(ttlSeconds)*time.Second))
	}

	// Handle metadata cache
	if mc, ok := config["metadataCache"].(bool); ok {
		opts = append(opts, cachefs.WithMetadataCache(mc))
	}

	return cachefs.New(underlying, opts...), nil
}

// ============================================================================
// EncryptFS - Encryption Wrapper
// ============================================================================

func registerEncryptFS() {
	Register("encryptfs", newEncryptFS, NodeSchema{
		Type:        "encryptfs",
		Description: "Encryption filesystem wrapper",
		Fields: []SchemaField{
			{
				Name:        "cipher",
				Type:        "select",
				Required:    false,
				Default:     "AES-256-GCM",
				Options:     []string{"AES-256-GCM", "ChaCha20-Poly1305"},
				Description: "Encryption cipher suite",
			},
			{
				Name:        "password",
				Type:        "string",
				Required:    true,
				Description: "Encryption password (use env var for production)",
			},
			{
				Name:        "kdfMemory",
				Type:        "int",
				Required:    false,
				Default:     65536, // 64MB
				Description: "KDF memory in KB for Argon2id",
			},
			{
				Name:        "kdfIterations",
				Type:        "int",
				Required:    false,
				Default:     3,
				Description: "KDF iterations for Argon2id",
			},
		},
	})
}

func newEncryptFS(config map[string]interface{}, underlying absfs.FileSystem) (absfs.FileSystem, error) {
	if underlying == nil {
		return nil, fmt.Errorf("encryptfs requires an underlying filesystem")
	}

	password, ok := config["password"].(string)
	if !ok || password == "" {
		return nil, fmt.Errorf("encryptfs requires 'password' config")
	}

	// Parse cipher
	cipher := encryptfs.CipherAES256GCM
	if cipherStr, ok := config["cipher"].(string); ok {
		switch cipherStr {
		case "ChaCha20-Poly1305":
			cipher = encryptfs.CipherChaCha20Poly1305
		}
	}

	// Parse KDF params
	memory := 65536
	iterations := 3
	if m, ok := config["kdfMemory"]; ok {
		switch v := m.(type) {
		case int:
			memory = v
		case float64:
			memory = int(v)
		}
	}
	if i, ok := config["kdfIterations"]; ok {
		switch v := i.(type) {
		case int:
			iterations = v
		case float64:
			iterations = int(v)
		}
	}

	encConfig := &encryptfs.Config{
		Cipher: cipher,
		KeyProvider: encryptfs.NewPasswordKeyProvider(
			[]byte(password),
			encryptfs.Argon2idParams{
				Memory:      uint32(memory),
				Iterations:  uint32(iterations),
				Parallelism: 4,
			},
		),
	}

	return encryptfs.New(underlying, encConfig)
}

// ============================================================================
// MetricsFS - Metrics Wrapper
// ============================================================================

func registerMetricsFS() {
	Register("metricsfs", newMetricsFS, NodeSchema{
		Type:        "metricsfs",
		Description: "Metrics collection wrapper",
		Fields: []SchemaField{
			{
				Name:        "enablePrometheus",
				Type:        "bool",
				Required:    false,
				Default:     false,
				Description: "Enable Prometheus metrics",
			},
		},
	})
}

func newMetricsFS(config map[string]interface{}, underlying absfs.FileSystem) (absfs.FileSystem, error) {
	if underlying == nil {
		return nil, fmt.Errorf("metricsfs requires an underlying filesystem")
	}

	// Wrap the metricsfs with ExtendFiler to get full FileSystem interface
	mfs := metricsfs.New(underlying)
	return absfs.ExtendFiler(mfs), nil
}
