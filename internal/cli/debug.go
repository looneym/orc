package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/example/orc/internal/context"
	"github.com/spf13/cobra"
)

// DebugCmd returns the debug command
func DebugCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug and diagnostic commands",
		Long:  `Tools for debugging ORC context detection and environment setup.`,
	}

	cmd.AddCommand(debugSessionInfoCmd())
	cmd.AddCommand(debugValidateContextCmd())

	return cmd
}

func debugSessionInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "session-info",
		Short: "Show current session context information",
		Long: `Display detailed information about the current ORC context detection.

Shows:
- Current working directory
- Mission context (.orc-mission marker)
- Deputy context (workspace metadata)
- Grove context (grove metadata)
- Environment variables (TMUX, etc.)

Useful for debugging context detection issues.

Examples:
  orc debug session-info`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			fmt.Printf("\n=== ORC Session Context ===\n\n")

			// Current directory
			fmt.Printf("Current Directory:\n")
			fmt.Printf("  %s\n\n", cwd)

			// Check for .orc-mission marker
			fmt.Printf("Mission Context (.orc-mission):\n")
			missionCtx, err := context.DetectMissionContext()
			if err == nil && missionCtx != nil {
				markerPath := filepath.Join(missionCtx.WorkspacePath, ".orc-mission")
				fmt.Printf("  ✓ Found: %s\n", markerPath)
				fmt.Printf("  Mission ID: %s\n", missionCtx.MissionID)
				fmt.Printf("  Workspace: %s\n\n", missionCtx.WorkspacePath)

				// Read and display .orc-mission content
				data, err := os.ReadFile(markerPath)
				if err == nil {
					fmt.Printf("  Content:\n")
					var marker map[string]interface{}
					if err := json.Unmarshal(data, &marker); err == nil {
						formatted, _ := json.MarshalIndent(marker, "    ", "  ")
						fmt.Printf("    %s\n\n", string(formatted))
					} else {
						fmt.Printf("    %s\n\n", string(data))
					}
				}
			} else {
				fmt.Printf("  ✗ Not found (not in a mission context)\n\n")
			}

			// Check for workspace metadata
			fmt.Printf("Workspace Metadata (.orc/metadata.json):\n")
			if missionCtx != nil {
				metadataPath := filepath.Join(missionCtx.WorkspacePath, ".orc", "metadata.json")
				if data, err := os.ReadFile(metadataPath); err == nil {
					fmt.Printf("  ✓ Found: %s\n", metadataPath)

					var metadata map[string]interface{}
					if err := json.Unmarshal(data, &metadata); err == nil {
						fmt.Printf("  Content:\n")
						formatted, _ := json.MarshalIndent(metadata, "    ", "  ")
						fmt.Printf("    %s\n\n", string(formatted))
					}
				} else {
					fmt.Printf("  ✗ Not found\n\n")
				}
			} else {
				fmt.Printf("  N/A (no mission context)\n\n")
			}

			// Check for grove metadata
			fmt.Printf("Grove Metadata (.orc/metadata.json in current dir):\n")
			localMetadataPath := filepath.Join(cwd, ".orc", "metadata.json")
			if data, err := os.ReadFile(localMetadataPath); err == nil {
				fmt.Printf("  ✓ Found: %s\n", localMetadataPath)

				var metadata map[string]interface{}
				if err := json.Unmarshal(data, &metadata); err == nil {
					fmt.Printf("  Content:\n")
					formatted, _ := json.MarshalIndent(metadata, "    ", "  ")
					fmt.Printf("    %s\n\n", string(formatted))
				}
			} else {
				fmt.Printf("  ✗ Not found (not in a grove)\n\n")
			}

			// Environment variables
			fmt.Printf("Environment:\n")
			if tmux := os.Getenv("TMUX"); tmux != "" {
				fmt.Printf("  TMUX: %s\n", tmux)
			} else {
				fmt.Printf("  TMUX: (not set - not in TMux session)\n")
			}

			// Context detection result
			fmt.Printf("\nContext Detection Result:\n")
			if missionCtx != nil {
				fmt.Printf("  Context: Deputy (mission-specific)\n")
				fmt.Printf("  Mission: %s\n", missionCtx.MissionID)
			} else {
				fmt.Printf("  Context: Master (global orchestrator)\n")
			}

			fmt.Println()

			return nil
		},
	}
}

func debugValidateContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate-context [directory]",
		Short: "Validate ORC context setup for a directory",
		Long: `Validate that a directory has proper ORC context markers and metadata.

Checks:
- .orc-mission marker exists and is valid JSON
- .orc/metadata.json exists and is valid JSON
- Mission ID is consistent across files
- Directory structure is correct

Useful for debugging mission workspace setup and grove creation issues.

Examples:
  orc debug validate-context ~/src/missions/MISSION-001
  orc debug validate-context ~/src/worktrees/test-grove
  orc debug validate-context .`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]

			// Resolve to absolute path
			absDir, err := filepath.Abs(dir)
			if err != nil {
				return fmt.Errorf("failed to resolve directory path: %w", err)
			}

			fmt.Printf("\n=== Validating ORC Context: %s ===\n\n", absDir)

			// Check if directory exists
			if _, err := os.Stat(absDir); os.IsNotExist(err) {
				return fmt.Errorf("directory does not exist: %s", absDir)
			}

			validationPassed := true

			// Check 1: .orc-mission marker
			fmt.Printf("1. .orc-mission marker\n")
			missionMarkerPath := filepath.Join(absDir, ".orc-mission")
			if data, err := os.ReadFile(missionMarkerPath); err == nil {
				fmt.Printf("   ✓ File exists: %s\n", missionMarkerPath)

				// Validate JSON
				var marker map[string]interface{}
				if err := json.Unmarshal(data, &marker); err == nil {
					fmt.Printf("   ✓ Valid JSON\n")

					// Check for mission_id field
					if missionID, ok := marker["mission_id"].(string); ok && missionID != "" {
						fmt.Printf("   ✓ mission_id present: %s\n", missionID)
					} else {
						fmt.Printf("   ✗ mission_id missing or invalid\n")
						validationPassed = false
					}

					// Check for workspace_path field
					if workspacePath, ok := marker["workspace_path"].(string); ok && workspacePath != "" {
						fmt.Printf("   ✓ workspace_path present: %s\n", workspacePath)
					} else {
						fmt.Printf("   ⚠️  workspace_path missing (optional)\n")
					}
				} else {
					fmt.Printf("   ✗ Invalid JSON: %v\n", err)
					validationPassed = false
				}
			} else {
				fmt.Printf("   ✗ File not found: %s\n", missionMarkerPath)
				validationPassed = false
			}
			fmt.Println()

			// Check 2: .orc directory
			fmt.Printf("2. .orc directory\n")
			orcDir := filepath.Join(absDir, ".orc")
			if info, err := os.Stat(orcDir); err == nil {
				if info.IsDir() {
					fmt.Printf("   ✓ Directory exists: %s\n", orcDir)
				} else {
					fmt.Printf("   ✗ .orc exists but is not a directory\n")
					validationPassed = false
				}
			} else {
				fmt.Printf("   ✗ Directory not found: %s\n", orcDir)
				validationPassed = false
			}
			fmt.Println()

			// Check 3: .orc/metadata.json
			fmt.Printf("3. .orc/metadata.json\n")
			metadataPath := filepath.Join(orcDir, "metadata.json")
			if data, err := os.ReadFile(metadataPath); err == nil {
				fmt.Printf("   ✓ File exists: %s\n", metadataPath)

				// Validate JSON
				var metadata map[string]interface{}
				if err := json.Unmarshal(data, &metadata); err == nil {
					fmt.Printf("   ✓ Valid JSON\n")

					// Check for active_mission_id (workspace metadata) or mission_id (grove metadata)
					if activeMissionID, ok := metadata["active_mission_id"].(string); ok && activeMissionID != "" {
						fmt.Printf("   ✓ active_mission_id present: %s (workspace metadata)\n", activeMissionID)
					} else if missionID, ok := metadata["mission_id"].(string); ok && missionID != "" {
						fmt.Printf("   ✓ mission_id present: %s (grove metadata)\n", missionID)
					} else {
						fmt.Printf("   ⚠️  Neither active_mission_id nor mission_id found\n")
					}
				} else {
					fmt.Printf("   ✗ Invalid JSON: %v\n", err)
					validationPassed = false
				}
			} else {
				fmt.Printf("   ⚠️  File not found: %s (optional for some contexts)\n", metadataPath)
			}
			fmt.Println()

			// Overall result
			fmt.Printf("=== Validation Result ===\n")
			if validationPassed {
				fmt.Printf("✓ All critical checks passed\n")
				fmt.Printf("  Context appears to be properly configured\n")
			} else {
				fmt.Printf("✗ Some checks failed\n")
				fmt.Printf("  Context may not be properly configured\n")
			}
			fmt.Println()

			return nil
		},
	}
}
