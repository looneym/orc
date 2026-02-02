package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// GatehouseServiceImpl implements the GatehouseService interface.
type GatehouseServiceImpl struct {
	gatehouseRepo secondary.GatehouseRepository
	workshopRepo  secondary.WorkshopRepository
}

// NewGatehouseService creates a new GatehouseService with injected dependencies.
func NewGatehouseService(gatehouseRepo secondary.GatehouseRepository, workshopRepo ...secondary.WorkshopRepository) *GatehouseServiceImpl {
	service := &GatehouseServiceImpl{
		gatehouseRepo: gatehouseRepo,
	}
	// Optional workshop repo for EnsureAllWorkshopsHaveGatehouses
	if len(workshopRepo) > 0 {
		service.workshopRepo = workshopRepo[0]
	}
	return service
}

// GetGatehouse retrieves a gatehouse by ID.
func (s *GatehouseServiceImpl) GetGatehouse(ctx context.Context, gatehouseID string) (*primary.Gatehouse, error) {
	record, err := s.gatehouseRepo.GetByID(ctx, gatehouseID)
	if err != nil {
		return nil, err
	}
	return s.recordToGatehouse(record), nil
}

// GetGatehouseByWorkshop retrieves a gatehouse by workshop ID.
func (s *GatehouseServiceImpl) GetGatehouseByWorkshop(ctx context.Context, workshopID string) (*primary.Gatehouse, error) {
	record, err := s.gatehouseRepo.GetByWorkshop(ctx, workshopID)
	if err != nil {
		return nil, err
	}
	return s.recordToGatehouse(record), nil
}

// ListGatehouses lists gatehouses with optional filters.
func (s *GatehouseServiceImpl) ListGatehouses(ctx context.Context, filters primary.GatehouseFilters) ([]*primary.Gatehouse, error) {
	records, err := s.gatehouseRepo.List(ctx, secondary.GatehouseFilters{
		WorkshopID: filters.WorkshopID,
		Status:     filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list gatehouses: %w", err)
	}

	gatehouses := make([]*primary.Gatehouse, len(records))
	for i, r := range records {
		gatehouses[i] = s.recordToGatehouse(r)
	}
	return gatehouses, nil
}

// CreateGatehouse creates a new gatehouse for a workshop.
// Returns error if workshop already has a gatehouse.
func (s *GatehouseServiceImpl) CreateGatehouse(ctx context.Context, workshopID string) (*primary.Gatehouse, error) {
	// Check workshop exists
	exists, err := s.gatehouseRepo.WorkshopExists(ctx, workshopID)
	if err != nil {
		return nil, fmt.Errorf("failed to check workshop exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("workshop %s not found", workshopID)
	}

	// Check no existing gatehouse
	hasGatehouse, err := s.gatehouseRepo.WorkshopHasGatehouse(ctx, workshopID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing gatehouse: %w", err)
	}
	if hasGatehouse {
		return nil, fmt.Errorf("workshop %s already has a gatehouse", workshopID)
	}

	// Generate next ID
	id, err := s.gatehouseRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate gatehouse ID: %w", err)
	}

	// Create in DB
	record := &secondary.GatehouseRecord{
		ID:         id,
		WorkshopID: workshopID,
		Status:     "active",
	}
	if err := s.gatehouseRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create gatehouse: %w", err)
	}

	return s.recordToGatehouse(record), nil
}

// EnsureAllWorkshopsHaveGatehouses creates gatehouses for any workshops missing them.
// Returns a list of newly created gatehouse IDs.
func (s *GatehouseServiceImpl) EnsureAllWorkshopsHaveGatehouses(ctx context.Context) ([]string, error) {
	if s.workshopRepo == nil {
		return nil, fmt.Errorf("workshop repository not available")
	}

	workshops, err := s.workshopRepo.List(ctx, secondary.WorkshopFilters{})
	if err != nil {
		return nil, fmt.Errorf("failed to list workshops: %w", err)
	}

	var created []string
	for _, ws := range workshops {
		hasGatehouse, err := s.gatehouseRepo.WorkshopHasGatehouse(ctx, ws.ID)
		if err != nil {
			return created, fmt.Errorf("failed to check gatehouse for %s: %w", ws.ID, err)
		}
		if hasGatehouse {
			continue
		}

		// Create gatehouse
		gatehouse, err := s.CreateGatehouse(ctx, ws.ID)
		if err != nil {
			return created, fmt.Errorf("failed to create gatehouse for %s: %w", ws.ID, err)
		}
		created = append(created, gatehouse.ID)
	}

	return created, nil
}

// Helper methods

func (s *GatehouseServiceImpl) recordToGatehouse(r *secondary.GatehouseRecord) *primary.Gatehouse {
	return &primary.Gatehouse{
		ID:         r.ID,
		WorkshopID: r.WorkshopID,
		Status:     r.Status,
		FocusedID:  r.FocusedID,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

// UpdateFocusedID sets or clears the focused container ID for a gatehouse.
func (s *GatehouseServiceImpl) UpdateFocusedID(ctx context.Context, gatehouseID, focusedID string) error {
	return s.gatehouseRepo.UpdateFocusedID(ctx, gatehouseID, focusedID)
}

// GetFocusedID returns the currently focused container ID for a gatehouse.
func (s *GatehouseServiceImpl) GetFocusedID(ctx context.Context, gatehouseID string) (string, error) {
	record, err := s.gatehouseRepo.GetByID(ctx, gatehouseID)
	if err != nil {
		return "", err
	}
	return record.FocusedID, nil
}

// Ensure GatehouseServiceImpl implements the interface
var _ primary.GatehouseService = (*GatehouseServiceImpl)(nil)
