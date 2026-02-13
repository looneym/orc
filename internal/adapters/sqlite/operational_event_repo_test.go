package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestOperationalEventRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewOperationalEventRepository(db)
	ctx := context.Background()

	// Create test fixtures: factory -> workshop
	seedFactory(t, db, "FACT-001", "Test Factory")
	seedWorkshop(t, db, "WORK-001", "FACT-001", "Test Workshop")

	tests := []struct {
		name   string
		record *secondary.OperationalEventRecord
		checks func(t *testing.T, got *secondary.OperationalEventRecord)
	}{
		{
			name: "creates event with all fields",
			record: &secondary.OperationalEventRecord{
				ID:         "OE-0001",
				WorkshopID: "WORK-001",
				ActorID:    "BENCH-014",
				Source:     "hook",
				Version:    "1.0",
				Level:      "info",
				Message:    "Hook invocation complete",
				DataJSON:   `{"duration_ms":42}`,
			},
			checks: func(t *testing.T, got *secondary.OperationalEventRecord) {
				if got.WorkshopID != "WORK-001" {
					t.Errorf("WorkshopID = %q, want %q", got.WorkshopID, "WORK-001")
				}
				if got.ActorID != "BENCH-014" {
					t.Errorf("ActorID = %q, want %q", got.ActorID, "BENCH-014")
				}
				if got.Source != "hook" {
					t.Errorf("Source = %q, want %q", got.Source, "hook")
				}
				if got.Version != "1.0" {
					t.Errorf("Version = %q, want %q", got.Version, "1.0")
				}
				if got.Level != "info" {
					t.Errorf("Level = %q, want %q", got.Level, "info")
				}
				if got.Message != "Hook invocation complete" {
					t.Errorf("Message = %q, want %q", got.Message, "Hook invocation complete")
				}
				if got.DataJSON != `{"duration_ms":42}` {
					t.Errorf("DataJSON = %q, want %q", got.DataJSON, `{"duration_ms":42}`)
				}
			},
		},
		{
			name: "creates event with nullable fields null",
			record: &secondary.OperationalEventRecord{
				ID:      "OE-0002",
				Source:  "system",
				Level:   "warn",
				Message: "System event without workshop",
				// WorkshopID, ActorID, Version, DataJSON all empty (null)
			},
			checks: func(t *testing.T, got *secondary.OperationalEventRecord) {
				if got.WorkshopID != "" {
					t.Errorf("WorkshopID = %q, want empty", got.WorkshopID)
				}
				if got.ActorID != "" {
					t.Errorf("ActorID = %q, want empty", got.ActorID)
				}
				if got.Version != "" {
					t.Errorf("Version = %q, want empty", got.Version)
				}
				if got.DataJSON != "" {
					t.Errorf("DataJSON = %q, want empty", got.DataJSON)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.record)
			if err != nil {
				t.Fatalf("Create failed: %v", err)
			}

			// Retrieve and verify via List (OperationalEventRepository has no GetByID)
			list, err := repo.List(ctx, secondary.OperationalEventFilters{})
			if err != nil {
				t.Fatalf("List failed: %v", err)
			}

			var got *secondary.OperationalEventRecord
			for _, e := range list {
				if e.ID == tt.record.ID {
					got = e
					break
				}
			}
			if got == nil {
				t.Fatalf("event %s not found in list", tt.record.ID)
			}

			tt.checks(t, got)
		})
	}
}

func TestOperationalEventRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewOperationalEventRepository(db)
	ctx := context.Background()

	// Setup
	seedFactory(t, db, "FACT-001", "Test Factory")
	seedWorkshop(t, db, "WORK-001", "FACT-001", "Workshop 1")
	seedWorkshop(t, db, "WORK-002", "FACT-001", "Workshop 2")

	repo.Create(ctx, &secondary.OperationalEventRecord{ID: "OE-0001", WorkshopID: "WORK-001", Source: "hook", Level: "info", Message: "Hook started"})
	repo.Create(ctx, &secondary.OperationalEventRecord{ID: "OE-0002", WorkshopID: "WORK-001", Source: "system", Level: "warn", Message: "System warning"})
	repo.Create(ctx, &secondary.OperationalEventRecord{ID: "OE-0003", WorkshopID: "WORK-002", Source: "hook", Level: "error", Message: "Hook failed"})

	tests := []struct {
		name      string
		filters   secondary.OperationalEventFilters
		wantCount int
		wantFirst string
	}{
		{
			name:      "lists all events",
			filters:   secondary.OperationalEventFilters{},
			wantCount: 3,
		},
		{
			name:      "filters by workshop_id",
			filters:   secondary.OperationalEventFilters{WorkshopID: "WORK-001"},
			wantCount: 2,
		},
		{
			name:      "filters by source",
			filters:   secondary.OperationalEventFilters{Source: "hook"},
			wantCount: 2,
		},
		{
			name:      "filters by level",
			filters:   secondary.OperationalEventFilters{Level: "error"},
			wantCount: 1,
			wantFirst: "OE-0003",
		},
		{
			name:      "applies limit",
			filters:   secondary.OperationalEventFilters{Limit: 2},
			wantCount: 2,
		},
		{
			name:      "combines filters",
			filters:   secondary.OperationalEventFilters{WorkshopID: "WORK-001", Source: "hook"},
			wantCount: 1,
			wantFirst: "OE-0001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list, err := repo.List(ctx, tt.filters)
			if err != nil {
				t.Fatalf("List failed: %v", err)
			}
			if len(list) != tt.wantCount {
				t.Errorf("len = %d, want %d", len(list), tt.wantCount)
			}
			if tt.wantFirst != "" && len(list) > 0 && list[0].ID != tt.wantFirst {
				t.Errorf("first ID = %q, want %q", list[0].ID, tt.wantFirst)
			}
		})
	}
}

func TestOperationalEventRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewOperationalEventRepository(db)
	ctx := context.Background()

	t.Run("returns OE-0001 for empty table", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "OE-0001" {
			t.Errorf("ID = %q, want %q", id, "OE-0001")
		}
	})

	t.Run("increments after creating events", func(t *testing.T) {
		repo.Create(ctx, &secondary.OperationalEventRecord{
			ID:      "OE-0001",
			Source:  "system",
			Level:   "info",
			Message: "Test event",
		})

		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "OE-0002" {
			t.Errorf("ID = %q, want %q", id, "OE-0002")
		}
	})
}

func TestOperationalEventRepository_PruneOlderThan(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewOperationalEventRepository(db)
	ctx := context.Background()

	// Insert events - all will have current timestamp
	repo.Create(ctx, &secondary.OperationalEventRecord{ID: "OE-0001", Source: "system", Level: "info", Message: "Event 1"})
	repo.Create(ctx, &secondary.OperationalEventRecord{ID: "OE-0002", Source: "system", Level: "info", Message: "Event 2"})

	t.Run("prunes nothing when events are recent", func(t *testing.T) {
		count, err := repo.PruneOlderThan(ctx, 1)
		if err != nil {
			t.Fatalf("PruneOlderThan failed: %v", err)
		}
		if count != 0 {
			t.Errorf("pruned = %d, want 0", count)
		}

		// Verify events still exist
		list, err := repo.List(ctx, secondary.OperationalEventFilters{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("prunes old events", func(t *testing.T) {
		// Manually backdate an event
		db.ExecContext(ctx, "UPDATE operational_events SET timestamp = datetime('now', '-10 days') WHERE id = ?", "OE-0001")

		count, err := repo.PruneOlderThan(ctx, 5)
		if err != nil {
			t.Fatalf("PruneOlderThan failed: %v", err)
		}
		if count != 1 {
			t.Errorf("pruned = %d, want 1", count)
		}

		// Verify only recent event remains
		list, err := repo.List(ctx, secondary.OperationalEventFilters{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "OE-0002" {
			t.Errorf("remaining ID = %q, want %q", list[0].ID, "OE-0002")
		}
	})
}
