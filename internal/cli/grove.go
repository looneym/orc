package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/looneym/orc/internal/models"
	"github.com/spf13/cobra"
)

// GroveCmd returns the grove command
func GroveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grove",
		Short: "Manage groves (worktrees)",
		Long:  `Create and manage groves (isolated development workspaces).`,
	}

	cmd.AddCommand(groveCreateCmd())
	cmd.AddCommand(groveListCmd())

	return cmd
}

func groveCreateCmd() *cobra.Command {
	var expeditionID string

	cmd := &cobra.Command{
		Use:   "create [grove-name]",
		Short: "Create a new grove",
		Long:  `Create a new grove (worktree) and optionally associate it with an expedition.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groveName := args[0]

			// Default path (user can customize this later)
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			path := fmt.Sprintf("%s/src/worktrees/%s", home, groveName)

			var expID *string
			if expeditionID != "" {
				expID = &expeditionID
			}

			grove, err := models.CreateGrove(groveName, path, expID)
			if err != nil {
				return fmt.Errorf("failed to create grove: %w", err)
			}

			fmt.Printf("✓ Created grove %s\n", grove.ID)
			fmt.Printf("  Path: %s\n", grove.Path)
			if grove.ExpeditionID.Valid {
				fmt.Printf("  Expedition: %s\n", grove.ExpeditionID.String)
			}
			fmt.Println()
			fmt.Println("Note: Grove path is tracked but directory is not created automatically.")
			fmt.Println("Create the worktree manually or use your worktree creation tool.")

			return nil
		},
	}

	cmd.Flags().StringVarP(&expeditionID, "expedition", "e", "", "Associate grove with expedition")

	return cmd
}

func groveListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all groves",
		Long:  `List all groves with their current status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			groves, err := models.ListGroves()
			if err != nil {
				return fmt.Errorf("failed to list groves: %w", err)
			}

			if len(groves) == 0 {
				fmt.Println("No groves found.")
				fmt.Println()
				fmt.Println("Create your first grove:")
				fmt.Println("  orc grove create my-grove --expedition EXP-001")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tSTATUS\tEXPEDITION\tPATH")
			fmt.Fprintln(w, "--\t------\t----------\t----")

			for _, grove := range groves {
				exp := "-"
				if grove.ExpeditionID.Valid {
					exp = grove.ExpeditionID.String
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					grove.ID,
					grove.Status,
					exp,
					grove.Path,
				)
			}

			w.Flush()
			return nil
		},
	}
}
