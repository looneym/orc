// Package cli provides thin CLI adapters that translate between CLI concerns
// and application services. Adapters handle argument parsing, output formatting,
// but delegate business logic to services.
package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/example/orc/internal/ports/primary"
)

// CommissionAdapter is a thin adapter that translates CLI operations to CommissionService calls.
// It depends only on the CommissionService interface, enabling easy testing with mocks.
type CommissionAdapter struct {
	service primary.CommissionService
	out     io.Writer
}

// NewCommissionAdapter creates a new CommissionAdapter with the given service.
func NewCommissionAdapter(service primary.CommissionService, out io.Writer) *CommissionAdapter {
	return &CommissionAdapter{
		service: service,
		out:     out,
	}
}

// Create creates a new commission.
func (a *CommissionAdapter) Create(ctx context.Context, title, description string) error {
	resp, err := a.service.CreateCommission(ctx, primary.CreateCommissionRequest{
		Title:       title,
		Description: description,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(a.out, "✓ Created commission %s: %s\n", resp.CommissionID, resp.Commission.Title)
	return nil
}

// List lists commissions with optional status filter.
func (a *CommissionAdapter) List(ctx context.Context, status string) error {
	commissions, err := a.service.ListCommissions(ctx, primary.CommissionFilters{
		Status: status,
	})
	if err != nil {
		return fmt.Errorf("failed to list commissions: %w", err)
	}

	if len(commissions) == 0 {
		fmt.Fprintln(a.out, "No commissions found")
		return nil
	}

	fmt.Fprintf(a.out, "\n%-15s %-10s %s\n", "ID", "STATUS", "TITLE")
	fmt.Fprintln(a.out, "────────────────────────────────────────────────────────────────")
	for _, c := range commissions {
		fmt.Fprintf(a.out, "%-15s %-10s %s\n", c.ID, c.Status, c.Title)
	}
	fmt.Fprintln(a.out)

	return nil
}

// Show displays details for a single commission.
// Note: Related entities (shipments, tomes) are fetched separately by the CLI layer.
func (a *CommissionAdapter) Show(ctx context.Context, commissionID string) (*primary.Commission, error) {
	commission, err := a.service.GetCommission(ctx, commissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get commission: %w", err)
	}

	fmt.Fprintf(a.out, "\nCommission: %s\n", commission.ID)
	fmt.Fprintf(a.out, "Title:   %s\n", commission.Title)
	fmt.Fprintf(a.out, "Status:  %s\n", commission.Status)
	if commission.Description != "" {
		fmt.Fprintf(a.out, "Description: %s\n", commission.Description)
	}
	fmt.Fprintf(a.out, "Created: %s\n", commission.CreatedAt)
	if commission.CompletedAt != "" {
		fmt.Fprintf(a.out, "Completed: %s\n", commission.CompletedAt)
	}
	fmt.Fprintln(a.out)

	return commission, nil
}

// Update updates a commission's title and/or description.
func (a *CommissionAdapter) Update(ctx context.Context, commissionID, title, description string) error {
	if title == "" && description == "" {
		return fmt.Errorf("must specify at least --title or --description")
	}

	err := a.service.UpdateCommission(ctx, primary.UpdateCommissionRequest{
		CommissionID: commissionID,
		Title:        title,
		Description:  description,
	})
	if err != nil {
		return fmt.Errorf("failed to update commission: %w", err)
	}

	fmt.Fprintf(a.out, "✓ Commission %s updated\n", commissionID)
	return nil
}

// Complete marks a commission as complete.
func (a *CommissionAdapter) Complete(ctx context.Context, commissionID string) error {
	err := a.service.CompleteCommission(ctx, commissionID)
	if err != nil {
		return err
	}

	fmt.Fprintf(a.out, "✓ Commission %s marked as complete\n", commissionID)
	return nil
}

// Archive archives a commission.
func (a *CommissionAdapter) Archive(ctx context.Context, commissionID string) error {
	err := a.service.ArchiveCommission(ctx, commissionID)
	if err != nil {
		return err
	}

	fmt.Fprintf(a.out, "✓ Commission %s archived\n", commissionID)
	return nil
}

// Delete deletes a commission.
func (a *CommissionAdapter) Delete(ctx context.Context, commissionID string, force bool) error {
	// Get commission details before deleting (for output)
	commission, err := a.service.GetCommission(ctx, commissionID)
	if err != nil {
		return fmt.Errorf("failed to get commission: %w", err)
	}

	err = a.service.DeleteCommission(ctx, primary.DeleteCommissionRequest{
		CommissionID: commissionID,
		Force:        force,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(a.out, "✓ Deleted commission %s: %s\n", commission.ID, commission.Title)
	return nil
}

// Pin pins a commission.
func (a *CommissionAdapter) Pin(ctx context.Context, commissionID string) error {
	err := a.service.PinCommission(ctx, commissionID)
	if err != nil {
		return fmt.Errorf("failed to pin commission: %w", err)
	}

	fmt.Fprintf(a.out, "✓ Commission %s pinned 📌\n", commissionID)
	return nil
}

// Unpin unpins a commission.
func (a *CommissionAdapter) Unpin(ctx context.Context, commissionID string) error {
	err := a.service.UnpinCommission(ctx, commissionID)
	if err != nil {
		return fmt.Errorf("failed to unpin commission: %w", err)
	}

	fmt.Fprintf(a.out, "✓ Commission %s unpinned\n", commissionID)
	return nil
}
