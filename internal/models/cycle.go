package models

import (
	"database/sql"
	"time"
)

type Cycle struct {
	ID             string
	ShipmentID     string
	SequenceNumber int64
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	StartedAt      sql.NullTime
	CompletedAt    sql.NullTime
}

// Cycle status constants
const (
	CycleStatusDraft        = "draft"
	CycleStatusApproved     = "approved"
	CycleStatusImplementing = "implementing"
	CycleStatusReview       = "review"
	CycleStatusComplete     = "complete"
	CycleStatusBlocked      = "blocked"
	CycleStatusClosed       = "closed"
	CycleStatusFailed       = "failed"
)
