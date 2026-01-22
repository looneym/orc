package models

import (
	"database/sql"
	"time"
)

type Tome struct {
	ID                  string
	ComcommissionID     string
	Title               string
	Description         sql.NullString
	Status              string
	AssignedWorkbenchID sql.NullString
	Pinned              bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
	CompletedAt         sql.NullTime
}

// Tome status constants
const (
	TomeStatusActive   = "active"
	TomeStatusPaused   = "paused"
	TomeStatusComplete = "complete"
)
