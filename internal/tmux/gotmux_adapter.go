package tmux

import (
	"fmt"
	"strings"

	"github.com/GianlucaP106/gotmux/gotmux"
)

// GotmuxAdapter wraps gotmux library for session lifecycle management.
// For the default server (Socket==""), it uses the gotmux library.
// For custom sockets, it uses Server.cmd() for all operations since
// gotmux only supports DefaultTmux().
type GotmuxAdapter struct {
	tmux   *gotmux.Tmux // used when Socket=="" (default server)
	server *Server      // used for socket-aware operations
}

// NewGotmuxAdapter creates a new gotmux adapter targeting the given socket.
// An empty socket means the default tmux server (uses gotmux library).
// A non-empty socket means a custom server (uses Server.cmd() for operations).
func NewGotmuxAdapter(socket string) (*GotmuxAdapter, error) {
	server := &Server{Socket: socket}
	if socket == "" {
		// Default server: use gotmux library
		tmux, err := gotmux.DefaultTmux()
		if err != nil {
			return nil, fmt.Errorf("failed to create tmux client: %w", err)
		}
		return &GotmuxAdapter{tmux: tmux, server: server}, nil
	}
	// Custom socket: gotmux not available, use Server.cmd() only
	return &GotmuxAdapter{server: server}, nil
}

// CreateWorkbenchSession creates a tmux session with a single-pane workbench window.
// The single pane runs "orc connect" (the goblin process).
func (g *GotmuxAdapter) CreateWorkbenchSession(sessionName, workbenchName, workbenchPath, workbenchID, workshopID string) error {
	if g.tmux != nil {
		return g.createWorkbenchSessionGotmux(sessionName, workbenchName, workbenchPath, workbenchID, workshopID)
	}
	return g.createWorkbenchSessionCmd(sessionName, workbenchName, workbenchPath, workbenchID, workshopID)
}

// createWorkbenchSessionGotmux uses gotmux library (default server only).
func (g *GotmuxAdapter) createWorkbenchSessionGotmux(sessionName, workbenchName, workbenchPath, workbenchID, workshopID string) error {
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
	return g.setupWorkbenchPanesGotmux(firstWindow, workbenchName, workbenchPath, workbenchID, workshopID)
}

// createWorkbenchSessionCmd uses Server.cmd() (works with any socket).
func (g *GotmuxAdapter) createWorkbenchSessionCmd(sessionName, workbenchName, workbenchPath, workbenchID, workshopID string) error {
	// Create session
	if err := g.server.cmd("new-session", "-d", "-s", sessionName, "-c", workbenchPath).Run(); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Set up the single-pane goblin window
	return g.setupWorkbenchPanesCmd(sessionName, workbenchName, workbenchID, workshopID)
}

// AddWorkbenchWindow creates a new single-pane window on an existing session.
// The pane runs "orc connect" (the goblin process).
func (g *GotmuxAdapter) AddWorkbenchWindow(session interface{}, workbenchName, workbenchPath, workbenchID, workshopID string) error {
	if g.tmux != nil {
		if gotmuxSession, ok := session.(*gotmux.Session); ok {
			return g.addWorkbenchWindowGotmux(gotmuxSession, workbenchName, workbenchPath, workbenchID, workshopID)
		}
	}
	// For custom socket or string session name, use cmd-based approach
	sessionName, _ := session.(string)
	return g.addWorkbenchWindowCmd(sessionName, workbenchName, workbenchPath, workbenchID, workshopID)
}

// addWorkbenchWindowGotmux uses gotmux library (default server only).
func (g *GotmuxAdapter) addWorkbenchWindowGotmux(session *gotmux.Session, workbenchName, workbenchPath, workbenchID, workshopID string) error {
	// Create new window
	window, err := session.NewWindow(&gotmux.NewWindowOptions{
		WindowName:     workbenchName,
		StartDirectory: workbenchPath,
		DoNotAttach:    true,
	})
	if err != nil {
		return fmt.Errorf("failed to create window %s: %w", workbenchName, err)
	}

	return g.setupWorkbenchPanesGotmux(window, workbenchName, workbenchPath, workbenchID, workshopID)
}

// addWorkbenchWindowCmd uses Server.cmd() (works with any socket).
func (g *GotmuxAdapter) addWorkbenchWindowCmd(sessionName, workbenchName, workbenchPath, workbenchID, workshopID string) error {
	if err := g.server.cmd("new-window", "-t", sessionName, "-n", workbenchName, "-c", workbenchPath, "-d").Run(); err != nil {
		return fmt.Errorf("failed to create window %s: %w", workbenchName, err)
	}

	return g.setupWorkbenchPanesCmd(sessionName+":"+workbenchName, workbenchName, workbenchID, workshopID)
}

// setupWorkbenchPanesGotmux configures a window as a single-pane goblin workspace using gotmux.
func (g *GotmuxAdapter) setupWorkbenchPanesGotmux(window *gotmux.Window, workbenchName, _ string, workbenchID, workshopID string) error {
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
	if err := g.server.cmd("respawn-pane", "-t", goblinPane.Id, "-k", "orc", "connect").Run(); err != nil {
		return fmt.Errorf("failed to respawn goblin pane: %w", err)
	}

	// Set tmux pane options for identity
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

// setupWorkbenchPanesCmd configures a window as a single-pane goblin workspace using Server.cmd().
// target is in format "session:window" (e.g., "WORK-005:orc-45")
func (g *GotmuxAdapter) setupWorkbenchPanesCmd(target, workbenchName, workbenchID, workshopID string) error {
	// Rename window
	if err := g.server.cmd("rename-window", "-t", target, workbenchName).Run(); err != nil {
		return fmt.Errorf("failed to rename window: %w", err)
	}

	// Get the pane ID of the first (only) pane
	output, err := g.server.cmd("list-panes", "-t", target, "-F", "#{pane_id}").Output()
	if err != nil {
		return fmt.Errorf("failed to list panes: %w", err)
	}
	paneID := strings.TrimSpace(strings.Split(strings.TrimSpace(string(output)), "\n")[0])
	if paneID == "" {
		return fmt.Errorf("no panes found in window")
	}

	// Make "orc connect" the root process
	if err := g.server.cmd("respawn-pane", "-t", paneID, "-k", "orc", "connect").Run(); err != nil {
		return fmt.Errorf("failed to respawn goblin pane: %w", err)
	}

	// Set tmux pane options
	g.server.cmd("set-option", "-t", paneID, "-p", "@pane_role", "goblin").Run()
	g.server.cmd("set-option", "-t", paneID, "-p", "@bench_id", workbenchID).Run()
	g.server.cmd("set-option", "-t", paneID, "-p", "@workshop_id", workshopID).Run()

	return nil
}

// GetSession returns a gotmux Session by name, or nil if not found.
// For custom sockets, returns nil (callers should use string-based session names).
func (g *GotmuxAdapter) GetSession(name string) (*gotmux.Session, error) {
	if g.tmux != nil {
		sessions, _ := g.tmux.ListSessions()
		for _, s := range sessions {
			if s.Name == name {
				return s, nil
			}
		}
		return nil, nil
	}
	// Custom socket: check if session exists via cmd
	if g.server.SessionExists(name) {
		// Return nil session but no error — caller can use string-based operations
		return nil, nil
	}
	return nil, nil
}

// SessionExists checks if a tmux session exists
func (g *GotmuxAdapter) SessionExists(name string) bool {
	if g.tmux != nil {
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
	return g.server.SessionExists(name)
}

// KillSession terminates a tmux session
func (g *GotmuxAdapter) KillSession(name string) error {
	if g.tmux != nil {
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
	return g.server.KillSession(name)
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

	if g.tmux != nil {
		return g.planApplyGotmux(sessionName, workbenches, plan)
	}
	return g.planApplyCmd(sessionName, workbenches, plan)
}

// planApplyGotmux uses gotmux library for planning (default server).
func (g *GotmuxAdapter) planApplyGotmux(sessionName string, workbenches []DesiredWorkbench, plan *ApplyPlan) (*ApplyPlan, error) {
	// Check if session exists
	session, err := g.GetSession(sessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check session: %w", err)
	}

	if session == nil {
		return g.planNoSession(sessionName, workbenches, plan), nil
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

// planApplyCmd uses Server.cmd() for planning (works with any socket).
func (g *GotmuxAdapter) planApplyCmd(sessionName string, workbenches []DesiredWorkbench, plan *ApplyPlan) (*ApplyPlan, error) {
	if !g.server.SessionExists(sessionName) {
		return g.planNoSession(sessionName, workbenches, plan), nil
	}

	// Session exists — reconcile windows
	plan.SessionExists = true

	windows, err := g.server.ListWindows(sessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to list windows: %w", err)
	}

	existingWindows := make(map[string]bool, len(windows))
	for _, w := range windows {
		existingWindows[w] = true
	}

	// Check which workbenches need windows
	for _, wb := range workbenches {
		if !existingWindows[wb.Name] {
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
		paneCount := g.server.GetPaneCount(sessionName, w)
		plan.WindowSummary = append(plan.WindowSummary, WindowStatus{
			Name:      w,
			PaneCount: paneCount,
			Healthy:   paneCount > 0,
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

// planNoSession builds a plan when no session exists.
func (g *GotmuxAdapter) planNoSession(sessionName string, workbenches []DesiredWorkbench, plan *ApplyPlan) *ApplyPlan {
	plan.SessionExists = false

	if len(workbenches) == 0 {
		return plan
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

	return plan
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
		if g.tmux != nil {
			session, err := g.GetSession(action.SessionName)
			if err != nil {
				return fmt.Errorf("failed to get session: %w", err)
			}
			if session == nil {
				return fmt.Errorf("session %s not found", action.SessionName)
			}
			return g.AddWorkbenchWindow(session, action.WorkbenchName, action.WorkbenchPath, action.WorkbenchID, action.WorkshopID)
		}
		return g.AddWorkbenchWindow(action.SessionName, action.WorkbenchName, action.WorkbenchPath, action.WorkbenchID, action.WorkshopID)

	case ActionApplyEnrichment:
		g.server.ApplyGlobalBindings()
		return g.server.EnrichSession(action.SessionName)

	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// AttachInstructions returns instructions for attaching to a session
func (g *GotmuxAdapter) AttachInstructions(sessionName string) string {
	socketInfo := ""
	if g.server.Socket != "" {
		socketInfo = fmt.Sprintf(" -L %s", g.server.Socket)
	}
	return fmt.Sprintf("Attach to session: tmux%s attach -t %s\n\n"+
		"Window Layout:\n"+
		"  Each window has a single goblin pane (orc connect)\n\n"+
		"TMux Commands:\n"+
		"  Switch windows: Ctrl+b then window number (1, 2, 3...)\n"+
		"  Detach session: Ctrl+b then d\n"+
		"  Open desk: Double-click status bar or Ctrl+b then u\n",
		socketInfo, sessionName)
}

// Server returns the underlying Server for this adapter.
func (g *GotmuxAdapter) Server() *Server {
	return g.server
}
