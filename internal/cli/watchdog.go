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

var watchdogCmd = &cobra.Command{
	Use:   "watchdog",
	Short: "Manage watchdogs",
	Long:  "List and view watchdogs (IMP monitors) in the ORC ledger",
}

var watchdogListCmd = &cobra.Command{
	Use:   "list",
	Short: "List watchdogs",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		workbenchID, _ := cmd.Flags().GetString("workbench")
		status, _ := cmd.Flags().GetString("status")

		watchdogs, err := wire.WatchdogService().ListWatchdogs(ctx, primary.WatchdogFilters{
			WorkbenchID: workbenchID,
			Status:      status,
		})
		if err != nil {
			return fmt.Errorf("failed to list watchdogs: %w", err)
		}

		if len(watchdogs) == 0 {
			fmt.Println("No watchdogs found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tWORKBENCH\tSTATUS\tCREATED")
		fmt.Fprintln(w, "--\t---------\t------\t-------")
		for _, item := range watchdogs {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				item.ID,
				item.WorkbenchID,
				item.Status,
				item.CreatedAt,
			)
		}
		w.Flush()
		return nil
	},
}

var watchdogShowCmd = &cobra.Command{
	Use:   "show [watchdog-id]",
	Short: "Show watchdog details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		watchdogID := args[0]

		watchdog, err := wire.WatchdogService().GetWatchdog(ctx, watchdogID)
		if err != nil {
			return fmt.Errorf("watchdog not found: %w", err)
		}

		fmt.Printf("Watchdog: %s\n", watchdog.ID)
		fmt.Printf("Workbench: %s\n", watchdog.WorkbenchID)
		fmt.Printf("Status: %s\n", watchdog.Status)
		fmt.Printf("Created: %s\n", watchdog.CreatedAt)
		fmt.Printf("Updated: %s\n", watchdog.UpdatedAt)

		return nil
	},
}

func init() {
	// watchdog list flags
	watchdogListCmd.Flags().String("workbench", "", "Filter by workbench ID")
	watchdogListCmd.Flags().StringP("status", "s", "", "Filter by status (active|inactive)")

	// Register subcommands
	watchdogCmd.AddCommand(watchdogListCmd)
	watchdogCmd.AddCommand(watchdogShowCmd)
}

// WatchdogCmd returns the watchdog command
func WatchdogCmd() *cobra.Command {
	return watchdogCmd
}
