package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

var workOrderCmd = &cobra.Command{
	Use:   "work_order",
	Short: "Manage work orders",
	Long:  "Create, list, update, and manage work orders in the ORC ledger",
}

var workOrderCreateCmd = &cobra.Command{
	Use:   "create [outcome]",
	Short: "Create a new work order",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID, _ := cmd.Flags().GetString("shipment-id")
		acceptanceCriteria, _ := cmd.Flags().GetString("acceptance-criteria")

		if shipmentID == "" {
			return fmt.Errorf("--shipment-id flag is required")
		}

		req := primary.CreateWorkOrderRequest{
			ShipmentID:         shipmentID,
			Outcome:            args[0],
			AcceptanceCriteria: acceptanceCriteria,
		}

		resp, err := wire.WorkOrderService().CreateWorkOrder(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to create work order: %w", err)
		}

		workOrder := resp.WorkOrder
		fmt.Printf("✓ Created work order %s\n", workOrder.ID)
		fmt.Printf("  Outcome: %s\n", workOrder.Outcome)
		fmt.Printf("  Shipment: %s\n", workOrder.ShipmentID)
		fmt.Printf("  Status: %s\n", workOrder.Status)
		return nil
	},
}

var workOrderListCmd = &cobra.Command{
	Use:   "list",
	Short: "List work orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID, _ := cmd.Flags().GetString("shipment-id")
		status, _ := cmd.Flags().GetString("status")

		workOrders, err := wire.WorkOrderService().ListWorkOrders(ctx, primary.WorkOrderFilters{
			ShipmentID: shipmentID,
			Status:     status,
		})
		if err != nil {
			return fmt.Errorf("failed to list work orders: %w", err)
		}

		if len(workOrders) == 0 {
			fmt.Println("No work orders found")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tSHIPMENT\tOUTCOME\tSTATUS")
		fmt.Fprintln(w, "--\t--------\t-------\t------")
		for _, item := range workOrders {
			// Truncate outcome for display
			outcome := item.Outcome
			if len(outcome) > 40 {
				outcome = outcome[:37] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				item.ID,
				item.ShipmentID,
				outcome,
				item.Status,
			)
		}
		w.Flush()
		return nil
	},
}

var workOrderShowCmd = &cobra.Command{
	Use:   "show [work-order-id]",
	Short: "Show work order details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		workOrderID := args[0]

		workOrder, err := wire.WorkOrderService().GetWorkOrder(ctx, workOrderID)
		if err != nil {
			return fmt.Errorf("work order not found: %w", err)
		}

		fmt.Printf("Work Order: %s\n", workOrder.ID)
		fmt.Printf("Shipment: %s\n", workOrder.ShipmentID)
		fmt.Printf("Outcome: %s\n", workOrder.Outcome)
		if workOrder.AcceptanceCriteria != "" {
			fmt.Printf("Acceptance Criteria: %s\n", workOrder.AcceptanceCriteria)
		}
		fmt.Printf("Status: %s\n", workOrder.Status)
		fmt.Printf("Created: %s\n", workOrder.CreatedAt)
		fmt.Printf("Updated: %s\n", workOrder.UpdatedAt)

		return nil
	},
}

var workOrderActivateCmd = &cobra.Command{
	Use:   "activate [work-order-id]",
	Short: "Activate a draft work order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		workOrderID := args[0]

		err := wire.WorkOrderService().ActivateWorkOrder(ctx, workOrderID)
		if err != nil {
			return fmt.Errorf("failed to activate work order: %w", err)
		}

		fmt.Printf("✓ Work order %s activated\n", workOrderID)
		return nil
	},
}

var workOrderCompleteCmd = &cobra.Command{
	Use:   "complete [work-order-id]",
	Short: "Complete an active work order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		workOrderID := args[0]

		err := wire.WorkOrderService().CompleteWorkOrder(ctx, workOrderID)
		if err != nil {
			return fmt.Errorf("failed to complete work order: %w", err)
		}

		fmt.Printf("✓ Work order %s completed\n", workOrderID)
		return nil
	},
}

var workOrderDeleteCmd = &cobra.Command{
	Use:   "delete [work-order-id]",
	Short: "Delete a work order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		workOrderID := args[0]

		err := wire.WorkOrderService().DeleteWorkOrder(ctx, workOrderID)
		if err != nil {
			return fmt.Errorf("failed to delete work order: %w", err)
		}

		fmt.Printf("✓ Work order %s deleted\n", workOrderID)
		return nil
	},
}

func init() {
	// work_order create flags
	workOrderCreateCmd.Flags().String("shipment-id", "", "Shipment ID (required)")
	workOrderCreateCmd.Flags().String("acceptance-criteria", "", "Acceptance criteria (JSON array)")
	workOrderCreateCmd.MarkFlagRequired("shipment-id")

	// work_order list flags
	workOrderListCmd.Flags().String("shipment-id", "", "Filter by shipment")
	workOrderListCmd.Flags().StringP("status", "s", "", "Filter by status")

	// Register subcommands
	workOrderCmd.AddCommand(workOrderCreateCmd)
	workOrderCmd.AddCommand(workOrderListCmd)
	workOrderCmd.AddCommand(workOrderShowCmd)
	workOrderCmd.AddCommand(workOrderActivateCmd)
	workOrderCmd.AddCommand(workOrderCompleteCmd)
	workOrderCmd.AddCommand(workOrderDeleteCmd)
}

// WorkOrderCmd returns the work_order command
func WorkOrderCmd() *cobra.Command {
	return workOrderCmd
}
