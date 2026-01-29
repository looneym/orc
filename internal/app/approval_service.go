package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// ApprovalServiceImpl implements the ApprovalService interface.
type ApprovalServiceImpl struct {
	approvalRepo secondary.ApprovalRepository
}

// NewApprovalService creates a new ApprovalService with injected dependencies.
func NewApprovalService(approvalRepo secondary.ApprovalRepository) *ApprovalServiceImpl {
	return &ApprovalServiceImpl{
		approvalRepo: approvalRepo,
	}
}

// GetApproval retrieves an approval by ID.
func (s *ApprovalServiceImpl) GetApproval(ctx context.Context, approvalID string) (*primary.Approval, error) {
	record, err := s.approvalRepo.GetByID(ctx, approvalID)
	if err != nil {
		return nil, err
	}
	return s.recordToApproval(record), nil
}

// GetApprovalByPlan retrieves an approval by plan ID.
func (s *ApprovalServiceImpl) GetApprovalByPlan(ctx context.Context, planID string) (*primary.Approval, error) {
	record, err := s.approvalRepo.GetByPlan(ctx, planID)
	if err != nil {
		return nil, err
	}
	return s.recordToApproval(record), nil
}

// ListApprovals lists approvals with optional filters.
func (s *ApprovalServiceImpl) ListApprovals(ctx context.Context, filters primary.ApprovalFilters) ([]*primary.Approval, error) {
	records, err := s.approvalRepo.List(ctx, secondary.ApprovalFilters{
		TaskID:  filters.TaskID,
		Outcome: filters.Outcome,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list approvals: %w", err)
	}

	approvals := make([]*primary.Approval, len(records))
	for i, r := range records {
		approvals[i] = s.recordToApproval(r)
	}
	return approvals, nil
}

// Helper methods

func (s *ApprovalServiceImpl) recordToApproval(r *secondary.ApprovalRecord) *primary.Approval {
	return &primary.Approval{
		ID:             r.ID,
		PlanID:         r.PlanID,
		TaskID:         r.TaskID,
		Mechanism:      r.Mechanism,
		ReviewerInput:  r.ReviewerInput,
		ReviewerOutput: r.ReviewerOutput,
		Outcome:        r.Outcome,
		CreatedAt:      r.CreatedAt,
	}
}

// Ensure ApprovalServiceImpl implements the interface
var _ primary.ApprovalService = (*ApprovalServiceImpl)(nil)
