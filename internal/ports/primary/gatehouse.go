package primary

import "context"

// GatehouseService defines the primary port for gatehouse operations.
// Gatehouses are 1:1 with workshops (Goblin seat).
type GatehouseService interface {
	// GetGatehouse retrieves a gatehouse by ID.
	GetGatehouse(ctx context.Context, gatehouseID string) (*Gatehouse, error)

	// GetGatehouseByWorkshop retrieves a gatehouse by workshop ID.
	GetGatehouseByWorkshop(ctx context.Context, workshopID string) (*Gatehouse, error)

	// ListGatehouses lists gatehouses with optional filters.
	ListGatehouses(ctx context.Context, filters GatehouseFilters) ([]*Gatehouse, error)
}

// Gatehouse represents a gatehouse entity at the port boundary.
type Gatehouse struct {
	ID         string
	WorkshopID string
	Status     string
	CreatedAt  string
	UpdatedAt  string
}

// GatehouseFilters contains filter options for listing gatehouses.
type GatehouseFilters struct {
	WorkshopID string
	Status     string
}

// Gatehouse status constants
const (
	GatehouseStatusActive = "active"
)
