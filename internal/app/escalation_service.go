package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// EscalationServiceImpl implements the EscalationService interface.
type EscalationServiceImpl struct {
	escalationRepo secondary.EscalationRepository
}

// NewEscalationService creates a new EscalationService with injected dependencies.
func NewEscalationService(escalationRepo secondary.EscalationRepository) *EscalationServiceImpl {
	return &EscalationServiceImpl{
		escalationRepo: escalationRepo,
	}
}

// GetEscalation retrieves an escalation by ID.
func (s *EscalationServiceImpl) GetEscalation(ctx context.Context, escalationID string) (*primary.Escalation, error) {
	record, err := s.escalationRepo.GetByID(ctx, escalationID)
	if err != nil {
		return nil, err
	}
	return s.recordToEscalation(record), nil
}

// ListEscalations lists escalations with optional filters.
func (s *EscalationServiceImpl) ListEscalations(ctx context.Context, filters primary.EscalationFilters) ([]*primary.Escalation, error) {
	records, err := s.escalationRepo.List(ctx, secondary.EscalationFilters{
		Status:        filters.Status,
		TargetActorID: filters.TargetActorID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list escalations: %w", err)
	}

	escalations := make([]*primary.Escalation, len(records))
	for i, r := range records {
		escalations[i] = s.recordToEscalation(r)
	}
	return escalations, nil
}

// Helper methods

func (s *EscalationServiceImpl) recordToEscalation(r *secondary.EscalationRecord) *primary.Escalation {
	return &primary.Escalation{
		ID:            r.ID,
		ApprovalID:    r.ApprovalID,
		PlanID:        r.PlanID,
		TaskID:        r.TaskID,
		Reason:        r.Reason,
		Status:        r.Status,
		RoutingRule:   r.RoutingRule,
		OriginActorID: r.OriginActorID,
		TargetActorID: r.TargetActorID,
		Resolution:    r.Resolution,
		ResolvedBy:    r.ResolvedBy,
		CreatedAt:     r.CreatedAt,
		ResolvedAt:    r.ResolvedAt,
	}
}

// Ensure EscalationServiceImpl implements the interface
var _ primary.EscalationService = (*EscalationServiceImpl)(nil)
