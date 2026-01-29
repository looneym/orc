package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestManifestRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewManifestRepository(db)
	ctx := context.Background()

	// Create test fixtures: commission -> shipment
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test Shipment", "draft")

	t.Run("creates manifest successfully", func(t *testing.T) {
		record := &secondary.ManifestRecord{
			ID:         "MAN-001",
			ShipmentID: "SHIP-001",
			CreatedBy:  "ORC",
			Status:     "draft",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "MAN-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.ShipmentID != "SHIP-001" {
			t.Errorf("ShipmentID = %q, want %q", got.ShipmentID, "SHIP-001")
		}
		if got.CreatedBy != "ORC" {
			t.Errorf("CreatedBy = %q, want %q", got.CreatedBy, "ORC")
		}
		if got.Status != "draft" {
			t.Errorf("Status = %q, want %q", got.Status, "draft")
		}
	})

	t.Run("creates manifest with optional fields", func(t *testing.T) {
		db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-002", "COMM-001", "Test 2", "draft")

		record := &secondary.ManifestRecord{
			ID:            "MAN-002",
			ShipmentID:    "SHIP-002",
			CreatedBy:     "ORC",
			Attestation:   "All tasks verified",
			Tasks:         `["TASK-001","TASK-002"]`,
			OrderingNotes: "Start with TASK-001",
			Status:        "draft",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "MAN-002")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.Attestation != "All tasks verified" {
			t.Errorf("Attestation = %q, want %q", got.Attestation, "All tasks verified")
		}
		if got.Tasks != `["TASK-001","TASK-002"]` {
			t.Errorf("Tasks = %q, want %q", got.Tasks, `["TASK-001","TASK-002"]`)
		}
		if got.OrderingNotes != "Start with TASK-001" {
			t.Errorf("OrderingNotes = %q, want %q", got.OrderingNotes, "Start with TASK-001")
		}
	})

	t.Run("enforces unique shipment constraint", func(t *testing.T) {
		record := &secondary.ManifestRecord{
			ID:         "MAN-003",
			ShipmentID: "SHIP-001", // Same shipment as MAN-001
			CreatedBy:  "ORC",
			Status:     "draft",
		}

		err := repo.Create(ctx, record)
		if err == nil {
			t.Fatal("Expected error for duplicate shipment, got nil")
		}
	})
}

func TestManifestRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewManifestRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "draft")

	repo.Create(ctx, &secondary.ManifestRecord{
		ID:         "MAN-001",
		ShipmentID: "SHIP-001",
		CreatedBy:  "ORC",
		Status:     "draft",
	})

	t.Run("finds manifest by ID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, "MAN-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if got.ID != "MAN-001" {
			t.Errorf("ID = %q, want %q", got.ID, "MAN-001")
		}
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "MAN-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestManifestRepository_GetByShipment(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewManifestRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "draft")

	repo.Create(ctx, &secondary.ManifestRecord{
		ID:         "MAN-001",
		ShipmentID: "SHIP-001",
		CreatedBy:  "ORC",
		Status:     "draft",
	})

	t.Run("finds manifest by shipment", func(t *testing.T) {
		got, err := repo.GetByShipment(ctx, "SHIP-001")
		if err != nil {
			t.Fatalf("GetByShipment failed: %v", err)
		}
		if got == nil {
			t.Fatal("expected manifest, got nil")
		}
		if got.ID != "MAN-001" {
			t.Errorf("ID = %q, want %q", got.ID, "MAN-001")
		}
	})

	t.Run("returns error for shipment without manifest", func(t *testing.T) {
		_, err := repo.GetByShipment(ctx, "SHIP-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestManifestRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewManifestRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test 1", "draft")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-002", "COMM-001", "Test 2", "draft")

	repo.Create(ctx, &secondary.ManifestRecord{ID: "MAN-001", ShipmentID: "SHIP-001", CreatedBy: "ORC", Status: "draft"})
	repo.Create(ctx, &secondary.ManifestRecord{ID: "MAN-002", ShipmentID: "SHIP-002", CreatedBy: "ORC", Status: "launched"})

	t.Run("lists all manifests", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.ManifestFilters{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("filters by shipment_id", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.ManifestFilters{ShipmentID: "SHIP-002"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "MAN-002" {
			t.Errorf("ID = %q, want %q", list[0].ID, "MAN-002")
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.ManifestFilters{Status: "launched"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "MAN-002" {
			t.Errorf("ID = %q, want %q", list[0].ID, "MAN-002")
		}
	})
}

func TestManifestRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewManifestRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "draft")

	repo.Create(ctx, &secondary.ManifestRecord{
		ID:         "MAN-001",
		ShipmentID: "SHIP-001",
		CreatedBy:  "ORC",
		Status:     "draft",
	})

	t.Run("updates attestation", func(t *testing.T) {
		err := repo.Update(ctx, &secondary.ManifestRecord{
			ID:          "MAN-001",
			Attestation: "Updated attestation",
		})
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "MAN-001")
		if got.Attestation != "Updated attestation" {
			t.Errorf("Attestation = %q, want %q", got.Attestation, "Updated attestation")
		}
	})

	t.Run("returns error for non-existent manifest", func(t *testing.T) {
		err := repo.Update(ctx, &secondary.ManifestRecord{
			ID:          "MAN-999",
			Attestation: "Will fail",
		})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestManifestRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewManifestRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "draft")

	repo.Create(ctx, &secondary.ManifestRecord{ID: "MAN-001", ShipmentID: "SHIP-001", CreatedBy: "ORC", Status: "draft"})

	t.Run("deletes manifest", func(t *testing.T) {
		err := repo.Delete(ctx, "MAN-001")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = repo.GetByID(ctx, "MAN-001")
		if err == nil {
			t.Error("expected error after delete, got nil")
		}
	})

	t.Run("returns error for non-existent manifest", func(t *testing.T) {
		err := repo.Delete(ctx, "MAN-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestManifestRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewManifestRepository(db)
	ctx := context.Background()

	t.Run("returns MAN-001 for empty table", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "MAN-001" {
			t.Errorf("ID = %q, want %q", id, "MAN-001")
		}
	})
}

func TestManifestRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewManifestRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "draft")

	repo.Create(ctx, &secondary.ManifestRecord{
		ID:         "MAN-001",
		ShipmentID: "SHIP-001",
		CreatedBy:  "ORC",
		Status:     "draft",
	})

	t.Run("updates to launched", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "MAN-001", "launched")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "MAN-001")
		if got.Status != "launched" {
			t.Errorf("Status = %q, want %q", got.Status, "launched")
		}
	})

	t.Run("returns error for non-existent manifest", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "MAN-999", "launched")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestManifestRepository_ShipmentExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewManifestRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "draft")

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

func TestManifestRepository_ShipmentHasManifest(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewManifestRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test 1", "draft")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-002", "COMM-001", "Test 2", "draft")

	repo.Create(ctx, &secondary.ManifestRecord{
		ID:         "MAN-001",
		ShipmentID: "SHIP-001",
		CreatedBy:  "ORC",
		Status:     "draft",
	})

	t.Run("returns true when shipment has manifest", func(t *testing.T) {
		has, err := repo.ShipmentHasManifest(ctx, "SHIP-001")
		if err != nil {
			t.Fatalf("ShipmentHasManifest failed: %v", err)
		}
		if !has {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false when shipment has no manifest", func(t *testing.T) {
		has, err := repo.ShipmentHasManifest(ctx, "SHIP-002")
		if err != nil {
			t.Fatalf("ShipmentHasManifest failed: %v", err)
		}
		if has {
			t.Error("expected false, got true")
		}
	})
}
