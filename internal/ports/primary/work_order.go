package primary

import "context"

// WorkOrderService defines the primary port for workOrder operations.
type WorkOrderService interface {
	// CreateWorkOrder creates a new workOrder.
	CreateWorkOrder(ctx context.Context, req CreateWorkOrderRequest) (*CreateWorkOrderResponse, error)

	// GetWorkOrder retrieves a workOrder by ID.
	GetWorkOrder(ctx context.Context, workOrderID string) (*WorkOrder, error)

	// GetWorkOrderByShipment retrieves a workOrder by shipment ID.
	GetWorkOrderByShipment(ctx context.Context, shipmentID string) (*WorkOrder, error)

	// ListWorkOrders lists work_orders with optional filters.
	ListWorkOrders(ctx context.Context, filters WorkOrderFilters) ([]*WorkOrder, error)

	// UpdateWorkOrder updates a workOrder.
	UpdateWorkOrder(ctx context.Context, req UpdateWorkOrderRequest) error

	// DeleteWorkOrder deletes a workOrder.
	DeleteWorkOrder(ctx context.Context, workOrderID string) error

	// ActivateWorkOrder transitions a work order from draft to active.
	ActivateWorkOrder(ctx context.Context, workOrderID string) error

	// CompleteWorkOrder transitions a work order from active to complete.
	CompleteWorkOrder(ctx context.Context, workOrderID string) error
}

// CreateWorkOrderRequest contains parameters for creating a workOrder.
type CreateWorkOrderRequest struct {
	ShipmentID         string
	Outcome            string
	AcceptanceCriteria string // Optional JSON array
}

// CreateWorkOrderResponse contains the result of creating a workOrder.
type CreateWorkOrderResponse struct {
	WorkOrderID string
	WorkOrder   *WorkOrder
}

// UpdateWorkOrderRequest contains parameters for updating a workOrder.
type UpdateWorkOrderRequest struct {
	WorkOrderID        string
	Outcome            string
	AcceptanceCriteria string
}

// WorkOrder represents a workOrder entity at the port boundary.
type WorkOrder struct {
	ID                 string
	ShipmentID         string
	Outcome            string
	AcceptanceCriteria string // JSON array
	Status             string
	CreatedAt          string
	UpdatedAt          string
}

// WorkOrderFilters contains filter options for listing work_orders.
type WorkOrderFilters struct {
	ShipmentID string
	Status     string
}

// WorkOrder status constants
const (
	WorkOrderStatusDraft    = "draft"
	WorkOrderStatusActive   = "active"
	WorkOrderStatusComplete = "complete"
)
