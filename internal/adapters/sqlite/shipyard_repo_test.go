package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestShipyardRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
	ctx := context.Background()

	// Seed required commission
	seedCommission(t, db, "COMM-001", "Test Commission")

	shipyard := &secondary.ShipyardRecord{
		CommissionID: "COMM-001",
	}

	err := repo.Create(ctx, shipyard)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify ID was generated
	if shipyard.ID == "" {
		t.Error("expected ID to be generated")
	}
	if shipyard.ID != "YARD-001" {
		t.Errorf("expected ID 'YARD-001', got %q", shipyard.ID)
	}

	// Verify round-trip
	got, err := repo.GetByID(ctx, shipyard.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.CommissionID != "COMM-001" {
		t.Errorf("expected commission_id 'COMM-001', got %q", got.CommissionID)
	}
}

func TestShipyardRepository_Create_CommissionNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
	ctx := context.Background()

	shipyard := &secondary.ShipyardRecord{
		CommissionID: "COMM-999",
	}

	err := repo.Create(ctx, shipyard)
	if err == nil {
		t.Error("expected error for non-existent commission")
	}
}

func TestShipyardRepository_Create_DuplicateCommission(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
	ctx := context.Background()

	// Seed required commission
	seedCommission(t, db, "COMM-001", "Test Commission")

	// Create first shipyard
	shipyard1 := &secondary.ShipyardRecord{CommissionID: "COMM-001"}
	err := repo.Create(ctx, shipyard1)
	if err != nil {
		t.Fatalf("First Create failed: %v", err)
	}

	// Try to create second shipyard for same commission
	shipyard2 := &secondary.ShipyardRecord{CommissionID: "COMM-001"}
	err = repo.Create(ctx, shipyard2)
	if err == nil {
		t.Error("expected error for duplicate commission")
	}
}

func TestShipyardRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
	ctx := context.Background()

	// Seed commission and shipyard
	seedCommission(t, db, "COMM-001", "Test Commission")
	seedShipyard(t, db, "YARD-001", "COMM-001")

	got, err := repo.GetByID(ctx, "YARD-001")
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

func TestShipyardRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "YARD-999")
	if err == nil {
		t.Error("expected error for non-existent shipyard")
	}
}

func TestShipyardRepository_GetByCommissionID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
	ctx := context.Background()

	// Seed commission and shipyard
	seedCommission(t, db, "COMM-001", "Test Commission")
	seedShipyard(t, db, "YARD-001", "COMM-001")

	got, err := repo.GetByCommissionID(ctx, "COMM-001")
	if err != nil {
		t.Fatalf("GetByCommissionID failed: %v", err)
	}
	if got.ID != "YARD-001" {
		t.Errorf("expected ID 'YARD-001', got %q", got.ID)
	}
}

func TestShipyardRepository_GetByCommissionID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
	ctx := context.Background()

	// Seed commission without shipyard
	seedCommission(t, db, "COMM-001", "Test Commission")

	_, err := repo.GetByCommissionID(ctx, "COMM-001")
	if err == nil {
		t.Error("expected error for commission without shipyard")
	}
}

func TestShipyardRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
	ctx := context.Background()

	// First ID with empty table
	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if nextID != "YARD-001" {
		t.Errorf("expected 'YARD-001', got %q", nextID)
	}

	// Seed some shipyards
	seedCommission(t, db, "COMM-001", "Test Commission 1")
	seedCommission(t, db, "COMM-002", "Test Commission 2")
	seedShipyard(t, db, "YARD-001", "COMM-001")
	seedShipyard(t, db, "YARD-002", "COMM-002")

	// Next ID after existing
	nextID, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if nextID != "YARD-003" {
		t.Errorf("expected 'YARD-003', got %q", nextID)
	}
}

func TestShipyardRepository_CommissionExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
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

// seedShipyard inserts a test shipyard.
func seedShipyard(t *testing.T, db *sql.DB, id, commissionID string) {
	t.Helper()
	_, err := db.Exec("INSERT INTO shipyards (id, commission_id) VALUES (?, ?)", id, commissionID)
	if err != nil {
		t.Fatalf("failed to seed shipyard: %v", err)
	}
}
