// Package infra contains pure business logic for infrastructure planning.
package infra

// PlanInput contains pre-fetched data for infrastructure plan generation.
// All values must be gathered by the caller - no I/O in the planner.
type PlanInput struct {
	WorkshopID   string
	WorkshopName string
	FactoryID    string
	FactoryName  string

	// Gatehouse state
	GatehouseID           string
	GatehousePath         string
	GatehousePathExists   bool
	GatehouseConfigExists bool

	// Workbench state
	Workbenches []WorkbenchPlanInput

	// Orphan state (exist on disk but not in DB)
	OrphanWorkbenches []WorkbenchPlanInput
	OrphanGatehouses  []GatehousePlanInput

	// TMux state
	TMuxSessionExists     bool              // Session found by ORC_WORKSHOP_ID
	TMuxActualSessionName string            // Actual session name (may differ after renames)
	TMuxExistingWindows   []string          // Window names currently in session
	TMuxExpectedWindows   []TMuxWindowInput // Windows that should exist (from workbenches)
}

// GatehousePlanInput contains pre-fetched data for a single gatehouse.
type GatehousePlanInput struct {
	PlaceID string // From config.json
	Path    string
}

// WorkbenchPlanInput contains pre-fetched data for a single workbench.
type WorkbenchPlanInput struct {
	ID             string
	Name           string
	WorktreePath   string
	RepoName       string
	HomeBranch     string
	WorktreeExists bool
	ConfigExists   bool
}

// TMuxWindowInput contains pre-fetched data for an expected tmux window.
type TMuxWindowInput struct {
	Name string // Window name (usually workbench name)
	Path string // Working directory for the window
}

// Plan describes infrastructure state for a workshop.
type Plan struct {
	WorkshopID   string
	WorkshopName string
	FactoryID    string
	FactoryName  string

	Gatehouse   *GatehouseOp
	Workbenches []WorkbenchOp

	// Orphans (exist on disk but not in DB)
	OrphanWorkbenches []WorkbenchOp
	OrphanGatehouses  []GatehouseOp

	// TMux state
	TMuxSession *TMuxSessionOp
}

// GatehouseOp describes gatehouse infrastructure state.
type GatehouseOp struct {
	ID           string
	Path         string
	Exists       bool
	ConfigExists bool
}

// WorkbenchOp describes workbench infrastructure state.
type WorkbenchOp struct {
	ID           string
	Name         string
	Path         string
	Exists       bool
	ConfigExists bool
	RepoName     string
	Branch       string
}

// TMuxSessionOp describes tmux session infrastructure state.
type TMuxSessionOp struct {
	SessionName   string
	Exists        bool
	Windows       []TMuxWindowOp
	OrphanWindows []TMuxWindowOp // Windows that exist but shouldn't (workbench deleted/archived)
}

// TMuxWindowOp describes tmux window infrastructure state.
type TMuxWindowOp struct {
	Name   string
	Path   string
	Exists bool
}

// GeneratePlan creates an infrastructure plan.
// This is a pure function - all input data must be pre-fetched.
func GeneratePlan(input PlanInput) Plan {
	plan := Plan{
		WorkshopID:   input.WorkshopID,
		WorkshopName: input.WorkshopName,
		FactoryID:    input.FactoryID,
		FactoryName:  input.FactoryName,
	}

	// Gatehouse
	plan.Gatehouse = &GatehouseOp{
		ID:           input.GatehouseID,
		Path:         input.GatehousePath,
		Exists:       input.GatehousePathExists,
		ConfigExists: input.GatehouseConfigExists,
	}

	// Workbenches
	for _, wb := range input.Workbenches {
		plan.Workbenches = append(plan.Workbenches, WorkbenchOp{
			ID:           wb.ID,
			Name:         wb.Name,
			Path:         wb.WorktreePath,
			Exists:       wb.WorktreeExists,
			ConfigExists: wb.ConfigExists,
			RepoName:     wb.RepoName,
			Branch:       wb.HomeBranch,
		})
	}

	// Orphan workbenches (exist on disk but not in DB)
	for _, wb := range input.OrphanWorkbenches {
		plan.OrphanWorkbenches = append(plan.OrphanWorkbenches, WorkbenchOp{
			ID:           wb.ID,
			Name:         wb.Name,
			Path:         wb.WorktreePath,
			Exists:       true, // By definition, orphans exist on disk
			ConfigExists: true,
		})
	}

	// Orphan gatehouses
	for _, gh := range input.OrphanGatehouses {
		plan.OrphanGatehouses = append(plan.OrphanGatehouses, GatehouseOp{
			ID:           gh.PlaceID,
			Path:         gh.Path,
			Exists:       true,
			ConfigExists: true,
		})
	}

	// TMux session state
	plan.TMuxSession = buildTMuxSessionOp(input)

	return plan
}

// buildTMuxSessionOp creates the TMux session operation plan.
func buildTMuxSessionOp(input PlanInput) *TMuxSessionOp {
	sessionOp := &TMuxSessionOp{
		SessionName: input.TMuxActualSessionName,
		Exists:      input.TMuxSessionExists,
	}

	// Build set of existing windows for O(1) lookup
	existingSet := make(map[string]bool)
	for _, w := range input.TMuxExistingWindows {
		existingSet[w] = true
	}

	// Build set of expected window names for orphan detection
	expectedSet := make(map[string]bool)
	for _, w := range input.TMuxExpectedWindows {
		expectedSet[w.Name] = true
	}

	// Check each expected window
	for _, expected := range input.TMuxExpectedWindows {
		sessionOp.Windows = append(sessionOp.Windows, TMuxWindowOp{
			Name:   expected.Name,
			Path:   expected.Path,
			Exists: existingSet[expected.Name],
		})
	}

	// Find orphan windows (exist but not expected)
	for _, windowName := range input.TMuxExistingWindows {
		if !expectedSet[windowName] {
			sessionOp.OrphanWindows = append(sessionOp.OrphanWindows, TMuxWindowOp{
				Name:   windowName,
				Exists: true,
			})
		}
	}

	return sessionOp
}
