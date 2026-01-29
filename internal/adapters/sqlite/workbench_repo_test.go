package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestWorkbenchRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	// Seed required factory and workshop
	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "test-workshop")

	workbench := &secondary.WorkbenchRecord{
		WorkshopID:   "SHOP-001",
		Name:         "test-workbench",
		WorktreePath: "/tmp/test-workbench",
	}

	err := repo.Create(ctx, workbench)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify ID was generated
	if workbench.ID == "" {
		t.Error("expected ID to be generated")
	}

	// Verify round-trip
	got, err := repo.GetByID(ctx, workbench.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Name != "test-workbench" {
		t.Errorf("expected name 'test-workbench', got %q", got.Name)
	}
	if got.Status != "active" {
		t.Errorf("expected default status 'active', got %q", got.Status)
	}
}

func TestWorkbenchRepository_Create_WorkshopNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	workbench := &secondary.WorkbenchRecord{
		WorkshopID:   "SHOP-999",
		Name:         "test-workbench",
		WorktreePath: "/tmp/test",
	}

	err := repo.Create(ctx, workbench)
	if err == nil {
		t.Error("expected error for non-existent workshop")
	}
}

func TestWorkbenchRepository_Create_WithOptionalFields(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	// Seed required factory and workshop
	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "test-workshop")

	workbench := &secondary.WorkbenchRecord{
		WorkshopID:    "SHOP-001",
		Name:          "test-workbench",
		WorktreePath:  "/tmp/test-workbench",
		RepoID:        "REPO-001",
		HomeBranch:    "main",
		CurrentBranch: "feature/test",
	}

	err := repo.Create(ctx, workbench)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify optional fields
	got, _ := repo.GetByID(ctx, workbench.ID)
	if got.RepoID != "REPO-001" {
		t.Errorf("expected RepoID 'REPO-001', got %q", got.RepoID)
	}
	if got.HomeBranch != "main" {
		t.Errorf("expected HomeBranch 'main', got %q", got.HomeBranch)
	}
	if got.CurrentBranch != "feature/test" {
		t.Errorf("expected CurrentBranch 'feature/test', got %q", got.CurrentBranch)
	}
}

func TestWorkbenchRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	// Seed workbench (includes factory and workshop)
	seedWorkbench(t, db, "BENCH-001", "", "test-workbench")

	got, err := repo.GetByID(ctx, "BENCH-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Name != "test-workbench" {
		t.Errorf("expected name 'test-workbench', got %q", got.Name)
	}
	if got.CreatedAt == "" {
		t.Error("expected CreatedAt to be set")
	}
	if got.UpdatedAt == "" {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestWorkbenchRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "BENCH-999")
	if err == nil {
		t.Error("expected error for non-existent workbench")
	}
}

func TestWorkbenchRepository_GetByPath(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	// Seed workbench with known path
	seedWorkbench(t, db, "BENCH-001", "", "test-workbench")

	got, err := repo.GetByPath(ctx, "/tmp/test/BENCH-001")
	if err != nil {
		t.Fatalf("GetByPath failed: %v", err)
	}
	if got.ID != "BENCH-001" {
		t.Errorf("expected ID 'BENCH-001', got %q", got.ID)
	}
}

func TestWorkbenchRepository_GetByPath_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	_, err := repo.GetByPath(ctx, "/nonexistent/path")
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestWorkbenchRepository_GetByWorkshop(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	// Seed factory, workshops, and workbenches
	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "workshop-1")
	seedWorkshop(t, db, "SHOP-002", "FACT-001", "workshop-2")
	_, _ = db.Exec("INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, 'active')", "BENCH-001", "SHOP-001", "bench-1", "/tmp/bench-1")
	_, _ = db.Exec("INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, 'active')", "BENCH-002", "SHOP-001", "bench-2", "/tmp/bench-2")
	_, _ = db.Exec("INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, 'active')", "BENCH-003", "SHOP-002", "bench-3", "/tmp/bench-3")

	workbenches, err := repo.GetByWorkshop(ctx, "SHOP-001")
	if err != nil {
		t.Fatalf("GetByWorkshop failed: %v", err)
	}
	if len(workbenches) != 2 {
		t.Errorf("expected 2 workbenches for SHOP-001, got %d", len(workbenches))
	}
}

func TestWorkbenchRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	// Seed factory, workshop, and workbenches
	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "test-workshop")
	_, _ = db.Exec("INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, 'active')", "BENCH-001", "SHOP-001", "bench-1", "/tmp/bench-1")
	_, _ = db.Exec("INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, 'active')", "BENCH-002", "SHOP-001", "bench-2", "/tmp/bench-2")
	_, _ = db.Exec("INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, 'active')", "BENCH-003", "SHOP-001", "bench-3", "/tmp/bench-3")

	workbenches, err := repo.List(ctx, "")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(workbenches) != 3 {
		t.Errorf("expected 3 workbenches, got %d", len(workbenches))
	}
}

func TestWorkbenchRepository_List_FilterByWorkshop(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	// Seed factory, workshops, and workbenches
	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "workshop-1")
	seedWorkshop(t, db, "SHOP-002", "FACT-001", "workshop-2")
	_, _ = db.Exec("INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, 'active')", "BENCH-001", "SHOP-001", "bench-1", "/tmp/bench-1")
	_, _ = db.Exec("INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, 'active')", "BENCH-002", "SHOP-001", "bench-2", "/tmp/bench-2")
	_, _ = db.Exec("INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, 'active')", "BENCH-003", "SHOP-002", "bench-3", "/tmp/bench-3")

	workbenches, err := repo.List(ctx, "SHOP-001")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(workbenches) != 2 {
		t.Errorf("expected 2 workbenches for SHOP-001, got %d", len(workbenches))
	}
}

func TestWorkbenchRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	// Seed workbench
	seedWorkbench(t, db, "BENCH-001", "", "original")

	// Update it
	err := repo.Update(ctx, &secondary.WorkbenchRecord{
		ID:   "BENCH-001",
		Name: "updated",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	got, _ := repo.GetByID(ctx, "BENCH-001")
	if got.Name != "updated" {
		t.Errorf("expected name 'updated', got %q", got.Name)
	}
}

func TestWorkbenchRepository_Update_MultipleFields(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	// Seed workbench
	seedWorkbench(t, db, "BENCH-001", "", "test")

	// Update multiple fields (status must be 'active' or 'archived' per CHECK constraint)
	err := repo.Update(ctx, &secondary.WorkbenchRecord{
		ID:            "BENCH-001",
		WorktreePath:  "/new/path",
		Status:        "archived",
		HomeBranch:    "develop",
		CurrentBranch: "feature/new",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	got, _ := repo.GetByID(ctx, "BENCH-001")
	if got.WorktreePath != "/new/path" {
		t.Errorf("expected path '/new/path', got %q", got.WorktreePath)
	}
	if got.Status != "archived" {
		t.Errorf("expected status 'archived', got %q", got.Status)
	}
	if got.HomeBranch != "develop" {
		t.Errorf("expected HomeBranch 'develop', got %q", got.HomeBranch)
	}
	if got.CurrentBranch != "feature/new" {
		t.Errorf("expected CurrentBranch 'feature/new', got %q", got.CurrentBranch)
	}
}

func TestWorkbenchRepository_Update_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, &secondary.WorkbenchRecord{
		ID:   "BENCH-999",
		Name: "updated",
	})
	if err == nil {
		t.Error("expected error for non-existent workbench")
	}
}

func TestWorkbenchRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	// Seed workbench
	seedWorkbench(t, db, "BENCH-001", "", "to-delete")

	// Delete it
	err := repo.Delete(ctx, "BENCH-001")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(ctx, "BENCH-001")
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestWorkbenchRepository_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "BENCH-999")
	if err == nil {
		t.Error("expected error for non-existent workbench")
	}
}

func TestWorkbenchRepository_Rename(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	// Seed workbench
	seedWorkbench(t, db, "BENCH-001", "", "original")

	// Rename it
	err := repo.Rename(ctx, "BENCH-001", "renamed")
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	// Verify rename
	got, _ := repo.GetByID(ctx, "BENCH-001")
	if got.Name != "renamed" {
		t.Errorf("expected name 'renamed', got %q", got.Name)
	}
}

func TestWorkbenchRepository_Rename_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	err := repo.Rename(ctx, "BENCH-999", "renamed")
	if err == nil {
		t.Error("expected error for non-existent workbench")
	}
}

func TestWorkbenchRepository_UpdatePath(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	// Seed workbench
	seedWorkbench(t, db, "BENCH-001", "", "test")

	// Update path
	err := repo.UpdatePath(ctx, "BENCH-001", "/new/path")
	if err != nil {
		t.Fatalf("UpdatePath failed: %v", err)
	}

	// Verify update
	got, _ := repo.GetByID(ctx, "BENCH-001")
	if got.WorktreePath != "/new/path" {
		t.Errorf("expected path '/new/path', got %q", got.WorktreePath)
	}
}

func TestWorkbenchRepository_UpdatePath_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	err := repo.UpdatePath(ctx, "BENCH-999", "/new/path")
	if err == nil {
		t.Error("expected error for non-existent workbench")
	}
}

func TestWorkbenchRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	// First ID should be BENCH-001
	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "BENCH-001" {
		t.Errorf("expected BENCH-001, got %s", id)
	}

	// Seed a workbench
	seedWorkbench(t, db, "BENCH-001", "", "test")

	// Next ID should be BENCH-002
	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "BENCH-002" {
		t.Errorf("expected BENCH-002, got %s", id)
	}
}

func TestWorkbenchRepository_WorkshopExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkbenchRepository(db)
	ctx := context.Background()

	// Workshop doesn't exist
	exists, err := repo.WorkshopExists(ctx, "SHOP-001")
	if err != nil {
		t.Fatalf("WorkshopExists failed: %v", err)
	}
	if exists {
		t.Error("expected workshop to not exist")
	}

	// Seed factory and workshop
	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "test-workshop")

	// Workshop exists
	exists, err = repo.WorkshopExists(ctx, "SHOP-001")
	if err != nil {
		t.Fatalf("WorkshopExists failed: %v", err)
	}
	if !exists {
		t.Error("expected workshop to exist")
	}
}
