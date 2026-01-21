package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	coreworkbench "github.com/example/orc/internal/core/workbench"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// WorkbenchServiceImpl implements the WorkbenchService interface.
type WorkbenchServiceImpl struct {
	workbenchRepo secondary.WorkbenchRepository
	workshopRepo  secondary.WorkshopRepository
	agentProvider secondary.AgentIdentityProvider
	executor      EffectExecutor
}

// NewWorkbenchService creates a new WorkbenchService with injected dependencies.
func NewWorkbenchService(
	workbenchRepo secondary.WorkbenchRepository,
	workshopRepo secondary.WorkshopRepository,
	agentProvider secondary.AgentIdentityProvider,
	executor EffectExecutor,
) *WorkbenchServiceImpl {
	return &WorkbenchServiceImpl{
		workbenchRepo: workbenchRepo,
		workshopRepo:  workshopRepo,
		agentProvider: agentProvider,
		executor:      executor,
	}
}

// CreateWorkbench creates a new workbench.
func (s *WorkbenchServiceImpl) CreateWorkbench(ctx context.Context, req primary.CreateWorkbenchRequest) (*primary.CreateWorkbenchResponse, error) {
	// 1. Get agent identity
	identity, err := s.agentProvider.GetCurrentIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent identity: %w", err)
	}

	// 2. Check if workshop exists
	workshopExists, err := s.workbenchRepo.WorkshopExists(ctx, req.WorkshopID)
	if err != nil {
		return nil, fmt.Errorf("failed to check workshop: %w", err)
	}

	// 3. Guard check
	guardCtx := coreworkbench.CreateWorkbenchContext{
		GuardContext: coreworkbench.GuardContext{
			AgentType:  coreworkbench.AgentType(identity.Type),
			AgentID:    identity.FullID,
			WorkshopID: req.WorkshopID,
		},
		WorkshopExists: workshopExists,
	}
	if result := coreworkbench.CanCreateWorkbench(guardCtx); !result.Allowed {
		return nil, result.Error()
	}

	// 4. Build workbench path
	basePath := req.BasePath
	if basePath == "" {
		basePath = s.defaultBasePath()
	}

	// Get workshop for path naming
	workshop, err := s.workshopRepo.GetByID(ctx, req.WorkshopID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workshop: %w", err)
	}

	workbenchPath := filepath.Join(basePath, fmt.Sprintf("%s-%s", workshop.FactoryID, req.Name))

	// 5. Create workbench record in DB
	record := &secondary.WorkbenchRecord{
		Name:         req.Name,
		WorkshopID:   req.WorkshopID,
		RepoID:       req.RepoID,
		WorktreePath: workbenchPath,
		Status:       "active",
	}
	if err := s.workbenchRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create workbench: %w", err)
	}

	return &primary.CreateWorkbenchResponse{
		WorkbenchID: record.ID,
		Workbench:   s.recordToWorkbench(record),
		Path:        workbenchPath,
	}, nil
}

// OpenWorkbench opens a workbench in TMux.
func (s *WorkbenchServiceImpl) OpenWorkbench(ctx context.Context, req primary.OpenWorkbenchRequest) (*primary.OpenWorkbenchResponse, error) {
	// 1. Fetch workbench
	workbench, err := s.workbenchRepo.GetByID(ctx, req.WorkbenchID)
	if err != nil {
		return nil, fmt.Errorf("workbench not found: %w", err)
	}

	// 2. Check path exists
	pathExists := s.pathExists(workbench.WorktreePath)

	// 3. Check TMux session (via environment)
	inTMux := os.Getenv("TMUX") != ""

	// 4. Guard check
	guardCtx := coreworkbench.OpenWorkbenchContext{
		WorkbenchID:     req.WorkbenchID,
		WorkbenchExists: true,
		PathExists:      pathExists,
		InTMuxSession:   inTMux,
	}
	if result := coreworkbench.CanOpenWorkbench(guardCtx); !result.Allowed {
		return nil, result.Error()
	}

	// 5. Get TMux session info
	sessionName := s.getTMuxSession()

	return &primary.OpenWorkbenchResponse{
		Workbench:   s.recordToWorkbench(workbench),
		SessionName: sessionName,
		WindowName:  workbench.Name,
	}, nil
}

// GetWorkbench retrieves a workbench by ID.
func (s *WorkbenchServiceImpl) GetWorkbench(ctx context.Context, workbenchID string) (*primary.Workbench, error) {
	record, err := s.workbenchRepo.GetByID(ctx, workbenchID)
	if err != nil {
		return nil, fmt.Errorf("workbench not found: %w", err)
	}
	return s.recordToWorkbench(record), nil
}

// GetWorkbenchByPath retrieves a workbench by its filesystem path.
func (s *WorkbenchServiceImpl) GetWorkbenchByPath(ctx context.Context, path string) (*primary.Workbench, error) {
	record, err := s.workbenchRepo.GetByPath(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("workbench not found at path: %w", err)
	}
	return s.recordToWorkbench(record), nil
}

// UpdateWorkbenchPath updates the filesystem path of a workbench.
func (s *WorkbenchServiceImpl) UpdateWorkbenchPath(ctx context.Context, workbenchID, newPath string) error {
	// Verify workbench exists
	_, err := s.workbenchRepo.GetByID(ctx, workbenchID)
	if err != nil {
		return fmt.Errorf("workbench not found: %w", err)
	}

	return s.workbenchRepo.UpdatePath(ctx, workbenchID, newPath)
}

// ListWorkbenches lists workbenches with optional filters.
func (s *WorkbenchServiceImpl) ListWorkbenches(ctx context.Context, filters primary.WorkbenchFilters) ([]*primary.Workbench, error) {
	records, err := s.workbenchRepo.List(ctx, filters.WorkshopID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workbenches: %w", err)
	}

	workbenches := make([]*primary.Workbench, len(records))
	for i, r := range records {
		workbenches[i] = s.recordToWorkbench(r)
	}
	return workbenches, nil
}

// RenameWorkbench renames a workbench.
func (s *WorkbenchServiceImpl) RenameWorkbench(ctx context.Context, req primary.RenameWorkbenchRequest) error {
	// 1. Check workbench exists
	_, err := s.workbenchRepo.GetByID(ctx, req.WorkbenchID)
	workbenchExists := err == nil

	// 2. Guard check
	if result := coreworkbench.CanRenameWorkbench(workbenchExists, req.WorkbenchID); !result.Allowed {
		return result.Error()
	}

	// 3. Update name
	return s.workbenchRepo.Rename(ctx, req.WorkbenchID, req.NewName)
}

// DeleteWorkbench deletes a workbench.
func (s *WorkbenchServiceImpl) DeleteWorkbench(ctx context.Context, req primary.DeleteWorkbenchRequest) error {
	// 1. Fetch workbench
	_, err := s.workbenchRepo.GetByID(ctx, req.WorkbenchID)
	if err != nil {
		return fmt.Errorf("workbench not found: %w", err)
	}

	// 2. Count active work (simplified - could add task repo)
	activeTaskCount := 0

	// 3. Guard check
	guardCtx := coreworkbench.DeleteWorkbenchContext{
		WorkbenchID:     req.WorkbenchID,
		ActiveTaskCount: activeTaskCount,
		ForceDelete:     req.Force,
	}
	if result := coreworkbench.CanDeleteWorkbench(guardCtx); !result.Allowed {
		return result.Error()
	}

	// 4. Delete from database
	return s.workbenchRepo.Delete(ctx, req.WorkbenchID)
}

// Helper methods

func (s *WorkbenchServiceImpl) recordToWorkbench(r *secondary.WorkbenchRecord) *primary.Workbench {
	return &primary.Workbench{
		ID:         r.ID,
		Name:       r.Name,
		WorkshopID: r.WorkshopID,
		RepoID:     r.RepoID,
		Path:       r.WorktreePath,
		Status:     r.Status,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

func (s *WorkbenchServiceImpl) defaultBasePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "src", "worktrees")
}

func (s *WorkbenchServiceImpl) pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (s *WorkbenchServiceImpl) getTMuxSession() string {
	// In production, would parse TMUX env var or run tmux display-message
	return "orc"
}

// Ensure WorkbenchServiceImpl implements the interface
var _ primary.WorkbenchService = (*WorkbenchServiceImpl)(nil)
