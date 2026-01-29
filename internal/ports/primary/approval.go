package primary

import "context"

// ApprovalService defines the primary port for approval operations.
// Approvals are 1:1 with plans.
type ApprovalService interface {
	// GetApproval retrieves an approval by ID.
	GetApproval(ctx context.Context, approvalID string) (*Approval, error)

	// GetApprovalByPlan retrieves an approval by plan ID.
	GetApprovalByPlan(ctx context.Context, planID string) (*Approval, error)

	// ListApprovals lists approvals with optional filters.
	ListApprovals(ctx context.Context, filters ApprovalFilters) ([]*Approval, error)
}

// Approval represents an approval entity at the port boundary.
type Approval struct {
	ID             string
	PlanID         string
	TaskID         string
	Mechanism      string // 'subagent' or 'manual'
	ReviewerInput  string
	ReviewerOutput string
	Outcome        string // 'approved' or 'escalated'
	CreatedAt      string
}

// ApprovalFilters contains filter options for listing approvals.
type ApprovalFilters struct {
	TaskID  string
	Outcome string
}

// Approval outcome constants
const (
	ApprovalOutcomeApproved  = "approved"
	ApprovalOutcomeEscalated = "escalated"
)

// Approval mechanism constants
const (
	ApprovalMechanismSubagent = "subagent"
	ApprovalMechanismManual   = "manual"
)
