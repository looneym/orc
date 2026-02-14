package tmux

import (
	"fmt"
	"os/exec"

	"github.com/GianlucaP106/gotmux/gotmux"
)

// GotmuxAdapter wraps gotmux library for session lifecycle management
type GotmuxAdapter struct {
	tmux *gotmux.Tmux
}

// NewGotmuxAdapter creates a new gotmux adapter
func NewGotmuxAdapter() (*GotmuxAdapter, error) {
	tmux, err := gotmux.DefaultTmux()
	if err != nil {
		return nil, fmt.Errorf("failed to create tmux client: %w", err)
	}
	return &GotmuxAdapter{
		tmux: tmux,
	}, nil
}

// CreateWorkbenchSession creates a tmux session with a single-pane workbench window.
// The single pane runs "orc connect" (the goblin process).
// Uses NewSession + setupWorkbenchPanes for the initial window.
func (g *GotmuxAdapter) CreateWorkbenchSession(sessionName, workbenchName, workbenchPath, workbenchID, workshopID string) error {
	// Create session with plain shell (setupWorkbenchPanes handles pane setup)
	session, err := g.tmux.NewSession(&gotmux.SessionOptions{
		Name:           sessionName,
		StartDirectory: workbenchPath,
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Get the auto-created first window
	windows, err := session.ListWindows()
	if err != nil {
		return fmt.Errorf("failed to list windows: %w", err)
	}
	if len(windows) == 0 {
		return fmt.Errorf("no windows found in new session")
	}
	firstWindow := windows[0]

	// Set up the single-pane goblin window
	return g.setupWorkbenchPanes(firstWindow, workbenchName, workbenchPath, workbenchID, workshopID)
}

// AddWorkbenchWindow creates a new single-pane window on an existing session.
// The pane runs "orc connect" (the goblin process).
// Pane options (@pane_role, @bench_id, @workshop_id) are set on the goblin pane.
func (g *GotmuxAdapter) AddWorkbenchWindow(session *gotmux.Session, workbenchName, workbenchPath, workbenchID, workshopID string) error {
	// Create new window (no ShellCommand available on NewWindowOptions)
	window, err := session.NewWindow(&gotmux.NewWindowOptions{
		WindowName:     workbenchName,
		StartDirectory: workbenchPath,
		DoNotAttach:    true,
	})
	if err != nil {
		return fmt.Errorf("failed to create window %s: %w", workbenchName, err)
	}

	return g.setupWorkbenchPanes(window, workbenchName, workbenchPath, workbenchID, workshopID)
}

// setupWorkbenchPanes configures a window as a single-pane goblin workspace.
// The window must already exist with at least one pane. It will be renamed to workbenchName
// and the single pane will run "orc connect" as its root process.
func (g *GotmuxAdapter) setupWorkbenchPanes(window *gotmux.Window, workbenchName, workbenchPath, workbenchID, workshopID string) error {
	// Rename window to workbench name
	if err := window.Rename(workbenchName); err != nil {
		return fmt.Errorf("failed to rename window: %w", err)
	}

	// Get the single pane (starts as plain shell)
	panes, err := window.ListPanes()
	if err != nil || len(panes) == 0 {
		return fmt.Errorf("failed to get initial pane: %w", err)
	}
	goblinPane := panes[0]

	// Make "orc connect" the root process via respawn-pane -k
	// (NewWindowOptions doesn't support ShellCommand, so we respawn)
	if err := exec.Command("tmux", "respawn-pane", "-t", goblinPane.Id, "-k", "orc", "connect").Run(); err != nil {
		return fmt.Errorf("failed to respawn goblin pane: %w", err)
	}

	// Set tmux pane options for identity
	// @pane_role is authoritative for pane identity (readable via #{@pane_role})
	// @bench_id and @workshop_id provide context without shell env vars
	if err := goblinPane.SetOption("@pane_role", "goblin"); err != nil {
		return fmt.Errorf("failed to set @pane_role=goblin: %w", err)
	}
	if err := goblinPane.SetOption("@bench_id", workbenchID); err != nil {
		return fmt.Errorf("failed to set @bench_id on goblin pane: %w", err)
	}
	if err := goblinPane.SetOption("@workshop_id", workshopID); err != nil {
		return fmt.Errorf("failed to set @workshop_id on goblin pane: %w", err)
	}

	return nil
}

// GetSession returns a gotmux Session by name, or nil if not found.
// Returns (nil, nil) when the tmux server is not running, allowing callers
// like PlanApply to treat a dead server as "no sessions exist".
func (g *GotmuxAdapter) GetSession(name string) (*gotmux.Session, error) {
	sessions, _ := g.tmux.ListSessions()
	// gotmux returns an error when the tmux server isn't running.
	// Treat this as "no sessions" — downstream CreateSession will start a new server.
	for _, s := range sessions {
		if s.Name == name {
			return s, nil
		}
	}
	return nil, nil
}

// SessionExists checks if a tmux session exists
func (g *GotmuxAdapter) SessionExists(name string) bool {
	sessions, err := g.tmux.ListSessions()
	if err != nil {
		return false
	}
	for _, s := range sessions {
		if s.Name == name {
			return true
		}
	}
	return false
}

// KillSession terminates a tmux session
func (g *GotmuxAdapter) KillSession(name string) error {
	sessions, err := g.tmux.ListSessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}
	for _, s := range sessions {
		if s.Name == name {
			return s.Kill()
		}
	}
	return fmt.Errorf("session %s not found", name)
}

// ApplyActionType identifies the kind of reconciliation action.
type ApplyActionType string

const (
	ActionCreateSession   ApplyActionType = "CreateSession"
	ActionAddWindow       ApplyActionType = "AddWindow"
	ActionApplyEnrichment ApplyActionType = "ApplyEnrichment"
)

// ApplyAction represents a single reconciliation action in the plan.
type ApplyAction struct {
	Type        ApplyActionType
	Description string

	// Context fields used during execution (not all are set for every action type)
	SessionName   string
	WorkbenchName string
	WorkbenchPath string
	WorkbenchID   string
	WorkshopID    string
}

// DesiredWorkbench describes a workbench that should exist as a window.
type DesiredWorkbench struct {
	Name       string
	Path       string
	ID         string
	WorkshopID string
}

// ApplyPlan contains the full reconciliation plan.
type ApplyPlan struct {
	SessionName   string
	SessionExists bool
	Actions       []ApplyAction
	WindowSummary []WindowStatus
}

// WindowStatus summarizes a window's current state for display.
type WindowStatus struct {
	Name      string
	PaneCount int
	Healthy   bool
}

// PlanApply compares desired state (workbenches) to actual tmux state and returns actions.
func (g *GotmuxAdapter) PlanApply(sessionName string, workbenches []DesiredWorkbench) (*ApplyPlan, error) {
	plan := &ApplyPlan{SessionName: sessionName}

	// Check if session exists
	session, err := g.GetSession(sessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check session: %w", err)
	}

	if session == nil {
		// Session doesn't exist — need to create everything
		plan.SessionExists = false

		if len(workbenches) == 0 {
			return plan, nil
		}

		// First workbench creates the session
		first := workbenches[0]
		plan.Actions = append(plan.Actions, ApplyAction{
			Type:          ActionCreateSession,
			Description:   fmt.Sprintf("Create session %s with window %s", sessionName, first.Name),
			SessionName:   sessionName,
			WorkbenchName: first.Name,
			WorkbenchPath: first.Path,
			WorkbenchID:   first.ID,
			WorkshopID:    first.WorkshopID,
		})

		// Remaining workbenches get added as windows
		for _, wb := range workbenches[1:] {
			plan.Actions = append(plan.Actions, ApplyAction{
				Type:          ActionAddWindow,
				Description:   fmt.Sprintf("Add window %s (%s)", wb.Name, wb.ID),
				SessionName:   sessionName,
				WorkbenchName: wb.Name,
				WorkbenchPath: wb.Path,
				WorkbenchID:   wb.ID,
				WorkshopID:    wb.WorkshopID,
			})
		}

		// Always enrich at the end
		plan.Actions = append(plan.Actions, ApplyAction{
			Type:        ActionApplyEnrichment,
			Description: "Apply ORC enrichment (bindings, pane titles)",
			SessionName: sessionName,
		})

		return plan, nil
	}

	// Session exists — reconcile windows
	plan.SessionExists = true

	windows, err := session.ListWindows()
	if err != nil {
		return nil, fmt.Errorf("failed to list windows: %w", err)
	}

	// Build set of existing window names
	existingWindows := make(map[string]*gotmux.Window, len(windows))
	for _, w := range windows {
		existingWindows[w.Name] = w
	}

	// Check which workbenches need windows
	for _, wb := range workbenches {
		if _, exists := existingWindows[wb.Name]; !exists {
			plan.Actions = append(plan.Actions, ApplyAction{
				Type:          ActionAddWindow,
				Description:   fmt.Sprintf("Add window %s (%s)", wb.Name, wb.ID),
				SessionName:   sessionName,
				WorkbenchName: wb.Name,
				WorkbenchPath: wb.Path,
				WorkbenchID:   wb.ID,
				WorkshopID:    wb.WorkshopID,
			})
		}
	}

	// Summarize existing windows
	for _, w := range windows {
		panes, err := w.ListPanes()
		if err != nil {
			continue
		}
		plan.WindowSummary = append(plan.WindowSummary, WindowStatus{
			Name:      w.Name,
			PaneCount: len(panes),
			Healthy:   len(panes) > 0,
		})
	}

	// Always enrich at the end
	plan.Actions = append(plan.Actions, ApplyAction{
		Type:        ActionApplyEnrichment,
		Description: "Apply ORC enrichment (bindings, pane titles)",
		SessionName: sessionName,
	})

	return plan, nil
}

// ExecutePlan executes all actions in a plan sequentially.
func (g *GotmuxAdapter) ExecutePlan(plan *ApplyPlan) error {
	for _, action := range plan.Actions {
		if err := g.executeAction(action); err != nil {
			return fmt.Errorf("action %s failed: %w", action.Type, err)
		}
	}
	return nil
}

// executeAction dispatches a single reconciliation action.
func (g *GotmuxAdapter) executeAction(action ApplyAction) error {
	switch action.Type {
	case ActionCreateSession:
		return g.CreateWorkbenchSession(action.SessionName, action.WorkbenchName, action.WorkbenchPath, action.WorkbenchID, action.WorkshopID)

	case ActionAddWindow:
		session, err := g.GetSession(action.SessionName)
		if err != nil {
			return fmt.Errorf("failed to get session: %w", err)
		}
		if session == nil {
			return fmt.Errorf("session %s not found", action.SessionName)
		}
		return g.AddWorkbenchWindow(session, action.WorkbenchName, action.WorkbenchPath, action.WorkbenchID, action.WorkshopID)

	case ActionApplyEnrichment:
		ApplyGlobalBindings()
		return EnrichSession(action.SessionName)

	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// AttachInstructions returns instructions for attaching to a session
func (g *GotmuxAdapter) AttachInstructions(sessionName string) string {
	return fmt.Sprintf("Attach to session: tmux attach -t %s\n\n"+
		"Window Layout:\n"+
		"  Each window has a single goblin pane (orc connect)\n\n"+
		"TMux Commands:\n"+
		"  Switch windows: Ctrl+b then window number (1, 2, 3...)\n"+
		"  Detach session: Ctrl+b then d\n"+
		"  Open desk: Double-click status bar or Ctrl+b then u\n",
		sessionName)
}
