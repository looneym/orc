// Package task contains the pure business logic for task operations.
// Guards are pure functions that evaluate preconditions without side effects.
package task

import "fmt"

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

// CreateTaskContext provides context for task creation guards.
type CreateTaskContext struct {
	CommissionID     string
	CommissionExists bool
	ShipmentID       string // optional, empty if not specified
	ShipmentExists   bool   // only checked if ShipmentID != ""
}

// CloseTaskContext provides context for task close guards.
type CloseTaskContext struct {
	TaskID   string
	IsPinned bool
}

// StatusTransitionContext provides context for status transition guards.
type StatusTransitionContext struct {
	TaskID string
	Status string // "open", "in-progress", "blocked", "closed"
}

// TagTaskContext provides context for tag operation guards.
type TagTaskContext struct {
	TaskID          string
	ExistingTagID   string // empty if no tag
	ExistingTagName string
}

// CanCreateTask evaluates whether a task can be created.
// Rules:
// - Commission must exist
// - Shipment must exist (if shipment_id provided)
func CanCreateTask(ctx CreateTaskContext) GuardResult {
	// Rule 1: Commission must exist
	if !ctx.CommissionExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("commission %s not found", ctx.CommissionID),
		}
	}

	// Rule 2: Shipment must exist (if provided)
	if ctx.ShipmentID != "" && !ctx.ShipmentExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("shipment %s not found", ctx.ShipmentID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanCloseTask evaluates whether a task can be closed.
// Rules:
// - Task must not be pinned
func CanCloseTask(ctx CloseTaskContext) GuardResult {
	if ctx.IsPinned {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cannot close pinned task %s. Unpin first with: orc task unpin %s", ctx.TaskID, ctx.TaskID),
		}
	}

	return GuardResult{Allowed: true}
}

// DeleteTaskContext provides context for task deletion guards.
type DeleteTaskContext struct {
	TaskID     string
	TaskExists bool
	Force      bool
}

// CanDeleteTask evaluates whether a task can be deleted.
// Rules:
// - --force flag required (escape hatch protection)
// - Task must exist
func CanDeleteTask(ctx DeleteTaskContext) GuardResult {
	if !ctx.Force {
		return GuardResult{
			Allowed: false,
			Reason:  "task deletion requires --force flag (this is an escape hatch)",
		}
	}

	if !ctx.TaskExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("task %s not found", ctx.TaskID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanTagTask evaluates whether a tag can be added to a task.
// Rules:
// - Task must not already have a tag (one tag per task limit)
func CanTagTask(ctx TagTaskContext) GuardResult {
	if ctx.ExistingTagID != "" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("task %s already has tag '%s' (one tag per task limit)\nRemove existing tag first with: orc task untag %s", ctx.TaskID, ctx.ExistingTagName, ctx.TaskID),
		}
	}

	return GuardResult{Allowed: true}
}
