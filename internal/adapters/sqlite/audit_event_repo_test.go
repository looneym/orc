package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestWorkshopEventRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopEventRepository(db)
	ctx := context.Background()

	// Create test fixtures: factory -> workshop
	seedFactory(t, db, "FACT-001", "Test Factory")
	seedWorkshop(t, db, "WORK-001", "FACT-001", "Test Workshop")

	t.Run("creates event with all fields", func(t *testing.T) {
		record := &secondary.AuditEventRecord{
			ID:         "WE-0001",
			WorkshopID: "WORK-001",
			ActorID:    "BENCH-014",
			Source:     "orc",
			Version:    "1.0",
			EntityType: "task",
			EntityID:   "TASK-001",
			Action:     "update",
			FieldName:  "status",
			OldValue:   "ready",
			NewValue:   "in_progress",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "WE-0001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.WorkshopID != "WORK-001" {
			t.Errorf("WorkshopID = %q, want %q", got.WorkshopID, "WORK-001")
		}
		if got.ActorID != "BENCH-014" {
			t.Errorf("ActorID = %q, want %q", got.ActorID, "BENCH-014")
		}
		if got.Source != "orc" {
			t.Errorf("Source = %q, want %q", got.Source, "orc")
		}
		if got.Version != "1.0" {
			t.Errorf("Version = %q, want %q", got.Version, "1.0")
		}
		if got.EntityType != "task" {
			t.Errorf("EntityType = %q, want %q", got.EntityType, "task")
		}
		if got.EntityID != "TASK-001" {
			t.Errorf("EntityID = %q, want %q", got.EntityID, "TASK-001")
		}
		if got.Action != "update" {
			t.Errorf("Action = %q, want %q", got.Action, "update")
		}
		if got.FieldName != "status" {
			t.Errorf("FieldName = %q, want %q", got.FieldName, "status")
		}
		if got.OldValue != "ready" {
			t.Errorf("OldValue = %q, want %q", got.OldValue, "ready")
		}
		if got.NewValue != "in_progress" {
			t.Errorf("NewValue = %q, want %q", got.NewValue, "in_progress")
		}
	})

	t.Run("creates event with nullable fields null", func(t *testing.T) {
		record := &secondary.AuditEventRecord{
			ID:         "WE-0002",
			EntityType: "shipment",
			EntityID:   "SHIP-001",
			Action:     "create",
			// WorkshopID, ActorID, Source, Version, FieldName, OldValue, NewValue all empty (null)
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "WE-0002")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.WorkshopID != "" {
			t.Errorf("WorkshopID = %q, want empty", got.WorkshopID)
		}
		if got.ActorID != "" {
			t.Errorf("ActorID = %q, want empty", got.ActorID)
		}
		if got.Source != "" {
			t.Errorf("Source = %q, want empty", got.Source)
		}
		if got.Version != "" {
			t.Errorf("Version = %q, want empty", got.Version)
		}
		if got.FieldName != "" {
			t.Errorf("FieldName = %q, want empty", got.FieldName)
		}
		if got.OldValue != "" {
			t.Errorf("OldValue = %q, want empty", got.OldValue)
		}
		if got.NewValue != "" {
			t.Errorf("NewValue = %q, want empty", got.NewValue)
		}
	})
}

func TestWorkshopEventRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopEventRepository(db)
	ctx := context.Background()

	// Setup
	seedFactory(t, db, "FACT-001", "Test Factory")
	seedWorkshop(t, db, "WORK-001", "FACT-001", "Test Workshop")

	repo.Create(ctx, &secondary.AuditEventRecord{
		ID:         "WE-0001",
		WorkshopID: "WORK-001",
		EntityType: "task",
		EntityID:   "TASK-001",
		Action:     "create",
	})

	t.Run("finds event by ID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, "WE-0001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if got.ID != "WE-0001" {
			t.Errorf("ID = %q, want %q", got.ID, "WE-0001")
		}
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "WE-9999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestWorkshopEventRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopEventRepository(db)
	ctx := context.Background()

	// Setup
	seedFactory(t, db, "FACT-001", "Test Factory")
	seedWorkshop(t, db, "WORK-001", "FACT-001", "Workshop 1")
	seedWorkshop(t, db, "WORK-002", "FACT-001", "Workshop 2")

	repo.Create(ctx, &secondary.AuditEventRecord{ID: "WE-0001", WorkshopID: "WORK-001", ActorID: "BENCH-014", Source: "orc", EntityType: "task", EntityID: "TASK-001", Action: "create"})
	repo.Create(ctx, &secondary.AuditEventRecord{ID: "WE-0002", WorkshopID: "WORK-001", ActorID: "BENCH-014", Source: "imp", EntityType: "task", EntityID: "TASK-001", Action: "update"})
	repo.Create(ctx, &secondary.AuditEventRecord{ID: "WE-0003", WorkshopID: "WORK-002", ActorID: "BENCH-003", Source: "orc", EntityType: "shipment", EntityID: "SHIP-001", Action: "delete"})

	t.Run("lists all events", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.AuditEventFilters{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 3 {
			t.Errorf("len = %d, want 3", len(list))
		}
	})

	t.Run("filters by workshop_id", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.AuditEventFilters{WorkshopID: "WORK-001"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("filters by entity_type", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.AuditEventFilters{EntityType: "shipment"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "WE-0003" {
			t.Errorf("ID = %q, want %q", list[0].ID, "WE-0003")
		}
	})

	t.Run("filters by entity_id", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.AuditEventFilters{EntityID: "TASK-001"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("filters by actor_id", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.AuditEventFilters{ActorID: "BENCH-003"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
	})

	t.Run("filters by action", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.AuditEventFilters{Action: "update"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "WE-0002" {
			t.Errorf("ID = %q, want %q", list[0].ID, "WE-0002")
		}
	})

	t.Run("filters by source", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.AuditEventFilters{Source: "orc"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("applies limit", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.AuditEventFilters{Limit: 2})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("combines filters", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.AuditEventFilters{WorkshopID: "WORK-001", Action: "create"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "WE-0001" {
			t.Errorf("ID = %q, want %q", list[0].ID, "WE-0001")
		}
	})
}

func TestWorkshopEventRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopEventRepository(db)
	ctx := context.Background()

	t.Run("returns WE-0001 for empty table", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "WE-0001" {
			t.Errorf("ID = %q, want %q", id, "WE-0001")
		}
	})

	t.Run("increments after creating events", func(t *testing.T) {
		// Setup
		seedFactory(t, db, "FACT-001", "Test Factory")
		seedWorkshop(t, db, "WORK-001", "FACT-001", "Test Workshop")

		repo.Create(ctx, &secondary.AuditEventRecord{
			ID:         "WE-0001",
			WorkshopID: "WORK-001",
			EntityType: "task",
			EntityID:   "TASK-001",
			Action:     "create",
		})

		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "WE-0002" {
			t.Errorf("ID = %q, want %q", id, "WE-0002")
		}
	})
}

func TestWorkshopEventRepository_WorkshopExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopEventRepository(db)
	ctx := context.Background()

	// Setup
	seedFactory(t, db, "FACT-001", "Test Factory")
	seedWorkshop(t, db, "WORK-001", "FACT-001", "Test Workshop")

	t.Run("returns true for existing workshop", func(t *testing.T) {
		exists, err := repo.WorkshopExists(ctx, "WORK-001")
		if err != nil {
			t.Fatalf("WorkshopExists failed: %v", err)
		}
		if !exists {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false for non-existent workshop", func(t *testing.T) {
		exists, err := repo.WorkshopExists(ctx, "WORK-999")
		if err != nil {
			t.Fatalf("WorkshopExists failed: %v", err)
		}
		if exists {
			t.Error("expected false, got true")
		}
	})
}
