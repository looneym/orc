package models

import (
	"database/sql"
	"time"
)

type WorkOrder struct {
	ID                 string
	ShipmentID         string
	Outcome            string
	AcceptanceCriteria sql.NullString
	Status             string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// WorkOrder status constants
const (
	WorkOrderStatusDraft    = "draft"
	WorkOrderStatusActive   = "active"
	WorkOrderStatusComplete = "complete"
)
