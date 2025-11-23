// Package absfs provides a minimal filesystem abstraction interface
// This is a POC implementation - in production, this would come from github.com/absfs/absfs
package absfs

import (
	"io"
	"io/fs"
	"time"
)

// FileSystem is the core abstraction for all filesystem implementations
type FileSystem interface {
	// Open opens a file for reading
	Open(name string) (File, error)

	// Create creates a file for writing
	Create(name string) (File, error)

	// Stat returns file info
	Stat(name string) (fs.FileInfo, error)

	// ReadDir reads directory contents
	ReadDir(name string) ([]fs.DirEntry, error)

	// Remove removes a file or empty directory
	Remove(name string) error

	// Mkdir creates a directory
	Mkdir(name string, perm fs.FileMode) error
}

// File represents an open file
type File interface {
	io.Reader
	io.Writer
	io.Closer
	io.Seeker

	Stat() (fs.FileInfo, error)
}

// FileInfo minimal implementation
type FileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (fi FileInfo) Name() string       { return fi.name }
func (fi FileInfo) Size() int64        { return fi.size }
func (fi FileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi FileInfo) ModTime() time.Time { return fi.modTime }
func (fi FileInfo) IsDir() bool        { return fi.isDir }
func (fi FileInfo) Sys() interface{}   { return nil }

// NewFileInfo creates a new FileInfo
func NewFileInfo(name string, size int64, mode fs.FileMode, modTime time.Time, isDir bool) fs.FileInfo {
	return FileInfo{
		name:    name,
		size:    size,
		mode:    mode,
		modTime: modTime,
		isDir:   isDir,
	}
}
