package cli

import (
	"fmt"

	"github.com/looneym/orc/internal/models"
	"github.com/spf13/cobra"
)

// SummaryCmd returns the summary command
func SummaryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "summary",
		Short: "Show summary of all open missions, operations, and expeditions",
		Long: `Display a high-level overview of all work in progress:
- Open missions with their active operations
- Open expeditions

This provides a global view of all work across ORC.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get all non-complete missions
			missions, err := models.ListMissions("")
			if err != nil {
				return fmt.Errorf("failed to list missions: %w", err)
			}

			// Filter to open missions (not complete or archived)
			var openMissions []*models.Mission
			for _, m := range missions {
				if m.Status != "complete" && m.Status != "archived" {
					openMissions = append(openMissions, m)
				}
			}

			if len(openMissions) == 0 {
				fmt.Println("No open missions")
				return nil
			}

			fmt.Println("📊 ORC Summary - Open Work")
			fmt.Println()

			// Display each mission with its operations in tree format
			for i, mission := range openMissions {
				// Display mission
				statusEmoji := getStatusEmoji(mission.Status)
				fmt.Printf("%s %s - %s [%s]\n", statusEmoji, mission.ID, mission.Title, mission.Status)

				// Get operations for this mission
				operations, err := models.ListOperations(mission.ID, "")
				if err != nil {
					return fmt.Errorf("failed to list operations for %s: %w", mission.ID, err)
				}

				// Filter to non-complete operations
				var activeOps []*models.Operation
				for _, op := range operations {
					if op.Status != "complete" && op.Status != "cancelled" {
						activeOps = append(activeOps, op)
					}
				}

				if len(activeOps) > 0 {
					for j, op := range activeOps {
						opEmoji := getStatusEmoji(op.Status)

						// Determine tree characters
						var prefix string
						if j < len(activeOps)-1 {
							prefix = "├── "
						} else {
							prefix = "└── "
						}

						fmt.Printf("%s%s %s - %s [%s]\n", prefix, opEmoji, op.ID, op.Title, op.Status)

						// Get work orders for this operation
						workOrders, err := models.ListWorkOrders(op.ID, "")
						if err == nil && len(workOrders) > 0 {
							var activeWOs []*models.WorkOrder
							for _, wo := range workOrders {
								if wo.Status != "complete" && wo.Status != "cancelled" {
									activeWOs = append(activeWOs, wo)
								}
							}

							if len(activeWOs) > 0 {
								for k, wo := range activeWOs {
									woEmoji := getStatusEmoji(wo.Status)
									var woPrefix string
									if j < len(activeOps)-1 {
										woPrefix = "│   "
									} else {
										woPrefix = "    "
									}
									if k < len(activeWOs)-1 {
										woPrefix += "├── "
									} else {
										woPrefix += "└── "
									}
									fmt.Printf("%s%s %s - %s [%s]\n", woPrefix, woEmoji, wo.ID, wo.Title, wo.Status)
								}
							}
						}
					}
				} else {
					fmt.Println("└── (No active operations)")
				}

				// Add spacing between missions
				if i < len(openMissions)-1 {
					fmt.Println()
				}
			}

			fmt.Println()

			// Show open expeditions
			expeditions, err := models.ListExpeditions()
			if err != nil {
				return fmt.Errorf("failed to list expeditions: %w", err)
			}

			var openExpeditions []*models.Expedition
			for _, exp := range expeditions {
				if exp.Status != "complete" {
					openExpeditions = append(openExpeditions, exp)
				}
			}

			if len(openExpeditions) > 0 {
				fmt.Println("🌲 Open Expeditions")
				fmt.Println()

				for _, exp := range openExpeditions {
					expEmoji := getStatusEmoji(exp.Status)
					impInfo := ""
					if exp.AssignedIMP.Valid {
						impInfo = fmt.Sprintf(" [IMP: %s]", exp.AssignedIMP.String)
					}
					woInfo := ""
					if exp.WorkOrderID.Valid {
						woInfo = fmt.Sprintf(" → %s", exp.WorkOrderID.String)
					}
					fmt.Printf("%s %s - %s [%s]%s%s\n", expEmoji, exp.ID, exp.Name, exp.Status, impInfo, woInfo)
				}
				fmt.Println()
			}

			return nil
		},
	}
}

func getStatusEmoji(status string) string {
	switch status {
	case "active", "in_progress":
		return "🚀"
	case "planning":
		return "📋"
	case "paused":
		return "⏸️"
	case "backlog":
		return "📦"
	case "next":
		return "⏭️"
	case "complete":
		return "✅"
	case "cancelled", "archived":
		return "🗄️"
	default:
		return "•"
	}
}
