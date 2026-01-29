package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/core/receipt"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// ReceiptServiceImpl implements the ReceiptService interface.
type ReceiptServiceImpl struct {
	recRepo secondary.ReceiptRepository
}

// NewReceiptService creates a new ReceiptService with injected dependencies.
func NewReceiptService(recRepo secondary.ReceiptRepository) *ReceiptServiceImpl {
	return &ReceiptServiceImpl{
		recRepo: recRepo,
	}
}

// CreateReceipt creates a new receipt.
func (s *ReceiptServiceImpl) CreateReceipt(ctx context.Context, req primary.CreateReceiptRequest) (*primary.CreateReceiptResponse, error) {
	// Validate task exists
	taskExists, err := s.recRepo.TaskExists(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate task: %w", err)
	}

	// Check if task already has a receipt (1:1 relationship)
	taskHasReceipt, err := s.recRepo.TaskHasReceipt(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing receipt: %w", err)
	}

	// Build guard context and evaluate
	guardCtx := receipt.CreateRECContext{
		TaskID:           req.TaskID,
		TaskExists:       taskExists,
		TaskHasReceipt:   taskHasReceipt,
		DeliveredOutcome: req.DeliveredOutcome,
	}

	result := receipt.CanCreateREC(guardCtx)
	if !result.Allowed {
		return nil, fmt.Errorf("%s", result.Reason)
	}

	// Get next ID
	nextID, err := s.recRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate REC ID: %w", err)
	}

	// Create record
	record := &secondary.ReceiptRecord{
		ID:                nextID,
		TaskID:            req.TaskID,
		DeliveredOutcome:  req.DeliveredOutcome,
		Evidence:          req.Evidence,
		VerificationNotes: req.VerificationNotes,
		Status:            "draft",
	}

	if err := s.recRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create REC: %w", err)
	}

	// Fetch created REC
	created, err := s.recRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created REC: %w", err)
	}

	return &primary.CreateReceiptResponse{
		ReceiptID: created.ID,
		Receipt:   s.recordToREC(created),
	}, nil
}

// GetReceipt retrieves a receipt by ID.
func (s *ReceiptServiceImpl) GetReceipt(ctx context.Context, recID string) (*primary.Receipt, error) {
	record, err := s.recRepo.GetByID(ctx, recID)
	if err != nil {
		return nil, err
	}
	return s.recordToREC(record), nil
}

// GetReceiptByTask retrieves a receipt by task ID.
func (s *ReceiptServiceImpl) GetReceiptByTask(ctx context.Context, taskID string) (*primary.Receipt, error) {
	record, err := s.recRepo.GetByTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return s.recordToREC(record), nil
}

// ListReceipts lists receipts with optional filters.
func (s *ReceiptServiceImpl) ListReceipts(ctx context.Context, filters primary.ReceiptFilters) ([]*primary.Receipt, error) {
	records, err := s.recRepo.List(ctx, secondary.ReceiptFilters{
		TaskID: filters.TaskID,
		Status: filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list RECs: %w", err)
	}

	recs := make([]*primary.Receipt, len(records))
	for i, r := range records {
		recs[i] = s.recordToREC(r)
	}
	return recs, nil
}

// UpdateReceipt updates a receipt.
func (s *ReceiptServiceImpl) UpdateReceipt(ctx context.Context, req primary.UpdateReceiptRequest) error {
	// Verify REC exists and is in draft status
	record, err := s.recRepo.GetByID(ctx, req.ReceiptID)
	if err != nil {
		return err
	}

	if record.Status != "draft" {
		return fmt.Errorf("cannot update REC %s: only draft RECs can be updated (current status: %s)", req.ReceiptID, record.Status)
	}

	updateRecord := &secondary.ReceiptRecord{
		ID:                req.ReceiptID,
		DeliveredOutcome:  req.DeliveredOutcome,
		Evidence:          req.Evidence,
		VerificationNotes: req.VerificationNotes,
	}
	return s.recRepo.Update(ctx, updateRecord)
}

// DeleteReceipt deletes a receipt.
func (s *ReceiptServiceImpl) DeleteReceipt(ctx context.Context, recID string) error {
	return s.recRepo.Delete(ctx, recID)
}

// SubmitReceipt transitions a receipt from draft to submitted.
func (s *ReceiptServiceImpl) SubmitReceipt(ctx context.Context, recID string) error {
	// Get current REC
	record, err := s.recRepo.GetByID(ctx, recID)
	if err != nil {
		return err
	}

	// Build guard context and evaluate
	guardCtx := receipt.StatusTransitionContext{
		RECID:         recID,
		CurrentStatus: record.Status,
	}

	result := receipt.CanSubmit(guardCtx)
	if !result.Allowed {
		return fmt.Errorf("%s", result.Reason)
	}

	return s.recRepo.UpdateStatus(ctx, recID, "submitted")
}

// VerifyReceipt transitions a receipt from submitted to verified.
func (s *ReceiptServiceImpl) VerifyReceipt(ctx context.Context, recID string) error {
	// Get current REC
	record, err := s.recRepo.GetByID(ctx, recID)
	if err != nil {
		return err
	}

	// Build guard context and evaluate
	guardCtx := receipt.StatusTransitionContext{
		RECID:         recID,
		CurrentStatus: record.Status,
	}

	result := receipt.CanVerify(guardCtx)
	if !result.Allowed {
		return fmt.Errorf("%s", result.Reason)
	}

	return s.recRepo.UpdateStatus(ctx, recID, "verified")
}

// Helper methods

func (s *ReceiptServiceImpl) recordToREC(r *secondary.ReceiptRecord) *primary.Receipt {
	return &primary.Receipt{
		ID:                r.ID,
		TaskID:            r.TaskID,
		DeliveredOutcome:  r.DeliveredOutcome,
		Evidence:          r.Evidence,
		VerificationNotes: r.VerificationNotes,
		Status:            r.Status,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}

// Ensure ReceiptServiceImpl implements the interface
var _ primary.ReceiptService = (*ReceiptServiceImpl)(nil)
