package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/example/orc/internal/config"
)

// MissionContext represents mission context information
type MissionContext struct {
	MissionID     string    `json:"mission_id"`
	WorkspacePath string    `json:"workspace_path"`
	IsMaster      bool      `json:"is_master"`
	CreatedAt     time.Time `json:"created_at"`
}

const markerFileName = ".orc-mission"

// DetectMissionContext checks if we're in a mission context
// by looking for .orc/config.json (or legacy .orc-mission) in current directory or parents
func DetectMissionContext() (*MissionContext, error) {
	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Walk up directory tree looking for config
	for {
		cfg, err := config.LoadConfigWithFallback(dir)
		if err == nil && cfg.Type == config.TypeMission {
			// Found mission config - convert to MissionContext
			createdAt, _ := time.Parse(time.RFC3339, cfg.Mission.CreatedAt)
			return &MissionContext{
				MissionID:     cfg.Mission.MissionID,
				WorkspacePath: cfg.Mission.WorkspacePath,
				IsMaster:      cfg.Mission.IsMaster,
				CreatedAt:     createdAt,
			}, nil
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding config
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

// WriteMissionContext creates a .orc/config.json file for mission context
func WriteMissionContext(workspacePath, missionID string) error {
	return WriteMissionConfig(workspacePath, missionID, false)
}

// WriteMissionConfig creates a .orc/config.json file with full control over fields
func WriteMissionConfig(workspacePath, missionID string, isMaster bool) error {
	cfg := &config.Config{
		Version: "1.0",
		Type:    config.TypeMission,
		Mission: &config.MissionConfig{
			MissionID:     missionID,
			WorkspacePath: workspacePath,
			IsMaster:      isMaster,
			CreatedAt:     time.Now().Format(time.RFC3339),
		},
	}

	return config.SaveConfig(workspacePath, cfg)
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
					if len(content) > 20 && (content[:20] == "module github.com/lo" || content[:30] == "module github.com/example/orc") {
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
