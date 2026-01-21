package primary

import "context"

// FactoryService defines the primary port for factory operations.
// A Factory is a TMux session - the persistent runtime environment.
type FactoryService interface {
	// CreateFactory creates a new factory.
	CreateFactory(ctx context.Context, req CreateFactoryRequest) (*CreateFactoryResponse, error)

	// GetFactory retrieves a factory by ID.
	GetFactory(ctx context.Context, factoryID string) (*Factory, error)

	// GetFactoryByName retrieves a factory by its unique name.
	GetFactoryByName(ctx context.Context, name string) (*Factory, error)

	// ListFactories lists factories with optional filters.
	ListFactories(ctx context.Context, filters FactoryFilters) ([]*Factory, error)

	// UpdateFactory updates factory name.
	UpdateFactory(ctx context.Context, req UpdateFactoryRequest) error

	// DeleteFactory deletes a factory.
	DeleteFactory(ctx context.Context, req DeleteFactoryRequest) error
}

// CreateFactoryRequest contains parameters for creating a factory.
type CreateFactoryRequest struct {
	Name string
}

// CreateFactoryResponse contains the result of creating a factory.
type CreateFactoryResponse struct {
	FactoryID string
	Factory   *Factory
}

// Factory represents a factory entity at the port boundary.
// A Factory is a TMux session - the persistent runtime environment.
type Factory struct {
	ID        string
	Name      string
	Status    string
	CreatedAt string
	UpdatedAt string
}

// FactoryFilters contains filter options for listing factories.
type FactoryFilters struct {
	Status string
	Limit  int
}

// UpdateFactoryRequest contains parameters for updating a factory.
type UpdateFactoryRequest struct {
	FactoryID string
	Name      string
}

// DeleteFactoryRequest contains parameters for deleting a factory.
type DeleteFactoryRequest struct {
	FactoryID string
	Force     bool
}
