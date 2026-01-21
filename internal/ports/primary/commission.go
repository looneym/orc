// Package primary defines the primary ports (driving adapters) for the application.
// These are the interfaces through which the outside world drives the application.
package primary

import "context"

// CommissionService defines the primary port for commission operations.
// This interface documents the intended contract for commission management.
// Implementations will live in the application layer, adapters in CLI/API layers.
type CommissionService interface {
	// CreateCommission creates a new commission with the given parameters.
	CreateCommission(ctx context.Context, req CreateCommissionRequest) (*CreateCommissionResponse, error)

	// StartCommission begins execution of a commission (activates it).
	StartCommission(ctx context.Context, req StartCommissionRequest) (*StartCommissionResponse, error)

	// LaunchCommission creates and immediately starts a commission.
	LaunchCommission(ctx context.Context, req LaunchCommissionRequest) (*LaunchCommissionResponse, error)

	// GetCommission retrieves a commission by ID.
	GetCommission(ctx context.Context, commissionID string) (*Commission, error)

	// ListCommissions lists commissions with optional filters.
	ListCommissions(ctx context.Context, filters CommissionFilters) ([]*Commission, error)

	// CompleteCommission marks a commission as complete.
	CompleteCommission(ctx context.Context, commissionID string) error

	// ArchiveCommission archives a completed commission.
	ArchiveCommission(ctx context.Context, commissionID string) error

	// UpdateCommission updates commission title and/or description.
	UpdateCommission(ctx context.Context, req UpdateCommissionRequest) error

	// DeleteCommission deletes a commission.
	DeleteCommission(ctx context.Context, req DeleteCommissionRequest) error

	// PinCommission pins a commission to prevent completion/archival.
	PinCommission(ctx context.Context, commissionID string) error

	// UnpinCommission unpins a commission.
	UnpinCommission(ctx context.Context, commissionID string) error
}

// CreateCommissionRequest contains parameters for creating a commission.
type CreateCommissionRequest struct {
	Title       string
	Description string
}

// CreateCommissionResponse contains the result of creating a commission.
type CreateCommissionResponse struct {
	CommissionID string
	Commission   *Commission
}

// StartCommissionRequest contains parameters for starting a commission.
type StartCommissionRequest struct {
	CommissionID string
}

// StartCommissionResponse contains the result of starting a commission.
type StartCommissionResponse struct {
	Commission *Commission
}

// LaunchCommissionRequest contains parameters for launching a commission.
type LaunchCommissionRequest struct {
	Title       string
	Description string
}

// LaunchCommissionResponse contains the result of launching a commission.
type LaunchCommissionResponse struct {
	CommissionID string
	Commission   *Commission
}

// Commission represents a commission entity at the port boundary.
type Commission struct {
	ID          string
	Title       string
	Description string
	Status      string
	CreatedAt   string
	StartedAt   string
	CompletedAt string
}

// CommissionFilters contains filter options for listing commissions.
type CommissionFilters struct {
	Status string
	Limit  int
}

// UpdateCommissionRequest contains parameters for updating a commission.
type UpdateCommissionRequest struct {
	CommissionID string
	Title        string
	Description  string
}

// DeleteCommissionRequest contains parameters for deleting a commission.
type DeleteCommissionRequest struct {
	CommissionID string
	Force        bool
}
