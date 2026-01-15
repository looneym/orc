package cli

import (
	"fmt"

	"github.com/example/orc/internal/models"
	"github.com/spf13/cobra"
)

var rabbitHoleCmd = &cobra.Command{
	Use:   "rabbit-hole",
	Short: "Manage rabbit holes (grouping layer within epics)",
	Long:  "Create, list, and manage rabbit holes in the ORC ledger",
}

var rabbitHoleCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new rabbit hole under an epic",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		epicID, _ := cmd.Flags().GetString("epic")
		description, _ := cmd.Flags().GetString("description")

		if epicID == "" {
			return fmt.Errorf("--epic flag is required")
		}

		rh, err := models.CreateRabbitHole(epicID, title, description)
		if err != nil {
			return fmt.Errorf("failed to create rabbit hole: %w", err)
		}

		fmt.Printf("✓ Created rabbit hole %s: %s\n", rh.ID, rh.Title)
		fmt.Printf("  Under epic: %s\n", rh.EpicID)
		fmt.Println()
		fmt.Println("💡 Next step:")
		fmt.Printf("   orc task create \"Task title\" --rabbit-hole %s\n", rh.ID)
		return nil
	},
}

var rabbitHoleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List rabbit holes",
	RunE: func(cmd *cobra.Command, args []string) error {
		epicID, _ := cmd.Flags().GetString("epic")
		status, _ := cmd.Flags().GetString("status")

		rhs, err := models.ListRabbitHoles(epicID, status)
		if err != nil {
			return fmt.Errorf("failed to list rabbit holes: %w", err)
		}

		if len(rhs) == 0 {
			fmt.Println("No rabbit holes found")
			return nil
		}

		fmt.Printf("Found %d rabbit hole(s):\n\n", len(rhs))
		for _, rh := range rhs {
			statusIcon := getStatusIcon(rh.Status)
			pinnedIcon := ""
			if rh.Pinned {
				pinnedIcon = " 📌"
			}
			fmt.Printf("%s %s: %s [%s]%s\n", statusIcon, rh.ID, rh.Title, rh.Status, pinnedIcon)
			fmt.Printf("   Epic: %s\n", rh.EpicID)

			// Show task count
			tasks, _ := models.GetRabbitHoleTasks(rh.ID)
			fmt.Printf("   Tasks: %d\n", len(tasks))
			fmt.Println()
		}
		return nil
	},
}

var rabbitHoleShowCmd = &cobra.Command{
	Use:   "show [rabbit-hole-id]",
	Short: "Show rabbit hole details with tasks",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rhID := args[0]

		rh, err := models.GetRabbitHole(rhID)
		if err != nil {
			return fmt.Errorf("rabbit hole not found: %w", err)
		}

		// Display rabbit hole details
		fmt.Printf("Rabbit Hole: %s\n", rh.ID)
		fmt.Printf("Title: %s\n", rh.Title)
		if rh.Description.Valid {
			fmt.Printf("Description: %s\n", rh.Description.String)
		}
		fmt.Printf("Status: %s\n", rh.Status)
		fmt.Printf("Epic: %s\n", rh.EpicID)
		if rh.Priority.Valid {
			fmt.Printf("Priority: %s\n", rh.Priority.String)
		}
		if rh.Pinned {
			fmt.Printf("Pinned: yes\n")
		}
		fmt.Printf("Created: %s\n", rh.CreatedAt.Format("2006-01-02 15:04"))
		if rh.CompletedAt.Valid {
			fmt.Printf("Completed: %s\n", rh.CompletedAt.Time.Format("2006-01-02 15:04"))
		}
		fmt.Println()

		// Display tasks
		tasks, err := models.GetRabbitHoleTasks(rh.ID)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}

		fmt.Printf("Tasks (%d):\n", len(tasks))
		for _, task := range tasks {
			statusIcon := getStatusIcon(task.Status)
			fmt.Printf("  %s %s: %s [%s]\n", statusIcon, task.ID, task.Title, task.Status)
		}

		return nil
	},
}

var rabbitHoleCompleteCmd = &cobra.Command{
	Use:   "complete [rabbit-hole-id]",
	Short: "Mark rabbit hole as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rhID := args[0]

		err := models.CompleteRabbitHole(rhID)
		if err != nil {
			return fmt.Errorf("failed to complete rabbit hole: %w", err)
		}

		fmt.Printf("✓ Rabbit hole %s marked as complete\n", rhID)
		return nil
	},
}

var rabbitHoleUpdateCmd = &cobra.Command{
	Use:   "update [rabbit-hole-id]",
	Short: "Update rabbit hole title and/or description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rhID := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		if title == "" && description == "" {
			return fmt.Errorf("must specify --title and/or --description")
		}

		err := models.UpdateRabbitHole(rhID, title, description)
		if err != nil {
			return fmt.Errorf("failed to update rabbit hole: %w", err)
		}

		fmt.Printf("✓ Rabbit hole %s updated\n", rhID)
		return nil
	},
}

func init() {
	// rabbit-hole create flags
	rabbitHoleCreateCmd.Flags().String("epic", "", "Epic ID (required)")
	rabbitHoleCreateCmd.MarkFlagRequired("epic")
	rabbitHoleCreateCmd.Flags().StringP("description", "d", "", "Rabbit hole description")

	// rabbit-hole list flags
	rabbitHoleListCmd.Flags().String("epic", "", "Filter by epic")
	rabbitHoleListCmd.Flags().StringP("status", "s", "", "Filter by status")

	// rabbit-hole update flags
	rabbitHoleUpdateCmd.Flags().String("title", "", "New title")
	rabbitHoleUpdateCmd.Flags().StringP("description", "d", "", "New description")

	// Register subcommands
	rabbitHoleCmd.AddCommand(rabbitHoleCreateCmd)
	rabbitHoleCmd.AddCommand(rabbitHoleListCmd)
	rabbitHoleCmd.AddCommand(rabbitHoleShowCmd)
	rabbitHoleCmd.AddCommand(rabbitHoleCompleteCmd)
	rabbitHoleCmd.AddCommand(rabbitHoleUpdateCmd)
}

// RabbitHoleCmd returns the rabbit-hole command
func RabbitHoleCmd() *cobra.Command {
	return rabbitHoleCmd
}
