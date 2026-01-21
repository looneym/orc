// Package workshop contains the pure business logic for workshop operations.
// Guards are pure functions that evaluate preconditions without side effects.
package workshop

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

// CreateWorkshopContext provides context for workshop creation guards.
type CreateWorkshopContext struct {
	FactoryID     string
	FactoryExists bool
}

// CanCreateWorkshop evaluates whether a workshop can be created.
// Rules:
// - Factory must exist
func CanCreateWorkshop(ctx CreateWorkshopContext) GuardResult {
	if !ctx.FactoryExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("factory %s not found", ctx.FactoryID),
		}
	}

	return GuardResult{Allowed: true}
}

// DeleteWorkshopContext provides context for workshop deletion guards.
type DeleteWorkshopContext struct {
	WorkshopID     string
	WorkshopExists bool
	WorkbenchCount int
	ForceDelete    bool
}

// CanDeleteWorkshop evaluates whether a workshop can be deleted.
// Rules:
// - Workshop must exist
// - Workshop with workbenches requires --force
func CanDeleteWorkshop(ctx DeleteWorkshopContext) GuardResult {
	if !ctx.WorkshopExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("workshop %s not found", ctx.WorkshopID),
		}
	}

	if ctx.WorkbenchCount > 0 && !ctx.ForceDelete {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("workshop %s has %d workbenches. Use --force to delete anyway", ctx.WorkshopID, ctx.WorkbenchCount),
		}
	}

	return GuardResult{Allowed: true}
}
