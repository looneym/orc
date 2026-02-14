// Package models contains domain types for ORC entities.
// SQL persistence has been moved to internal/adapters/sqlite/*.go
package models

import (
	"database/sql"
	"time"
)

// Shipment represents a shipment entity.
// This is the domain type used within the models package.
// For persistence, use the repository interfaces in ports/secondary.
type Shipment struct {
	ID                  string
	CommissionID        string
	Title               string
	Description         sql.NullString
	Status              string
	ClosedReason        sql.NullString // Why was this shipment closed (completed, abandoned, etc.)
	AssignedWorkbenchID sql.NullString
	RepoID              sql.NullString // REPO-xxx - linked repository for branch ownership
	Branch              sql.NullString // Owned branch (e.g., ml/SHIP-001-feature-name)
	Pinned              bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
	CompletedAt         sql.NullTime
}

// Shipment status constants - simplified lifecycle
// Flow: draft → ready → in-progress → closed
const (
	ShipmentStatusDraft      = "draft"
	ShipmentStatusReady      = "ready"
	ShipmentStatusInProgress = "in-progress"
	ShipmentStatusClosed     = "closed"
)
