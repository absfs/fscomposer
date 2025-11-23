package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/absfs/fscomposer/engine"
	"github.com/absfs/fscomposer/registry"
)

const version = "0.1.0-poc"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "validate":
		if err := validateCommand(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "build":
		if err := buildCommand(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "nodes":
		if err := nodesCommand(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "info":
		if err := infoCommand(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "version":
		fmt.Printf("fscomposer version %s (POC)\n", version)

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("fscomposer - Visual Filesystem Composition Studio")
	fmt.Printf("Version: %s\n\n", version)
	fmt.Println("Usage:")
	fmt.Println("  fscomposer <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  validate <spec.yaml>     Validate a composition spec")
	fmt.Println("  build <spec.yaml>        Build and test a composition")
	fmt.Println("  nodes [list|<type>]      Show available node types or details")
	fmt.Println("  info <spec.yaml>         Show composition information")
	fmt.Println("  version                  Show version information")
	fmt.Println("  help                     Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  fscomposer validate examples/encrypted-s3.yaml")
	fmt.Println("  fscomposer build examples/encrypted-s3.yaml")
	fmt.Println("  fscomposer nodes list")
	fmt.Println("  fscomposer nodes cachefs")
}

// validateCommand validates a composition spec
func validateCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: fscomposer validate <spec.yaml>")
	}

	filename := args[0]

	fmt.Printf("Validating: %s\n", filename)

	// Parse the spec
	spec, err := engine.ParseFile(filename)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	fmt.Printf("✓ Spec format valid\n")

	// Validate
	validator := engine.NewValidator(spec)
	if err := validator.ValidateAll(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	fmt.Printf("✓ All node types registered\n")
	fmt.Printf("✓ No cycles detected\n")
	fmt.Printf("✓ Connection types compatible\n")
	fmt.Printf("✓ Node configurations valid\n")
	fmt.Println()
	fmt.Printf("Composition '%s' is valid!\n", spec.Name)
	fmt.Printf("  Nodes: %d\n", len(spec.Nodes))
	fmt.Printf("  Connections: %d\n", len(spec.Connections))
	fmt.Printf("  Mount: %s (%s)\n", spec.Mount.Type, spec.Mount.Root)

	return nil
}

// buildCommand builds and tests a composition
func buildCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: fscomposer build <spec.yaml>")
	}

	filename := args[0]

	fmt.Printf("Building: %s\n", filename)

	// Parse the spec
	spec, err := engine.ParseFile(filename)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	fmt.Printf("✓ Spec parsed\n")

	// Build the filesystem stack
	builder := engine.NewBuilder(spec)
	fs, err := builder.Build()
	if err != nil {
		return fmt.Errorf("build error: %w", err)
	}

	fmt.Printf("✓ Filesystem stack built\n")

	// Test basic operations
	fmt.Println("\nTesting basic operations...")

	// Test: Create a test file
	testFile := "test.txt"
	testData := []byte("Hello from fscomposer POC!")

	f, err := fs.Create(testFile)
	if err != nil {
		return fmt.Errorf("failed to create test file: %w", err)
	}

	n, err := f.Write(testData)
	if err != nil {
		f.Close()
		return fmt.Errorf("failed to write test file: %w", err)
	}
	f.Close()

	fmt.Printf("✓ Created and wrote %d bytes to %s\n", n, testFile)

	// Test: Read the file back
	f, err = fs.Open(testFile)
	if err != nil {
		return fmt.Errorf("failed to open test file: %w", err)
	}

	readData := make([]byte, len(testData))
	n, err = f.Read(readData)
	f.Close()
	if err != nil {
		return fmt.Errorf("failed to read test file: %w", err)
	}

	fmt.Printf("✓ Read %d bytes from %s\n", n, testFile)

	if string(readData[:n]) != string(testData) {
		return fmt.Errorf("data mismatch: got %q, want %q", readData[:n], testData)
	}

	fmt.Printf("✓ Data integrity verified\n")

	// Test: Stat the file
	info, err := fs.Stat(testFile)
	if err != nil {
		return fmt.Errorf("failed to stat test file: %w", err)
	}

	fmt.Printf("✓ File stat: %s (%d bytes)\n", info.Name(), info.Size())

	// Test: Remove the file
	if err := fs.Remove(testFile); err != nil {
		return fmt.Errorf("failed to remove test file: %w", err)
	}

	fmt.Printf("✓ Removed %s\n", testFile)

	fmt.Println()
	fmt.Printf("✓ All tests passed!\n")
	fmt.Printf("\nComposition '%s' is working correctly.\n", spec.Name)
	fmt.Println("Stack composition:")

	// Print the node chain
	printNodeChain(spec)

	return nil
}

// nodesCommand lists available node types or shows details
func nodesCommand(args []string) error {
	if len(args) == 0 || args[0] == "list" {
		// List all node types
		types := registry.ListTypes()
		sort.Strings(types)

		fmt.Println("Available node types:")
		fmt.Println()

		// Group by category
		backends := []string{}
		wrappers := []string{}

		for _, t := range types {
			if engine.IsBackendNode(t) {
				backends = append(backends, t)
			} else {
				wrappers = append(wrappers, t)
			}
		}

		if len(backends) > 0 {
			fmt.Println("Backends (Data Sources):")
			for _, t := range backends {
				schema, _ := registry.GetSchema(t)
				fmt.Printf("  %-12s  %s\n", t, schema.Description)
			}
			fmt.Println()
		}

		if len(wrappers) > 0 {
			fmt.Println("Wrappers (Middleware):")
			for _, t := range wrappers {
				schema, _ := registry.GetSchema(t)
				fmt.Printf("  %-12s  %s\n", t, schema.Description)
			}
		}

		return nil
	}

	// Show details for a specific node type
	nodeType := args[0]
	schema, err := registry.GetSchema(nodeType)
	if err != nil {
		return err
	}

	fmt.Printf("Node Type: %s\n", schema.Type)
	fmt.Printf("Description: %s\n", schema.Description)
	fmt.Println()

	if len(schema.Fields) > 0 {
		fmt.Println("Configuration Fields:")
		for _, field := range schema.Fields {
			required := ""
			if field.Required {
				required = " (required)"
			}
			fmt.Printf("  %s: %s%s\n", field.Name, field.Type, required)
			if field.Description != "" {
				fmt.Printf("    %s\n", field.Description)
			}
			if field.Default != nil {
				fmt.Printf("    Default: %v\n", field.Default)
			}
			if len(field.Options) > 0 {
				fmt.Printf("    Options: %s\n", strings.Join(field.Options, ", "))
			}
		}
	} else {
		fmt.Println("No configuration required.")
	}

	return nil
}

// infoCommand shows information about a composition
func infoCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: fscomposer info <spec.yaml>")
	}

	filename := args[0]

	// Parse the spec
	spec, err := engine.ParseFile(filename)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	fmt.Printf("Composition: %s\n", spec.Name)
	if spec.Description != "" {
		fmt.Printf("Description: %s\n", spec.Description)
	}
	fmt.Printf("Version: %s\n", spec.Version)
	fmt.Println()

	fmt.Printf("Nodes (%d):\n", len(spec.Nodes))
	for _, node := range spec.Nodes {
		fmt.Printf("  - %s (%s)\n", node.ID, node.Type)
	}
	fmt.Println()

	fmt.Printf("Connections (%d):\n", len(spec.Connections))
	for _, conn := range spec.Connections {
		fmt.Printf("  %s -> %s\n", conn.From, conn.To)
	}
	fmt.Println()

	fmt.Printf("Mount:\n")
	fmt.Printf("  Type: %s\n", spec.Mount.Type)
	fmt.Printf("  Root: %s\n", spec.Mount.Root)
	if spec.Mount.Path != "" {
		fmt.Printf("  Path: %s\n", spec.Mount.Path)
	}
	if spec.Mount.Port != 0 {
		fmt.Printf("  Port: %d\n", spec.Mount.Port)
	}

	return nil
}

// printNodeChain prints the chain of nodes in a composition
func printNodeChain(spec *engine.CompositionSpec) {
	// Find backend nodes (nodes with no incoming connections)
	backends := []string{}
	for _, node := range spec.Nodes {
		incoming := spec.GetIncomingConnections(node.ID)
		if len(incoming) == 0 {
			backends = append(backends, node.ID)
		}
	}

	// For each backend, trace the chain to the mount
	for _, backendID := range backends {
		chain := []string{}
		visited := make(map[string]bool)

		current := backendID
		for current != "" {
			if visited[current] {
				break // Avoid infinite loops
			}
			visited[current] = true

			node := spec.GetNode(current)
			if node != nil {
				chain = append(chain, fmt.Sprintf("%s (%s)", node.ID, node.Type))
			}

			// Find next node in chain
			outgoing := spec.GetOutgoingConnections(current)
			if len(outgoing) > 0 {
				current = outgoing[0].To
			} else {
				current = ""
			}
		}

		if len(chain) > 0 {
			fmt.Printf("  %s\n", strings.Join(chain, " → "))
		}
	}
}
