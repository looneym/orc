package primary

import "context"

// EventService defines the primary port for unified event operations.
type EventService interface {
	// ListEvents retrieves events matching the given filters.
	// Queries both audit and operational event repos, merge-sorted by timestamp descending.
	ListEvents(ctx context.Context, filters EventFilters) ([]*Event, error)

	// GetEvent retrieves a single event by ID.
	GetEvent(ctx context.Context, id string) (*Event, error)

	// PruneEvents deletes events older than the specified number of days from both repos.
	// Returns the combined count of deleted entries.
	PruneEvents(ctx context.Context, olderThanDays int) (int, error)
}

// Event represents a unified event at the port boundary.
// Audit events have empty Level/Message/Data fields.
// Operational events have empty EntityType/EntityID/Action/FieldName/OldValue/NewValue fields.
type Event struct {
	ID         string
	WorkshopID string
	Timestamp  string
	ActorID    string
	Source     string // Event source (e.g., "ledger", "hook", "system")
	Version    string // Schema version

	// Audit event fields (zero values for operational events)
	EntityType string
	EntityID   string
	Action     string // 'create', 'update', 'delete'
	FieldName  string // For updates only
	OldValue   string
	NewValue   string

	// Operational event fields (zero values for audit events)
	Level   string // "debug", "info", "warn", "error"
	Message string
	Data    string // JSON payload

	CreatedAt string
}

// EventFilters contains filter options for querying events.
type EventFilters struct {
	WorkshopID string
	Source     string
	Level      string
	EventType  string // "audit", "ops", "all" (default: "all")
	ActorID    string
	EntityID   string
	Limit      int
}
