package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MissionContext represents the .orc-mission marker file
type MissionContext struct {
	MissionID     string    `json:"mission_id"`
	WorkspacePath string    `json:"workspace_path"`
	CreatedAt     time.Time `json:"created_at"`
}

const markerFileName = ".orc-mission"

// DetectMissionContext checks if we're in a deputy ORC context
// by looking for .orc-mission file in current directory or parents
func DetectMissionContext() (*MissionContext, error) {
	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Walk up directory tree looking for .orc-mission
	for {
		markerPath := filepath.Join(dir, markerFileName)
		if _, err := os.Stat(markerPath); err == nil {
			// Found it - read and parse
			return ReadMissionContext(markerPath)
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding marker
			return nil, nil
		}
		dir = parent
	}
}

// ReadMissionContext reads and parses a .orc-mission file
func ReadMissionContext(path string) (*MissionContext, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read mission context: %w", err)
	}

	var ctx MissionContext
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, fmt.Errorf("failed to parse mission context: %w", err)
	}

	return &ctx, nil
}

// WriteMissionContext creates a .orc-mission marker file
func WriteMissionContext(workspacePath, missionID string) error {
	ctx := MissionContext{
		MissionID:     missionID,
		WorkspacePath: workspacePath,
		CreatedAt:     time.Now(),
	}

	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal mission context: %w", err)
	}

	markerPath := filepath.Join(workspacePath, markerFileName)
	if err := os.WriteFile(markerPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write mission context: %w", err)
	}

	return nil
}

// GetContextMissionID returns the mission ID from context, or empty string if not in deputy context
func GetContextMissionID() string {
	ctx, err := DetectMissionContext()
	if err != nil || ctx == nil {
		return ""
	}
	return ctx.MissionID
}

// IsDeputyContext returns true if we're running in a deputy ORC context
func IsDeputyContext() bool {
	ctx, _ := DetectMissionContext()
	return ctx != nil
}

// IsOrcSourceDirectory checks if the current directory is the ORC source code directory
// Used to prevent deputy ORCs from modifying the orchestrator source
func IsOrcSourceDirectory() bool {
	// Check for key ORC source files
	markers := []string{"cmd/orc/main.go", "internal/db/schema.go", "go.mod"}

	for _, marker := range markers {
		if _, err := os.Stat(marker); err == nil {
			// Check if go.mod contains ORC module
			if marker == "go.mod" {
				data, err := os.ReadFile(marker)
				if err == nil && len(data) > 0 {
					// Simple check for orc module name
					content := string(data)
					if len(content) > 20 && (content[:20] == "module github.com/lo" || content[:30] == "module github.com/looneym/orc") {
						return true
					}
				}
			} else {
				return true
			}
		}
	}

	return false
}
