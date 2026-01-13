package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/looneym/orc/internal/models"
	"github.com/spf13/cobra"
)

// ExpeditionCmd returns the expedition command
func ExpeditionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "expedition",
		Short: "Manage expeditions",
		Long:  `Create, list, and manage Forest Factory expeditions.`,
	}

	cmd.AddCommand(expeditionCreateCmd())
	cmd.AddCommand(expeditionListCmd())
	cmd.AddCommand(expeditionShowCmd())

	return cmd
}

func expeditionCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new expedition",
		Long:  `Create a new expedition with the given name.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			exp, err := models.CreateExpedition(name)
			if err != nil {
				return fmt.Errorf("failed to create expedition: %w", err)
			}

			fmt.Printf("✓ Created expedition %s: %s\n", exp.ID, exp.Name)
			fmt.Printf("  Status: %s\n", exp.Status)
			fmt.Println()
			fmt.Println("Next steps:")
			fmt.Printf("  orc grove create <grove-name> --expedition %s\n", exp.ID)
			fmt.Printf("  orc expedition show %s\n", exp.ID)

			return nil
		},
	}
}

func expeditionListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all expeditions",
		Long:  `List all expeditions with their current status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			expeditions, err := models.ListExpeditions()
			if err != nil {
				return fmt.Errorf("failed to list expeditions: %w", err)
			}

			if len(expeditions) == 0 {
				fmt.Println("No expeditions found.")
				fmt.Println()
				fmt.Println("Create your first expedition:")
				fmt.Println("  orc expedition create \"My First Expedition\"")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tSTATUS\tIMP\tCREATED")
			fmt.Fprintln(w, "--\t----\t------\t---\t-------")

			for _, exp := range expeditions {
				imp := "-"
				if exp.AssignedIMP.Valid {
					imp = exp.AssignedIMP.String
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					exp.ID,
					exp.Name,
					exp.Status,
					imp,
					exp.CreatedAt.Format("2006-01-02"),
				)
			}

			w.Flush()
			return nil
		},
	}
}

func expeditionShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [expedition-id]",
		Short: "Show expedition details",
		Long:  `Show detailed information about an expedition.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			exp, err := models.GetExpedition(id)
			if err != nil {
				return fmt.Errorf("failed to get expedition: %w", err)
			}

			fmt.Printf("Expedition: %s\n", exp.ID)
			fmt.Printf("Name:       %s\n", exp.Name)
			fmt.Printf("Status:     %s\n", exp.Status)

			if exp.AssignedIMP.Valid {
				fmt.Printf("IMP:        %s\n", exp.AssignedIMP.String)
			} else {
				fmt.Printf("IMP:        (unassigned)\n")
			}

			if exp.WorkOrderID.Valid {
				fmt.Printf("Work Order: %s\n", exp.WorkOrderID.String)
			}

			fmt.Printf("Created:    %s\n", exp.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Updated:    %s\n", exp.UpdatedAt.Format("2006-01-02 15:04:05"))

			// Show associated groves
			groves, err := models.GetGrovesByExpedition(id)
			if err != nil {
				return fmt.Errorf("failed to get groves: %w", err)
			}

			if len(groves) > 0 {
				fmt.Println()
				fmt.Println("Groves:")
				for _, grove := range groves {
					fmt.Printf("  - %s (%s)\n", grove.ID, grove.Status)
					fmt.Printf("    Path: %s\n", grove.Path)
				}
			}

			return nil
		},
	}
}
