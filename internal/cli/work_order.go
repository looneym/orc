package cli

import (
	"fmt"

	"github.com/example/orc/internal/context"
	"github.com/example/orc/internal/models"
	"github.com/spf13/cobra"
)

var workOrderCmd = &cobra.Command{
	Use:   "work-order",
	Short: "Manage work orders (individual tasks)",
	Long:  "Create, list, claim, and complete work orders in the ORC ledger",
}

var workOrderCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new work order",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		missionID, _ := cmd.Flags().GetString("mission")
		description, _ := cmd.Flags().GetString("description")
		contextRef, _ := cmd.Flags().GetString("context-ref")
		parentID, _ := cmd.Flags().GetString("parent")

		// Smart default: use deputy context if available, otherwise MISSION-001
		if missionID == "" {
			if ctxMissionID := context.GetContextMissionID(); ctxMissionID != "" {
				missionID = ctxMissionID
				fmt.Printf("ℹ️  Using mission from context: %s\n", missionID)
			} else {
				missionID = "MISSION-001"
			}
		}

		wo, err := models.CreateWorkOrder(missionID, title, description, contextRef, parentID)
		if err != nil {
			return fmt.Errorf("failed to create work order: %w", err)
		}

		fmt.Printf("✓ Created work order %s: %s\n", wo.ID, wo.Title)
		fmt.Printf("  Under mission: %s\n", wo.MissionID)
		if wo.ParentID.Valid {
			fmt.Printf("  Parent: %s\n", wo.ParentID.String)
		}
		if wo.ContextRef.Valid {
			fmt.Printf("  Context: %s\n", wo.ContextRef.String)
		}
		return nil
	},
}

var workOrderListCmd = &cobra.Command{
	Use:   "list",
	Short: "List work orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		missionID, _ := cmd.Flags().GetString("mission")
		status, _ := cmd.Flags().GetString("status")

		// Smart default: use deputy context if available and no mission specified
		if missionID == "" {
			if ctxMissionID := context.GetContextMissionID(); ctxMissionID != "" {
				missionID = ctxMissionID
			}
		}

		orders, err := models.ListWorkOrders(missionID, status)
		if err != nil {
			return fmt.Errorf("failed to list work orders: %w", err)
		}

		if len(orders) == 0 {
			fmt.Println("No work orders found")
			return nil
		}

		fmt.Printf("\n%-15s %-15s %-10s %-15s %s\n", "ID", "MISSION", "STATUS", "GROVE", "TITLE")
		fmt.Println("─────────────────────────────────────────────────────────────────────────────────")
		for _, wo := range orders {
			grove := "-"
			if wo.AssignedGroveID.Valid {
				grove = wo.AssignedGroveID.String
			}
			fmt.Printf("%-15s %-15s %-10s %-15s %s\n", wo.ID, wo.MissionID, wo.Status, grove, wo.Title)
		}
		fmt.Println()

		return nil
	},
}

var workOrderShowCmd = &cobra.Command{
	Use:   "show [work-order-id]",
	Short: "Show work order details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		wo, err := models.GetWorkOrder(id)
		if err != nil {
			return fmt.Errorf("failed to get work order: %w", err)
		}

		fmt.Printf("\nWork Order: %s\n", wo.ID)
		fmt.Printf("Mission:    %s\n", wo.MissionID)
		fmt.Printf("Title:      %s\n", wo.Title)
		fmt.Printf("Status:     %s\n", wo.Status)
		if wo.Type.Valid {
			fmt.Printf("Type:       %s\n", wo.Type.String)
		}
		if wo.Priority.Valid {
			fmt.Printf("Priority:   %s\n", wo.Priority.String)
		}
		if wo.ParentID.Valid {
			fmt.Printf("Parent:     %s\n", wo.ParentID.String)
		}
		if wo.Description.Valid {
			fmt.Printf("Description: %s\n", wo.Description.String)
		}
		if wo.AssignedGroveID.Valid {
			fmt.Printf("Grove:      %s\n", wo.AssignedGroveID.String)
		}
		if wo.ContextRef.Valid {
			fmt.Printf("Context:    %s\n", wo.ContextRef.String)
		}
		fmt.Printf("Created:    %s\n", wo.CreatedAt.Format("2006-01-02 15:04"))
		if wo.ClaimedAt.Valid {
			fmt.Printf("Claimed:    %s\n", wo.ClaimedAt.Time.Format("2006-01-02 15:04"))
		}
		if wo.CompletedAt.Valid {
			fmt.Printf("Completed:  %s\n", wo.CompletedAt.Time.Format("2006-01-02 15:04"))
		}
		fmt.Println()

		return nil
	},
}

var workOrderClaimCmd = &cobra.Command{
	Use:   "claim [work-order-id]",
	Short: "Claim a work order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		groveID, _ := cmd.Flags().GetString("grove")

		err := models.ClaimWorkOrder(id, groveID)
		if err != nil {
			return fmt.Errorf("failed to claim work order: %w", err)
		}

		if groveID != "" {
			fmt.Printf("✓ Work order %s claimed by %s\n", id, groveID)
		} else {
			fmt.Printf("✓ Work order %s claimed by IMP-UNKNOWN\n", id)
		}
		fmt.Printf("  Status: implement\n")
		return nil
	},
}

var workOrderCompleteCmd = &cobra.Command{
	Use:   "complete [work-order-id]",
	Short: "Mark a work order as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		err := models.CompleteWorkOrder(id)
		if err != nil {
			return fmt.Errorf("failed to complete work order: %w", err)
		}

		fmt.Printf("✓ Work order %s marked as complete\n", id)
		return nil
	},
}

var workOrderSetParentCmd = &cobra.Command{
	Use:   "set-parent [work-order-id]",
	Short: "Set or update the parent of a work order",
	Long: `Set or update the parent work order to create an epic hierarchy.

Examples:
  orc work-order set-parent WO-015 --parent WO-010  # Make WO-015 a child of WO-010
  orc work-order set-parent WO-015 --parent ""      # Remove parent (make top-level)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		parentID, _ := cmd.Flags().GetString("parent")

		err := models.SetParentWorkOrder(id, parentID)
		if err != nil {
			return fmt.Errorf("failed to set parent: %w", err)
		}

		if parentID != "" {
			fmt.Printf("✓ Work order %s parent set to %s\n", id, parentID)
		} else {
			fmt.Printf("✓ Work order %s parent removed (now top-level)\n", id)
		}
		return nil
	},
}

var workOrderPinCmd = &cobra.Command{
	Use:   "pin [work-order-id]",
	Short: "Pin a work order to keep it visible",
	Long: `Pin a work order to show it in a special section at the top of the summary.

Useful for long-running epics or important work streams that need to stay visible
even when other work orders are in progress.

Examples:
  orc work-order pin WO-031  # Pin epic
  orc work-order pin WO-061  # Pin important work stream`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		err := models.PinWorkOrder(id)
		if err != nil {
			return fmt.Errorf("failed to pin work order: %w", err)
		}

		fmt.Printf("📌 Work order %s pinned\n", id)
		fmt.Printf("  Will appear in pinned section at top of summary\n")
		return nil
	},
}

var workOrderUnpinCmd = &cobra.Command{
	Use:   "unpin [work-order-id]",
	Short: "Unpin a work order",
	Long: `Unpin a work order to remove it from the pinned section.

Examples:
  orc work-order unpin WO-031`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		err := models.UnpinWorkOrder(id)
		if err != nil {
			return fmt.Errorf("failed to unpin work order: %w", err)
		}

		fmt.Printf("✓ Work order %s unpinned\n", id)
		return nil
	},
}

var workOrderUpdateCmd = &cobra.Command{
	Use:   "update [work-order-id]",
	Short: "Update work order title and/or description",
	Long: `Update the title and/or description of an existing work order.

Examples:
  orc work-order update WO-042 --title "New Title"
  orc work-order update WO-042 --description "New description"
  orc work-order update WO-042 --title "New Title" --description "New description"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		if title == "" && description == "" {
			return fmt.Errorf("must specify at least --title or --description")
		}

		err := models.UpdateWorkOrder(id, title, description)
		if err != nil {
			return fmt.Errorf("failed to update work order: %w", err)
		}

		fmt.Printf("✓ Work order %s updated\n", id)
		if title != "" {
			fmt.Printf("  Title: %s\n", title)
		}
		if description != "" {
			fmt.Printf("  Description: %s\n", description)
		}
		return nil
	},
}

var workOrderDiscoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover ready work orders and claim them",
	Long: `Discover work orders that are ready to be claimed in the current mission.

This command helps deputies autonomously find and claim work. It:
1. Lists all ready work orders in the current mission
2. Shows priorities and determines if grove/IMP is needed
3. Allows interactive or automatic claiming
4. Provides next steps for execution

Examples:
  orc work-order discover              # Interactive discovery
  orc work-order discover --claim WO-140  # Directly claim specific work order
  orc work-order discover --auto-claim    # Automatically claim highest priority`,
	RunE: func(cmd *cobra.Command, args []string) error {
		autoClaim, _ := cmd.Flags().GetBool("auto-claim")
		claimID, _ := cmd.Flags().GetString("claim")

		// Get current mission context
		missionID := context.GetContextMissionID()
		if missionID == "" {
			return fmt.Errorf("no mission context detected - run from a mission or grove directory")
		}

		fmt.Printf("🔍 Discovering work for %s...\n\n", missionID)

		// List ready work orders
		orders, err := models.ListWorkOrders(missionID, "ready")
		if err != nil {
			return fmt.Errorf("failed to list work orders: %w", err)
		}

		if len(orders) == 0 {
			fmt.Println("✓ No work available")
			return nil
		}

		// Display ready work orders
		fmt.Printf("📦 Ready Work Orders (%d):\n", len(orders))
		for _, wo := range orders {
			needsGrove := classifyWorkOrder(wo)
			groveIndicator := ""
			if needsGrove {
				groveIndicator = " [needs grove]"
			} else {
				groveIndicator = " [local]"
			}

			priority := "medium"
			if wo.Priority.Valid {
				priority = wo.Priority.String
			}

			fmt.Printf("  %s: %s [priority: %s]%s\n", wo.ID, wo.Title, priority, groveIndicator)
		}
		fmt.Println()

		// Handle direct claim
		if claimID != "" {
			return claimAndReportWorkOrder(claimID)
		}

		// Handle auto-claim
		if autoClaim {
			// Claim highest priority (first in list)
			highestPriority := orders[0]
			fmt.Printf("🎯 Auto-claiming: %s\n\n", highestPriority.ID)
			return claimAndReportWorkOrder(highestPriority.ID)
		}

		// Interactive mode - prompt to claim
		fmt.Printf("💡 To claim a work order:\n")
		fmt.Printf("   orc work-order discover --claim %s\n", orders[0].ID)
		fmt.Printf("   orc work-order claim %s\n", orders[0].ID)
		fmt.Println()
		fmt.Printf("💡 To auto-claim highest priority:\n")
		fmt.Printf("   orc work-order discover --auto-claim\n")
		fmt.Println()

		return nil
	},
}

// classifyWorkOrder determines if a work order needs a grove/IMP
func classifyWorkOrder(wo *models.WorkOrder) bool {
	// Simple heuristic based on keywords in title
	implementationKeywords := []string{
		"implement", "add", "create", "build", "write",
		"refactor", "fix", "update", "modify", "develop",
	}

	titleLower := fmt.Sprintf("%s %s", wo.Title, wo.Description.String)
	titleLower = fmt.Sprintf("%s", titleLower) // Make lowercase comparison

	for _, keyword := range implementationKeywords {
		if containsIgnoreCase(titleLower, keyword) {
			return true
		}
	}

	// Check for file patterns in description
	if wo.Description.Valid {
		filePatterns := []string{".go", ".py", ".js", ".ts", ".java", ".rb", ".c", ".cpp"}
		for _, pattern := range filePatterns {
			if containsIgnoreCase(wo.Description.String, pattern) {
				return true
			}
		}
	}

	return false
}

// containsIgnoreCase checks if a string contains a substring case-insensitively
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && findIgnoreCase(s, substr) >= 0
}

// findIgnoreCase finds a substring in a string case-insensitively
func findIgnoreCase(s, substr string) int {
	sLen := len(s)
	subLen := len(substr)
	if subLen > sLen {
		return -1
	}
	for i := 0; i <= sLen-subLen; i++ {
		match := true
		for j := 0; j < subLen; j++ {
			if toLower(s[i+j]) != toLower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// toLower converts a byte to lowercase
func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

// claimAndReportWorkOrder claims a work order and reports next steps
func claimAndReportWorkOrder(id string) error {
	// Get work order details
	wo, err := models.GetWorkOrder(id)
	if err != nil {
		return fmt.Errorf("failed to get work order: %w", err)
	}

	// Claim it
	err = models.ClaimWorkOrder(id, "")
	if err != nil {
		return fmt.Errorf("failed to claim work order: %w", err)
	}

	fmt.Printf("✓ Claimed %s: %s\n", wo.ID, wo.Title)
	fmt.Printf("  Status: implement\n\n")

	// Determine next steps
	needsGrove := classifyWorkOrder(wo)
	if needsGrove {
		fmt.Printf("🌳 This work order requires code changes (needs grove/IMP)\n\n")
		fmt.Printf("Next steps:\n")
		fmt.Printf("  1. Assign to IMP: orc assign %s --grove <grove-id>\n", wo.ID)
		fmt.Printf("  2. Or create grove: orc grove create <name> --mission %s\n", wo.MissionID)
		fmt.Printf("  3. IMP will implement the changes\n")
		fmt.Println()
		fmt.Printf("⚠️  Note: IMP assignment protocol (WO-145) not yet implemented\n")
		fmt.Printf("   For now, manually switch to grove window and work on %s\n", wo.ID)
	} else {
		fmt.Printf("📝 This work order can be executed locally (no grove needed)\n\n")
		fmt.Printf("Next steps:\n")
		fmt.Printf("  1. Execute the work order tasks\n")
		fmt.Printf("  2. Mark complete: orc work-order complete %s\n", wo.ID)
	}
	fmt.Println()

	return nil
}

// WorkOrderCmd returns the work-order command
func WorkOrderCmd() *cobra.Command {
	// Add flags
	workOrderCreateCmd.Flags().StringP("mission", "m", "", "Mission ID (defaults to MISSION-001)")
	workOrderCreateCmd.Flags().StringP("description", "d", "", "Work order description")
	workOrderCreateCmd.Flags().StringP("context-ref", "c", "", "Graphiti context reference (e.g., graphiti:episode-uuid)")
	workOrderCreateCmd.Flags().StringP("parent", "p", "", "Parent work order ID (for creating sub-tasks)")

	workOrderListCmd.Flags().StringP("mission", "m", "", "Filter by mission ID")
	workOrderListCmd.Flags().StringP("status", "s", "", "Filter by status (ready, design, implement, deploy, blocked, paused, complete)")

	workOrderClaimCmd.Flags().StringP("grove", "g", "", "Grove ID claiming the work order")

	workOrderSetParentCmd.Flags().StringP("parent", "p", "", "Parent work order ID (empty string to remove parent)")
	workOrderSetParentCmd.MarkFlagRequired("parent")

	workOrderUpdateCmd.Flags().StringP("title", "t", "", "New title for work order")
	workOrderUpdateCmd.Flags().StringP("description", "d", "", "New description for work order")

	workOrderDiscoverCmd.Flags().BoolP("auto-claim", "a", false, "Automatically claim highest priority work order")
	workOrderDiscoverCmd.Flags().StringP("claim", "c", "", "Directly claim specific work order ID")

	// Add subcommands
	workOrderCmd.AddCommand(workOrderCreateCmd)
	workOrderCmd.AddCommand(workOrderListCmd)
	workOrderCmd.AddCommand(workOrderShowCmd)
	workOrderCmd.AddCommand(workOrderClaimCmd)
	workOrderCmd.AddCommand(workOrderCompleteCmd)
	workOrderCmd.AddCommand(workOrderSetParentCmd)
	workOrderCmd.AddCommand(workOrderPinCmd)
	workOrderCmd.AddCommand(workOrderUnpinCmd)
	workOrderCmd.AddCommand(workOrderUpdateCmd)
	workOrderCmd.AddCommand(workOrderDiscoverCmd)

	return workOrderCmd
}
