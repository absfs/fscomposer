//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/absfs/fscomposer/engine"
	"github.com/absfs/fusefs"
)

// mountCommand mounts a composition via FUSE
func mountCommand(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: fscomposer mount <spec.yaml> <mountpoint>")
	}

	specFile := args[0]
	mountpoint := args[1]

	fmt.Printf("Mounting: %s at %s\n", specFile, mountpoint)

	// Parse the spec
	spec, err := engine.ParseFile(specFile)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	fmt.Printf("✓ Spec parsed: %s\n", spec.Name)

	// Build the filesystem stack
	builder := engine.NewBuilder(spec)
	fs, err := builder.Build()
	if err != nil {
		return fmt.Errorf("build error: %w", err)
	}

	fmt.Printf("✓ Filesystem stack built\n")
	fmt.Println("\nStack composition:")
	printNodeChain(spec)
	fmt.Println()

	// Mount via FUSE
	opts := fusefs.DefaultMountOptions(mountpoint)
	opts.FSName = spec.Name
	if spec.Description != "" {
		opts.FSName = spec.Name + " - " + spec.Description
	}

	fmt.Printf("Mounting at %s...\n", mountpoint)
	fmt.Println("Press Ctrl+C to unmount")

	fuseFS, err := fusefs.Mount(fs, opts)
	if err != nil {
		return fmt.Errorf("failed to mount: %w", err)
	}

	// Set up signal handling for graceful unmount
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("✓ Mounted successfully at %s\n", mountpoint)
	fmt.Println("\nFilesystem is now available. Press Ctrl+C to unmount.")

	// Wait for interrupt signal
	<-sigChan

	fmt.Println("\nUnmounting...")
	if err := fuseFS.Unmount(); err != nil {
		return fmt.Errorf("failed to unmount: %w", err)
	}

	fmt.Println("✓ Unmounted successfully")
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
