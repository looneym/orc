// Package cycleworkorder contains the pure business logic for cycle work order operations.
// Guards are pure functions that evaluate preconditions without side effects.
package cycleworkorder

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

// CreateCWOContext provides context for CWO creation guards.
type CreateCWOContext struct {
	CycleID     string
	CycleExists bool
	CycleHasCWO bool
	Outcome     string
	ShipmentID  string
}

// StatusTransitionContext provides context for status transition guards.
type StatusTransitionContext struct {
	CWOID         string
	CurrentStatus string
	Outcome       string
	CycleExists   bool
	CycleStatus   string
}

// CanCreateCWO evaluates whether a CWO can be created.
// Rules:
// - Cycle must exist
// - Cycle must not already have a CWO (1:1 constraint)
// - Outcome must not be empty
func CanCreateCWO(ctx CreateCWOContext) GuardResult {
	// Rule 1: Cycle must exist
	if !ctx.CycleExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cycle %s not found", ctx.CycleID),
		}
	}

	// Rule 2: Cycle must not already have a CWO
	if ctx.CycleHasCWO {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cycle %s already has a CWO", ctx.CycleID),
		}
	}

	// Rule 3: Outcome must not be empty
	if strings.TrimSpace(ctx.Outcome) == "" {
		return GuardResult{
			Allowed: false,
			Reason:  "outcome cannot be empty",
		}
	}

	return GuardResult{Allowed: true}
}

// CanApprove evaluates whether a CWO can be approved.
// Rules:
// - CWO must be in draft status
// - Outcome must not be empty
// - Cycle must exist
func CanApprove(ctx StatusTransitionContext) GuardResult {
	// Rule 1: Must be in draft status
	if ctx.CurrentStatus != "draft" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only approve draft CWOs (current status: %s)", ctx.CurrentStatus),
		}
	}

	// Rule 2: Outcome must not be empty
	if strings.TrimSpace(ctx.Outcome) == "" {
		return GuardResult{
			Allowed: false,
			Reason:  "cannot approve CWO: outcome is empty",
		}
	}

	// Rule 3: Cycle must exist
	if !ctx.CycleExists {
		return GuardResult{
			Allowed: false,
			Reason:  "cannot approve CWO: parent cycle no longer exists",
		}
	}

	return GuardResult{Allowed: true}
}

// CanComplete evaluates whether a CWO can be completed.
// Rules:
// - CWO must be in active status
// - Parent cycle must exist
func CanComplete(ctx StatusTransitionContext) GuardResult {
	// Rule 1: Must be in active status
	if ctx.CurrentStatus != "active" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only complete active CWOs (current status: %s)", ctx.CurrentStatus),
		}
	}

	// Rule 2: Cycle must exist
	if !ctx.CycleExists {
		return GuardResult{
			Allowed: false,
			Reason:  "cannot complete CWO: parent cycle no longer exists",
		}
	}

	return GuardResult{Allowed: true}
}
