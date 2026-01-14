package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/example/orc/internal/context"
)

// AgentType represents the type of agent
type AgentType string

const (
	AgentTypeDeputy AgentType = "DEPUTY"
	AgentTypeIMP    AgentType = "IMP"
)

// AgentIdentity represents a parsed agent ID
type AgentIdentity struct {
	Type      AgentType
	ID        string // Mission ID for deputy, Grove ID for IMP
	FullID    string // Complete ID like "DEPUTY-MISSION-001" or "IMP-GROVE-001"
	MissionID string // Mission ID (same as ID for deputy, extracted from grove for IMP)
}

// GetCurrentAgentID detects the current agent identity based on working directory context
func GetCurrentAgentID() (*AgentIdentity, error) {
	// Check mission context first
	missionCtx, err := context.DetectMissionContext()
	if err != nil {
		return nil, fmt.Errorf("failed to detect mission context: %w", err)
	}

	if missionCtx == nil {
		return nil, fmt.Errorf("no agent context detected - not in a mission workspace or grove")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	// If at workspace root, we're deputy
	if cwd == missionCtx.WorkspacePath {
		return &AgentIdentity{
			Type:      AgentTypeDeputy,
			ID:        missionCtx.MissionID,
			FullID:    fmt.Sprintf("DEPUTY-%s", missionCtx.MissionID),
			MissionID: missionCtx.MissionID,
		}, nil
	}

	// Otherwise check for grove metadata
	metadataPath := filepath.Join(cwd, ".orc", "metadata.json")
	data, err := os.ReadFile(metadataPath)
	if err == nil {
		var metadata struct {
			GroveID   string `json:"grove_id"`
			MissionID string `json:"mission_id"`
		}
		if err := json.Unmarshal(data, &metadata); err == nil && metadata.GroveID != "" {
			return &AgentIdentity{
				Type:      AgentTypeIMP,
				ID:        metadata.GroveID,
				FullID:    fmt.Sprintf("IMP-%s", metadata.GroveID),
				MissionID: metadata.MissionID,
			}, nil
		}
	}

	// Fallback: in mission context but not at workspace root = deputy
	return &AgentIdentity{
		Type:      AgentTypeDeputy,
		ID:        missionCtx.MissionID,
		FullID:    fmt.Sprintf("DEPUTY-%s", missionCtx.MissionID),
		MissionID: missionCtx.MissionID,
	}, nil
}

// ParseAgentID parses an agent ID string like "DEPUTY-MISSION-001" or "IMP-GROVE-001"
func ParseAgentID(agentID string) (*AgentIdentity, error) {
	parts := strings.SplitN(agentID, "-", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid agent ID format: %s (expected TYPE-ID)", agentID)
	}

	agentType := AgentType(parts[0])
	id := parts[1]

	switch agentType {
	case AgentTypeDeputy:
		return &AgentIdentity{
			Type:      AgentTypeDeputy,
			ID:        id,
			FullID:    agentID,
			MissionID: id,
		}, nil
	case AgentTypeIMP:
		// For IMP, we need to extract mission ID from grove ID
		// Grove IDs are like GROVE-001, we need to look up the mission
		// For now, return partial identity (caller must resolve mission)
		return &AgentIdentity{
			Type:   AgentTypeIMP,
			ID:     id,
			FullID: agentID,
		}, nil
	default:
		return nil, fmt.Errorf("unknown agent type: %s", agentType)
	}
}

// ResolveTMuxTarget converts an agent ID to a tmux target string
func ResolveTMuxTarget(agentID string, groveName string) (string, error) {
	identity, err := ParseAgentID(agentID)
	if err != nil {
		return "", err
	}

	if identity.Type == AgentTypeDeputy {
		// Deputy always in window 1, pane 1
		return fmt.Sprintf("orc-%s:1.1", identity.ID), nil
	}

	// For IMP, need grove name and mission ID
	if identity.MissionID == "" || groveName == "" {
		return "", fmt.Errorf("IMP target requires mission ID and grove name")
	}

	// Window named by grove, pane 2 is Claude
	return fmt.Sprintf("orc-%s:%s.2", identity.MissionID, groveName), nil
}
