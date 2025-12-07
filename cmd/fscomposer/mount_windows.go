//go:build windows

package main

import (
	"fmt"

	"github.com/absfs/fscomposer/engine"
)

// mountCommand is not supported on Windows
func mountCommand(args []string) error {
	return fmt.Errorf("mount command is not supported on Windows (FUSE is Unix-only)")
}

// printNodeChain is a stub for Windows - not used since mount is unsupported
func printNodeChain(spec *engine.CompositionSpec) {
	// Not used on Windows
}
