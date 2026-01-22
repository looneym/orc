package primary

import "context"

// CycleWorkOrderService defines the primary port for cycle work order operations.
type CycleWorkOrderService interface {
	// CreateCycleWorkOrder creates a new cycle work order.
	CreateCycleWorkOrder(ctx context.Context, req CreateCycleWorkOrderRequest) (*CreateCycleWorkOrderResponse, error)

	// GetCycleWorkOrder retrieves a cycle work order by ID.
	GetCycleWorkOrder(ctx context.Context, cwoID string) (*CycleWorkOrder, error)

	// GetCycleWorkOrderByCycle retrieves a cycle work order by cycle ID.
	GetCycleWorkOrderByCycle(ctx context.Context, cycleID string) (*CycleWorkOrder, error)

	// ListCycleWorkOrders lists cycle work orders with optional filters.
	ListCycleWorkOrders(ctx context.Context, filters CycleWorkOrderFilters) ([]*CycleWorkOrder, error)

	// UpdateCycleWorkOrder updates a cycle work order.
	UpdateCycleWorkOrder(ctx context.Context, req UpdateCycleWorkOrderRequest) error

	// DeleteCycleWorkOrder deletes a cycle work order.
	DeleteCycleWorkOrder(ctx context.Context, cwoID string) error

	// ActivateCycleWorkOrder transitions a cycle work order from draft to active.
	ActivateCycleWorkOrder(ctx context.Context, cwoID string) error

	// CompleteCycleWorkOrder transitions a cycle work order from active to complete.
	CompleteCycleWorkOrder(ctx context.Context, cwoID string) error
}

// CreateCycleWorkOrderRequest contains parameters for creating a cycle work order.
type CreateCycleWorkOrderRequest struct {
	CycleID            string
	Outcome            string
	AcceptanceCriteria string // Optional JSON array
}

// CreateCycleWorkOrderResponse contains the result of creating a cycle work order.
type CreateCycleWorkOrderResponse struct {
	CycleWorkOrderID string
	CycleWorkOrder   *CycleWorkOrder
}

// UpdateCycleWorkOrderRequest contains parameters for updating a cycle work order.
type UpdateCycleWorkOrderRequest struct {
	CycleWorkOrderID   string
	Outcome            string
	AcceptanceCriteria string
}

// CycleWorkOrder represents a cycle work order entity at the port boundary.
type CycleWorkOrder struct {
	ID                 string
	CycleID            string
	ShipmentID         string
	Outcome            string
	AcceptanceCriteria string // JSON array
	Status             string
	CreatedAt          string
	UpdatedAt          string
}

// CycleWorkOrderFilters contains filter options for listing cycle work orders.
type CycleWorkOrderFilters struct {
	CycleID    string
	ShipmentID string
	Status     string
}

// CycleWorkOrder status constants
const (
	CycleWorkOrderStatusDraft    = "draft"
	CycleWorkOrderStatusActive   = "active"
	CycleWorkOrderStatusComplete = "complete"
)
