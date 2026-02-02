package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestGatehouseRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewGatehouseRepository(db)
	ctx := context.Background()

	// Create test fixtures: factory -> workshop
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test Factory", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test Workshop", "active")

	t.Run("creates gatehouse successfully", func(t *testing.T) {
		record := &secondary.GatehouseRecord{
			ID:         "GATE-001",
			WorkshopID: "WORK-001",
			Status:     "active",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "GATE-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.WorkshopID != "WORK-001" {
			t.Errorf("WorkshopID = %q, want %q", got.WorkshopID, "WORK-001")
		}
		if got.Status != "active" {
			t.Errorf("Status = %q, want %q", got.Status, "active")
		}
	})

	t.Run("enforces unique workshop constraint", func(t *testing.T) {
		record := &secondary.GatehouseRecord{
			ID:         "GATE-002",
			WorkshopID: "WORK-001", // Same workshop as GATE-001
			Status:     "active",
		}

		err := repo.Create(ctx, record)
		if err == nil {
			t.Fatal("Expected error for duplicate workshop, got nil")
		}
	})
}

func TestGatehouseRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewGatehouseRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")

	repo.Create(ctx, &secondary.GatehouseRecord{
		ID:         "GATE-001",
		WorkshopID: "WORK-001",
		Status:     "active",
	})

	t.Run("finds gatehouse by ID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, "GATE-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if got.ID != "GATE-001" {
			t.Errorf("ID = %q, want %q", got.ID, "GATE-001")
		}
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "GATE-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestGatehouseRepository_GetByWorkshop(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewGatehouseRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")

	repo.Create(ctx, &secondary.GatehouseRecord{
		ID:         "GATE-001",
		WorkshopID: "WORK-001",
		Status:     "active",
	})

	t.Run("finds gatehouse by workshop", func(t *testing.T) {
		got, err := repo.GetByWorkshop(ctx, "WORK-001")
		if err != nil {
			t.Fatalf("GetByWorkshop failed: %v", err)
		}
		if got == nil {
			t.Fatal("expected gatehouse, got nil")
		}
		if got.ID != "GATE-001" {
			t.Errorf("ID = %q, want %q", got.ID, "GATE-001")
		}
	})

	t.Run("returns error for workshop without gatehouse", func(t *testing.T) {
		_, err := repo.GetByWorkshop(ctx, "WORK-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestGatehouseRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewGatehouseRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test 1", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-002", "FACT-001", "Test 2", "active")

	repo.Create(ctx, &secondary.GatehouseRecord{ID: "GATE-001", WorkshopID: "WORK-001", Status: "active"})
	repo.Create(ctx, &secondary.GatehouseRecord{ID: "GATE-002", WorkshopID: "WORK-002", Status: "active"})

	t.Run("lists all gatehouses", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.GatehouseFilters{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("filters by workshop_id", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.GatehouseFilters{WorkshopID: "WORK-002"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "GATE-002" {
			t.Errorf("ID = %q, want %q", list[0].ID, "GATE-002")
		}
	})
}

func TestGatehouseRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewGatehouseRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")

	repo.Create(ctx, &secondary.GatehouseRecord{ID: "GATE-001", WorkshopID: "WORK-001", Status: "active"})

	t.Run("deletes gatehouse", func(t *testing.T) {
		err := repo.Delete(ctx, "GATE-001")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = repo.GetByID(ctx, "GATE-001")
		if err == nil {
			t.Error("expected error after delete, got nil")
		}
	})

	t.Run("returns error for non-existent gatehouse", func(t *testing.T) {
		err := repo.Delete(ctx, "GATE-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestGatehouseRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewGatehouseRepository(db)
	ctx := context.Background()

	t.Run("returns GATE-001 for empty table", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "GATE-001" {
			t.Errorf("ID = %q, want %q", id, "GATE-001")
		}
	})
}

func TestGatehouseRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewGatehouseRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")

	repo.Create(ctx, &secondary.GatehouseRecord{
		ID:         "GATE-001",
		WorkshopID: "WORK-001",
		Status:     "active",
	})

	t.Run("updates status", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "GATE-001", "inactive")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "GATE-001")
		if got.Status != "inactive" {
			t.Errorf("Status = %q, want %q", got.Status, "inactive")
		}
	})

	t.Run("returns error for non-existent gatehouse", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "GATE-999", "active")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestGatehouseRepository_WorkshopExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewGatehouseRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")

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

func TestGatehouseRepository_WorkshopHasGatehouse(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewGatehouseRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test 1", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-002", "FACT-001", "Test 2", "active")

	repo.Create(ctx, &secondary.GatehouseRecord{
		ID:         "GATE-001",
		WorkshopID: "WORK-001",
		Status:     "active",
	})

	t.Run("returns true when workshop has gatehouse", func(t *testing.T) {
		has, err := repo.WorkshopHasGatehouse(ctx, "WORK-001")
		if err != nil {
			t.Fatalf("WorkshopHasGatehouse failed: %v", err)
		}
		if !has {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false when workshop has no gatehouse", func(t *testing.T) {
		has, err := repo.WorkshopHasGatehouse(ctx, "WORK-002")
		if err != nil {
			t.Fatalf("WorkshopHasGatehouse failed: %v", err)
		}
		if has {
			t.Error("expected false, got true")
		}
	})
}

func TestGatehouseRepository_UpdateFocusedID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewGatehouseRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")

	repo.Create(ctx, &secondary.GatehouseRecord{
		ID:         "GATE-001",
		WorkshopID: "WORK-001",
		Status:     "active",
	})

	t.Run("sets focused ID", func(t *testing.T) {
		err := repo.UpdateFocusedID(ctx, "GATE-001", "SHIP-042")
		if err != nil {
			t.Fatalf("UpdateFocusedID failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "GATE-001")
		if got.FocusedID != "SHIP-042" {
			t.Errorf("FocusedID = %q, want %q", got.FocusedID, "SHIP-042")
		}
	})

	t.Run("clears focused ID with empty string", func(t *testing.T) {
		// First set it
		repo.UpdateFocusedID(ctx, "GATE-001", "COMM-001")

		// Then clear it
		err := repo.UpdateFocusedID(ctx, "GATE-001", "")
		if err != nil {
			t.Fatalf("UpdateFocusedID failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "GATE-001")
		if got.FocusedID != "" {
			t.Errorf("FocusedID = %q, want empty", got.FocusedID)
		}
	})

	t.Run("returns error for non-existent gatehouse", func(t *testing.T) {
		err := repo.UpdateFocusedID(ctx, "GATE-999", "SHIP-001")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
