package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestWatchdogRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWatchdogRepository(db)
	ctx := context.Background()

	// Create test fixtures: factory -> workshop -> workbench
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test Factory", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test Workshop", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test Workbench", "/tmp/test", "active")

	t.Run("creates watchdog successfully", func(t *testing.T) {
		record := &secondary.WatchdogRecord{
			ID:          "WATCH-001",
			WorkbenchID: "BENCH-001",
			Status:      "inactive",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "WATCH-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.WorkbenchID != "BENCH-001" {
			t.Errorf("WorkbenchID = %q, want %q", got.WorkbenchID, "BENCH-001")
		}
		if got.Status != "inactive" {
			t.Errorf("Status = %q, want %q", got.Status, "inactive")
		}
	})

	t.Run("enforces unique workbench constraint", func(t *testing.T) {
		record := &secondary.WatchdogRecord{
			ID:          "WATCH-002",
			WorkbenchID: "BENCH-001", // Same workbench as WATCH-001
			Status:      "inactive",
		}

		err := repo.Create(ctx, record)
		if err == nil {
			t.Fatal("Expected error for duplicate workbench, got nil")
		}
	})
}

func TestWatchdogRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWatchdogRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/tmp/test", "active")

	repo.Create(ctx, &secondary.WatchdogRecord{
		ID:          "WATCH-001",
		WorkbenchID: "BENCH-001",
		Status:      "inactive",
	})

	t.Run("finds watchdog by ID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, "WATCH-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if got.ID != "WATCH-001" {
			t.Errorf("ID = %q, want %q", got.ID, "WATCH-001")
		}
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "WATCH-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestWatchdogRepository_GetByWorkbench(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWatchdogRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/tmp/test", "active")

	repo.Create(ctx, &secondary.WatchdogRecord{
		ID:          "WATCH-001",
		WorkbenchID: "BENCH-001",
		Status:      "inactive",
	})

	t.Run("finds watchdog by workbench", func(t *testing.T) {
		got, err := repo.GetByWorkbench(ctx, "BENCH-001")
		if err != nil {
			t.Fatalf("GetByWorkbench failed: %v", err)
		}
		if got == nil {
			t.Fatal("expected watchdog, got nil")
		}
		if got.ID != "WATCH-001" {
			t.Errorf("ID = %q, want %q", got.ID, "WATCH-001")
		}
	})

	t.Run("returns error for workbench without watchdog", func(t *testing.T) {
		_, err := repo.GetByWorkbench(ctx, "BENCH-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestWatchdogRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWatchdogRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test 1", "/tmp/test1", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-002", "WORK-001", "Test 2", "/tmp/test2", "active")

	repo.Create(ctx, &secondary.WatchdogRecord{ID: "WATCH-001", WorkbenchID: "BENCH-001", Status: "inactive"})
	repo.Create(ctx, &secondary.WatchdogRecord{ID: "WATCH-002", WorkbenchID: "BENCH-002", Status: "active"})

	t.Run("lists all watchdogs", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.WatchdogFilters{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.WatchdogFilters{Status: "active"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "WATCH-002" {
			t.Errorf("ID = %q, want %q", list[0].ID, "WATCH-002")
		}
	})

	t.Run("filters by workbench_id", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.WatchdogFilters{WorkbenchID: "BENCH-001"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "WATCH-001" {
			t.Errorf("ID = %q, want %q", list[0].ID, "WATCH-001")
		}
	})
}

func TestWatchdogRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWatchdogRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/tmp/test", "active")

	repo.Create(ctx, &secondary.WatchdogRecord{ID: "WATCH-001", WorkbenchID: "BENCH-001", Status: "inactive"})

	t.Run("deletes watchdog", func(t *testing.T) {
		err := repo.Delete(ctx, "WATCH-001")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = repo.GetByID(ctx, "WATCH-001")
		if err == nil {
			t.Error("expected error after delete, got nil")
		}
	})

	t.Run("returns error for non-existent watchdog", func(t *testing.T) {
		err := repo.Delete(ctx, "WATCH-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestWatchdogRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWatchdogRepository(db)
	ctx := context.Background()

	t.Run("returns WATCH-001 for empty table", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "WATCH-001" {
			t.Errorf("ID = %q, want %q", id, "WATCH-001")
		}
	})
}

func TestWatchdogRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWatchdogRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/tmp/test", "active")

	repo.Create(ctx, &secondary.WatchdogRecord{
		ID:          "WATCH-001",
		WorkbenchID: "BENCH-001",
		Status:      "inactive",
	})

	t.Run("updates to active", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "WATCH-001", "active")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "WATCH-001")
		if got.Status != "active" {
			t.Errorf("Status = %q, want %q", got.Status, "active")
		}
	})

	t.Run("returns error for non-existent watchdog", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "WATCH-999", "active")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestWatchdogRepository_WorkbenchExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWatchdogRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/tmp/test", "active")

	t.Run("returns true for existing workbench", func(t *testing.T) {
		exists, err := repo.WorkbenchExists(ctx, "BENCH-001")
		if err != nil {
			t.Fatalf("WorkbenchExists failed: %v", err)
		}
		if !exists {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false for non-existent workbench", func(t *testing.T) {
		exists, err := repo.WorkbenchExists(ctx, "BENCH-999")
		if err != nil {
			t.Fatalf("WorkbenchExists failed: %v", err)
		}
		if exists {
			t.Error("expected false, got true")
		}
	})
}

func TestWatchdogRepository_WorkbenchHasWatchdog(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWatchdogRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test 1", "/tmp/test1", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-002", "WORK-001", "Test 2", "/tmp/test2", "active")

	repo.Create(ctx, &secondary.WatchdogRecord{
		ID:          "WATCH-001",
		WorkbenchID: "BENCH-001",
		Status:      "inactive",
	})

	t.Run("returns true when workbench has watchdog", func(t *testing.T) {
		has, err := repo.WorkbenchHasWatchdog(ctx, "BENCH-001")
		if err != nil {
			t.Fatalf("WorkbenchHasWatchdog failed: %v", err)
		}
		if !has {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false when workbench has no watchdog", func(t *testing.T) {
		has, err := repo.WorkbenchHasWatchdog(ctx, "BENCH-002")
		if err != nil {
			t.Fatalf("WorkbenchHasWatchdog failed: %v", err)
		}
		if has {
			t.Error("expected false, got true")
		}
	})
}
