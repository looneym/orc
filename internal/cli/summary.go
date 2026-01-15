package cli

import (
	"fmt"

	"github.com/example/orc/internal/context"
	"github.com/example/orc/internal/models"
	"github.com/spf13/cobra"
)

// SummaryCmd returns the summary command
func SummaryCmd() *cobra.Command {
	var showAll bool

	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show summary of all open missions and work orders",
		Long: `Show a hierarchical summary of missions and their work orders.

Filtering:
  --scope all      Show all missions (default)
  --scope current  Show only current mission (requires mission context)
  --hide paused    Hide paused items
  --hide blocked   Hide blocked items
  --hide paused,blocked  Hide multiple statuses

Examples:
  orc summary                       # all missions, all statuses
  orc summary --scope current       # current mission only
  orc summary --hide paused         # hide paused work
  orc summary --scope current --hide paused,blocked  # focused view`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get scope from flag (default: "all")
			scope, _ := cmd.Flags().GetString("scope")

			// Validate scope value
			if scope != "all" && scope != "current" {
				return fmt.Errorf("invalid scope: must be 'all' or 'current'")
			}

			// Determine mission filter based on scope
			var filterMissionID string
			if scope == "current" {
				// Get current mission from context
				missionCtx, _ := context.DetectMissionContext()
				if missionCtx == nil || missionCtx.MissionID == "" {
					return fmt.Errorf("--scope current requires being in a mission context (no .orc-mission file found)")
				}
				filterMissionID = missionCtx.MissionID
				fmt.Printf("📊 ORC Summary - %s (Current Mission)\n", filterMissionID)
			} else {
				// scope == "all"
				fmt.Println("📊 ORC Summary - Open Work")
			}
			fmt.Println()

			// Get all non-complete missions
			missions, err := models.ListMissions("")
			if err != nil {
				return fmt.Errorf("failed to list missions: %w", err)
			}

			// Get hide statuses from flag
			hideStatuses, _ := cmd.Flags().GetStringSlice("hide")
			hideMap := make(map[string]bool)
			for _, status := range hideStatuses {
				hideMap[status] = true
			}

			// Filter to open missions (not complete or archived)
			var openMissions []*models.Mission
			for _, m := range missions {
				// Always hide complete and archived
				if m.Status == "complete" || m.Status == "archived" {
					continue
				}
				// Hide if in hide list
				if hideMap[m.Status] {
					continue
				}
				// If in deputy context and not showing all, filter to this mission
				if filterMissionID != "" && m.ID != filterMissionID {
					continue
				}
				openMissions = append(openMissions, m)
			}

			if len(openMissions) == 0 {
				if filterMissionID != "" {
					fmt.Printf("No open work orders for %s\n", filterMissionID)
				} else {
					fmt.Println("No open missions")
				}
				return nil
			}

			// Display each mission with its work orders in tree format
			for i, mission := range openMissions {
				// Display mission
				statusEmoji := getStatusEmoji(mission.Status)
				fmt.Printf("%s %s - %s [%s]\n", statusEmoji, mission.ID, mission.Title, mission.Status)
				fmt.Println("│") // Empty line with vertical continuation after mission header

				// Get work orders for this mission
				workOrders, err := models.ListWorkOrders(mission.ID, "")
				if err != nil {
					return fmt.Errorf("failed to list work orders for %s: %w", mission.ID, err)
				}

				// Filter to non-complete work orders
				var activeWOs []*models.WorkOrder
				for _, wo := range workOrders {
					// Always hide complete
					if wo.Status == "complete" {
						continue
					}
					// Hide if in hide list
					if hideMap[wo.Status] {
						continue
					}
					activeWOs = append(activeWOs, wo)
				}

				if len(activeWOs) > 0 {
					// Build children map
					childrenMap := make(map[string][]*models.WorkOrder)
					for _, wo := range activeWOs {
						if wo.ParentID.Valid {
							children := childrenMap[wo.ParentID.String]
							children = append(children, wo)
							childrenMap[wo.ParentID.String] = children
						}
					}

					// Separate epics (have children) from standalone
					epics := []*models.WorkOrder{}
					standalone := []*models.WorkOrder{}
					for _, wo := range activeWOs {
						if wo.ParentID.Valid {
							// This is a child, skip
							continue
						}
						if len(childrenMap[wo.ID]) > 0 {
							epics = append(epics, wo)
						} else {
							standalone = append(standalone, wo)
						}
					}

					// Display epics first with empty lines between them
					for _, epic := range epics {
						epicEmoji := getStatusEmoji(epic.Status)
						// Add pin emoji if pinned
						if epic.Pinned {
							epicEmoji = "📌" + epicEmoji
						}
						groveInfo := ""
						if epic.AssignedGroveID.Valid {
							groveInfo = fmt.Sprintf(" [Grove: %s]", epic.AssignedGroveID.String)
						}
						fmt.Printf("├── %s %s - %s [%s]%s\n", epicEmoji, epic.ID, epic.Title, epic.Status, groveInfo)

						// Display children (no empty lines between children)
						children := childrenMap[epic.ID]
						for k, child := range children {
							childEmoji := getStatusEmoji(child.Status)
							// Add pin emoji if pinned
							if child.Pinned {
								childEmoji = "📌" + childEmoji
							}
							var childPrefix string
							if k < len(children)-1 {
								childPrefix = "│   ├── "
							} else {
								childPrefix = "│   └── "
							}
							childGroveInfo := ""
							if child.AssignedGroveID.Valid {
								childGroveInfo = fmt.Sprintf(" [Grove: %s]", child.AssignedGroveID.String)
							}
							fmt.Printf("%s%s %s - %s [%s]%s\n", childPrefix, childEmoji, child.ID, child.Title, child.Status, childGroveInfo)
						}
						// Empty line after each epic with vertical continuation
						fmt.Println("│")
					}

					// Display standalone work orders with empty lines between them
					for j, wo := range standalone {
						woEmoji := getStatusEmoji(wo.Status)
						// Add pin emoji if pinned
						if wo.Pinned {
							woEmoji = "📌" + woEmoji
						}
						groveInfo := ""
						if wo.AssignedGroveID.Valid {
							groveInfo = fmt.Sprintf(" [Grove: %s]", wo.AssignedGroveID.String)
						}
						// Use └── for last standalone, ├── for others
						var prefix string
						if j < len(standalone)-1 {
							prefix = "├── "
						} else {
							prefix = "└── "
						}
						fmt.Printf("%s%s %s - %s [%s]%s\n", prefix, woEmoji, wo.ID, wo.Title, wo.Status, groveInfo)
						// Add vertical continuation line between standalone orders (not after last)
						if j < len(standalone)-1 {
							fmt.Println("│")
						}
					}
				} else {
					fmt.Println("└── (No active work orders)")
				}

				// Add spacing between missions
				if i < len(openMissions)-1 {
					fmt.Println()
				}
			}

			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all missions (override deputy scoping)")
	cmd.Flags().StringP("scope", "s", "all", "Scope: 'all' or 'current'")
	cmd.Flags().StringSlice("hide", []string{}, "Hide work orders with these statuses (comma-separated: paused,blocked)")

	return cmd
}

func getStatusEmoji(status string) string {
	switch status {
	case "ready":
		return "📦"
	case "paused":
		return "💤"
	case "design":
		return "📐"
	case "implement":
		return "🔨"
	case "deploy":
		return "🚀"
	case "blocked":
		return "🚫"
	case "complete":
		return "✓"
	default:
		return "📦" // default to ready
	}
}
