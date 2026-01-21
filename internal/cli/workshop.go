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

// WorkshopCmd returns the workshop command
func WorkshopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workshop",
		Short: "Manage workshops (persistent places)",
		Long:  `Create and manage workshops - persistent places within factories that host workbenches.`,
	}

	cmd.AddCommand(workshopCreateCmd())
	cmd.AddCommand(workshopListCmd())
	cmd.AddCommand(workshopShowCmd())
	cmd.AddCommand(workshopDeleteCmd())

	return cmd
}

func workshopCreateCmd() *cobra.Command {
	var factoryID string
	var name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new workshop in a factory",
		Long: `Create a new workshop within a factory.

A Workshop is a persistent named place within a Factory. Workshops
have atmospheric names from a pool (e.g., "Ironmoss Forge", "Blackpine Foundry").
If no name is provided, one will be assigned from the pool.

Examples:
  orc workshop create --factory FACT-001
  orc workshop create --factory FACT-001 --name "Custom Workshop"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if factoryID == "" {
				return fmt.Errorf("--factory flag is required")
			}

			resp, err := wire.WorkshopService().CreateWorkshop(ctx, primary.CreateWorkshopRequest{
				FactoryID: factoryID,
				Name:      name,
			})
			if err != nil {
				return fmt.Errorf("failed to create workshop: %w", err)
			}

			fmt.Printf("Created workshop %s: %s\n", resp.WorkshopID, resp.Workshop.Name)
			fmt.Printf("  Factory: %s\n", resp.Workshop.FactoryID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&factoryID, "factory", "f", "", "Factory ID (required)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Workshop name (optional - uses name pool if empty)")

	return cmd
}

func workshopListCmd() *cobra.Command {
	var factoryID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all workshops",
		Long:  `List all workshops with their current status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			workshops, err := wire.WorkshopService().ListWorkshops(ctx, primary.WorkshopFilters{
				FactoryID: factoryID,
			})
			if err != nil {
				return fmt.Errorf("failed to list workshops: %w", err)
			}

			if len(workshops) == 0 {
				fmt.Println("No workshops found.")
				fmt.Println()
				fmt.Println("Create your first workshop:")
				fmt.Println("  orc workshop create --factory FACT-001")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tFACTORY\tSTATUS")
			fmt.Fprintln(w, "--\t----\t-------\t------")

			for _, ws := range workshops {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					ws.ID,
					ws.Name,
					ws.FactoryID,
					ws.Status,
				)
			}

			w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVarP(&factoryID, "factory", "f", "", "Filter by factory ID")

	return cmd
}

func workshopShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [workshop-id]",
		Short: "Show workshop details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			workshop, err := wire.WorkshopService().GetWorkshop(ctx, args[0])
			if err != nil {
				return fmt.Errorf("workshop not found: %w", err)
			}

			fmt.Printf("Workshop: %s\n", workshop.ID)
			fmt.Printf("Name: %s\n", workshop.Name)
			fmt.Printf("Factory: %s\n", workshop.FactoryID)
			fmt.Printf("Status: %s\n", workshop.Status)
			fmt.Printf("Created: %s\n", workshop.CreatedAt)

			return nil
		},
	}
}

func workshopDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [workshop-id]",
		Short: "Delete a workshop",
		Long: `Delete a workshop from the database.

WARNING: This is a destructive operation. Workshops with workbenches
require the --force flag.

Examples:
  orc workshop delete WORK-001
  orc workshop delete WORK-001 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			workshopID := args[0]

			err := wire.WorkshopService().DeleteWorkshop(ctx, primary.DeleteWorkshopRequest{
				WorkshopID: workshopID,
				Force:      force,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Deleted workshop %s\n", workshopID)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force delete even with workbenches")

	return cmd
}
