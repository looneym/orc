package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestWorkOrderRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkOrderRepository(db)
	ctx := context.Background()

	// Create test shipment
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test Shipment", "active")

	t.Run("creates work order successfully", func(t *testing.T) {
		record := &secondary.WorkOrderRecord{
			ID:         "WO-001",
			ShipmentID: "SHIP-001",
			Outcome:    "Implement feature X",
			Status:     "draft",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "WO-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.Outcome != "Implement feature X" {
			t.Errorf("Outcome = %q, want %q", got.Outcome, "Implement feature X")
		}
		if got.Status != "draft" {
			t.Errorf("Status = %q, want %q", got.Status, "draft")
		}
	})

	t.Run("creates work order with acceptance criteria", func(t *testing.T) {
		db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-002", "COMM-001", "Test 2", "active")

		record := &secondary.WorkOrderRecord{
			ID:                 "WO-002",
			ShipmentID:         "SHIP-002",
			Outcome:            "Implement feature Y",
			AcceptanceCriteria: `["Tests pass", "No lint errors"]`,
			Status:             "draft",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "WO-002")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.AcceptanceCriteria != `["Tests pass", "No lint errors"]` {
			t.Errorf("AcceptanceCriteria = %q, want %q", got.AcceptanceCriteria, `["Tests pass", "No lint errors"]`)
		}
	})

	t.Run("enforces unique shipment constraint", func(t *testing.T) {
		record := &secondary.WorkOrderRecord{
			ID:         "WO-003",
			ShipmentID: "SHIP-001", // Same shipment as WO-001
			Outcome:    "Duplicate",
			Status:     "draft",
		}

		err := repo.Create(ctx, record)
		if err == nil {
			t.Fatal("Expected error for duplicate shipment, got nil")
		}
	})
}

func TestWorkOrderRepository_GetByShipment(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkOrderRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")

	repo.Create(ctx, &secondary.WorkOrderRecord{
		ID:         "WO-001",
		ShipmentID: "SHIP-001",
		Outcome:    "Test outcome",
		Status:     "draft",
	})

	t.Run("finds work order by shipment", func(t *testing.T) {
		got, err := repo.GetByShipment(ctx, "SHIP-001")
		if err != nil {
			t.Fatalf("GetByShipment failed: %v", err)
		}
		if got == nil {
			t.Fatal("expected work order, got nil")
		}
		if got.ID != "WO-001" {
			t.Errorf("ID = %q, want %q", got.ID, "WO-001")
		}
	})

	t.Run("returns error for shipment without work order", func(t *testing.T) {
		_, err := repo.GetByShipment(ctx, "SHIP-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestWorkOrderRepository_ShipmentHasWorkOrder(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkOrderRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test 1", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-002", "COMM-001", "Test 2", "active")

	repo.Create(ctx, &secondary.WorkOrderRecord{
		ID:         "WO-001",
		ShipmentID: "SHIP-001",
		Outcome:    "Test",
		Status:     "draft",
	})

	t.Run("returns true when shipment has work order", func(t *testing.T) {
		has, err := repo.ShipmentHasWorkOrder(ctx, "SHIP-001")
		if err != nil {
			t.Fatalf("ShipmentHasWorkOrder failed: %v", err)
		}
		if !has {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false when shipment has no work order", func(t *testing.T) {
		has, err := repo.ShipmentHasWorkOrder(ctx, "SHIP-002")
		if err != nil {
			t.Fatalf("ShipmentHasWorkOrder failed: %v", err)
		}
		if has {
			t.Error("expected false, got true")
		}
	})
}

func TestWorkOrderRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkOrderRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")

	repo.Create(ctx, &secondary.WorkOrderRecord{
		ID:         "WO-001",
		ShipmentID: "SHIP-001",
		Outcome:    "Test",
		Status:     "draft",
	})

	t.Run("updates to active", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "WO-001", "active")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "WO-001")
		if got.Status != "active" {
			t.Errorf("Status = %q, want %q", got.Status, "active")
		}
	})

	t.Run("updates to complete", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "WO-001", "complete")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "WO-001")
		if got.Status != "complete" {
			t.Errorf("Status = %q, want %q", got.Status, "complete")
		}
	})

	t.Run("returns error for non-existent work order", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "WO-999", "active")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestWorkOrderRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkOrderRepository(db)
	ctx := context.Background()

	t.Run("returns WO-001 for empty table", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "WO-001" {
			t.Errorf("ID = %q, want %q", id, "WO-001")
		}
	})
}

func TestWorkOrderRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkOrderRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test 1", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-002", "COMM-001", "Test 2", "active")

	repo.Create(ctx, &secondary.WorkOrderRecord{ID: "WO-001", ShipmentID: "SHIP-001", Outcome: "Test 1", Status: "draft"})
	repo.Create(ctx, &secondary.WorkOrderRecord{ID: "WO-002", ShipmentID: "SHIP-002", Outcome: "Test 2", Status: "active"})

	t.Run("lists all work orders", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.WorkOrderFilters{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.WorkOrderFilters{Status: "draft"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "WO-001" {
			t.Errorf("ID = %q, want %q", list[0].ID, "WO-001")
		}
	})

	t.Run("filters by shipment", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.WorkOrderFilters{ShipmentID: "SHIP-002"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "WO-002" {
			t.Errorf("ID = %q, want %q", list[0].ID, "WO-002")
		}
	})
}

func TestWorkOrderRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewWorkOrderRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")

	repo.Create(ctx, &secondary.WorkOrderRecord{ID: "WO-001", ShipmentID: "SHIP-001", Outcome: "Test", Status: "draft"})

	t.Run("deletes work order", func(t *testing.T) {
		err := repo.Delete(ctx, "WO-001")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = repo.GetByID(ctx, "WO-001")
		if err == nil {
			t.Error("expected error after delete, got nil")
		}
	})

	t.Run("returns error for non-existent work order", func(t *testing.T) {
		err := repo.Delete(ctx, "WO-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
