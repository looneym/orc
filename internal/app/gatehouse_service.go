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
}

// NewGatehouseService creates a new GatehouseService with injected dependencies.
func NewGatehouseService(gatehouseRepo secondary.GatehouseRepository) *GatehouseServiceImpl {
	return &GatehouseServiceImpl{
		gatehouseRepo: gatehouseRepo,
	}
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

// Helper methods

func (s *GatehouseServiceImpl) recordToGatehouse(r *secondary.GatehouseRecord) *primary.Gatehouse {
	return &primary.Gatehouse{
		ID:         r.ID,
		WorkshopID: r.WorkshopID,
		Status:     r.Status,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

// Ensure GatehouseServiceImpl implements the interface
var _ primary.GatehouseService = (*GatehouseServiceImpl)(nil)
