// Package commission contains the pure business logic for commission operations.
// This is part of the Functional Core - no I/O, only pure functions.
package commission

import "time"

// CommissionStatus represents the possible states of a commission.
type CommissionStatus string

const (
	StatusActive   CommissionStatus = "active"
	StatusPaused   CommissionStatus = "paused"
	StatusComplete CommissionStatus = "complete"
	StatusArchived CommissionStatus = "archived"
)

// StatusTransitionResult contains the result of a status transition.
// This is a value object that captures both the new status and any
// side effects (like setting CompletedAt timestamp).
type StatusTransitionResult struct {
	NewStatus   CommissionStatus
	CompletedAt *time.Time // Set when transitioning to complete status
}

// ApplyStatusTransition applies a status transition and returns the result.
// This is a pure function that captures the business rule:
// - When status becomes "complete", CompletedAt should be set to the current time.
// The caller should pass the current time to enable testing.
func ApplyStatusTransition(newStatus CommissionStatus, now time.Time) StatusTransitionResult {
	result := StatusTransitionResult{
		NewStatus: newStatus,
	}

	if newStatus == StatusComplete {
		result.CompletedAt = &now
	}

	return result
}

// InitialStatus returns the initial status for a new commission.
// This is a pure function that defines the business rule for new commissions.
func InitialStatus() CommissionStatus {
	return StatusActive
}
