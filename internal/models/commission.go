// Package models contains domain types for ORC entities.
// SQL persistence has been moved to internal/adapters/sqlite/*.go
package models

import (
	"database/sql"
	"time"
)

// Commission represents a commission entity.
// This is the domain type used within the models package.
// For persistence, use the repository interfaces in ports/secondary.
type Commission struct {
	ID          string
	Title       string
	Description sql.NullString
	Status      string
	Pinned      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt sql.NullTime
}

// CommissionStatus constants
const (
	CommissionStatusActive   = "active"
	CommissionStatusComplete = "complete"
	CommissionStatusArchived = "archived"
)
