package cli

import (
	"fmt"

	"github.com/looneym/orc/internal/models"
	"github.com/spf13/cobra"
)

var missionCmd = &cobra.Command{
	Use:   "mission",
	Short: "Manage missions (strategic work streams)",
	Long:  "Create, list, and manage missions in the ORC ledger",
}

var missionCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new mission",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		description, _ := cmd.Flags().GetString("description")

		mission, err := models.CreateMission(title, description)
		if err != nil {
			return fmt.Errorf("failed to create mission: %w", err)
		}

		fmt.Printf("✓ Created mission %s: %s\n", mission.ID, mission.Title)
		return nil
	},
}

var missionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List missions",
	RunE: func(cmd *cobra.Command, args []string) error {
		status, _ := cmd.Flags().GetString("status")

		missions, err := models.ListMissions(status)
		if err != nil {
			return fmt.Errorf("failed to list missions: %w", err)
		}

		if len(missions) == 0 {
			fmt.Println("No missions found")
			return nil
		}

		fmt.Printf("\n%-15s %-10s %s\n", "ID", "STATUS", "TITLE")
		fmt.Println("────────────────────────────────────────────────────────────────")
		for _, m := range missions {
			fmt.Printf("%-15s %-10s %s\n", m.ID, m.Status, m.Title)
		}
		fmt.Println()

		return nil
	},
}

var missionShowCmd = &cobra.Command{
	Use:   "show [mission-id]",
	Short: "Show mission details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		mission, err := models.GetMission(id)
		if err != nil {
			return fmt.Errorf("failed to get mission: %w", err)
		}

		fmt.Printf("\nMission: %s\n", mission.ID)
		fmt.Printf("Title:   %s\n", mission.Title)
		fmt.Printf("Status:  %s\n", mission.Status)
		if mission.Description.Valid {
			fmt.Printf("Description: %s\n", mission.Description.String)
		}
		fmt.Printf("Created: %s\n", mission.CreatedAt.Format("2006-01-02 15:04"))
		if mission.CompletedAt.Valid {
			fmt.Printf("Completed: %s\n", mission.CompletedAt.Time.Format("2006-01-02 15:04"))
		}
		fmt.Println()

		// List operations under this mission
		operations, err := models.ListOperations(id, "")
		if err == nil && len(operations) > 0 {
			fmt.Println("Operations:")
			for _, op := range operations {
				fmt.Printf("  - %s [%s] %s\n", op.ID, op.Status, op.Title)
			}
			fmt.Println()
		}

		return nil
	},
}

var missionCompleteCmd = &cobra.Command{
	Use:   "complete [mission-id]",
	Short: "Mark a mission as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		err := models.UpdateMissionStatus(id, "complete")
		if err != nil {
			return fmt.Errorf("failed to complete mission: %w", err)
		}

		fmt.Printf("✓ Mission %s marked as complete\n", id)
		return nil
	},
}

var missionUpdateCmd = &cobra.Command{
	Use:   "update [mission-id]",
	Short: "Update mission title and/or description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		if title == "" && description == "" {
			return fmt.Errorf("must specify at least --title or --description")
		}

		err := models.UpdateMission(id, title, description)
		if err != nil {
			return fmt.Errorf("failed to update mission: %w", err)
		}

		fmt.Printf("✓ Mission %s updated\n", id)
		return nil
	},
}

// MissionCmd returns the mission command
func MissionCmd() *cobra.Command {
	// Add flags
	missionCreateCmd.Flags().StringP("description", "d", "", "Mission description")
	missionListCmd.Flags().StringP("status", "s", "", "Filter by status (active, paused, complete, archived)")
	missionUpdateCmd.Flags().StringP("title", "t", "", "New mission title")
	missionUpdateCmd.Flags().StringP("description", "d", "", "New mission description")

	// Add subcommands
	missionCmd.AddCommand(missionCreateCmd)
	missionCmd.AddCommand(missionListCmd)
	missionCmd.AddCommand(missionShowCmd)
	missionCmd.AddCommand(missionCompleteCmd)
	missionCmd.AddCommand(missionUpdateCmd)

	return missionCmd
}
