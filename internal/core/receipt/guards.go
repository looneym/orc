// Package receipt contains the pure business logic for receipt operations.
// Guards are pure functions that evaluate preconditions without side effects.
package receipt

import (
	"fmt"
	"strings"
)

// GuardResult represents the outcome of a guard evaluation.
type GuardResult struct {
	Allowed bool
	Reason  string
}

// Error converts the guard result to an error if not allowed.
func (r GuardResult) Error() error {
	if r.Allowed {
		return nil
	}
	return fmt.Errorf("%s", r.Reason)
}

// CreateRECContext provides context for REC creation guards.
type CreateRECContext struct {
	TaskID           string
	TaskExists       bool
	TaskHasReceipt   bool
	DeliveredOutcome string
}

// StatusTransitionContext provides context for status transition guards.
type StatusTransitionContext struct {
	RECID         string
	CurrentStatus string
}

// CanCreateREC evaluates whether a REC can be created.
// Rules:
// - Task must exist
// - Task must not already have a REC (1:1 constraint)
// - DeliveredOutcome must not be empty
func CanCreateREC(ctx CreateRECContext) GuardResult {
	// Rule 1: Task must exist
	if !ctx.TaskExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("task %s not found", ctx.TaskID),
		}
	}

	// Rule 2: Task must not already have a REC
	if ctx.TaskHasReceipt {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("task %s already has a receipt", ctx.TaskID),
		}
	}

	// Rule 3: DeliveredOutcome must not be empty
	if strings.TrimSpace(ctx.DeliveredOutcome) == "" {
		return GuardResult{
			Allowed: false,
			Reason:  "delivered outcome cannot be empty",
		}
	}

	return GuardResult{Allowed: true}
}

// CanSubmit evaluates whether a REC can be submitted.
// Rules:
// - REC must be in draft status
func CanSubmit(ctx StatusTransitionContext) GuardResult {
	// Rule 1: Must be in draft status
	if ctx.CurrentStatus != "draft" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only submit draft RECs (current status: %s)", ctx.CurrentStatus),
		}
	}

	return GuardResult{Allowed: true}
}

// CanVerify evaluates whether a REC can be verified.
// Rules:
// - REC must be in submitted status
func CanVerify(ctx StatusTransitionContext) GuardResult {
	// Rule 1: Must be in submitted status
	if ctx.CurrentStatus != "submitted" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only verify submitted RECs (current status: %s)", ctx.CurrentStatus),
		}
	}

	return GuardResult{Allowed: true}
}
