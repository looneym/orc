package agent

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/context"
)

// AgentType represents the type of agent
type AgentType string

const (
	AgentTypeMaster AgentType = "MASTER"
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

	// If we're NOT in a mission context, check if we're master ORC
	if missionCtx == nil {
		tmuxSession := os.Getenv("TMUX")
		if tmuxSession != "" {
			sessionName := getCurrentTmuxSession()
			if sessionName == "ORC" {
				// Master ORC working outside mission contexts
				return &AgentIdentity{
					Type:      AgentTypeMaster,
					ID:        "ORC",
					FullID:    "MASTER-ORC",
					MissionID: "", // Master doesn't belong to a mission
				}, nil
			}
		}
		return nil, fmt.Errorf("no agent context detected - not in a mission workspace or grove")
	}

	// We're in a mission context - check tmux to see if we're master visiting
	// or an actual deputy for this mission
	tmuxSession := os.Getenv("TMUX")
	if tmuxSession != "" {
		sessionName := getCurrentTmuxSession()
		// Master ORC session, but working in a mission directory
		if sessionName == "ORC" {
			// Master visiting mission - use mission for scoping but identify as master
			return &AgentIdentity{
				Type:      AgentTypeMaster,
				ID:        "ORC",
				FullID:    "MASTER-ORC",
				MissionID: missionCtx.MissionID, // Use mission for message scoping
			}, nil
		}
		// Deputy session for this specific mission
		if sessionName == fmt.Sprintf("orc-%s", missionCtx.MissionID) {
			// Actual deputy for this mission
			return &AgentIdentity{
				Type:      AgentTypeDeputy,
				ID:        missionCtx.MissionID,
				FullID:    fmt.Sprintf("DEPUTY-%s", missionCtx.MissionID),
				MissionID: missionCtx.MissionID,
			}, nil
		}
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

	// Otherwise check for grove config
	cfg, err := config.LoadConfigWithFallback(cwd)
	if err == nil && cfg.Type == config.TypeGrove {
		return &AgentIdentity{
			Type:      AgentTypeIMP,
			ID:        cfg.Grove.GroveID,
			FullID:    fmt.Sprintf("IMP-%s", cfg.Grove.GroveID),
			MissionID: cfg.Grove.MissionID,
		}, nil
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
	case AgentTypeMaster:
		return &AgentIdentity{
			Type:      AgentTypeMaster,
			ID:        id,
			FullID:    agentID,
			MissionID: "", // Master doesn't belong to a specific mission
		}, nil
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

// getCurrentTmuxSession returns the current tmux session name, or empty string if not in tmux
func getCurrentTmuxSession() string {
	cmd := exec.Command("tmux", "display-message", "-p", "#S")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
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

	if identity.Type == AgentTypeMaster {
		// Master ORC in ORC session, window 1, pane 1
		return "ORC:1.1", nil
	}

	// For IMP, need grove name and mission ID
	if identity.MissionID == "" || groveName == "" {
		return "", fmt.Errorf("IMP target requires mission ID and grove name")
	}

	// Window named by grove, pane 2 is Claude
	return fmt.Sprintf("orc-%s:%s.2", identity.MissionID, groveName), nil
}
