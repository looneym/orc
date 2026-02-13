package primary

import "context"

// ShipmentService defines the primary port for shipment operations.
type ShipmentService interface {
	// CreateShipment creates a new shipment for a commission.
	CreateShipment(ctx context.Context, req CreateShipmentRequest) (*CreateShipmentResponse, error)

	// GetShipment retrieves a shipment by ID.
	GetShipment(ctx context.Context, shipmentID string) (*Shipment, error)

	// ListShipments lists shipments with optional filters.
	ListShipments(ctx context.Context, filters ShipmentFilters) ([]*Shipment, error)

	// CompleteShipment marks a shipment as closed.
	// If force is true, closes even if tasks are not closed.
	CompleteShipment(ctx context.Context, shipmentID string, force bool) error

	// UpdateShipment updates a shipment's title and/or description.
	UpdateShipment(ctx context.Context, req UpdateShipmentRequest) error

	// PinShipment pins a shipment to prevent closing.
	PinShipment(ctx context.Context, shipmentID string) error

	// UnpinShipment unpins a shipment.
	UnpinShipment(ctx context.Context, shipmentID string) error

	// AssignShipmentToWorkbench assigns a shipment to a workbench.
	AssignShipmentToWorkbench(ctx context.Context, shipmentID, workbenchID string) error

	// GetShipmentsByWorkbench retrieves shipments assigned to a workbench.
	GetShipmentsByWorkbench(ctx context.Context, workbenchID string) ([]*Shipment, error)

	// GetShipmentTasks retrieves all tasks for a shipment.
	GetShipmentTasks(ctx context.Context, shipmentID string) ([]*Task, error)

	// DeleteShipment deletes a shipment.
	DeleteShipment(ctx context.Context, shipmentID string) error

	// UpdateStatus sets a shipment's status directly.
	UpdateStatus(ctx context.Context, shipmentID, status string) error

	// SetStatus sets a shipment's status with escape hatch protection.
	// If force is true, allows backwards transitions.
	SetStatus(ctx context.Context, shipmentID, status string, force bool) error

	// MoveShipmentToCommission moves a shipment and its children to a different commission.
	MoveShipmentToCommission(ctx context.Context, shipmentID, targetCommissionID string) (*MoveShipmentResult, error)
}

// MoveShipmentResult contains the counts of cascaded children updated during a move.
type MoveShipmentResult struct {
	TasksUpdated int
	NotesUpdated int
	PRsUpdated   int
}

// CreateShipmentRequest contains parameters for creating a shipment.
type CreateShipmentRequest struct {
	CommissionID string
	Title        string
	Description  string
	RepoID       string // Optional - link shipment to a repository for branch ownership
	Branch       string // Optional - override auto-generated branch name
	SpecNoteID   string // Optional: NOTE-xxx - spec note that generated this shipment
}

// CreateShipmentResponse contains the result of creating a shipment.
type CreateShipmentResponse struct {
	ShipmentID string
	Shipment   *Shipment
}

// UpdateShipmentRequest contains parameters for updating a shipment.
type UpdateShipmentRequest struct {
	ShipmentID  string
	Title       string
	Description string
	Branch      string
}

// Shipment represents a shipment entity at the port boundary.
// Status lifecycle: draft -> ready -> in-progress -> closed
type Shipment struct {
	ID                  string
	CommissionID        string
	Title               string
	Description         string
	Status              string // draft, ready, in-progress, closed
	AssignedWorkbenchID string
	RepoID              string // Linked repository for branch ownership
	Branch              string // Owned branch (e.g., ml/SHIP-001-feature-name)
	Pinned              bool
	SpecNoteID          string // Spec note that generated this shipment (NOTE-xxx)
	CreatedAt           string
	UpdatedAt           string
	CompletedAt         string
}

// ShipmentFilters contains filter options for listing shipments.
type ShipmentFilters struct {
	CommissionID string
	Status       string
}
