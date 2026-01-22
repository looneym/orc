package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestCycleRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleRepository(db)
	ctx := context.Background()

	// Create test shipment
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test Shipment", "active")

	t.Run("creates cycle successfully", func(t *testing.T) {
		record := &secondary.CycleRecord{
			ID:             "CYC-001",
			ShipmentID:     "SHIP-001",
			SequenceNumber: 1,
			Status:         "queued",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "CYC-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.SequenceNumber != 1 {
			t.Errorf("SequenceNumber = %d, want %d", got.SequenceNumber, 1)
		}
		if got.Status != "queued" {
			t.Errorf("Status = %q, want %q", got.Status, "queued")
		}
	})

	t.Run("enforces unique shipment+sequence constraint", func(t *testing.T) {
		record := &secondary.CycleRecord{
			ID:             "CYC-002",
			ShipmentID:     "SHIP-001",
			SequenceNumber: 1, // Same as CYC-001
			Status:         "queued",
		}

		err := repo.Create(ctx, record)
		if err == nil {
			t.Fatal("Expected error for duplicate shipment+sequence, got nil")
		}
	})

	t.Run("allows same sequence for different shipments", func(t *testing.T) {
		db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-002", "COMM-001", "Test 2", "active")

		record := &secondary.CycleRecord{
			ID:             "CYC-003",
			ShipmentID:     "SHIP-002",
			SequenceNumber: 1, // Same sequence but different shipment
			Status:         "queued",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	})
}

func TestCycleRepository_GetNextSequenceNumber(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test 1", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-002", "COMM-001", "Test 2", "active")

	t.Run("returns 1 for shipment with no cycles", func(t *testing.T) {
		seq, err := repo.GetNextSequenceNumber(ctx, "SHIP-001")
		if err != nil {
			t.Fatalf("GetNextSequenceNumber failed: %v", err)
		}
		if seq != 1 {
			t.Errorf("seq = %d, want 1", seq)
		}
	})

	t.Run("returns next sequence after creating cycles", func(t *testing.T) {
		repo.Create(ctx, &secondary.CycleRecord{ID: "CYC-001", ShipmentID: "SHIP-001", SequenceNumber: 1, Status: "queued"})
		repo.Create(ctx, &secondary.CycleRecord{ID: "CYC-002", ShipmentID: "SHIP-001", SequenceNumber: 2, Status: "queued"})

		seq, err := repo.GetNextSequenceNumber(ctx, "SHIP-001")
		if err != nil {
			t.Fatalf("GetNextSequenceNumber failed: %v", err)
		}
		if seq != 3 {
			t.Errorf("seq = %d, want 3", seq)
		}
	})

	t.Run("returns 1 for different shipment", func(t *testing.T) {
		seq, err := repo.GetNextSequenceNumber(ctx, "SHIP-002")
		if err != nil {
			t.Fatalf("GetNextSequenceNumber failed: %v", err)
		}
		if seq != 1 {
			t.Errorf("seq = %d, want 1", seq)
		}
	})
}

func TestCycleRepository_GetActiveCycle(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test 1", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-002", "COMM-001", "Test 2", "active")

	repo.Create(ctx, &secondary.CycleRecord{ID: "CYC-001", ShipmentID: "SHIP-001", SequenceNumber: 1, Status: "complete"})
	repo.Create(ctx, &secondary.CycleRecord{ID: "CYC-002", ShipmentID: "SHIP-001", SequenceNumber: 2, Status: "active"})
	repo.Create(ctx, &secondary.CycleRecord{ID: "CYC-003", ShipmentID: "SHIP-001", SequenceNumber: 3, Status: "queued"})

	t.Run("returns active cycle", func(t *testing.T) {
		got, err := repo.GetActiveCycle(ctx, "SHIP-001")
		if err != nil {
			t.Fatalf("GetActiveCycle failed: %v", err)
		}
		if got == nil {
			t.Fatal("expected cycle, got nil")
		}
		if got.ID != "CYC-002" {
			t.Errorf("ID = %q, want %q", got.ID, "CYC-002")
		}
	})

	t.Run("returns nil for shipment without active cycle", func(t *testing.T) {
		got, err := repo.GetActiveCycle(ctx, "SHIP-002")
		if err != nil {
			t.Fatalf("GetActiveCycle failed: %v", err)
		}
		if got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})
}

func TestCycleRepository_GetByShipmentAndSequence(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")

	repo.Create(ctx, &secondary.CycleRecord{ID: "CYC-001", ShipmentID: "SHIP-001", SequenceNumber: 1, Status: "queued"})
	repo.Create(ctx, &secondary.CycleRecord{ID: "CYC-002", ShipmentID: "SHIP-001", SequenceNumber: 2, Status: "queued"})

	t.Run("finds cycle by shipment and sequence", func(t *testing.T) {
		got, err := repo.GetByShipmentAndSequence(ctx, "SHIP-001", 2)
		if err != nil {
			t.Fatalf("GetByShipmentAndSequence failed: %v", err)
		}
		if got.ID != "CYC-002" {
			t.Errorf("ID = %q, want %q", got.ID, "CYC-002")
		}
	})

	t.Run("returns error for non-existent sequence", func(t *testing.T) {
		_, err := repo.GetByShipmentAndSequence(ctx, "SHIP-001", 99)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestCycleRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")

	repo.Create(ctx, &secondary.CycleRecord{ID: "CYC-001", ShipmentID: "SHIP-001", SequenceNumber: 1, Status: "queued"})

	t.Run("updates to active with started_at", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "CYC-001", "active", true, false)
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "CYC-001")
		if got.Status != "active" {
			t.Errorf("Status = %q, want %q", got.Status, "active")
		}
		if got.StartedAt == "" {
			t.Error("StartedAt should be set")
		}
		if got.CompletedAt != "" {
			t.Error("CompletedAt should not be set")
		}
	})

	t.Run("updates to complete with completed_at", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "CYC-001", "complete", false, true)
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "CYC-001")
		if got.Status != "complete" {
			t.Errorf("Status = %q, want %q", got.Status, "complete")
		}
		if got.CompletedAt == "" {
			t.Error("CompletedAt should be set")
		}
	})

	t.Run("returns error for non-existent cycle", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "CYC-999", "active", true, false)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestCycleRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleRepository(db)
	ctx := context.Background()

	t.Run("returns CYC-001 for empty table", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "CYC-001" {
			t.Errorf("ID = %q, want %q", id, "CYC-001")
		}
	})
}

func TestCycleRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test 1", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-002", "COMM-001", "Test 2", "active")

	repo.Create(ctx, &secondary.CycleRecord{ID: "CYC-001", ShipmentID: "SHIP-001", SequenceNumber: 1, Status: "complete"})
	repo.Create(ctx, &secondary.CycleRecord{ID: "CYC-002", ShipmentID: "SHIP-001", SequenceNumber: 2, Status: "active"})
	repo.Create(ctx, &secondary.CycleRecord{ID: "CYC-003", ShipmentID: "SHIP-002", SequenceNumber: 1, Status: "queued"})

	t.Run("lists all cycles", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.CycleFilters{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 3 {
			t.Errorf("len = %d, want 3", len(list))
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.CycleFilters{Status: "active"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "CYC-002" {
			t.Errorf("ID = %q, want %q", list[0].ID, "CYC-002")
		}
	})

	t.Run("filters by shipment", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.CycleFilters{ShipmentID: "SHIP-001"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("orders by sequence number", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.CycleFilters{ShipmentID: "SHIP-001"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if list[0].SequenceNumber != 1 {
			t.Errorf("first cycle seq = %d, want 1", list[0].SequenceNumber)
		}
		if list[1].SequenceNumber != 2 {
			t.Errorf("second cycle seq = %d, want 2", list[1].SequenceNumber)
		}
	})
}

func TestCycleRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")

	repo.Create(ctx, &secondary.CycleRecord{ID: "CYC-001", ShipmentID: "SHIP-001", SequenceNumber: 1, Status: "queued"})

	t.Run("deletes cycle", func(t *testing.T) {
		err := repo.Delete(ctx, "CYC-001")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = repo.GetByID(ctx, "CYC-001")
		if err == nil {
			t.Error("expected error after delete, got nil")
		}
	})

	t.Run("returns error for non-existent cycle", func(t *testing.T) {
		err := repo.Delete(ctx, "CYC-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestCycleRepository_ShipmentExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")

	t.Run("returns true for existing shipment", func(t *testing.T) {
		exists, err := repo.ShipmentExists(ctx, "SHIP-001")
		if err != nil {
			t.Fatalf("ShipmentExists failed: %v", err)
		}
		if !exists {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false for non-existent shipment", func(t *testing.T) {
		exists, err := repo.ShipmentExists(ctx, "SHIP-999")
		if err != nil {
			t.Fatalf("ShipmentExists failed: %v", err)
		}
		if exists {
			t.Error("expected false, got true")
		}
	})
}
