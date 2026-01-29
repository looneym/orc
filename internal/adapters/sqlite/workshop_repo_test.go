package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestWorkshopRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	// Seed required factory
	seedFactory(t, db, "FACT-001", "test-factory")

	workshop := &secondary.WorkshopRecord{
		FactoryID: "FACT-001",
		Name:      "test-workshop",
	}

	err := repo.Create(ctx, workshop)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify ID was generated
	if workshop.ID == "" {
		t.Error("expected ID to be generated")
	}

	// Verify round-trip
	got, err := repo.GetByID(ctx, workshop.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Name != "test-workshop" {
		t.Errorf("expected name 'test-workshop', got %q", got.Name)
	}
	if got.Status != "active" {
		t.Errorf("expected default status 'active', got %q", got.Status)
	}
}

func TestWorkshopRepository_Create_GeneratesName(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	// Seed required factory
	seedFactory(t, db, "FACT-001", "test-factory")

	workshop := &secondary.WorkshopRecord{
		FactoryID: "FACT-001",
		// No name provided
	}

	err := repo.Create(ctx, workshop)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify name was generated
	if workshop.Name == "" {
		t.Error("expected name to be generated")
	}
}

func TestWorkshopRepository_Create_FactoryNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	workshop := &secondary.WorkshopRecord{
		FactoryID: "FACT-999",
		Name:      "test-workshop",
	}

	err := repo.Create(ctx, workshop)
	if err == nil {
		t.Error("expected error for non-existent factory")
	}
}

func TestWorkshopRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	// Seed required factory and workshop
	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "test-workshop")

	got, err := repo.GetByID(ctx, "SHOP-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Name != "test-workshop" {
		t.Errorf("expected name 'test-workshop', got %q", got.Name)
	}
	if got.CreatedAt == "" {
		t.Error("expected CreatedAt to be set")
	}
	if got.UpdatedAt == "" {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestWorkshopRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "SHOP-999")
	if err == nil {
		t.Error("expected error for non-existent workshop")
	}
}

func TestWorkshopRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	// Seed factory and workshops
	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "workshop-1")
	seedWorkshop(t, db, "SHOP-002", "FACT-001", "workshop-2")
	seedWorkshop(t, db, "SHOP-003", "FACT-001", "workshop-3")

	workshops, err := repo.List(ctx, secondary.WorkshopFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(workshops) != 3 {
		t.Errorf("expected 3 workshops, got %d", len(workshops))
	}
}

func TestWorkshopRepository_List_FilterByFactory(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	// Seed factories and workshops
	seedFactory(t, db, "FACT-001", "factory-1")
	seedFactory(t, db, "FACT-002", "factory-2")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "workshop-1")
	seedWorkshop(t, db, "SHOP-002", "FACT-001", "workshop-2")
	seedWorkshop(t, db, "SHOP-003", "FACT-002", "workshop-3")

	workshops, err := repo.List(ctx, secondary.WorkshopFilters{FactoryID: "FACT-001"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(workshops) != 2 {
		t.Errorf("expected 2 workshops for FACT-001, got %d", len(workshops))
	}
}

func TestWorkshopRepository_List_FilterByStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	// Seed factory and workshops
	seedFactory(t, db, "FACT-001", "test-factory")
	_, _ = db.Exec("INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "SHOP-001", "FACT-001", "active-shop", "active")
	_, _ = db.Exec("INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "SHOP-002", "FACT-001", "inactive-shop", "inactive")

	workshops, err := repo.List(ctx, secondary.WorkshopFilters{Status: "active"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(workshops) != 1 {
		t.Errorf("expected 1 active workshop, got %d", len(workshops))
	}
}

func TestWorkshopRepository_List_WithLimit(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	// Seed factory and workshops
	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "workshop-1")
	seedWorkshop(t, db, "SHOP-002", "FACT-001", "workshop-2")
	seedWorkshop(t, db, "SHOP-003", "FACT-001", "workshop-3")

	workshops, err := repo.List(ctx, secondary.WorkshopFilters{Limit: 2})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(workshops) != 2 {
		t.Errorf("expected 2 workshops with limit, got %d", len(workshops))
	}
}

func TestWorkshopRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	// Seed factory and workshop
	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "original")

	// Update it
	err := repo.Update(ctx, &secondary.WorkshopRecord{
		ID:   "SHOP-001",
		Name: "updated",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	got, _ := repo.GetByID(ctx, "SHOP-001")
	if got.Name != "updated" {
		t.Errorf("expected name 'updated', got %q", got.Name)
	}
}

func TestWorkshopRepository_Update_Status(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	// Seed factory and workshop
	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "test")

	// Update status (status must be 'active' or 'archived' per CHECK constraint)
	err := repo.Update(ctx, &secondary.WorkshopRecord{
		ID:     "SHOP-001",
		Status: "archived",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	got, _ := repo.GetByID(ctx, "SHOP-001")
	if got.Status != "archived" {
		t.Errorf("expected status 'archived', got %q", got.Status)
	}
}

func TestWorkshopRepository_Update_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, &secondary.WorkshopRecord{
		ID:   "SHOP-999",
		Name: "updated",
	})
	if err == nil {
		t.Error("expected error for non-existent workshop")
	}
}

func TestWorkshopRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	// Seed factory and workshop
	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "to-delete")

	// Delete it
	err := repo.Delete(ctx, "SHOP-001")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(ctx, "SHOP-001")
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestWorkshopRepository_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "SHOP-999")
	if err == nil {
		t.Error("expected error for non-existent workshop")
	}
}

func TestWorkshopRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	// First ID should be WORK-001 (workshop prefix is WORK, not SHOP)
	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "WORK-001" {
		t.Errorf("expected WORK-001, got %s", id)
	}

	// Seed factory and create a workshop
	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "WORK-001", "FACT-001", "test")

	// Next ID should be WORK-002
	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "WORK-002" {
		t.Errorf("expected WORK-002, got %s", id)
	}
}

func TestWorkshopRepository_CountWorkbenches(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	// Seed factory and workshop
	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "test")

	// Count should be 0
	count, err := repo.CountWorkbenches(ctx, "SHOP-001")
	if err != nil {
		t.Fatalf("CountWorkbenches failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 workbenches, got %d", count)
	}

	// Add workbenches
	_, _ = db.Exec("INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, 'active')", "BENCH-001", "SHOP-001", "bench-1", "/tmp/bench-1")
	_, _ = db.Exec("INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, 'active')", "BENCH-002", "SHOP-001", "bench-2", "/tmp/bench-2")

	// Count should be 2
	count, err = repo.CountWorkbenches(ctx, "SHOP-001")
	if err != nil {
		t.Fatalf("CountWorkbenches failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 workbenches, got %d", count)
	}
}

func TestWorkshopRepository_CountByFactory(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	// Seed factories and workshops
	seedFactory(t, db, "FACT-001", "factory-1")
	seedFactory(t, db, "FACT-002", "factory-2")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "workshop-1")
	seedWorkshop(t, db, "SHOP-002", "FACT-001", "workshop-2")
	seedWorkshop(t, db, "SHOP-003", "FACT-002", "workshop-3")

	// Count for FACT-001 should be 2
	count, err := repo.CountByFactory(ctx, "FACT-001")
	if err != nil {
		t.Fatalf("CountByFactory failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 workshops for FACT-001, got %d", count)
	}

	// Count for FACT-002 should be 1
	count, err = repo.CountByFactory(ctx, "FACT-002")
	if err != nil {
		t.Fatalf("CountByFactory failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 workshop for FACT-002, got %d", count)
	}
}

func TestWorkshopRepository_FactoryExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkshopRepository(db)
	ctx := context.Background()

	// Factory doesn't exist
	exists, err := repo.FactoryExists(ctx, "FACT-001")
	if err != nil {
		t.Fatalf("FactoryExists failed: %v", err)
	}
	if exists {
		t.Error("expected factory to not exist")
	}

	// Seed factory
	seedFactory(t, db, "FACT-001", "test-factory")

	// Factory exists
	exists, err = repo.FactoryExists(ctx, "FACT-001")
	if err != nil {
		t.Fatalf("FactoryExists failed: %v", err)
	}
	if !exists {
		t.Error("expected factory to exist")
	}
}
