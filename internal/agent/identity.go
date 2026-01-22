package agent

import (
	"fmt"
	"os"
	"strings"

	"github.com/example/orc/internal/config"
)

// AgentType represents the type of agent
type AgentType string

const (
	AgentTypeGoblin AgentType = "GOBLIN"
	AgentTypeIMP    AgentType = "IMP"
)

// AgentIdentity represents a parsed agent ID
type AgentIdentity struct {
	Type         AgentType
	ID           string // "GOBLIN" for orchestrator, Workbench ID for IMP
	FullID       string // Complete ID like "GOBLIN" or "IMP-BENCH-001"
	CommissionID string // Commission ID (empty for Goblin outside commission, populated for IMP)
}

// GetCurrentAgentID detects the current agent identity based on working directory context
// Simple logic: If has IMP role with workbench → IMP, otherwise → Goblin
func GetCurrentAgentID() (*AgentIdentity, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	// Check for config in current directory
	cfg, err := config.LoadConfig(cwd)
	if err == nil && cfg.Role == config.RoleIMP && cfg.WorkbenchID != "" {
		// We're in a workbench - this is an IMP
		return &AgentIdentity{
			Type:         AgentTypeIMP,
			ID:           cfg.WorkbenchID,
			FullID:       fmt.Sprintf("IMP-%s", cfg.WorkbenchID),
			CommissionID: cfg.CommissionID,
		}, nil
	}

	// Not in a workbench with IMP role - we're a Goblin (orchestrator)
	// Goblin can work anywhere: commission workspaces, ORC repo, anywhere
	commissionID := ""
	if cfg != nil {
		commissionID = cfg.CommissionID
	}

	return &AgentIdentity{
		Type:         AgentTypeGoblin,
		ID:           "GOBLIN",
		FullID:       "GOBLIN",
		CommissionID: commissionID, // Populated if in commission context, empty otherwise
	}, nil
}

// ParseAgentID parses an agent ID string like "GOBLIN" or "IMP-BENCH-001"
func ParseAgentID(agentID string) (*AgentIdentity, error) {
	// Special case: GOBLIN or ORC (backwards compat) has no parts
	if agentID == "GOBLIN" || agentID == "ORC" {
		return &AgentIdentity{
			Type:         AgentTypeGoblin,
			ID:           "GOBLIN",
			FullID:       "GOBLIN",
			CommissionID: "", // Goblin can work across commissions
		}, nil
	}

	parts := strings.SplitN(agentID, "-", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid agent ID format: %s (expected GOBLIN or IMP-WORKBENCH-ID)", agentID)
	}

	agentType := AgentType(parts[0])
	id := parts[1]

	switch agentType {
	case AgentTypeIMP:
		// For IMP, we need to extract commission ID from workbench ID
		// Workbench IDs are like BENCH-001, we need to look up the commission
		// For now, return partial identity (caller must resolve commission)
		return &AgentIdentity{
			Type:   AgentTypeIMP,
			ID:     id,
			FullID: agentID,
		}, nil
	default:
		return nil, fmt.Errorf("unknown agent type: %s (expected GOBLIN or IMP)", agentType)
	}
}

// ResolveTMuxTarget converts an agent ID to a tmux target string
func ResolveTMuxTarget(agentID string, workbenchName string) (string, error) {
	identity, err := ParseAgentID(agentID)
	if err != nil {
		return "", err
	}

	if identity.Type == AgentTypeGoblin {
		// Goblin always in ORC session, window 1, pane 1
		return "ORC:1.1", nil
	}

	// For IMP, need workbench name and commission ID
	if identity.CommissionID == "" || workbenchName == "" {
		return "", fmt.Errorf("IMP target requires commission ID and workbench name")
	}

	// Window named by workbench, pane 2 is Claude
	return fmt.Sprintf("orc-%s:%s.2", identity.CommissionID, workbenchName), nil
}
