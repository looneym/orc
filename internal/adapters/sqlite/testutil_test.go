// Package sqlite_test contains integration tests for SQLite repositories.
//
// # Schema Protection
//
// This file is the SINGLE POINT where the database schema is loaded for tests.
// All test setup functions use db.GetSchemaSQL() to ensure tests run against
// the authoritative schema, preventing drift between test and production.
//
// DO NOT hardcode CREATE TABLE statements in test files. `make schema-check`
// will fail if you do. Instead, use setupTestDB() and the seed* helpers.
package sqlite_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/example/orc/internal/db"
)

// setupTestDB creates an in-memory database with the authoritative schema.
// This is the single shared test database setup function for all repository tests.
// Uses db.GetSchemaSQL() to prevent test schemas from drifting.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Use the authoritative schema from schema.go
	_, err = testDB.Exec(db.GetSchemaSQL())
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	t.Cleanup(func() {
		testDB.Close()
	})

	return testDB
}

// seedCommission inserts a test commission and returns its ID.
func seedCommission(t *testing.T, db *sql.DB, id, title string) string {
	t.Helper()
	if id == "" {
		id = "COMM-001"
	}
	if title == "" {
		title = "Test Commission"
	}
	_, err := db.Exec("INSERT INTO commissions (id, title, status) VALUES (?, ?, 'active')", id, title)
	if err != nil {
		t.Fatalf("failed to seed commission: %v", err)
	}
	return id
}

// seedShipment inserts a test shipment and returns its ID.
func seedShipment(t *testing.T, db *sql.DB, id, commissionID, title string) string {
	t.Helper()
	if id == "" {
		id = "SHIP-001"
	}
	if commissionID == "" {
		commissionID = "COMM-001"
	}
	if title == "" {
		title = "Test Shipment"
	}
	_, err := db.Exec("INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, 'draft')", id, commissionID, title)
	if err != nil {
		t.Fatalf("failed to seed shipment: %v", err)
	}
	return id
}

// seedTask inserts a test task and returns its ID.
func seedTask(t *testing.T, db *sql.DB, id, commissionID, title string) string {
	t.Helper()
	if id == "" {
		id = "TASK-001"
	}
	if commissionID == "" {
		commissionID = "COMM-001"
	}
	if title == "" {
		title = "Test Task"
	}
	_, err := db.Exec("INSERT INTO tasks (id, commission_id, title, status) VALUES (?, ?, ?, 'ready')", id, commissionID, title)
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}
	return id
}

// seedFactory inserts a test factory (if not exists) and returns its ID.
func seedFactory(t *testing.T, db *sql.DB, id, name string) string {
	t.Helper()
	if id == "" {
		id = "FACT-001"
	}
	if name == "" {
		name = "test-factory"
	}
	_, err := db.Exec("INSERT OR IGNORE INTO factories (id, name, status) VALUES (?, ?, 'active')", id, name)
	if err != nil {
		t.Fatalf("failed to seed factory: %v", err)
	}
	return id
}

// seedWorkshop inserts a test workshop (if not exists) and returns its ID.
func seedWorkshop(t *testing.T, db *sql.DB, id, factoryID, name string) string {
	t.Helper()
	if id == "" {
		id = "SHOP-001"
	}
	if factoryID == "" {
		factoryID = "FACT-001"
	}
	if name == "" {
		name = "test-workshop"
	}
	_, err := db.Exec("INSERT OR IGNORE INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, 'active')", id, factoryID, name)
	if err != nil {
		t.Fatalf("failed to seed workshop: %v", err)
	}
	return id
}

// seedWorkbench inserts a test workbench (and required factory/workshop) and returns its ID.
// The commissionID parameter is legacy and ignored - workbenches now use workshop hierarchy.
func seedWorkbench(t *testing.T, db *sql.DB, id, _ /* commissionID */, name string) string {
	t.Helper()
	if id == "" {
		id = "BENCH-001"
	}
	// Seed factory and workshop if they don't exist
	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "test-workshop")

	if name == "" {
		name = "test-workbench"
	}
	path := "/tmp/test/" + id // Use ID to ensure unique paths
	_, err := db.Exec("INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, 'active')", id, "SHOP-001", name, path)
	if err != nil {
		t.Fatalf("failed to seed workbench: %v", err)
	}
	return id
}

// seedGatehouse inserts a test gatehouse and returns its ID.
func seedGatehouse(t *testing.T, db *sql.DB, id, workshopID string) string {
	t.Helper()
	if id == "" {
		id = "GATE-001"
	}
	if workshopID == "" {
		workshopID = "SHOP-001"
	}
	_, err := db.Exec("INSERT INTO gatehouses (id, workshop_id, status) VALUES (?, ?, 'active')", id, workshopID)
	if err != nil {
		t.Fatalf("failed to seed gatehouse: %v", err)
	}
	return id
}

// seedTag inserts a test tag and returns its ID.
func seedTag(t *testing.T, db *sql.DB, id, name string) string {
	t.Helper()
	if id == "" {
		id = "TAG-001"
	}
	if name == "" {
		name = "test-tag"
	}
	_, err := db.Exec("INSERT INTO tags (id, name) VALUES (?, ?)", id, name)
	if err != nil {
		t.Fatalf("failed to seed tag: %v", err)
	}
	return id
}
