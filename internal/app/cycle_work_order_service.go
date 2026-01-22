package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/core/cycleworkorder"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// CycleWorkOrderServiceImpl implements the CycleWorkOrderService interface.
type CycleWorkOrderServiceImpl struct {
	cwoRepo      secondary.CycleWorkOrderRepository
	cycleService primary.CycleService
}

// NewCycleWorkOrderService creates a new CycleWorkOrderService with injected dependencies.
func NewCycleWorkOrderService(cwoRepo secondary.CycleWorkOrderRepository, cycleService primary.CycleService) *CycleWorkOrderServiceImpl {
	return &CycleWorkOrderServiceImpl{
		cwoRepo:      cwoRepo,
		cycleService: cycleService,
	}
}

// CreateCycleWorkOrder creates a new cycle work order.
func (s *CycleWorkOrderServiceImpl) CreateCycleWorkOrder(ctx context.Context, req primary.CreateCycleWorkOrderRequest) (*primary.CreateCycleWorkOrderResponse, error) {
	// Validate cycle exists
	cycleExists, err := s.cwoRepo.CycleExists(ctx, req.CycleID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate cycle: %w", err)
	}

	// Check if cycle already has a CWO (1:1 relationship)
	cycleHasCWO, err := s.cwoRepo.CycleHasCWO(ctx, req.CycleID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing CWO: %w", err)
	}

	// Build guard context and evaluate
	guardCtx := cycleworkorder.CreateCWOContext{
		CycleID:     req.CycleID,
		CycleExists: cycleExists,
		CycleHasCWO: cycleHasCWO,
		Outcome:     req.Outcome,
	}

	result := cycleworkorder.CanCreateCWO(guardCtx)
	if !result.Allowed {
		return nil, fmt.Errorf("%s", result.Reason)
	}

	// Get shipment ID from cycle
	shipmentID, err := s.cwoRepo.GetCycleShipmentID(ctx, req.CycleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cycle shipment ID: %w", err)
	}

	// Get next ID
	nextID, err := s.cwoRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CWO ID: %w", err)
	}

	// Create record
	record := &secondary.CycleWorkOrderRecord{
		ID:                 nextID,
		CycleID:            req.CycleID,
		ShipmentID:         shipmentID,
		Outcome:            req.Outcome,
		AcceptanceCriteria: req.AcceptanceCriteria,
		Status:             "draft",
	}

	if err := s.cwoRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create CWO: %w", err)
	}

	// Fetch created CWO
	created, err := s.cwoRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created CWO: %w", err)
	}

	return &primary.CreateCycleWorkOrderResponse{
		CycleWorkOrderID: created.ID,
		CycleWorkOrder:   s.recordToCWO(created),
	}, nil
}

// GetCycleWorkOrder retrieves a cycle work order by ID.
func (s *CycleWorkOrderServiceImpl) GetCycleWorkOrder(ctx context.Context, cwoID string) (*primary.CycleWorkOrder, error) {
	record, err := s.cwoRepo.GetByID(ctx, cwoID)
	if err != nil {
		return nil, err
	}
	return s.recordToCWO(record), nil
}

// GetCycleWorkOrderByCycle retrieves a cycle work order by cycle ID.
func (s *CycleWorkOrderServiceImpl) GetCycleWorkOrderByCycle(ctx context.Context, cycleID string) (*primary.CycleWorkOrder, error) {
	record, err := s.cwoRepo.GetByCycle(ctx, cycleID)
	if err != nil {
		return nil, err
	}
	return s.recordToCWO(record), nil
}

// ListCycleWorkOrders lists cycle work orders with optional filters.
func (s *CycleWorkOrderServiceImpl) ListCycleWorkOrders(ctx context.Context, filters primary.CycleWorkOrderFilters) ([]*primary.CycleWorkOrder, error) {
	records, err := s.cwoRepo.List(ctx, secondary.CycleWorkOrderFilters{
		CycleID:    filters.CycleID,
		ShipmentID: filters.ShipmentID,
		Status:     filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list CWOs: %w", err)
	}

	cwos := make([]*primary.CycleWorkOrder, len(records))
	for i, r := range records {
		cwos[i] = s.recordToCWO(r)
	}
	return cwos, nil
}

// UpdateCycleWorkOrder updates a cycle work order.
func (s *CycleWorkOrderServiceImpl) UpdateCycleWorkOrder(ctx context.Context, req primary.UpdateCycleWorkOrderRequest) error {
	// Verify CWO exists and is in draft status
	record, err := s.cwoRepo.GetByID(ctx, req.CycleWorkOrderID)
	if err != nil {
		return err
	}

	if record.Status != "draft" {
		return fmt.Errorf("cannot update CWO %s: only draft CWOs can be updated (current status: %s)", req.CycleWorkOrderID, record.Status)
	}

	updateRecord := &secondary.CycleWorkOrderRecord{
		ID:                 req.CycleWorkOrderID,
		Outcome:            req.Outcome,
		AcceptanceCriteria: req.AcceptanceCriteria,
	}
	return s.cwoRepo.Update(ctx, updateRecord)
}

// DeleteCycleWorkOrder deletes a cycle work order.
func (s *CycleWorkOrderServiceImpl) DeleteCycleWorkOrder(ctx context.Context, cwoID string) error {
	return s.cwoRepo.Delete(ctx, cwoID)
}

// ApproveCycleWorkOrder transitions a cycle work order from draft to active.
// Also cascades: updates parent Cycle status to "approved".
func (s *CycleWorkOrderServiceImpl) ApproveCycleWorkOrder(ctx context.Context, cwoID string) error {
	// Get current CWO
	record, err := s.cwoRepo.GetByID(ctx, cwoID)
	if err != nil {
		return err
	}

	// Check cycle exists
	cycleExists, err := s.cwoRepo.CycleExists(ctx, record.CycleID)
	if err != nil {
		return fmt.Errorf("failed to validate cycle: %w", err)
	}

	// Build guard context and evaluate
	guardCtx := cycleworkorder.StatusTransitionContext{
		CWOID:         cwoID,
		CurrentStatus: record.Status,
		Outcome:       record.Outcome,
		CycleExists:   cycleExists,
	}

	result := cycleworkorder.CanApprove(guardCtx)
	if !result.Allowed {
		return fmt.Errorf("%s", result.Reason)
	}

	// Update CWO status to active
	if err := s.cwoRepo.UpdateStatus(ctx, cwoID, "active"); err != nil {
		return err
	}

	// CASCADE: Update parent Cycle status to "approved"
	if s.cycleService != nil {
		if err := s.cycleService.UpdateCycleStatus(ctx, record.CycleID, "approved"); err != nil {
			return fmt.Errorf("failed to cascade cycle status update: %w", err)
		}
	}

	return nil
}

// CompleteCycleWorkOrder transitions a cycle work order from active to complete.
func (s *CycleWorkOrderServiceImpl) CompleteCycleWorkOrder(ctx context.Context, cwoID string) error {
	// Get current CWO
	record, err := s.cwoRepo.GetByID(ctx, cwoID)
	if err != nil {
		return err
	}

	// Check cycle exists and get its status
	cycleExists, err := s.cwoRepo.CycleExists(ctx, record.CycleID)
	if err != nil {
		return fmt.Errorf("failed to validate cycle: %w", err)
	}

	var cycleStatus string
	if cycleExists {
		cycleStatus, err = s.cwoRepo.GetCycleStatus(ctx, record.CycleID)
		if err != nil {
			return fmt.Errorf("failed to get cycle status: %w", err)
		}
	}

	// Build guard context and evaluate
	guardCtx := cycleworkorder.StatusTransitionContext{
		CWOID:         cwoID,
		CurrentStatus: record.Status,
		Outcome:       record.Outcome,
		CycleExists:   cycleExists,
		CycleStatus:   cycleStatus,
	}

	result := cycleworkorder.CanComplete(guardCtx)
	if !result.Allowed {
		return fmt.Errorf("%s", result.Reason)
	}

	return s.cwoRepo.UpdateStatus(ctx, cwoID, "complete")
}

// Helper methods

func (s *CycleWorkOrderServiceImpl) recordToCWO(r *secondary.CycleWorkOrderRecord) *primary.CycleWorkOrder {
	return &primary.CycleWorkOrder{
		ID:                 r.ID,
		CycleID:            r.CycleID,
		ShipmentID:         r.ShipmentID,
		Outcome:            r.Outcome,
		AcceptanceCriteria: r.AcceptanceCriteria,
		Status:             r.Status,
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
	}
}

// Ensure CycleWorkOrderServiceImpl implements the interface
var _ primary.CycleWorkOrderService = (*CycleWorkOrderServiceImpl)(nil)
