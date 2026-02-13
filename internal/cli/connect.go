package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// ConnectCmd returns the connect command
func ConnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Launch Claude agent with boot instructions",
		Long: `Launch Claude Code with the /orc-prime skill for context bootstrapping.

This command is designed to be the root command for agent TMux panes, ensuring
that every time a pane spawns or respawns, Claude boots with proper context.

The boot sequence:
  1. orc connect (launches claude)
  2. SessionStart hook injects ORC context automatically
  3. Claude runs /orc-prime skill for full orientation
  4. Agent is ready to work

Usage:
  orc connect                    # Launch Claude (role from place_id)

TMux Integration:
  # Set as pane root command
  tmux send-keys -t session:window.pane "orc connect" C-m

  # Respawn pane with orc connect
  tmux respawn-pane -t session:window.pane -k "orc connect"`,
		RunE: runConnect,
	}

	cmd.Flags().Bool("dry-run", false, "Show command that would be executed without running it")

	return cmd
}

func runConnect(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	cwd, _ := os.Getwd()

	// The prime directive: Claude must run /orc-prime skill upon boot
	// The SessionStart hook also injects context, but /orc-prime ensures full orientation
	primeDirective := "Run /orc-prime to bootstrap project context. Do not greet the user first - just run the skill immediately."

	// Build claude command
	// Using "claude" assumes it's in PATH (standard Claude Code installation)
	claudeArgs := []string{primeDirective}
	claudeCmd := exec.Command("claude", claudeArgs...)

	// Pass through stdio for interactive session
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr

	// Set working directory to current directory
	claudeCmd.Dir = cwd

	if dryRun {
		fmt.Printf("Would execute: claude %q\n", primeDirective)
		fmt.Printf("Working directory: %s\n", claudeCmd.Dir)
		return nil
	}

	// Launch Claude
	return claudeCmd.Run()
}
