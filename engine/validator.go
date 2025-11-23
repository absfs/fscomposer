package engine

import (
	"fmt"
)

// Validator performs advanced validation on composition specs
type Validator struct {
	spec *CompositionSpec
}

// NewValidator creates a new validator for the given spec
func NewValidator(spec *CompositionSpec) *Validator {
	return &Validator{spec: spec}
}

// ValidateAll performs all validation checks
func (v *Validator) ValidateAll() error {
	// Basic validation first
	if err := v.spec.Validate(); err != nil {
		return err
	}

	// Check for cycles in the connection graph
	if err := v.DetectCycles(); err != nil {
		return err
	}

	// Validate connection types
	if err := v.ValidateConnectionTypes(); err != nil {
		return err
	}

	// Validate node configurations
	if err := v.ValidateNodeConfigs(); err != nil {
		return err
	}

	return nil
}

// DetectCycles checks for cycles in the connection graph
func (v *Validator) DetectCycles() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for _, node := range v.spec.Nodes {
		if !visited[node.ID] {
			if v.hasCycle(node.ID, visited, recStack) {
				return fmt.Errorf("cycle detected in connection graph involving node %s", node.ID)
			}
		}
	}

	return nil
}

// hasCycle is a recursive helper for cycle detection using DFS
func (v *Validator) hasCycle(nodeID string, visited, recStack map[string]bool) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	// Get outgoing connections (from this node)
	for _, conn := range v.spec.GetOutgoingConnections(nodeID) {
		targetID := conn.To

		if !visited[targetID] {
			if v.hasCycle(targetID, visited, recStack) {
				return true
			}
		} else if recStack[targetID] {
			// Back edge found - cycle detected
			return true
		}
	}

	recStack[nodeID] = false
	return false
}

// ValidateConnectionTypes ensures connections are valid for node types
func (v *Validator) ValidateConnectionTypes() error {
	for i, conn := range v.spec.Connections {
		fromNode := v.spec.GetNode(conn.From)
		toNode := v.spec.GetNode(conn.To)

		if fromNode == nil || toNode == nil {
			continue // Already validated in basic check
		}

		// Backend nodes should not receive incoming connections
		// (except from multiplexer nodes which reference them in config)
		if IsBackendNode(fromNode.Type) {
			incoming := v.spec.GetIncomingConnections(conn.From)
			if len(incoming) > 0 {
				return fmt.Errorf("connection %d: backend node %s (%s) cannot have incoming connections",
					i, conn.From, fromNode.Type)
			}
		}

		// Wrapper nodes (except multiplexers) should have exactly one incoming connection
		if IsWrapperNode(toNode.Type) && !IsMultiplexerNode(toNode.Type) {
			incoming := v.spec.GetIncomingConnections(conn.To)
			if len(incoming) > 1 {
				return fmt.Errorf("wrapper node %s (%s) can only have one incoming connection, has %d",
					conn.To, toNode.Type, len(incoming))
			}
		}

		// Multiplexer nodes (switchfs, unionfs) should not have incoming connections
		// because they reference their backends in config
		if IsMultiplexerNode(toNode.Type) {
			incoming := v.spec.GetIncomingConnections(conn.To)
			if len(incoming) > 0 && toNode.Type == NodeTypeSwitchFS {
				return fmt.Errorf("switchfs node %s should reference backends in config, not via incoming connections",
					conn.To)
			}
		}
	}

	return nil
}

// ValidateNodeConfigs validates node-specific configuration requirements
func (v *Validator) ValidateNodeConfigs() error {
	for _, node := range v.spec.Nodes {
		switch node.Type {
		case NodeTypeOSFS:
			if err := v.validateOSFSConfig(&node); err != nil {
				return err
			}
		case NodeTypeCacheFS:
			if err := v.validateCacheFSConfig(&node); err != nil {
				return err
			}
		case NodeTypeEncryptFS:
			if err := v.validateEncryptFSConfig(&node); err != nil {
				return err
			}
		case NodeTypeSwitchFS:
			if err := v.validateSwitchFSConfig(&node); err != nil {
				return err
			}
		}
		// Add more validators as needed
	}

	return nil
}

// Node-specific configuration validators

func (v *Validator) validateOSFSConfig(node *Node) error {
	if node.Config == nil {
		return fmt.Errorf("node %s: osfs requires 'root' config", node.ID)
	}
	root, ok := node.Config["root"].(string)
	if !ok || root == "" {
		return fmt.Errorf("node %s: osfs requires 'root' config (string path)", node.ID)
	}
	return nil
}

func (v *Validator) validateCacheFSConfig(node *Node) error {
	if node.Config == nil {
		return nil // Use defaults
	}

	// Validate size if provided
	if size, ok := node.Config["size"]; ok {
		if _, ok := size.(int); !ok {
			// Try float64 (YAML unmarshals numbers as float64)
			if sizeFloat, ok := size.(float64); !ok {
				return fmt.Errorf("node %s: cachefs 'size' must be a number", node.ID)
			} else if sizeFloat < 0 {
				return fmt.Errorf("node %s: cachefs 'size' must be positive", node.ID)
			}
		}
	}

	// Validate policy if provided
	if policy, ok := node.Config["policy"]; ok {
		policyStr, ok := policy.(string)
		if !ok {
			return fmt.Errorf("node %s: cachefs 'policy' must be a string", node.ID)
		}
		validPolicies := map[string]bool{"LRU": true, "LFU": true, "ARC": true}
		if !validPolicies[policyStr] {
			return fmt.Errorf("node %s: cachefs 'policy' must be one of: LRU, LFU, ARC", node.ID)
		}
	}

	return nil
}

func (v *Validator) validateEncryptFSConfig(node *Node) error {
	if node.Config == nil {
		return fmt.Errorf("node %s: encryptfs requires configuration", node.ID)
	}

	// Validate algorithm
	algorithm, ok := node.Config["algorithm"].(string)
	if !ok {
		return fmt.Errorf("node %s: encryptfs requires 'algorithm' config", node.ID)
	}
	validAlgos := map[string]bool{"AES-256-GCM": true, "ChaCha20-Poly1305": true}
	if !validAlgos[algorithm] {
		return fmt.Errorf("node %s: invalid encryption algorithm %s", node.ID, algorithm)
	}

	// Validate key source
	keySource, ok := node.Config["keySource"].(string)
	if !ok {
		return fmt.Errorf("node %s: encryptfs requires 'keySource' config", node.ID)
	}

	// Validate key source specific configs
	switch keySource {
	case "env":
		if _, ok := node.Config["keyEnv"].(string); !ok {
			return fmt.Errorf("node %s: encryptfs with keySource=env requires 'keyEnv'", node.ID)
		}
	case "file":
		if _, ok := node.Config["keyFile"].(string); !ok {
			return fmt.Errorf("node %s: encryptfs with keySource=file requires 'keyFile'", node.ID)
		}
	default:
		return fmt.Errorf("node %s: invalid keySource %s (must be env or file)", node.ID, keySource)
	}

	return nil
}

func (v *Validator) validateSwitchFSConfig(node *Node) error {
	if node.Config == nil {
		return fmt.Errorf("node %s: switchfs requires 'routes' config", node.ID)
	}

	routes, ok := node.Config["routes"]
	if !ok {
		return fmt.Errorf("node %s: switchfs requires 'routes' config", node.ID)
	}

	routeList, ok := routes.([]interface{})
	if !ok {
		return fmt.Errorf("node %s: switchfs 'routes' must be a list", node.ID)
	}

	if len(routeList) == 0 {
		return fmt.Errorf("node %s: switchfs requires at least one route", node.ID)
	}

	// Validate each route references an existing node
	for i, route := range routeList {
		routeMap, ok := route.(map[string]interface{})
		if !ok {
			return fmt.Errorf("node %s: route %d is not a valid map", node.ID, i)
		}

		target, ok := routeMap["target"].(string)
		if !ok {
			return fmt.Errorf("node %s: route %d missing 'target'", node.ID, i)
		}

		if v.spec.GetNode(target) == nil {
			return fmt.Errorf("node %s: route %d references unknown target node %s", node.ID, i, target)
		}

		if _, ok := routeMap["pattern"].(string); !ok {
			return fmt.Errorf("node %s: route %d missing 'pattern'", node.ID, i)
		}
	}

	return nil
}
