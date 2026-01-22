package primary

import "context"

// CycleService defines the primary port for cycle operations.
type CycleService interface {
	// CreateCycle creates a new cycle (sequence number auto-assigned).
	CreateCycle(ctx context.Context, req CreateCycleRequest) (*CreateCycleResponse, error)

	// GetCycle retrieves a cycle by ID.
	GetCycle(ctx context.Context, cycleID string) (*Cycle, error)

	// ListCycles lists cycles with optional filters.
	ListCycles(ctx context.Context, filters CycleFilters) ([]*Cycle, error)

	// DeleteCycle deletes a cycle.
	DeleteCycle(ctx context.Context, cycleID string) error

	// StartCycle transitions a cycle from queued to active.
	StartCycle(ctx context.Context, cycleID string) error

	// CompleteCycle transitions a cycle from active to complete.
	CompleteCycle(ctx context.Context, cycleID string) error

	// GetActiveCycle returns the active cycle for a shipment (if any).
	GetActiveCycle(ctx context.Context, shipmentID string) (*Cycle, error)
}

// CreateCycleRequest contains parameters for creating a cycle.
type CreateCycleRequest struct {
	ShipmentID string
	// SequenceNumber is auto-assigned based on existing cycles for the shipment
}

// CreateCycleResponse contains the result of creating a cycle.
type CreateCycleResponse struct {
	CycleID string
	Cycle   *Cycle
}

// Cycle represents a cycle entity at the port boundary.
type Cycle struct {
	ID             string
	ShipmentID     string
	SequenceNumber int64
	Status         string
	CreatedAt      string
	UpdatedAt      string
	StartedAt      string // Empty string means null
	CompletedAt    string // Empty string means null
}

// CycleFilters contains filter options for listing cycles.
type CycleFilters struct {
	ShipmentID string
	Status     string
}

// Cycle status constants
const (
	CycleStatusQueued   = "queued"
	CycleStatusActive   = "active"
	CycleStatusComplete = "complete"
)
