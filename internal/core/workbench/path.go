// Package workbench contains domain logic for workbench operations.
package workbench

import (
	"os"
	"path/filepath"
)

// DefaultBasePath is the directory under home where workbenches are created.
const DefaultBasePath = "wb"

// ComputePath returns the canonical filesystem path for a workbench.
// Path is deterministic: ~/wb/{workbenchName}
func ComputePath(workbenchName string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, DefaultBasePath, workbenchName)
}
