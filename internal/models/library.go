package models

import "time"

// Library represents a library entity (one per commission, auto-created).
// Libraries hold parked/reference tomes that aren't in active conclaves.
type Library struct {
	ID           string
	CommissionID string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
