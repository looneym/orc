package primary

import "context"

// WorkbenchService defines the primary port for workbench operations.
// A Workbench is a git worktree - replaces the Grove concept.
type WorkbenchService interface {
	// CreateWorkbench creates a new workbench in a workshop.
	CreateWorkbench(ctx context.Context, req CreateWorkbenchRequest) (*CreateWorkbenchResponse, error)

	// OpenWorkbench opens a workbench in a TMux window with IMP layout.
	OpenWorkbench(ctx context.Context, req OpenWorkbenchRequest) (*OpenWorkbenchResponse, error)

	// GetWorkbench retrieves a workbench by ID.
	GetWorkbench(ctx context.Context, workbenchID string) (*Workbench, error)

	// GetWorkbenchByPath retrieves a workbench by its filesystem path.
	GetWorkbenchByPath(ctx context.Context, path string) (*Workbench, error)

	// ListWorkbenches lists workbenches with optional filters.
	ListWorkbenches(ctx context.Context, filters WorkbenchFilters) ([]*Workbench, error)

	// RenameWorkbench renames a workbench.
	RenameWorkbench(ctx context.Context, req RenameWorkbenchRequest) error

	// UpdateWorkbenchPath updates the filesystem path of a workbench.
	UpdateWorkbenchPath(ctx context.Context, workbenchID, newPath string) error

	// DeleteWorkbench deletes a workbench.
	DeleteWorkbench(ctx context.Context, req DeleteWorkbenchRequest) error
}

// CreateWorkbenchRequest contains parameters for creating a workbench.
type CreateWorkbenchRequest struct {
	Name       string
	WorkshopID string   // Required
	RepoID     string   // Optional - link to repo
	Repos      []string // Optional repository names for worktree creation
	BasePath   string   // Optional, defaults to ~/src/worktrees
}

// CreateWorkbenchResponse contains the result of workbench creation.
type CreateWorkbenchResponse struct {
	WorkbenchID string
	Workbench   *Workbench
	Path        string // Materialized workbench path
}

// OpenWorkbenchRequest contains parameters for opening a workbench.
type OpenWorkbenchRequest struct {
	WorkbenchID string
}

// OpenWorkbenchResponse contains the result of opening a workbench.
type OpenWorkbenchResponse struct {
	Workbench   *Workbench
	SessionName string
	WindowName  string
}

// RenameWorkbenchRequest contains parameters for renaming a workbench.
type RenameWorkbenchRequest struct {
	WorkbenchID  string
	NewName      string
	UpdateConfig bool // Also update .orc/config.json
}

// DeleteWorkbenchRequest contains parameters for deleting a workbench.
type DeleteWorkbenchRequest struct {
	WorkbenchID    string
	Force          bool
	RemoveWorktree bool // Also remove filesystem worktree
}

// Workbench represents a workbench entity at the port boundary.
// A Workbench is a git worktree - replaces Grove.
type Workbench struct {
	ID         string
	Name       string
	WorkshopID string
	RepoID     string
	Path       string
	Status     string
	CreatedAt  string
	UpdatedAt  string
}

// WorkbenchFilters contains filter options for listing workbenches.
type WorkbenchFilters struct {
	WorkshopID string
	RepoID     string
	Status     string
	Limit      int
}
