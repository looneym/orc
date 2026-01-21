package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
	ctx "github.com/example/orc/internal/context"
	"github.com/example/orc/internal/wire"
)

// StatusCmd returns the status command
func StatusCmd() *cobra.Command {
	var showHandoff bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current work context from config.json",
		Long: `Display the current work context based on .orc/config.json:
- Active mission, shipments, and tasks
- Latest handoff ID and timestamp (use --handoff to see full note)

This provides a focused view of "where am I right now?"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if we're in a mission context first
			commissionCtx, _ := ctx.DetectCommissionContext()
			var activeCommissionID string
			var currentHandoffID string
			var lastUpdated string
			var currentFocus string

			if commissionCtx != nil {
				// Commission context - try to load config from workspace or current directory
				cfg, err := config.LoadConfig(commissionCtx.WorkspacePath)
				if err != nil {
					// Try current directory
					cwd, _ := os.Getwd()
					cfg, err = config.LoadConfig(cwd)
				}

				if err == nil {
					// Extract fields based on config type
					switch cfg.Type {
					case config.TypeGrove:
						activeCommissionID = cfg.Grove.CommissionID
						currentFocus = cfg.Grove.CurrentFocus
					case config.TypeCommission:
						activeCommissionID = cfg.Commission.CommissionID
						currentFocus = cfg.Commission.CurrentFocus
					case config.TypeGlobal:
						activeCommissionID = cfg.State.ActiveCommissionID
						currentHandoffID = cfg.State.CurrentHandoffID
						lastUpdated = cfg.State.LastUpdated
						currentFocus = cfg.State.CurrentFocus
					}
				}

				// If still no active mission, use mission from .orc-mission file
				if activeCommissionID == "" {
					activeCommissionID = commissionCtx.CommissionID
				}
				fmt.Println("🎯 ORC Status - Commission Context")
			} else {
				// Master context - read from global config.json
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("failed to get home directory: %w", err)
				}

				cfg, err := config.LoadConfig(homeDir)
				if err != nil {
					return fmt.Errorf("failed to read config.json: %w\nHint: Run 'orc init' if you haven't initialized ORC yet", err)
				}

				if cfg.State != nil {
					activeCommissionID = cfg.State.ActiveCommissionID
					currentHandoffID = cfg.State.CurrentHandoffID
					lastUpdated = cfg.State.LastUpdated
					currentFocus = cfg.State.CurrentFocus
				}

				fmt.Println("🎯 ORC Status - Current Context")
			}
			fmt.Println()

			// Display active commission
			if activeCommissionID != "" {
				commission, err := wire.CommissionService().GetCommission(context.Background(), activeCommissionID)
				if err != nil {
					fmt.Printf("❌ Commission: %s (error loading: %v)\n", activeCommissionID, err)
				} else {
					fmt.Printf("🎯 Commission: %s - %s [%s]\n", commission.ID, commission.Title, commission.Status)
					if commission.Description != "" {
						fmt.Printf("   %s\n", commission.Description)
					}
				}
			} else {
				fmt.Println("🎯 Commission: (none active)")
			}
			fmt.Println()

			// Display current focus if set
			if currentFocus != "" {
				containerType, title, status := GetFocusInfo(currentFocus)
				if containerType != "" {
					fmt.Printf("Focus: %s - %s [%s]\n", currentFocus, title, status)
					fmt.Printf("   (%s)\n", containerType)
				} else {
					fmt.Printf("Focus: %s (container not found)\n", currentFocus)
				}
				fmt.Println()
			}

			// Display latest handoff
			if currentHandoffID != "" {
				handoff, err := wire.HandoffService().GetHandoff(context.Background(), currentHandoffID)
				if err != nil {
					fmt.Printf("❌ Latest Handoff: %s (error loading: %v)\n", currentHandoffID, err)
				} else {
					fmt.Printf("📝 Latest Handoff: %s\n", handoff.ID)
					fmt.Printf("   Created: %s\n", handoff.CreatedAt)

					// Show full note if --handoff flag is set
					if showHandoff {
						fmt.Println()
						fmt.Println("--- HANDOFF NOTE ---")
						fmt.Println(handoff.HandoffNote)
						fmt.Println("--- END HANDOFF ---")
					}
				}
			} else {
				fmt.Println("📝 Latest Handoff: (none)")
			}
			fmt.Println()

			if lastUpdated != "" {
				fmt.Printf("Last updated: %s\n", lastUpdated)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&showHandoff, "handoff", "n", false, "Show full handoff note")

	return cmd
}
