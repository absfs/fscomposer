// Package registry manages filesystem node type registration and construction
package registry

import (
	"fmt"

	"github.com/absfs/fscomposer/absfs"
)

// NodeConstructor is a function that creates a filesystem instance from configuration
// For wrapper nodes, the underlying filesystem is provided
// For backend nodes, underlying is nil
type NodeConstructor func(config map[string]interface{}, underlying absfs.FileSystem) (absfs.FileSystem, error)

// Registry holds all registered node types
type Registry struct {
	constructors map[string]NodeConstructor
	schemas      map[string]NodeSchema
}

// NodeSchema describes the configuration schema for a node type
type NodeSchema struct {
	Type        string
	Description string
	Fields      []SchemaField
}

// SchemaField describes a configuration field
type SchemaField struct {
	Name        string
	Type        string // "string", "int", "bool", "select"
	Required    bool
	Default     interface{}
	Description string
	Options     []string // For "select" type
}

// New creates a new empty registry
func New() *Registry {
	return &Registry{
		constructors: make(map[string]NodeConstructor),
		schemas:      make(map[string]NodeSchema),
	}
}

// Register adds a node type to the registry
func (r *Registry) Register(nodeType string, constructor NodeConstructor, schema NodeSchema) {
	r.constructors[nodeType] = constructor
	r.schemas[nodeType] = schema
}

// Get returns the constructor for a node type
func (r *Registry) Get(nodeType string) (NodeConstructor, error) {
	constructor, ok := r.constructors[nodeType]
	if !ok {
		return nil, fmt.Errorf("unknown node type: %s", nodeType)
	}
	return constructor, nil
}

// GetSchema returns the schema for a node type
func (r *Registry) GetSchema(nodeType string) (NodeSchema, error) {
	schema, ok := r.schemas[nodeType]
	if !ok {
		return NodeSchema{}, fmt.Errorf("unknown node type: %s", nodeType)
	}
	return schema, nil
}

// IsRegistered returns true if the node type is registered
func (r *Registry) IsRegistered(nodeType string) bool {
	_, ok := r.constructors[nodeType]
	return ok
}

// ListTypes returns all registered node types
func (r *Registry) ListTypes() []string {
	var types []string
	for t := range r.constructors {
		types = append(types, t)
	}
	return types
}

// DefaultRegistry is the global registry instance
var DefaultRegistry = New()

// Register adds a node type to the default registry
func Register(nodeType string, constructor NodeConstructor, schema NodeSchema) {
	DefaultRegistry.Register(nodeType, constructor, schema)
}

// Get returns the constructor from the default registry
func Get(nodeType string) (NodeConstructor, error) {
	return DefaultRegistry.Get(nodeType)
}

// GetSchema returns the schema from the default registry
func GetSchema(nodeType string) (NodeSchema, error) {
	return DefaultRegistry.GetSchema(nodeType)
}

// IsRegistered checks the default registry
func IsRegistered(nodeType string) bool {
	return DefaultRegistry.IsRegistered(nodeType)
}

// ListTypes returns types from the default registry
func ListTypes() []string {
	return DefaultRegistry.ListTypes()
}
