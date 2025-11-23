package engine

import (
	"fmt"

	"github.com/absfs/fscomposer/absfs"
	"github.com/absfs/fscomposer/registry"
)

// Builder constructs filesystem stacks from composition specs
type Builder struct {
	spec     *CompositionSpec
	registry *registry.Registry
	built    map[string]absfs.FileSystem // Cache of built nodes
}

// NewBuilder creates a new builder for the given spec
func NewBuilder(spec *CompositionSpec) *Builder {
	return &Builder{
		spec:     spec,
		registry: registry.DefaultRegistry,
		built:    make(map[string]absfs.FileSystem),
	}
}

// Build constructs the complete filesystem stack
// Returns the root filesystem (the one specified in mount.root)
func (b *Builder) Build() (absfs.FileSystem, error) {
	// First, validate the spec
	validator := NewValidator(b.spec)
	if err := validator.ValidateAll(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get the root node ID from mount config
	rootNodeID := b.spec.Mount.Root
	if rootNodeID == "" {
		return nil, fmt.Errorf("mount root is empty")
	}

	// Build the filesystem for the root node
	fs, err := b.buildNode(rootNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to build root node %s: %w", rootNodeID, err)
	}

	return fs, nil
}

// buildNode recursively builds a node and all its dependencies
func (b *Builder) buildNode(nodeID string) (absfs.FileSystem, error) {
	// Check if already built
	if fs, ok := b.built[nodeID]; ok {
		return fs, nil
	}

	// Get node definition
	node := b.spec.GetNode(nodeID)
	if node == nil {
		return nil, fmt.Errorf("node %s not found in spec", nodeID)
	}

	// Get constructor from registry
	constructor, err := b.registry.Get(node.Type)
	if err != nil {
		return nil, fmt.Errorf("node %s: %w", nodeID, err)
	}

	// Determine the underlying filesystem
	var underlying absfs.FileSystem

	// Special handling for multiplexer nodes (switchfs, unionfs)
	// They don't use the connection graph, but reference backends in config
	if IsMultiplexerNode(node.Type) {
		// For now, we'll handle this in the constructor
		// The constructor will build referenced nodes via b.buildNode
		underlying = nil // Multiplexers handle their own dependencies
	} else {
		// For regular wrappers, get the underlying filesystem from incoming connection
		incoming := b.spec.GetIncomingConnections(nodeID)

		if len(incoming) == 0 {
			// Backend node - no underlying filesystem
			if !IsBackendNode(node.Type) {
				return nil, fmt.Errorf("wrapper node %s has no incoming connection", nodeID)
			}
			underlying = nil
		} else if len(incoming) == 1 {
			// Single underlying filesystem (typical wrapper)
			underlyingNodeID := incoming[0].From
			var err error
			underlying, err = b.buildNode(underlyingNodeID)
			if err != nil {
				return nil, fmt.Errorf("failed to build underlying node %s for %s: %w",
					underlyingNodeID, nodeID, err)
			}
		} else {
			return nil, fmt.Errorf("node %s has multiple incoming connections (should have been caught by validator)", nodeID)
		}
	}

	// Construct the filesystem
	fs, err := constructor(node.Config, underlying)
	if err != nil {
		return nil, fmt.Errorf("failed to construct node %s (%s): %w", nodeID, node.Type, err)
	}

	// Cache the built filesystem
	b.built[nodeID] = fs

	return fs, nil
}

// GetBuiltNode returns a previously built node by ID
func (b *Builder) GetBuiltNode(nodeID string) (absfs.FileSystem, bool) {
	fs, ok := b.built[nodeID]
	return fs, ok
}

// BuildAll builds all nodes in the spec (useful for validation)
func (b *Builder) BuildAll() error {
	for _, node := range b.spec.Nodes {
		if _, err := b.buildNode(node.ID); err != nil {
			return err
		}
	}
	return nil
}
