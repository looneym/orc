// Package shipment contains the pure business logic for shipment operations.
// Guards are pure functions that evaluate preconditions without side effects.
package shipment

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

// CreateShipmentContext provides context for shipment creation guards.
type CreateShipmentContext struct {
	CommissionID     string
	CommissionExists bool
}

// TaskSummary contains minimal task info for guard evaluation.
type TaskSummary struct {
	ID     string
	Status string
}

// CloseShipmentContext provides context for shipment close guards.
type CloseShipmentContext struct {
	ShipmentID      string
	IsPinned        bool
	Tasks           []TaskSummary
	ForceCompletion bool // Skip task check if explicitly forced
}

// StatusTransitionContext provides context for status transition guards.
type StatusTransitionContext struct {
	ShipmentID    string
	Status        string // "draft", "ready", "in-progress", "closed"
	OpenTaskCount int    // count of non-closed tasks
}

// AssignWorkbenchContext provides context for workbench assignment guards.
type AssignWorkbenchContext struct {
	ShipmentID            string
	WorkbenchID           string
	ShipmentExists        bool
	WorkbenchAssignedToID string // ID of shipment workbench is assigned to, empty if unassigned
}

// CanCreateShipment evaluates whether a shipment can be created.
// Rules:
// - Commission must exist
func CanCreateShipment(ctx CreateShipmentContext) GuardResult {
	if !ctx.CommissionExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("commission %s not found", ctx.CommissionID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanCloseShipment evaluates whether a shipment can be closed.
// Rules:
// - Shipment must not be pinned
// - All tasks must be closed (unless forced)
func CanCloseShipment(ctx CloseShipmentContext) GuardResult {
	if ctx.IsPinned {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cannot close pinned shipment %s. Unpin first with: orc shipment unpin %s", ctx.ShipmentID, ctx.ShipmentID),
		}
	}

	// Check for non-closed tasks (unless force flag is set)
	if !ctx.ForceCompletion {
		var incomplete []string
		for _, t := range ctx.Tasks {
			if t.Status != "closed" {
				incomplete = append(incomplete, t.ID)
			}
		}
		if len(incomplete) > 0 {
			return GuardResult{
				Allowed: false,
				Reason: fmt.Sprintf("cannot close shipment: %d task(s) not closed (%s). Use --force to close anyway",
					len(incomplete), strings.Join(incomplete, ", ")),
			}
		}
	}

	return GuardResult{Allowed: true}
}

// OverrideStatusContext provides context for status override guards.
type OverrideStatusContext struct {
	ShipmentID    string
	CurrentStatus string
	NewStatus     string
	Force         bool
}

// statusOrder defines the progression order for shipment statuses.
// Lower index = earlier in lifecycle.
var statusOrder = map[string]int{
	"draft":       0,
	"ready":       1,
	"in-progress": 2,
	"closed":      3,
}

// ValidStatuses returns all valid shipment statuses.
func ValidStatuses() []string {
	return []string{"draft", "ready", "in-progress", "closed"}
}

// CanOverrideStatus evaluates whether a shipment status can be overridden.
// Rules:
// - New status must be valid
// - Backwards transitions require --force flag
func CanOverrideStatus(ctx OverrideStatusContext) GuardResult {
	// Rule 1: New status must be valid
	if _, ok := statusOrder[ctx.NewStatus]; !ok {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("invalid status '%s'. Valid statuses: %s", ctx.NewStatus, strings.Join(ValidStatuses(), ", ")),
		}
	}

	// Rule 2: Check for backwards transition
	currentIdx, currentOk := statusOrder[ctx.CurrentStatus]
	newIdx := statusOrder[ctx.NewStatus]

	// If current status is unknown, allow transition
	if !currentOk {
		return GuardResult{Allowed: true}
	}

	// Backwards transition requires force
	if newIdx < currentIdx && !ctx.Force {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("backwards transition from '%s' to '%s' requires --force flag", ctx.CurrentStatus, ctx.NewStatus),
		}
	}

	return GuardResult{Allowed: true}
}

// CanAssignWorkbench evaluates whether a workbench can be assigned to a shipment.
// Rules:
// - Shipment must exist
// - Workbench must not be assigned to another shipment
func CanAssignWorkbench(ctx AssignWorkbenchContext) GuardResult {
	// Rule 1: Shipment must exist
	if !ctx.ShipmentExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("shipment %s not found", ctx.ShipmentID),
		}
	}

	// Rule 2: Workbench must not be assigned to another shipment
	// If the workbench is assigned to this same shipment, that's OK (idempotent)
	if ctx.WorkbenchAssignedToID != "" && ctx.WorkbenchAssignedToID != ctx.ShipmentID {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("workbench already assigned to shipment %s", ctx.WorkbenchAssignedToID),
		}
	}

	return GuardResult{Allowed: true}
}
