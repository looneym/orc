package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// CycleServiceImpl implements the CycleService interface.
type CycleServiceImpl struct {
	cycleRepo secondary.CycleRepository
}

// NewCycleService creates a new CycleService with injected dependencies.
func NewCycleService(cycleRepo secondary.CycleRepository) *CycleServiceImpl {
	return &CycleServiceImpl{
		cycleRepo: cycleRepo,
	}
}

// CreateCycle creates a new cycle with auto-assigned sequence number.
func (s *CycleServiceImpl) CreateCycle(ctx context.Context, req primary.CreateCycleRequest) (*primary.CreateCycleResponse, error) {
	// Validate shipment exists
	exists, err := s.cycleRepo.ShipmentExists(ctx, req.ShipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate shipment: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("shipment %s not found", req.ShipmentID)
	}

	// Get next sequence number for this shipment
	seqNum, err := s.cycleRepo.GetNextSequenceNumber(ctx, req.ShipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next sequence number: %w", err)
	}

	// Get next ID
	nextID, err := s.cycleRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cycle ID: %w", err)
	}

	// Create record
	record := &secondary.CycleRecord{
		ID:             nextID,
		ShipmentID:     req.ShipmentID,
		SequenceNumber: seqNum,
		Status:         "draft",
	}

	if err := s.cycleRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create cycle: %w", err)
	}

	// Fetch created cycle
	created, err := s.cycleRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created cycle: %w", err)
	}

	return &primary.CreateCycleResponse{
		CycleID: created.ID,
		Cycle:   s.recordToCycle(created),
	}, nil
}

// GetCycle retrieves a cycle by ID.
func (s *CycleServiceImpl) GetCycle(ctx context.Context, cycleID string) (*primary.Cycle, error) {
	record, err := s.cycleRepo.GetByID(ctx, cycleID)
	if err != nil {
		return nil, err
	}
	return s.recordToCycle(record), nil
}

// ListCycles lists cycles with optional filters.
func (s *CycleServiceImpl) ListCycles(ctx context.Context, filters primary.CycleFilters) ([]*primary.Cycle, error) {
	records, err := s.cycleRepo.List(ctx, secondary.CycleFilters{
		ShipmentID: filters.ShipmentID,
		Status:     filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list cycles: %w", err)
	}

	cycles := make([]*primary.Cycle, len(records))
	for i, r := range records {
		cycles[i] = s.recordToCycle(r)
	}
	return cycles, nil
}

// DeleteCycle deletes a cycle.
func (s *CycleServiceImpl) DeleteCycle(ctx context.Context, cycleID string) error {
	return s.cycleRepo.Delete(ctx, cycleID)
}

// StartCycle transitions a cycle from queued to active.
func (s *CycleServiceImpl) StartCycle(ctx context.Context, cycleID string) error {
	// Verify cycle exists and is in queued status
	record, err := s.cycleRepo.GetByID(ctx, cycleID)
	if err != nil {
		return err
	}

	if record.Status != "queued" {
		return fmt.Errorf("cannot start cycle %s: current status is %s (must be queued)", cycleID, record.Status)
	}

	// Check if shipment already has an active cycle
	activeCycle, err := s.cycleRepo.GetActiveCycle(ctx, record.ShipmentID)
	if err != nil {
		return fmt.Errorf("failed to check for active cycle: %w", err)
	}
	if activeCycle != nil {
		return fmt.Errorf("shipment %s already has an active cycle: %s", record.ShipmentID, activeCycle.ID)
	}

	// Update status to active and set started_at
	return s.cycleRepo.UpdateStatus(ctx, cycleID, "active", true, false)
}

// CompleteCycle transitions a cycle from active to complete.
func (s *CycleServiceImpl) CompleteCycle(ctx context.Context, cycleID string) error {
	// Verify cycle exists and is in active status
	record, err := s.cycleRepo.GetByID(ctx, cycleID)
	if err != nil {
		return err
	}

	if record.Status != "active" {
		return fmt.Errorf("cannot complete cycle %s: current status is %s (must be active)", cycleID, record.Status)
	}

	// Update status to complete and set completed_at
	return s.cycleRepo.UpdateStatus(ctx, cycleID, "complete", false, true)
}

// GetActiveCycle returns the active cycle for a shipment (if any).
func (s *CycleServiceImpl) GetActiveCycle(ctx context.Context, shipmentID string) (*primary.Cycle, error) {
	record, err := s.cycleRepo.GetActiveCycle(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	return s.recordToCycle(record), nil
}

// UpdateCycleStatus updates the cycle status directly.
// Used by cascade updates from child entity state changes (CWO, Plan, CREC).
func (s *CycleServiceImpl) UpdateCycleStatus(ctx context.Context, cycleID string, status string) error {
	// Validate the status is allowed
	validStatuses := map[string]bool{
		"draft": true, "approved": true, "implementing": true,
		"review": true, "complete": true, "blocked": true, "closed": true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid cycle status: %s", status)
	}

	// Determine if we should set started_at or completed_at
	setStarted := status == "implementing"
	setCompleted := status == "complete"

	return s.cycleRepo.UpdateStatus(ctx, cycleID, status, setStarted, setCompleted)
}

// Helper methods

func (s *CycleServiceImpl) recordToCycle(r *secondary.CycleRecord) *primary.Cycle {
	return &primary.Cycle{
		ID:             r.ID,
		ShipmentID:     r.ShipmentID,
		SequenceNumber: r.SequenceNumber,
		Status:         r.Status,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
		StartedAt:      r.StartedAt,
		CompletedAt:    r.CompletedAt,
	}
}

// Ensure CycleServiceImpl implements the interface
var _ primary.CycleService = (*CycleServiceImpl)(nil)
