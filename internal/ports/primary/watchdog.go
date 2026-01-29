package primary

import "context"

// WatchdogService defines the primary port for watchdog operations.
// Watchdogs are 1:1 with workbenches (IMP monitors).
type WatchdogService interface {
	// GetWatchdog retrieves a watchdog by ID.
	GetWatchdog(ctx context.Context, watchdogID string) (*Watchdog, error)

	// GetWatchdogByWorkbench retrieves a watchdog by workbench ID.
	GetWatchdogByWorkbench(ctx context.Context, workbenchID string) (*Watchdog, error)

	// ListWatchdogs lists watchdogs with optional filters.
	ListWatchdogs(ctx context.Context, filters WatchdogFilters) ([]*Watchdog, error)
}

// Watchdog represents a watchdog entity at the port boundary.
type Watchdog struct {
	ID          string
	WorkbenchID string
	Status      string
	CreatedAt   string
	UpdatedAt   string
}

// WatchdogFilters contains filter options for listing watchdogs.
type WatchdogFilters struct {
	WorkbenchID string
	Status      string
}

// Watchdog status constants
const (
	WatchdogStatusActive   = "active"
	WatchdogStatusInactive = "inactive"
)
