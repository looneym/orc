package models

import "time"

// Shipyard represents a shipyard entity (one per commission, auto-created).
// Shipyards hold parked/future shipments that aren't in active conclaves.
type Shipyard struct {
	ID           string
	CommissionID string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
