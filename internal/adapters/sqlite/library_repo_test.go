package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestLibraryRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewLibraryRepository(db)
	ctx := context.Background()

	// Seed required commission
	seedCommission(t, db, "COMM-001", "Test Commission")

	library := &secondary.LibraryRecord{
		CommissionID: "COMM-001",
	}

	err := repo.Create(ctx, library)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify ID was generated
	if library.ID == "" {
		t.Error("expected ID to be generated")
	}
	if library.ID != "LIB-001" {
		t.Errorf("expected ID 'LIB-001', got %q", library.ID)
	}

	// Verify round-trip
	got, err := repo.GetByID(ctx, library.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.CommissionID != "COMM-001" {
		t.Errorf("expected commission_id 'COMM-001', got %q", got.CommissionID)
	}
}

func TestLibraryRepository_Create_CommissionNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewLibraryRepository(db)
	ctx := context.Background()

	library := &secondary.LibraryRecord{
		CommissionID: "COMM-999",
	}

	err := repo.Create(ctx, library)
	if err == nil {
		t.Error("expected error for non-existent commission")
	}
}

func TestLibraryRepository_Create_DuplicateCommission(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewLibraryRepository(db)
	ctx := context.Background()

	// Seed required commission
	seedCommission(t, db, "COMM-001", "Test Commission")

	// Create first library
	library1 := &secondary.LibraryRecord{CommissionID: "COMM-001"}
	err := repo.Create(ctx, library1)
	if err != nil {
		t.Fatalf("First Create failed: %v", err)
	}

	// Try to create second library for same commission
	library2 := &secondary.LibraryRecord{CommissionID: "COMM-001"}
	err = repo.Create(ctx, library2)
	if err == nil {
		t.Error("expected error for duplicate commission")
	}
}

func TestLibraryRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewLibraryRepository(db)
	ctx := context.Background()

	// Seed commission and library
	seedCommission(t, db, "COMM-001", "Test Commission")
	seedLibrary(t, db, "LIB-001", "COMM-001")

	got, err := repo.GetByID(ctx, "LIB-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.CommissionID != "COMM-001" {
		t.Errorf("expected commission_id 'COMM-001', got %q", got.CommissionID)
	}
	if got.CreatedAt == "" {
		t.Error("expected CreatedAt to be set")
	}
	if got.UpdatedAt == "" {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestLibraryRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewLibraryRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "LIB-999")
	if err == nil {
		t.Error("expected error for non-existent library")
	}
}

func TestLibraryRepository_GetByCommissionID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewLibraryRepository(db)
	ctx := context.Background()

	// Seed commission and library
	seedCommission(t, db, "COMM-001", "Test Commission")
	seedLibrary(t, db, "LIB-001", "COMM-001")

	got, err := repo.GetByCommissionID(ctx, "COMM-001")
	if err != nil {
		t.Fatalf("GetByCommissionID failed: %v", err)
	}
	if got.ID != "LIB-001" {
		t.Errorf("expected ID 'LIB-001', got %q", got.ID)
	}
}

func TestLibraryRepository_GetByCommissionID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewLibraryRepository(db)
	ctx := context.Background()

	// Seed commission without library
	seedCommission(t, db, "COMM-001", "Test Commission")

	_, err := repo.GetByCommissionID(ctx, "COMM-001")
	if err == nil {
		t.Error("expected error for commission without library")
	}
}

func TestLibraryRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewLibraryRepository(db)
	ctx := context.Background()

	// First ID with empty table
	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if nextID != "LIB-001" {
		t.Errorf("expected 'LIB-001', got %q", nextID)
	}

	// Seed some libraries
	seedCommission(t, db, "COMM-001", "Test Commission 1")
	seedCommission(t, db, "COMM-002", "Test Commission 2")
	seedLibrary(t, db, "LIB-001", "COMM-001")
	seedLibrary(t, db, "LIB-002", "COMM-002")

	// Next ID after existing
	nextID, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if nextID != "LIB-003" {
		t.Errorf("expected 'LIB-003', got %q", nextID)
	}
}

func TestLibraryRepository_CommissionExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewLibraryRepository(db)
	ctx := context.Background()

	// Commission doesn't exist
	exists, err := repo.CommissionExists(ctx, "COMM-001")
	if err != nil {
		t.Fatalf("CommissionExists failed: %v", err)
	}
	if exists {
		t.Error("expected commission to not exist")
	}

	// Seed commission
	seedCommission(t, db, "COMM-001", "Test Commission")

	// Commission exists
	exists, err = repo.CommissionExists(ctx, "COMM-001")
	if err != nil {
		t.Fatalf("CommissionExists failed: %v", err)
	}
	if !exists {
		t.Error("expected commission to exist")
	}
}

// seedLibrary inserts a test library.
func seedLibrary(t *testing.T, db *sql.DB, id, commissionID string) {
	t.Helper()
	_, err := db.Exec("INSERT INTO libraries (id, commission_id) VALUES (?, ?)", id, commissionID)
	if err != nil {
		t.Fatalf("failed to seed library: %v", err)
	}
}
