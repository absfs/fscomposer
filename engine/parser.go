package engine

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseFile parses a composition spec from a YAML file
func ParseFile(filename string) (*CompositionSpec, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return Parse(data)
}

// Parse parses a composition spec from YAML bytes
func Parse(data []byte) (*CompositionSpec, error) {
	var spec CompositionSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &spec, nil
}

// Validate performs basic validation on the spec
func (spec *CompositionSpec) Validate() error {
	if spec.Version == "" {
		return fmt.Errorf("version is required")
	}

	if spec.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(spec.Nodes) == 0 {
		return fmt.Errorf("at least one node is required")
	}

	// Validate nodes have unique IDs
	nodeIDs := make(map[string]bool)
	for _, node := range spec.Nodes {
		if node.ID == "" {
			return fmt.Errorf("node ID is required")
		}
		if nodeIDs[node.ID] {
			return fmt.Errorf("duplicate node ID: %s", node.ID)
		}
		nodeIDs[node.ID] = true

		if node.Type == "" {
			return fmt.Errorf("node %s: type is required", node.ID)
		}

		if !IsBackendNode(node.Type) && !IsWrapperNode(node.Type) {
			return fmt.Errorf("node %s: unknown type %s", node.ID, node.Type)
		}
	}

	// Validate connections reference existing nodes
	for i, conn := range spec.Connections {
		if !nodeIDs[conn.From] {
			return fmt.Errorf("connection %d: 'from' node %s not found", i, conn.From)
		}
		if !nodeIDs[conn.To] {
			return fmt.Errorf("connection %d: 'to' node %s not found", i, conn.To)
		}
	}

	// Validate mount references existing node
	if spec.Mount.Root == "" {
		return fmt.Errorf("mount root is required")
	}
	if !nodeIDs[spec.Mount.Root] {
		return fmt.Errorf("mount root %s not found", spec.Mount.Root)
	}

	if spec.Mount.Type == "" {
		return fmt.Errorf("mount type is required")
	}

	return nil
}

// GetNode returns the node with the given ID, or nil if not found
func (spec *CompositionSpec) GetNode(id string) *Node {
	for i := range spec.Nodes {
		if spec.Nodes[i].ID == id {
			return &spec.Nodes[i]
		}
	}
	return nil
}

// GetIncomingConnections returns all connections where the given node is the target (to)
func (spec *CompositionSpec) GetIncomingConnections(nodeID string) []Connection {
	var incoming []Connection
	for _, conn := range spec.Connections {
		if conn.To == nodeID {
			incoming = append(incoming, conn)
		}
	}
	return incoming
}

// GetOutgoingConnections returns all connections where the given node is the source (from)
func (spec *CompositionSpec) GetOutgoingConnections(nodeID string) []Connection {
	var outgoing []Connection
	for _, conn := range spec.Connections {
		if conn.From == nodeID {
			outgoing = append(outgoing, conn)
		}
	}
	return outgoing
}
