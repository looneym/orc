package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultWorkspacePath(t *testing.T) {
	path, err := DefaultWorkspacePath("COMM-001")
	if err != nil {
		t.Fatalf("DefaultWorkspacePath failed: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "src", "commissions", "COMM-001")

	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestParseWorkshopIDFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "standard workshop path",
			path:     "/Users/test/.orc/ws/WORK-003-myproject/",
			expected: "WORK-003",
		},
		{
			name:     "workshop path without trailing slash",
			path:     "/Users/test/.orc/ws/WORK-001",
			expected: "WORK-001",
		},
		{
			name:     "workshop path with suffix",
			path:     "/home/user/.orc/ws/WORK-042-some-name/subdir",
			expected: "WORK-042",
		},
		{
			name:     "no workshop ID in path",
			path:     "/Users/test/src/worktrees/myproject",
			expected: "",
		},
		{
			name:     "workbench path (not workshop)",
			path:     "/Users/test/src/worktrees/BENCH-014",
			expected: "",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseWorkshopIDFromPath(tt.path)
			if result != tt.expected {
				t.Errorf("ParseWorkshopIDFromPath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestMigrateConfig(t *testing.T) {
	tests := []struct {
		name           string
		initialConfig  map[string]any
		wantOldFocus   string
		wantModified   bool
		wantErr        bool
		checkNewConfig func(t *testing.T, cfg *Config)
	}{
		{
			name: "migrate IMP config with deprecated fields",
			initialConfig: map[string]any{
				"version":       "1.0",
				"role":          "IMP",
				"workbench_id":  "BENCH-014",
				"commission_id": "COMM-001",
				"current_focus": "SHIP-123",
			},
			wantOldFocus: "SHIP-123",
			wantModified: true,
			wantErr:      false,
			checkNewConfig: func(t *testing.T, cfg *Config) {
				// After migration, place_id should be BENCH-014
				if cfg.PlaceID != "BENCH-014" {
					t.Errorf("expected place_id BENCH-014, got %s", cfg.PlaceID)
				}
			},
		},
		{
			name: "already migrated config (new format)",
			initialConfig: map[string]any{
				"version":  "1.0",
				"place_id": "BENCH-014",
			},
			wantOldFocus: "",
			wantModified: false,
			wantErr:      false,
		},
		{
			name: "migrate Goblin config with commission_id (no place_id after migration)",
			initialConfig: map[string]any{
				"version":       "1.0",
				"role":          "GOBLIN",
				"commission_id": "COMM-001",
			},
			wantOldFocus: "",
			wantModified: true,
			wantErr:      false,
			checkNewConfig: func(t *testing.T, cfg *Config) {
				// Goblin configs without workshop_id can't auto-migrate to place_id
				// They require DB lookup for gatehouse ID
				// So place_id will be empty after v1 migration
				if cfg.PlaceID != "" {
					t.Errorf("expected empty place_id for Goblin v1 migration, got %s", cfg.PlaceID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "orc-config-test")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create .orc directory and config
			orcDir := filepath.Join(tmpDir, ".orc")
			if err := os.MkdirAll(orcDir, 0755); err != nil {
				t.Fatalf("failed to create .orc dir: %v", err)
			}

			configPath := filepath.Join(orcDir, "config.json")
			data, err := json.Marshal(tt.initialConfig)
			if err != nil {
				t.Fatalf("failed to marshal initial config: %v", err)
			}
			if err := os.WriteFile(configPath, data, 0644); err != nil {
				t.Fatalf("failed to write initial config: %v", err)
			}

			// Run migration
			oldFocus, modified, err := MigrateConfig(tmpDir)

			// Check error
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check old focus
			if oldFocus != tt.wantOldFocus {
				t.Errorf("oldFocus = %q, want %q", oldFocus, tt.wantOldFocus)
			}

			// Check modified
			if modified != tt.wantModified {
				t.Errorf("modified = %v, want %v", modified, tt.wantModified)
			}

			// If we have additional checks
			if tt.checkNewConfig != nil && !tt.wantErr {
				cfg, err := LoadConfig(tmpDir)
				if err != nil {
					t.Fatalf("failed to load config after migration: %v", err)
				}
				tt.checkNewConfig(t, cfg)
			}
		})
	}
}

func TestLoadConfig_BackwardCompatibility(t *testing.T) {
	// Test that loading an old config format works and migrates to place_id
	tmpDir, err := os.MkdirTemp("", "orc-config-compat")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .orc directory and old-format config
	orcDir := filepath.Join(tmpDir, ".orc")
	if err := os.MkdirAll(orcDir, 0755); err != nil {
		t.Fatalf("failed to create .orc dir: %v", err)
	}

	// Old IMP config with deprecated fields
	oldConfig := `{"version":"1.0","role":"IMP","workbench_id":"BENCH-001","commission_id":"COMM-001","current_focus":"SHIP-123"}`
	configPath := filepath.Join(orcDir, "config.json")
	if err := os.WriteFile(configPath, []byte(oldConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Should load without error and migrate to place_id format
	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed with old format: %v", err)
	}

	// Core fields should be populated
	if cfg.Version != "1.0" {
		t.Errorf("Version = %q, want 1.0", cfg.Version)
	}
	// IMP configs should be migrated to place_id
	if cfg.PlaceID != "BENCH-001" {
		t.Errorf("PlaceID = %q, want BENCH-001", cfg.PlaceID)
	}
}

func TestGetPlaceType(t *testing.T) {
	tests := []struct {
		placeID  string
		expected string
	}{
		{"BENCH-001", PlaceTypeWorkbench},
		{"BENCH-014", PlaceTypeWorkbench},
		{"GATE-001", PlaceTypeGatehouse},
		{"GATE-123", PlaceTypeGatehouse},
		{"", ""},
		{"WORK-001", ""},
		{"COMM-001", ""},
		{"SHIP-001", ""},
		{"BEN", ""},
	}

	for _, tt := range tests {
		t.Run(tt.placeID, func(t *testing.T) {
			result := GetPlaceType(tt.placeID)
			if result != tt.expected {
				t.Errorf("GetPlaceType(%q) = %q, want %q", tt.placeID, result, tt.expected)
			}
		})
	}
}

func TestGetRoleFromPlaceID(t *testing.T) {
	tests := []struct {
		placeID  string
		expected string
	}{
		{"BENCH-001", RoleIMP},
		{"GATE-001", RoleGoblin},
		{"", ""},
		{"WORK-001", ""},
	}

	for _, tt := range tests {
		t.Run(tt.placeID, func(t *testing.T) {
			result := GetRoleFromPlaceID(tt.placeID)
			if result != tt.expected {
				t.Errorf("GetRoleFromPlaceID(%q) = %q, want %q", tt.placeID, result, tt.expected)
			}
		})
	}
}
