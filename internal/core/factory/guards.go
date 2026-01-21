// Package factory contains the pure business logic for factory operations.
// Guards are pure functions that evaluate preconditions without side effects.
package factory

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

// CreateFactoryContext provides context for factory creation guards.
type CreateFactoryContext struct {
	Name       string
	NameExists bool
}

// CanCreateFactory evaluates whether a factory can be created.
// Rules:
// - Name must be unique
func CanCreateFactory(ctx CreateFactoryContext) GuardResult {
	if ctx.NameExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("factory with name %q already exists", ctx.Name),
		}
	}

	return GuardResult{Allowed: true}
}

// DeleteFactoryContext provides context for factory deletion guards.
type DeleteFactoryContext struct {
	FactoryID       string
	FactoryExists   bool
	WorkshopCount   int
	CommissionCount int
	ForceDelete     bool
}

// CanDeleteFactory evaluates whether a factory can be deleted.
// Rules:
// - Factory must exist
// - Factory with workshops or commissions requires --force
func CanDeleteFactory(ctx DeleteFactoryContext) GuardResult {
	if !ctx.FactoryExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("factory %s not found", ctx.FactoryID),
		}
	}

	if (ctx.WorkshopCount > 0 || ctx.CommissionCount > 0) && !ctx.ForceDelete {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("factory %s has %d workshops and %d commissions. Use --force to delete anyway", ctx.FactoryID, ctx.WorkshopCount, ctx.CommissionCount),
		}
	}

	return GuardResult{Allowed: true}
}
