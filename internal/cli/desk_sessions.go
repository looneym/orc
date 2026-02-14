package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

// DeskCmd returns the desk command
func DeskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "desk",
		Short: "Manage desk tmux servers",
		Long: `Manage the per-workbench desk tmux servers.

Each workbench can have a desk server (socket: {bench-name}-desk) that hosts
the ORC summary dashboard and scratch shell in a popup overlay.

Commands:
  list                   List all desk servers and their status
  kill <bench-name>      Kill a specific desk server
  kill --all             Kill all desk servers
  review <note-id>       Open a note for interactive review in vim

Examples:
  orc desk list
  orc desk kill orc-45
  orc desk kill --all
  orc desk review NOTE-123`,
	}

	cmd.AddCommand(deskListCmd())
	cmd.AddCommand(deskKillCmd())
	cmd.AddCommand(deskReviewCmd())

	return cmd
}

func deskListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all desk tmux servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			servers, err := wire.ListDeskServers()
			if err != nil {
				return fmt.Errorf("failed to scan desk servers: %w", err)
			}

			if len(servers) == 0 {
				fmt.Println("No desk servers found")
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

func deskKillCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "kill [bench-name]",
		Short: "Kill desk tmux server(s)",
		Long: `Kill a specific desk server by workbench name, or all with --all.

Examples:
  orc desk kill orc-45
  orc desk kill --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if all {
				killed, err := wire.KillAllDeskServers()
				if err != nil {
					return fmt.Errorf("failed to kill desk servers: %w", err)
				}
				fmt.Printf("Killed %d desk server(s)\n", killed)
				return nil
			}

			if len(args) == 0 {
				return fmt.Errorf("specify a workbench name or use --all")
			}

			benchName := args[0]
			if err := wire.KillDeskServer(benchName); err != nil {
				return err
			}
			fmt.Printf("Killed desk server for %s\n", benchName)
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Kill all desk servers")

	return cmd
}

func deskReviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "review [note-id]",
		Short: "Open a note for interactive review in vim",
		Long: `Open a note's content in vim for editing. On save and quit,
the updated content is persisted back to the database and an operational
event is emitted so automated processes can detect review completion.

This command is typically invoked by the desk TUI (via 'd' keybind) in an
ephemeral tmux window. It can also be run directly from the command line.

Examples:
  orc desk review NOTE-123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noteID := args[0]
			if !strings.HasPrefix(noteID, "NOTE-") {
				return fmt.Errorf("expected a NOTE-xxx ID, got: %s", noteID)
			}

			ctx := context.Background()

			// Fetch current note content
			note, err := wire.NoteService().GetNote(ctx, noteID)
			if err != nil {
				return fmt.Errorf("note not found: %w", err)
			}

			// Write content to temp file
			tmpFile, err := os.CreateTemp("", fmt.Sprintf("orc-review-%s-*.md", noteID))
			if err != nil {
				return fmt.Errorf("failed to create temp file: %w", err)
			}
			tmpPath := tmpFile.Name()
			defer os.Remove(tmpPath)

			// Write title as first line, then blank line, then content
			reviewContent := fmt.Sprintf("# %s\n\n%s", note.Title, note.Content)
			if _, err := tmpFile.WriteString(reviewContent); err != nil {
				tmpFile.Close()
				return fmt.Errorf("failed to write temp file: %w", err)
			}
			tmpFile.Close()

			// Record file checksum before editing to detect changes
			beforeContent, _ := os.ReadFile(tmpPath)

			// Open in vim
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vim"
			}
			editorCmd := exec.Command(editor, tmpPath)
			editorCmd.Stdin = os.Stdin
			editorCmd.Stdout = os.Stdout
			editorCmd.Stderr = os.Stderr
			if err := editorCmd.Run(); err != nil {
				return fmt.Errorf("editor exited with error: %w", err)
			}

			// Read back edited content
			afterContent, err := os.ReadFile(tmpPath)
			if err != nil {
				return fmt.Errorf("failed to read edited file: %w", err)
			}

			// Check if content changed
			if string(beforeContent) == string(afterContent) {
				fmt.Println("No changes made.")
				return nil
			}

			// Parse title and content back from the edited file
			// Format: "# Title\n\nContent..."
			newTitle, newContent := parseReviewFile(string(afterContent))

			// Update the note
			updateReq := primary.UpdateNoteRequest{
				NoteID: noteID,
			}
			if newTitle != note.Title {
				updateReq.Title = newTitle
			}
			if newContent != note.Content {
				updateReq.Content = newContent
			}

			if updateReq.Title != "" || updateReq.Content != "" {
				if err := wire.NoteService().UpdateNote(ctx, updateReq); err != nil {
					return fmt.Errorf("failed to update note: %w", err)
				}
				fmt.Printf("Updated %s\n", noteID)
			}

			// Emit operational event for review completion
			_ = wire.EventWriter().EmitOperational(ctx,
				"desk", "info",
				fmt.Sprintf("Review completed: %s", noteID),
				map[string]string{
					"entity_id": noteID,
					"action":    "review.complete",
				},
			)

			return nil
		},
	}
}

// parseReviewFile extracts title and content from the review file format.
// Expected format: "# Title\n\nContent..."
// Falls back to using entire content if no markdown heading found.
func parseReviewFile(text string) (string, string) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "# ") {
		return "", text
	}

	// Split at first blank line to separate title from content
	parts := strings.SplitN(text, "\n\n", 2)
	title := strings.TrimPrefix(parts[0], "# ")
	title = strings.TrimSpace(title)

	content := ""
	if len(parts) > 1 {
		content = strings.TrimSpace(parts[1])
	}

	return title, content
}
