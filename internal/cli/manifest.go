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

var manifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "Manage manifests",
	Long:  "List and view shipment manifests in the ORC ledger",
}

var manifestListCmd = &cobra.Command{
	Use:   "list",
	Short: "List manifests",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID, _ := cmd.Flags().GetString("shipment")
		status, _ := cmd.Flags().GetString("status")

		manifests, err := wire.ManifestService().ListManifests(ctx, primary.ManifestFilters{
			ShipmentID: shipmentID,
			Status:     status,
		})
		if err != nil {
			return fmt.Errorf("failed to list manifests: %w", err)
		}

		if len(manifests) == 0 {
			fmt.Println("No manifests found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tSHIPMENT\tCREATED BY\tSTATUS\tCREATED")
		fmt.Fprintln(w, "--\t--------\t----------\t------\t-------")
		for _, item := range manifests {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				item.ID,
				item.ShipmentID,
				item.CreatedBy,
				item.Status,
				item.CreatedAt,
			)
		}
		w.Flush()
		return nil
	},
}

var manifestShowCmd = &cobra.Command{
	Use:   "show [manifest-id]",
	Short: "Show manifest details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		manifestID := args[0]

		manifest, err := wire.ManifestService().GetManifest(ctx, manifestID)
		if err != nil {
			return fmt.Errorf("manifest not found: %w", err)
		}

		fmt.Printf("Manifest: %s\n", manifest.ID)
		fmt.Printf("Shipment: %s\n", manifest.ShipmentID)
		fmt.Printf("Created By: %s\n", manifest.CreatedBy)
		fmt.Printf("Status: %s\n", manifest.Status)
		if manifest.Attestation != "" {
			fmt.Printf("Attestation: %s\n", manifest.Attestation)
		}
		if manifest.Tasks != "" {
			fmt.Printf("Tasks: %s\n", manifest.Tasks)
		}
		if manifest.OrderingNotes != "" {
			fmt.Printf("Ordering Notes: %s\n", manifest.OrderingNotes)
		}
		fmt.Printf("Created: %s\n", manifest.CreatedAt)
		fmt.Printf("Updated: %s\n", manifest.UpdatedAt)

		return nil
	},
}

func init() {
	// manifest list flags
	manifestListCmd.Flags().String("shipment", "", "Filter by shipment ID")
	manifestListCmd.Flags().StringP("status", "s", "", "Filter by status (draft|launched)")

	// Register subcommands
	manifestCmd.AddCommand(manifestListCmd)
	manifestCmd.AddCommand(manifestShowCmd)
}

// ManifestCmd returns the manifest command
func ManifestCmd() *cobra.Command {
	return manifestCmd
}
