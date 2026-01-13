package cli

import (
	"fmt"

	"github.com/looneym/orc/internal/models"
	"github.com/spf13/cobra"
)

var workOrderCmd = &cobra.Command{
	Use:   "work-order",
	Short: "Manage work orders (individual tasks)",
	Long:  "Create, list, claim, and complete work orders in the ORC ledger",
}

var workOrderCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new work order",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		operationID, _ := cmd.Flags().GetString("operation")
		description, _ := cmd.Flags().GetString("description")
		contextRef, _ := cmd.Flags().GetString("context-ref")

		if operationID == "" {
			return fmt.Errorf("--operation flag is required")
		}

		wo, err := models.CreateWorkOrder(operationID, title, description, contextRef)
		if err != nil {
			return fmt.Errorf("failed to create work order: %w", err)
		}

		fmt.Printf("✓ Created work order %s: %s\n", wo.ID, wo.Title)
		fmt.Printf("  Under operation: %s\n", wo.OperationID)
		if wo.ContextRef.Valid {
			fmt.Printf("  Context: %s\n", wo.ContextRef.String)
		}
		return nil
	},
}

var workOrderListCmd = &cobra.Command{
	Use:   "list",
	Short: "List work orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		operationID, _ := cmd.Flags().GetString("operation")
		status, _ := cmd.Flags().GetString("status")

		orders, err := models.ListWorkOrders(operationID, status)
		if err != nil {
			return fmt.Errorf("failed to list work orders: %w", err)
		}

		if len(orders) == 0 {
			fmt.Println("No work orders found")
			return nil
		}

		fmt.Printf("\n%-15s %-15s %-10s %-15s %s\n", "ID", "OPERATION", "STATUS", "ASSIGNED", "TITLE")
		fmt.Println("─────────────────────────────────────────────────────────────────────────────────")
		for _, wo := range orders {
			assigned := "-"
			if wo.AssignedImp.Valid {
				assigned = wo.AssignedImp.String
			}
			fmt.Printf("%-15s %-15s %-10s %-15s %s\n", wo.ID, wo.OperationID, wo.Status, assigned, wo.Title)
		}
		fmt.Println()

		return nil
	},
}

var workOrderShowCmd = &cobra.Command{
	Use:   "show [work-order-id]",
	Short: "Show work order details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		wo, err := models.GetWorkOrder(id)
		if err != nil {
			return fmt.Errorf("failed to get work order: %w", err)
		}

		fmt.Printf("\nWork Order: %s\n", wo.ID)
		fmt.Printf("Operation:  %s\n", wo.OperationID)
		fmt.Printf("Title:      %s\n", wo.Title)
		fmt.Printf("Status:     %s\n", wo.Status)
		if wo.Description.Valid {
			fmt.Printf("Description: %s\n", wo.Description.String)
		}
		if wo.AssignedImp.Valid {
			fmt.Printf("Assigned:   %s\n", wo.AssignedImp.String)
		}
		if wo.ContextRef.Valid {
			fmt.Printf("Context:    %s\n", wo.ContextRef.String)
		}
		fmt.Printf("Created:    %s\n", wo.CreatedAt.Format("2006-01-02 15:04"))
		if wo.ClaimedAt.Valid {
			fmt.Printf("Claimed:    %s\n", wo.ClaimedAt.Time.Format("2006-01-02 15:04"))
		}
		if wo.CompletedAt.Valid {
			fmt.Printf("Completed:  %s\n", wo.CompletedAt.Time.Format("2006-01-02 15:04"))
		}
		fmt.Println()

		return nil
	},
}

var workOrderClaimCmd = &cobra.Command{
	Use:   "claim [work-order-id]",
	Short: "Claim a work order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		impName, _ := cmd.Flags().GetString("imp")

		if impName == "" {
			impName = "IMP-UNKNOWN"
		}

		err := models.ClaimWorkOrder(id, impName)
		if err != nil {
			return fmt.Errorf("failed to claim work order: %w", err)
		}

		fmt.Printf("✓ Work order %s claimed by %s\n", id, impName)
		fmt.Printf("  Status: in_progress\n")
		return nil
	},
}

var workOrderCompleteCmd = &cobra.Command{
	Use:   "complete [work-order-id]",
	Short: "Mark a work order as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		err := models.CompleteWorkOrder(id)
		if err != nil {
			return fmt.Errorf("failed to complete work order: %w", err)
		}

		fmt.Printf("✓ Work order %s marked as complete\n", id)
		return nil
	},
}

// WorkOrderCmd returns the work-order command
func WorkOrderCmd() *cobra.Command {
	// Add flags
	workOrderCreateCmd.Flags().StringP("operation", "o", "", "Operation ID (required)")
	workOrderCreateCmd.Flags().StringP("description", "d", "", "Work order description")
	workOrderCreateCmd.Flags().StringP("context-ref", "c", "", "Graphiti context reference (e.g., graphiti:episode-uuid)")

	workOrderListCmd.Flags().StringP("operation", "o", "", "Filter by operation ID")
	workOrderListCmd.Flags().StringP("status", "s", "", "Filter by status (backlog, next, in_progress, complete)")

	workOrderClaimCmd.Flags().StringP("imp", "i", "", "IMP name claiming the work order")

	// Add subcommands
	workOrderCmd.AddCommand(workOrderCreateCmd)
	workOrderCmd.AddCommand(workOrderListCmd)
	workOrderCmd.AddCommand(workOrderShowCmd)
	workOrderCmd.AddCommand(workOrderClaimCmd)
	workOrderCmd.AddCommand(workOrderCompleteCmd)

	return workOrderCmd
}
