package sqlite_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ctxutil"
	"github.com/example/orc/internal/ports/secondary"
)

func TestEventWriterAdapter_EmitAuditCreate(t *testing.T) {
	db := setupTestDB(t)

	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "test-workshop")
	seedWorkbench(t, db, "BENCH-014", "", "orc-014")

	eventRepo := sqlite.NewWorkshopEventRepository(db)
	opRepo := sqlite.NewOperationalEventRepository(db)
	benchRepo := sqlite.NewWorkbenchRepository(db, nil)
	writer := sqlite.NewEventWriterAdapter(eventRepo, opRepo, benchRepo, nil, "abc123")

	ctx := ctxutil.WithActorID(context.Background(), "IMP-BENCH-014")

	err := writer.EmitAuditCreate(ctx, "task", "TASK-001")
	if err != nil {
		t.Fatalf("EmitAuditCreate failed: %v", err)
	}

	events, err := eventRepo.List(ctx, secondary.AuditEventFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	e := events[0]
	if e.EntityType != "task" {
		t.Errorf("EntityType = %q, want %q", e.EntityType, "task")
	}
	if e.EntityID != "TASK-001" {
		t.Errorf("EntityID = %q, want %q", e.EntityID, "TASK-001")
	}
	if e.Action != "create" {
		t.Errorf("Action = %q, want %q", e.Action, "create")
	}
	if e.Source != "ledger" {
		t.Errorf("Source = %q, want %q", e.Source, "ledger")
	}
	if e.Version != "abc123" {
		t.Errorf("Version = %q, want %q", e.Version, "abc123")
	}
	if e.WorkshopID != "SHOP-001" {
		t.Errorf("WorkshopID = %q, want %q", e.WorkshopID, "SHOP-001")
	}
}

func TestEventWriterAdapter_EmitAuditUpdate(t *testing.T) {
	db := setupTestDB(t)

	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "test-workshop")
	seedWorkbench(t, db, "BENCH-014", "", "orc-014")

	eventRepo := sqlite.NewWorkshopEventRepository(db)
	opRepo := sqlite.NewOperationalEventRepository(db)
	benchRepo := sqlite.NewWorkbenchRepository(db, nil)
	writer := sqlite.NewEventWriterAdapter(eventRepo, opRepo, benchRepo, nil, "abc123")

	ctx := ctxutil.WithActorID(context.Background(), "IMP-BENCH-014")

	err := writer.EmitAuditUpdate(ctx, "shipment", "SHIP-001", "status", "draft", "ready")
	if err != nil {
		t.Fatalf("EmitAuditUpdate failed: %v", err)
	}

	events, err := eventRepo.List(ctx, secondary.AuditEventFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	e := events[0]
	if e.Action != "update" {
		t.Errorf("Action = %q, want %q", e.Action, "update")
	}
	if e.FieldName != "status" {
		t.Errorf("FieldName = %q, want %q", e.FieldName, "status")
	}
	if e.OldValue != "draft" {
		t.Errorf("OldValue = %q, want %q", e.OldValue, "draft")
	}
	if e.NewValue != "ready" {
		t.Errorf("NewValue = %q, want %q", e.NewValue, "ready")
	}
}

func TestEventWriterAdapter_EmitAuditDelete(t *testing.T) {
	db := setupTestDB(t)

	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "test-workshop")
	seedWorkbench(t, db, "BENCH-014", "", "orc-014")

	eventRepo := sqlite.NewWorkshopEventRepository(db)
	opRepo := sqlite.NewOperationalEventRepository(db)
	benchRepo := sqlite.NewWorkbenchRepository(db, nil)
	writer := sqlite.NewEventWriterAdapter(eventRepo, opRepo, benchRepo, nil, "abc123")

	ctx := ctxutil.WithActorID(context.Background(), "IMP-BENCH-014")

	err := writer.EmitAuditDelete(ctx, "note", "NOTE-001")
	if err != nil {
		t.Fatalf("EmitAuditDelete failed: %v", err)
	}

	events, err := eventRepo.List(ctx, secondary.AuditEventFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Action != "delete" {
		t.Errorf("Action = %q, want %q", events[0].Action, "delete")
	}
}

func TestEventWriterAdapter_EmitAuditCreate_NoWorkshop(t *testing.T) {
	db := setupTestDB(t)

	eventRepo := sqlite.NewWorkshopEventRepository(db)
	opRepo := sqlite.NewOperationalEventRepository(db)
	benchRepo := sqlite.NewWorkbenchRepository(db, nil)
	writer := sqlite.NewEventWriterAdapter(eventRepo, opRepo, benchRepo, nil, "abc123")

	// No actor in context — workshop cannot be resolved
	ctx := context.Background()

	err := writer.EmitAuditCreate(ctx, "task", "TASK-001")
	if err != nil {
		t.Fatalf("EmitAuditCreate should succeed (skip) without workshop: %v", err)
	}

	events, err := eventRepo.List(ctx, secondary.AuditEventFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events when no workshop, got %d", len(events))
	}
}

func TestEventWriterAdapter_EmitOperational(t *testing.T) {
	db := setupTestDB(t)

	seedFactory(t, db, "FACT-001", "test-factory")
	seedWorkshop(t, db, "SHOP-001", "FACT-001", "test-workshop")
	seedWorkbench(t, db, "BENCH-014", "", "orc-014")

	eventRepo := sqlite.NewWorkshopEventRepository(db)
	opRepo := sqlite.NewOperationalEventRepository(db)
	benchRepo := sqlite.NewWorkbenchRepository(db, nil)
	writer := sqlite.NewEventWriterAdapter(eventRepo, opRepo, benchRepo, nil, "abc123")

	ctx := ctxutil.WithActorID(context.Background(), "IMP-BENCH-014")

	data := map[string]string{"key1": "val1", "key2": "val2"}
	err := writer.EmitOperational(ctx, "poll", "info", "poll completed", data)
	if err != nil {
		t.Fatalf("EmitOperational failed: %v", err)
	}

	events, err := opRepo.List(ctx, secondary.OperationalEventFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	e := events[0]
	if e.Source != "poll" {
		t.Errorf("Source = %q, want %q", e.Source, "poll")
	}
	if e.Level != "info" {
		t.Errorf("Level = %q, want %q", e.Level, "info")
	}
	if e.Message != "poll completed" {
		t.Errorf("Message = %q, want %q", e.Message, "poll completed")
	}
	if e.Version != "abc123" {
		t.Errorf("Version = %q, want %q", e.Version, "abc123")
	}

	// Verify data JSON
	var gotData map[string]string
	if err := json.Unmarshal([]byte(e.DataJSON), &gotData); err != nil {
		t.Fatalf("failed to unmarshal DataJSON: %v", err)
	}
	if gotData["key1"] != "val1" || gotData["key2"] != "val2" {
		t.Errorf("DataJSON decoded = %v, want key1=val1 key2=val2", gotData)
	}
}

func TestEventWriterAdapter_EmitOperational_NoWorkshop(t *testing.T) {
	db := setupTestDB(t)

	eventRepo := sqlite.NewWorkshopEventRepository(db)
	opRepo := sqlite.NewOperationalEventRepository(db)
	benchRepo := sqlite.NewWorkbenchRepository(db, nil)
	writer := sqlite.NewEventWriterAdapter(eventRepo, opRepo, benchRepo, nil, "abc123")

	// No actor — operational events should still persist (workshopID will be empty)
	ctx := context.Background()

	err := writer.EmitOperational(ctx, "workbench", "warn", "no workshop context", nil)
	if err != nil {
		t.Fatalf("EmitOperational should succeed without workshop: %v", err)
	}

	events, err := opRepo.List(ctx, secondary.OperationalEventFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 operational event even without workshop, got %d", len(events))
	}

	if events[0].WorkshopID != "" {
		t.Errorf("WorkshopID = %q, want empty", events[0].WorkshopID)
	}
}

func TestEventWriterAdapter_EmitOperational_NilData(t *testing.T) {
	db := setupTestDB(t)

	eventRepo := sqlite.NewWorkshopEventRepository(db)
	opRepo := sqlite.NewOperationalEventRepository(db)
	writer := sqlite.NewEventWriterAdapter(eventRepo, opRepo, nil, nil, "abc123")

	ctx := context.Background()

	err := writer.EmitOperational(ctx, "ledger", "debug", "test msg", nil)
	if err != nil {
		t.Fatalf("EmitOperational with nil data failed: %v", err)
	}

	events, err := opRepo.List(ctx, secondary.OperationalEventFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].DataJSON != "" {
		t.Errorf("DataJSON = %q, want empty for nil data", events[0].DataJSON)
	}
}
