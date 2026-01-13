package cli

import (
	"fmt"

	"github.com/looneym/orc/internal/models"
	"github.com/spf13/cobra"
)

var operationCmd = &cobra.Command{
	Use:   "operation",
	Short: "Manage operations (tactical work groupings)",
	Long:  "Create, list, and manage operations in the ORC ledger",
}

var operationCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new operation",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error{
		title := args[0]
		missionID, _ := cmd.Flags().GetString("mission")
		description, _ := cmd.Flags().GetString("description")

		if missionID == "" {
			return fmt.Errorf("--mission flag is required")
		}

		operation, err := models.CreateOperation(missionID, title, description)
		if err != nil {
			return fmt.Errorf("failed to create operation: %w", err)
		}

		fmt.Printf("✓ Created operation %s: %s\n", operation.ID, operation.Title)
		fmt.Printf("  Under mission: %s\n", operation.MissionID)
		return nil
	},
}

var operationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List operations",
	RunE: func(cmd *cobra.Command, args []string) error {
		missionID, _ := cmd.Flags().GetString("mission")
		status, _ := cmd.Flags().GetString("status")

		operations, err := models.ListOperations(missionID, status)
		if err != nil {
			return fmt.Errorf("failed to list operations: %w", err)
		}

		if len(operations) == 0 {
			fmt.Println("No operations found")
			return nil
		}

		fmt.Printf("\n%-15s %-15s %-10s %s\n", "ID", "MISSION", "STATUS", "TITLE")
		fmt.Println("────────────────────────────────────────────────────────────────────────")
		for _, op := range operations {
			fmt.Printf("%-15s %-15s %-10s %s\n", op.ID, op.MissionID, op.Status, op.Title)
		}
		fmt.Println()

		return nil
	},
}

var operationShowCmd = &cobra.Command{
	Use:   "show [operation-id]",
	Short: "Show operation details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		operation, err := models.GetOperation(id)
		if err != nil {
			return fmt.Errorf("failed to get operation: %w", err)
		}

		fmt.Printf("\nOperation: %s\n", operation.ID)
		fmt.Printf("Mission:   %s\n", operation.MissionID)
		fmt.Printf("Title:     %s\n", operation.Title)
		fmt.Printf("Status:    %s\n", operation.Status)
		if operation.Description.Valid {
			fmt.Printf("Description: %s\n", operation.Description.String)
		}
		fmt.Printf("Created:   %s\n", operation.CreatedAt.Format("2006-01-02 15:04"))
		if operation.CompletedAt.Valid {
			fmt.Printf("Completed: %s\n", operation.CompletedAt.Time.Format("2006-01-02 15:04"))
		}
		fmt.Println()

		// List work orders under this operation
		workOrders, err := models.ListWorkOrders(id, "")
		if err == nil && len(workOrders) > 0 {
			fmt.Println("Work Orders:")
			for _, wo := range workOrders {
				fmt.Printf("  - %s [%s] %s\n", wo.ID, wo.Status, wo.Title)
			}
			fmt.Println()
		}

		return nil
	},
}

var operationCompleteCmd = &cobra.Command{
	Use:   "complete [operation-id]",
	Short: "Mark an operation as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		err := models.UpdateOperationStatus(id, "complete")
		if err != nil {
			return fmt.Errorf("failed to complete operation: %w", err)
		}

		fmt.Printf("✓ Operation %s marked as complete\n", id)
		return nil
	},
}

// OperationCmd returns the operation command
func OperationCmd() *cobra.Command {
	// Add flags
	operationCreateCmd.Flags().StringP("mission", "m", "", "Mission ID (required)")
	operationCreateCmd.Flags().StringP("description", "d", "", "Operation description")
	operationListCmd.Flags().StringP("mission", "m", "", "Filter by mission ID")
	operationListCmd.Flags().StringP("status", "s", "", "Filter by status (backlog, active, complete, cancelled)")

	// Add subcommands
	operationCmd.AddCommand(operationCreateCmd)
	operationCmd.AddCommand(operationListCmd)
	operationCmd.AddCommand(operationShowCmd)
	operationCmd.AddCommand(operationCompleteCmd)

	return operationCmd
}
