package primary

import "context"

// ManifestService defines the primary port for manifest operations.
// Manifests are 1:1 with shipments.
type ManifestService interface {
	// GetManifest retrieves a manifest by ID.
	GetManifest(ctx context.Context, manifestID string) (*Manifest, error)

	// GetManifestByShipment retrieves a manifest by shipment ID.
	GetManifestByShipment(ctx context.Context, shipmentID string) (*Manifest, error)

	// ListManifests lists manifests with optional filters.
	ListManifests(ctx context.Context, filters ManifestFilters) ([]*Manifest, error)
}

// Manifest represents a manifest entity at the port boundary.
type Manifest struct {
	ID            string
	ShipmentID    string
	CreatedBy     string
	Attestation   string // May be empty
	Tasks         string // JSON array of task IDs, may be empty
	OrderingNotes string // May be empty
	Status        string // 'draft' or 'launched'
	CreatedAt     string
	UpdatedAt     string
}

// ManifestFilters contains filter options for listing manifests.
type ManifestFilters struct {
	ShipmentID string
	Status     string
}

// Manifest status constants
const (
	ManifestStatusDraft    = "draft"
	ManifestStatusLaunched = "launched"
)
