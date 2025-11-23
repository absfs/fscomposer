// Package engine provides the composition engine for building filesystem stacks
package engine

// CompositionSpec represents a complete filesystem composition specification
type CompositionSpec struct {
	Version     string       `yaml:"version" json:"version"`
	Name        string       `yaml:"name" json:"name"`
	Description string       `yaml:"description,omitempty" json:"description,omitempty"`
	Nodes       []Node       `yaml:"nodes" json:"nodes"`
	Connections []Connection `yaml:"connections" json:"connections"`
	Mount       MountConfig  `yaml:"mount" json:"mount"`
}

// Node represents a single filesystem node (backend or wrapper)
type Node struct {
	ID     string                 `yaml:"id" json:"id"`
	Type   string                 `yaml:"type" json:"type"`
	Config map[string]interface{} `yaml:"config,omitempty" json:"config,omitempty"`
}

// Connection represents a data flow connection between nodes
// Data flows FROM the underlying filesystem TO the wrapper
type Connection struct {
	From string `yaml:"from" json:"from"`
	To   string `yaml:"to" json:"to"`
}

// MountConfig specifies how the composed filesystem should be mounted
type MountConfig struct {
	Type    string                 `yaml:"type" json:"type"` // fuse, webdav, nfs, api
	Path    string                 `yaml:"path,omitempty" json:"path,omitempty"`
	Port    int                    `yaml:"port,omitempty" json:"port,omitempty"`
	Root    string                 `yaml:"root" json:"root"` // Node ID to mount
	Export  string                 `yaml:"export,omitempty" json:"export,omitempty"`
	Options map[string]interface{} `yaml:"options,omitempty" json:"options,omitempty"`
}

// NodeType constants for type checking
const (
	// Backend node types (data sources)
	NodeTypeOSFS    = "osfs"
	NodeTypeMemFS   = "memfs"
	NodeTypeS3FS    = "s3fs"
	NodeTypeSFTPFS  = "sftpfs"
	NodeTypeWebDAVFS = "webdavfs"
	NodeTypeBoltFS  = "boltfs"
	NodeTypeHTTPFS  = "httpfs"

	// Wrapper node types (middleware)
	NodeTypeCacheFS   = "cachefs"
	NodeTypeEncryptFS = "encryptfs"
	NodeTypeCompressFS = "compressfs"
	NodeTypeRetryFS   = "retryfs"
	NodeTypeMetricsFS = "metricsfs"
	NodeTypeUnionFS   = "unionfs"
	NodeTypePermFS    = "permfs"
	NodeTypeQuotaFS   = "quotafs"
	NodeTypeSwitchFS  = "switchfs"
	NodeTypeLogFS     = "logfs"
)

// IsBackendNode returns true if the node type is a backend (data source)
func IsBackendNode(nodeType string) bool {
	backends := []string{
		NodeTypeOSFS, NodeTypeMemFS, NodeTypeS3FS, NodeTypeSFTPFS,
		NodeTypeWebDAVFS, NodeTypeBoltFS, NodeTypeHTTPFS,
	}
	for _, b := range backends {
		if nodeType == b {
			return true
		}
	}
	return false
}

// IsWrapperNode returns true if the node type is a wrapper (middleware)
func IsWrapperNode(nodeType string) bool {
	wrappers := []string{
		NodeTypeCacheFS, NodeTypeEncryptFS, NodeTypeCompressFS, NodeTypeRetryFS,
		NodeTypeMetricsFS, NodeTypeUnionFS, NodeTypePermFS, NodeTypeQuotaFS,
		NodeTypeSwitchFS, NodeTypeLogFS,
	}
	for _, w := range wrappers {
		if nodeType == w {
			return true
		}
	}
	return false
}

// IsMultiplexerNode returns true if the node type accepts multiple inputs
// These nodes reference their backends via config, not connections
func IsMultiplexerNode(nodeType string) bool {
	return nodeType == NodeTypeSwitchFS || nodeType == NodeTypeUnionFS
}
