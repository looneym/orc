package app

import (
	"context"
	"fmt"

	coreworkshop "github.com/example/orc/internal/core/workshop"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// WorkshopServiceImpl implements the WorkshopService interface.
type WorkshopServiceImpl struct {
	workshopRepo secondary.WorkshopRepository
}

// NewWorkshopService creates a new WorkshopService with injected dependencies.
func NewWorkshopService(workshopRepo secondary.WorkshopRepository) *WorkshopServiceImpl {
	return &WorkshopServiceImpl{
		workshopRepo: workshopRepo,
	}
}

// CreateWorkshop creates a new workshop in a factory.
func (s *WorkshopServiceImpl) CreateWorkshop(ctx context.Context, req primary.CreateWorkshopRequest) (*primary.CreateWorkshopResponse, error) {
	// 1. Check if factory exists
	factoryExists, err := s.workshopRepo.FactoryExists(ctx, req.FactoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to check factory: %w", err)
	}

	// 2. Guard check
	guardCtx := coreworkshop.CreateWorkshopContext{
		FactoryID:     req.FactoryID,
		FactoryExists: factoryExists,
	}
	if result := coreworkshop.CanCreateWorkshop(guardCtx); !result.Allowed {
		return nil, result.Error()
	}

	// 3. Create workshop record (ID and name generation handled by repo)
	record := &secondary.WorkshopRecord{
		FactoryID: req.FactoryID,
		Name:      req.Name, // May be empty - repo will use name pool
		Status:    "active",
	}
	if err := s.workshopRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create workshop: %w", err)
	}

	return &primary.CreateWorkshopResponse{
		WorkshopID: record.ID,
		Workshop:   s.recordToWorkshop(record),
	}, nil
}

// GetWorkshop retrieves a workshop by ID.
func (s *WorkshopServiceImpl) GetWorkshop(ctx context.Context, workshopID string) (*primary.Workshop, error) {
	record, err := s.workshopRepo.GetByID(ctx, workshopID)
	if err != nil {
		return nil, fmt.Errorf("workshop not found: %w", err)
	}
	return s.recordToWorkshop(record), nil
}

// ListWorkshops lists workshops with optional filters.
func (s *WorkshopServiceImpl) ListWorkshops(ctx context.Context, filters primary.WorkshopFilters) ([]*primary.Workshop, error) {
	records, err := s.workshopRepo.List(ctx, secondary.WorkshopFilters{
		FactoryID: filters.FactoryID,
		Status:    filters.Status,
		Limit:     filters.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list workshops: %w", err)
	}

	workshops := make([]*primary.Workshop, len(records))
	for i, r := range records {
		workshops[i] = s.recordToWorkshop(r)
	}
	return workshops, nil
}

// UpdateWorkshop updates workshop name.
func (s *WorkshopServiceImpl) UpdateWorkshop(ctx context.Context, req primary.UpdateWorkshopRequest) error {
	// 1. Check workshop exists
	_, err := s.workshopRepo.GetByID(ctx, req.WorkshopID)
	if err != nil {
		return fmt.Errorf("workshop not found: %w", err)
	}

	// 2. Update
	record := &secondary.WorkshopRecord{
		ID:   req.WorkshopID,
		Name: req.Name,
	}
	if err := s.workshopRepo.Update(ctx, record); err != nil {
		return fmt.Errorf("failed to update workshop: %w", err)
	}

	return nil
}

// DeleteWorkshop deletes a workshop.
func (s *WorkshopServiceImpl) DeleteWorkshop(ctx context.Context, req primary.DeleteWorkshopRequest) error {
	// 1. Check workshop exists
	_, err := s.workshopRepo.GetByID(ctx, req.WorkshopID)
	workshopExists := err == nil

	// 2. Count workbenches
	workbenchCount, _ := s.workshopRepo.CountWorkbenches(ctx, req.WorkshopID)

	// 3. Guard check
	guardCtx := coreworkshop.DeleteWorkshopContext{
		WorkshopID:     req.WorkshopID,
		WorkshopExists: workshopExists,
		WorkbenchCount: workbenchCount,
		ForceDelete:    req.Force,
	}
	if result := coreworkshop.CanDeleteWorkshop(guardCtx); !result.Allowed {
		return result.Error()
	}

	// 4. Delete
	return s.workshopRepo.Delete(ctx, req.WorkshopID)
}

// Helper methods

func (s *WorkshopServiceImpl) recordToWorkshop(r *secondary.WorkshopRecord) *primary.Workshop {
	return &primary.Workshop{
		ID:        r.ID,
		FactoryID: r.FactoryID,
		Name:      r.Name,
		Status:    r.Status,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// Ensure WorkshopServiceImpl implements the interface
var _ primary.WorkshopService = (*WorkshopServiceImpl)(nil)
