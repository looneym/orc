package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestFactoryRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewFactoryRepository(db)
	ctx := context.Background()

	factory := &secondary.FactoryRecord{
		ID:   "FACT-001",
		Name: "test-factory",
	}

	err := repo.Create(ctx, factory)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify round-trip
	got, err := repo.GetByID(ctx, factory.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Name != "test-factory" {
		t.Errorf("expected name 'test-factory', got %q", got.Name)
	}
	if got.Status != "active" {
		t.Errorf("expected default status 'active', got %q", got.Status)
	}
}

func TestFactoryRepository_Create_RequiresID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewFactoryRepository(db)
	ctx := context.Background()

	factory := &secondary.FactoryRecord{
		Name: "test-factory",
	}

	err := repo.Create(ctx, factory)
	if err == nil {
		t.Error("expected error for missing ID")
	}
}

func TestFactoryRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewFactoryRepository(db)
	ctx := context.Background()

	// Create a factory
	factory := &secondary.FactoryRecord{
		ID:   "FACT-001",
		Name: "test-factory",
	}
	_ = repo.Create(ctx, factory)

	// Retrieve it
	got, err := repo.GetByID(ctx, "FACT-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Name != "test-factory" {
		t.Errorf("expected name 'test-factory', got %q", got.Name)
	}
	if got.CreatedAt == "" {
		t.Error("expected CreatedAt to be set")
	}
	if got.UpdatedAt == "" {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestFactoryRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewFactoryRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "FACT-999")
	if err == nil {
		t.Error("expected error for non-existent factory")
	}
}

func TestFactoryRepository_GetByName(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewFactoryRepository(db)
	ctx := context.Background()

	// Create a factory
	factory := &secondary.FactoryRecord{
		ID:   "FACT-001",
		Name: "unique-name",
	}
	_ = repo.Create(ctx, factory)

	// Retrieve by name
	got, err := repo.GetByName(ctx, "unique-name")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if got.ID != "FACT-001" {
		t.Errorf("expected ID 'FACT-001', got %q", got.ID)
	}
}

func TestFactoryRepository_GetByName_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewFactoryRepository(db)
	ctx := context.Background()

	_, err := repo.GetByName(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent factory name")
	}
}

func TestFactoryRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewFactoryRepository(db)
	ctx := context.Background()

	// Create multiple factories
	_ = repo.Create(ctx, &secondary.FactoryRecord{ID: "FACT-001", Name: "factory-1"})
	_ = repo.Create(ctx, &secondary.FactoryRecord{ID: "FACT-002", Name: "factory-2"})
	_ = repo.Create(ctx, &secondary.FactoryRecord{ID: "FACT-003", Name: "factory-3"})

	factories, err := repo.List(ctx, secondary.FactoryFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(factories) != 3 {
		t.Errorf("expected 3 factories, got %d", len(factories))
	}
}

func TestFactoryRepository_List_FilterByStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewFactoryRepository(db)
	ctx := context.Background()

	// Create factories with different statuses
	_ = repo.Create(ctx, &secondary.FactoryRecord{ID: "FACT-001", Name: "factory-1", Status: "active"})
	_ = repo.Create(ctx, &secondary.FactoryRecord{ID: "FACT-002", Name: "factory-2", Status: "inactive"})

	// List only active
	factories, err := repo.List(ctx, secondary.FactoryFilters{Status: "active"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(factories) != 1 {
		t.Errorf("expected 1 active factory, got %d", len(factories))
	}
}

func TestFactoryRepository_List_WithLimit(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewFactoryRepository(db)
	ctx := context.Background()

	// Create multiple factories
	_ = repo.Create(ctx, &secondary.FactoryRecord{ID: "FACT-001", Name: "factory-1"})
	_ = repo.Create(ctx, &secondary.FactoryRecord{ID: "FACT-002", Name: "factory-2"})
	_ = repo.Create(ctx, &secondary.FactoryRecord{ID: "FACT-003", Name: "factory-3"})

	factories, err := repo.List(ctx, secondary.FactoryFilters{Limit: 2})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(factories) != 2 {
		t.Errorf("expected 2 factories with limit, got %d", len(factories))
	}
}

func TestFactoryRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewFactoryRepository(db)
	ctx := context.Background()

	// Create a factory
	_ = repo.Create(ctx, &secondary.FactoryRecord{ID: "FACT-001", Name: "original"})

	// Update it
	err := repo.Update(ctx, &secondary.FactoryRecord{
		ID:   "FACT-001",
		Name: "updated",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	got, _ := repo.GetByID(ctx, "FACT-001")
	if got.Name != "updated" {
		t.Errorf("expected name 'updated', got %q", got.Name)
	}
}

func TestFactoryRepository_Update_Status(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewFactoryRepository(db)
	ctx := context.Background()

	// Create a factory
	_ = repo.Create(ctx, &secondary.FactoryRecord{ID: "FACT-001", Name: "test"})

	// Update status (status must be 'active' or 'archived' per CHECK constraint)
	err := repo.Update(ctx, &secondary.FactoryRecord{
		ID:     "FACT-001",
		Status: "archived",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	got, _ := repo.GetByID(ctx, "FACT-001")
	if got.Status != "archived" {
		t.Errorf("expected status 'archived', got %q", got.Status)
	}
}

func TestFactoryRepository_Update_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewFactoryRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, &secondary.FactoryRecord{
		ID:   "FACT-999",
		Name: "updated",
	})
	if err == nil {
		t.Error("expected error for non-existent factory")
	}
}

func TestFactoryRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewFactoryRepository(db)
	ctx := context.Background()

	// Create a factory
	_ = repo.Create(ctx, &secondary.FactoryRecord{ID: "FACT-001", Name: "to-delete"})

	// Delete it
	err := repo.Delete(ctx, "FACT-001")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(ctx, "FACT-001")
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestFactoryRepository_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewFactoryRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "FACT-999")
	if err == nil {
		t.Error("expected error for non-existent factory")
	}
}

func TestFactoryRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewFactoryRepository(db)
	ctx := context.Background()

	// First ID should be FACT-001
	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "FACT-001" {
		t.Errorf("expected FACT-001, got %s", id)
	}

	// Create a factory
	_ = repo.Create(ctx, &secondary.FactoryRecord{ID: id, Name: "test"})

	// Next ID should be FACT-002
	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "FACT-002" {
		t.Errorf("expected FACT-002, got %s", id)
	}
}

func TestFactoryRepository_CountWorkshops(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewFactoryRepository(db)
	ctx := context.Background()

	// Create a factory
	_ = repo.Create(ctx, &secondary.FactoryRecord{ID: "FACT-001", Name: "test"})

	// Count should be 0
	count, err := repo.CountWorkshops(ctx, "FACT-001")
	if err != nil {
		t.Fatalf("CountWorkshops failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 workshops, got %d", count)
	}

	// Add workshops
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "workshop-1")
	seedWorkshop(t, db, "SHOP-002", "FACT-001", "workshop-2")

	// Count should be 2
	count, err = repo.CountWorkshops(ctx, "FACT-001")
	if err != nil {
		t.Fatalf("CountWorkshops failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 workshops, got %d", count)
	}
}

func TestFactoryRepository_CountCommissions(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewFactoryRepository(db)
	ctx := context.Background()

	// Create a factory
	_ = repo.Create(ctx, &secondary.FactoryRecord{ID: "FACT-001", Name: "test"})

	// Count should be 0
	count, err := repo.CountCommissions(ctx, "FACT-001")
	if err != nil {
		t.Fatalf("CountCommissions failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 commissions, got %d", count)
	}

	// Add commissions with factory_id
	_, _ = db.Exec("INSERT INTO commissions (id, factory_id, title, status) VALUES (?, ?, ?, 'active')", "COMM-001", "FACT-001", "Comm 1")
	_, _ = db.Exec("INSERT INTO commissions (id, factory_id, title, status) VALUES (?, ?, ?, 'active')", "COMM-002", "FACT-001", "Comm 2")

	// Count should be 2
	count, err = repo.CountCommissions(ctx, "FACT-001")
	if err != nil {
		t.Fatalf("CountCommissions failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 commissions, got %d", count)
	}
}
