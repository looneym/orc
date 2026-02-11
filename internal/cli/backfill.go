package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/db"
)

// BackfillCmd returns the backfill command
func BackfillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backfill",
		Short: "Data migration and backfill commands",
		Long:  `One-time data migration commands. Safe to run multiple times (idempotent).`,
	}

	cmd.AddCommand(backfillLibraryTomesCmd())
	cmd.AddCommand(backfillLifecycleStatusesCmd())

	return cmd
}

func backfillLibraryTomesCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "library-tomes",
		Short: "Migrate library tomes to commission root (orphan tomes)",
		Long: `Convert tomes with container_type='library' to orphan tomes at commission root.

This migration:
- Clears container_id and container_type for library tomes
- Library tomes become orphan tomes (visible in ROOT TOMES section)
- Safe to run multiple times (idempotent)

Run this after removing the Library entity from the schema.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.GetDB()
			if err != nil {
				return fmt.Errorf("failed to get database: %w", err)
			}

			// Count library tomes
			var count int
			err = database.QueryRow(`SELECT COUNT(*) FROM tomes WHERE container_type = 'library'`).Scan(&count)
			if err != nil {
				return fmt.Errorf("failed to count library tomes: %w", err)
			}

			if count == 0 {
				fmt.Println("No library tomes found. Migration already complete or no library tomes exist.")
				return nil
			}

			fmt.Printf("Found %d library tomes to migrate...\n", count)

			if dryRun {
				fmt.Println("\n[DRY RUN] No changes made. Run without --dry-run to apply migration.")
				return nil
			}

			// Migrate library tomes to commission root (orphan tomes)
			result, err := database.Exec(`
				UPDATE tomes
				SET container_id = NULL,
				    container_type = NULL,
				    conclave_id = NULL,
				    updated_at = CURRENT_TIMESTAMP
				WHERE container_type = 'library'
			`)
			if err != nil {
				return fmt.Errorf("failed to migrate library tomes: %w", err)
			}

			rowsAffected, _ := result.RowsAffected()
			fmt.Printf("\n✓ Migration complete: %d tomes migrated to commission root\n", rowsAffected)
			fmt.Println("  Library tomes now appear in ROOT TOMES section of orc summary")
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be migrated without making changes")

	return cmd
}

func backfillLifecycleStatusesCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "lifecycle-statuses",
		Short: "Migrate shipment and task statuses to simplified lifecycle",
		Long: `Migrate shipment and task statuses to the new simplified lifecycle.

Shipment status mapping:
  draft                         → draft
  exploring, synthesizing,
    specced, planned            → draft
  tasked, ready_for_imp         → ready
  implementing, auto_implementing → in-progress
  implemented, deployed,
    verified, complete          → closed

Task status mapping:
  ready                         → open
  in_progress                   → in-progress
  paused                        → open
  blocked                       → blocked
  complete                      → closed

Safe to run multiple times (idempotent).
Must run BEFORE the schema change that constrains statuses.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.GetDB()
			if err != nil {
				return fmt.Errorf("failed to get database: %w", err)
			}

			// --- Shipment status migration ---
			shipmentMappings := []struct {
				oldStatuses []string
				newStatus   string
			}{
				// draft stays draft (no-op, but included for reporting)
				{[]string{"exploring", "synthesizing", "specced", "planned"}, "draft"},
				{[]string{"tasked", "ready_for_imp"}, "ready"},
				{[]string{"implementing", "auto_implementing"}, "in-progress"},
				{[]string{"implemented", "deployed", "verified", "complete"}, "closed"},
			}

			fmt.Println("=== Shipment Status Migration ===")
			totalShipments := 0
			for _, m := range shipmentMappings {
				for _, old := range m.oldStatuses {
					var count int
					err := database.QueryRow(`SELECT COUNT(*) FROM shipments WHERE status = ?`, old).Scan(&count)
					if err != nil {
						return fmt.Errorf("failed to count shipments with status %q: %w", old, err)
					}
					if count > 0 {
						fmt.Printf("  %s → %s: %d shipments\n", old, m.newStatus, count)
						totalShipments += count
					}
				}
			}

			if totalShipments == 0 {
				fmt.Println("  No shipments need migration.")
			}

			// --- Task status migration ---
			taskMappings := []struct {
				oldStatus string
				newStatus string
			}{
				{"ready", "open"},
				// in_progress → in-progress (underscore to hyphen)
				{"in_progress", "in-progress"},
				{"paused", "open"},
				// blocked stays blocked (no-op)
				{"complete", "closed"},
			}

			fmt.Println("\n=== Task Status Migration ===")
			totalTasks := 0
			for _, m := range taskMappings {
				var count int
				err := database.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = ?`, m.oldStatus).Scan(&count)
				if err != nil {
					return fmt.Errorf("failed to count tasks with status %q: %w", m.oldStatus, err)
				}
				if count > 0 {
					fmt.Printf("  %s → %s: %d tasks\n", m.oldStatus, m.newStatus, count)
					totalTasks += count
				}
			}

			if totalTasks == 0 {
				fmt.Println("  No tasks need migration.")
			}

			if totalShipments == 0 && totalTasks == 0 {
				fmt.Println("\nNothing to migrate. Already up to date.")
				return nil
			}

			if dryRun {
				fmt.Println("\n[DRY RUN] No changes made. Run without --dry-run to apply migration.")
				return nil
			}

			// Apply shipment migrations in a transaction
			tx, err := database.Begin()
			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}
			defer tx.Rollback() //nolint:errcheck

			shipmentUpdated := int64(0)
			for _, m := range shipmentMappings {
				for _, old := range m.oldStatuses {
					result, err := tx.Exec(
						`UPDATE shipments SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE status = ?`,
						m.newStatus, old,
					)
					if err != nil {
						return fmt.Errorf("failed to migrate shipment status %q → %q: %w", old, m.newStatus, err)
					}
					n, _ := result.RowsAffected()
					shipmentUpdated += n
				}
			}

			// Set closed_reason for shipments moving to closed
			// (completed_at already set means it was naturally completed)
			_, err = tx.Exec(
				`UPDATE shipments SET updated_at = CURRENT_TIMESTAMP WHERE status = 'closed' AND completed_at IS NOT NULL`,
			)
			if err != nil {
				return fmt.Errorf("failed to update closed shipment timestamps: %w", err)
			}

			taskUpdated := int64(0)
			for _, m := range taskMappings {
				result, err := tx.Exec(
					`UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE status = ?`,
					m.newStatus, m.oldStatus,
				)
				if err != nil {
					return fmt.Errorf("failed to migrate task status %q → %q: %w", m.oldStatus, m.newStatus, err)
				}
				n, _ := result.RowsAffected()
				taskUpdated += n
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit migration: %w", err)
			}

			fmt.Printf("\n✓ Migration complete: %d shipments, %d tasks updated\n", shipmentUpdated, taskUpdated)
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be migrated without making changes")

	return cmd
}
