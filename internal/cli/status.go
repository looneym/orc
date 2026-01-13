package cli

import (
	"fmt"

	"github.com/looneym/orc/internal/models"
	"github.com/spf13/cobra"
)

// StatusCmd returns the status command
func StatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show ORC status overview",
		Long:  `Display an overview of all expeditions, groves, and work orders.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🌲 ORC Forest Factory Status")
			fmt.Println()

			// Count expeditions by status
			expeditions, err := models.ListExpeditions()
			if err != nil {
				return fmt.Errorf("failed to list expeditions: %w", err)
			}

			statusCounts := make(map[string]int)
			for _, exp := range expeditions {
				statusCounts[exp.Status]++
			}

			fmt.Printf("Expeditions: %d total\n", len(expeditions))
			if len(expeditions) > 0 {
				for status, count := range statusCounts {
					fmt.Printf("  - %s: %d\n", status, count)
				}
			}
			fmt.Println()

			// Count groves by status
			groves, err := models.ListGroves()
			if err != nil {
				return fmt.Errorf("failed to list groves: %w", err)
			}

			groveStatusCounts := make(map[string]int)
			for _, grove := range groves {
				groveStatusCounts[grove.Status]++
			}

			fmt.Printf("Groves: %d total\n", len(groves))
			if len(groves) > 0 {
				for status, count := range groveStatusCounts {
					fmt.Printf("  - %s: %d\n", status, count)
				}
			}
			fmt.Println()

			// Count work orders by status
			orders, err := models.ListWorkOrders("", "")
			if err != nil {
				return fmt.Errorf("failed to list work orders: %w", err)
			}

			orderStatusCounts := make(map[string]int)
			for _, wo := range orders {
				orderStatusCounts[wo.Status]++
			}

			fmt.Printf("Work Orders: %d total\n", len(orders))
			if len(orders) > 0 {
				for status, count := range orderStatusCounts {
					fmt.Printf("  - %s: %d\n", status, count)
				}
			}
			fmt.Println()

			// Show active expeditions
			var activeExpeditions []*models.Expedition
			for _, exp := range expeditions {
				if exp.Status == "active" {
					activeExpeditions = append(activeExpeditions, exp)
				}
			}

			if len(activeExpeditions) > 0 {
				fmt.Println("Active Expeditions:")
				for _, exp := range activeExpeditions {
					fmt.Printf("  - %s: %s", exp.ID, exp.Name)
					if exp.AssignedIMP.Valid {
						fmt.Printf(" (IMP: %s)", exp.AssignedIMP.String)
					}
					fmt.Println()
				}
				fmt.Println()
			}

			if len(expeditions) == 0 && len(groves) == 0 && len(orders) == 0 {
				fmt.Println("No activity yet. Get started:")
				fmt.Println("  orc expedition create \"My First Expedition\"")
			}

			return nil
		},
	}
}
