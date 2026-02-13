package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/wire"
)

// UtilsSessionsCmd returns the utils-sessions command
func UtilsSessionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "utils-sessions",
		Short: "Manage utils tmux servers",
		Long: `Manage the per-workbench utils tmux servers.

Each workbench can have a utils server (socket: {bench-name}-utils) that hosts
the ORC summary dashboard and scratch shell in a popup overlay.

Commands:
  list                   List all utils servers and their status
  kill <bench-name>      Kill a specific utils server
  kill --all             Kill all utils servers

Examples:
  orc utils-sessions list
  orc utils-sessions kill orc-45
  orc utils-sessions kill --all`,
	}

	cmd.AddCommand(utilsSessionsListCmd())
	cmd.AddCommand(utilsSessionsKillCmd())

	return cmd
}

func utilsSessionsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all utils tmux servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			servers, err := wire.ListUtilsServers()
			if err != nil {
				return fmt.Errorf("failed to scan utils servers: %w", err)
			}

			if len(servers) == 0 {
				fmt.Println("No utils servers found")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "WORKBENCH\tSOCKET\tSTATUS")
			for _, s := range servers {
				status := "dead"
				if s.Alive {
					status = "alive"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", s.BenchName, s.Socket, status)
			}
			return w.Flush()
		},
	}
}

func utilsSessionsKillCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "kill [bench-name]",
		Short: "Kill utils tmux server(s)",
		Long: `Kill a specific utils server by workbench name, or all with --all.

Examples:
  orc utils-sessions kill orc-45
  orc utils-sessions kill --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if all {
				killed, err := wire.KillAllUtilsServers()
				if err != nil {
					return fmt.Errorf("failed to kill utils servers: %w", err)
				}
				fmt.Printf("Killed %d utils server(s)\n", killed)
				return nil
			}

			if len(args) == 0 {
				return fmt.Errorf("specify a workbench name or use --all")
			}

			benchName := args[0]
			if err := wire.KillUtilsServer(benchName); err != nil {
				return err
			}
			fmt.Printf("Killed utils server for %s\n", benchName)
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Kill all utils servers")

	return cmd
}
