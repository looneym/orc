package primary

import "context"

// WorkshopService defines the primary port for workshop operations.
// A Workshop is a persistent place within a Factory, hosting Workbenches.
type WorkshopService interface {
	// CreateWorkshop creates a new workshop in a factory.
	CreateWorkshop(ctx context.Context, req CreateWorkshopRequest) (*CreateWorkshopResponse, error)

	// GetWorkshop retrieves a workshop by ID.
	GetWorkshop(ctx context.Context, workshopID string) (*Workshop, error)

	// ListWorkshops lists workshops with optional filters.
	ListWorkshops(ctx context.Context, filters WorkshopFilters) ([]*Workshop, error)

	// UpdateWorkshop updates workshop name.
	UpdateWorkshop(ctx context.Context, req UpdateWorkshopRequest) error

	// DeleteWorkshop deletes a workshop.
	DeleteWorkshop(ctx context.Context, req DeleteWorkshopRequest) error

	// OpenWorkshop launches a TMux session for the workshop.
	OpenWorkshop(ctx context.Context, req OpenWorkshopRequest) (*OpenWorkshopResponse, error)

	// CloseWorkshop kills the workshop's TMux session.
	CloseWorkshop(ctx context.Context, workshopID string) error
}

// CreateWorkshopRequest contains parameters for creating a workshop.
type CreateWorkshopRequest struct {
	FactoryID string
	Name      string // Optional - will use name pool if empty
}

// CreateWorkshopResponse contains the result of creating a workshop.
type CreateWorkshopResponse struct {
	WorkshopID string
	Workshop   *Workshop
}

// Workshop represents a workshop entity at the port boundary.
// A Workshop is a persistent place within a Factory.
type Workshop struct {
	ID        string
	FactoryID string
	Name      string
	Status    string
	CreatedAt string
	UpdatedAt string
}

// WorkshopFilters contains filter options for listing workshops.
type WorkshopFilters struct {
	FactoryID string
	Status    string
	Limit     int
}

// UpdateWorkshopRequest contains parameters for updating a workshop.
type UpdateWorkshopRequest struct {
	WorkshopID string
	Name       string
}

// DeleteWorkshopRequest contains parameters for deleting a workshop.
type DeleteWorkshopRequest struct {
	WorkshopID string
	Force      bool
}

// OpenWorkshopRequest contains parameters for opening a workshop TMux session.
type OpenWorkshopRequest struct {
	WorkshopID string
}

// OpenWorkshopResponse contains the result of opening a workshop.
type OpenWorkshopResponse struct {
	Workshop           *Workshop
	SessionName        string
	SessionAlreadyOpen bool
	AttachInstructions string
}
