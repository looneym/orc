package app

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockOperationalEventRepository implements secondary.OperationalEventRepository for testing.
type mockOperationalEventRepository struct {
	events map[string]*secondary.OperationalEventRecord
	nextID int
}

func newMockOperationalEventRepository() *mockOperationalEventRepository {
	return &mockOperationalEventRepository{
		events: make(map[string]*secondary.OperationalEventRecord),
		nextID: 1,
	}
}

func (m *mockOperationalEventRepository) Create(ctx context.Context, event *secondary.OperationalEventRecord) error {
	m.events[event.ID] = event
	return nil
}

func (m *mockOperationalEventRepository) List(ctx context.Context, filters secondary.OperationalEventFilters) ([]*secondary.OperationalEventRecord, error) {
	var result []*secondary.OperationalEventRecord
	for _, e := range m.events {
		if filters.WorkshopID != "" && e.WorkshopID != filters.WorkshopID {
			continue
		}
		if filters.Source != "" && e.Source != filters.Source {
			continue
		}
		if filters.Level != "" && e.Level != filters.Level {
			continue
		}
		result = append(result, e)
	}
	if filters.Limit > 0 && len(result) > filters.Limit {
		result = result[:filters.Limit]
	}
	return result, nil
}

func (m *mockOperationalEventRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("OE-%04d", id), nil
}

func (m *mockOperationalEventRepository) PruneOlderThan(ctx context.Context, days int) (int, error) {
	count := 0
	cutoff := time.Now().AddDate(0, 0, -days)
	for id, event := range m.events {
		ts, err := time.Parse(time.RFC3339, event.Timestamp)
		if err != nil {
			continue
		}
		if ts.Before(cutoff) {
			delete(m.events, id)
			count++
		}
	}
	return count, nil
}

func newTestEventService() (*EventServiceImpl, *mockWorkshopEventRepository, *mockOperationalEventRepository) {
	auditRepo := newMockWorkshopEventRepository()
	opsRepo := newMockOperationalEventRepository()
	service := NewEventService(auditRepo, opsRepo)
	return service, auditRepo, opsRepo
}

func TestEventService_ListEvents_All(t *testing.T) {
	service, auditRepo, opsRepo := newTestEventService()
	ctx := context.Background()

	auditRepo.events["WE-0001"] = &secondary.AuditEventRecord{
		ID: "WE-0001", WorkshopID: "SHOP-001", EntityType: "task",
		EntityID: "TASK-001", Action: "create", Timestamp: "2024-01-01T12:00:00Z",
	}
	opsRepo.events["OE-0001"] = &secondary.OperationalEventRecord{
		ID: "OE-0001", WorkshopID: "SHOP-001", Source: "hook",
		Level: "info", Message: "hook fired", Timestamp: "2024-01-01T12:01:00Z",
	}

	events, err := service.ListEvents(ctx, primary.EventFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	// Should be sorted descending by timestamp
	if events[0].ID != "OE-0001" {
		t.Errorf("expected first event OE-0001 (newer), got %s", events[0].ID)
	}
	if events[1].ID != "WE-0001" {
		t.Errorf("expected second event WE-0001 (older), got %s", events[1].ID)
	}
}

func TestEventService_ListEvents_AuditOnly(t *testing.T) {
	service, auditRepo, opsRepo := newTestEventService()
	ctx := context.Background()

	auditRepo.events["WE-0001"] = &secondary.AuditEventRecord{
		ID: "WE-0001", Timestamp: "2024-01-01T12:00:00Z", EntityType: "task", Action: "create",
	}
	opsRepo.events["OE-0001"] = &secondary.OperationalEventRecord{
		ID: "OE-0001", Timestamp: "2024-01-01T12:01:00Z", Source: "hook", Level: "info", Message: "test",
	}

	events, err := service.ListEvents(ctx, primary.EventFilters{EventType: "audit"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].ID != "WE-0001" {
		t.Errorf("expected WE-0001, got %s", events[0].ID)
	}
}

func TestEventService_ListEvents_OpsOnly(t *testing.T) {
	service, auditRepo, opsRepo := newTestEventService()
	ctx := context.Background()

	auditRepo.events["WE-0001"] = &secondary.AuditEventRecord{
		ID: "WE-0001", Timestamp: "2024-01-01T12:00:00Z", EntityType: "task", Action: "create",
	}
	opsRepo.events["OE-0001"] = &secondary.OperationalEventRecord{
		ID: "OE-0001", Timestamp: "2024-01-01T12:01:00Z", Source: "hook", Level: "info", Message: "test",
	}

	events, err := service.ListEvents(ctx, primary.EventFilters{EventType: "ops"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].ID != "OE-0001" {
		t.Errorf("expected OE-0001, got %s", events[0].ID)
	}
}

func TestEventService_ListEvents_WithLimit(t *testing.T) {
	service, auditRepo, opsRepo := newTestEventService()
	ctx := context.Background()

	auditRepo.events["WE-0001"] = &secondary.AuditEventRecord{
		ID: "WE-0001", Timestamp: "2024-01-01T12:00:00Z", EntityType: "task", Action: "create",
	}
	auditRepo.events["WE-0002"] = &secondary.AuditEventRecord{
		ID: "WE-0002", Timestamp: "2024-01-01T12:02:00Z", EntityType: "task", Action: "update",
	}
	opsRepo.events["OE-0001"] = &secondary.OperationalEventRecord{
		ID: "OE-0001", Timestamp: "2024-01-01T12:01:00Z", Source: "hook", Level: "info", Message: "test",
	}

	events, err := service.ListEvents(ctx, primary.EventFilters{Limit: 2})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events (limited), got %d", len(events))
	}
}

func TestEventService_ListEvents_MergeSort(t *testing.T) {
	service, auditRepo, opsRepo := newTestEventService()
	ctx := context.Background()

	// Interleaved timestamps
	auditRepo.events["WE-0001"] = &secondary.AuditEventRecord{
		ID: "WE-0001", Timestamp: "2024-01-01T12:00:00Z", EntityType: "task", Action: "create",
	}
	auditRepo.events["WE-0002"] = &secondary.AuditEventRecord{
		ID: "WE-0002", Timestamp: "2024-01-01T12:04:00Z", EntityType: "shipment", Action: "update",
	}
	opsRepo.events["OE-0001"] = &secondary.OperationalEventRecord{
		ID: "OE-0001", Timestamp: "2024-01-01T12:02:00Z", Source: "hook", Level: "info", Message: "first",
	}
	opsRepo.events["OE-0002"] = &secondary.OperationalEventRecord{
		ID: "OE-0002", Timestamp: "2024-01-01T12:06:00Z", Source: "system", Level: "warn", Message: "second",
	}

	events, err := service.ListEvents(ctx, primary.EventFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}
	// Should be: OE-0002 (12:06), WE-0002 (12:04), OE-0001 (12:02), WE-0001 (12:00)
	expectedOrder := []string{"OE-0002", "WE-0002", "OE-0001", "WE-0001"}
	for i, id := range expectedOrder {
		if events[i].ID != id {
			t.Errorf("position %d: expected %s, got %s", i, id, events[i].ID)
		}
	}
}

func TestEventService_GetEvent(t *testing.T) {
	service, auditRepo, _ := newTestEventService()
	ctx := context.Background()

	auditRepo.events["WE-0001"] = &secondary.AuditEventRecord{
		ID: "WE-0001", WorkshopID: "SHOP-001", Timestamp: "2024-01-01T12:00:00Z",
		EntityType: "task", EntityID: "TASK-001", Action: "create",
	}

	event, err := service.GetEvent(ctx, "WE-0001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if event.EntityID != "TASK-001" {
		t.Errorf("expected entityID 'TASK-001', got %q", event.EntityID)
	}
}

func TestEventService_GetEvent_NotFound(t *testing.T) {
	service, _, _ := newTestEventService()
	ctx := context.Background()

	_, err := service.GetEvent(ctx, "WE-9999")
	if err == nil {
		t.Error("expected error for non-existent event")
	}
}

func TestEventService_PruneEvents(t *testing.T) {
	service, auditRepo, opsRepo := newTestEventService()
	ctx := context.Background()

	oldTime := time.Now().AddDate(0, 0, -60).Format(time.RFC3339)
	newTime := time.Now().Format(time.RFC3339)

	auditRepo.events["WE-0001"] = &secondary.AuditEventRecord{
		ID: "WE-0001", Timestamp: oldTime, EntityType: "task",
	}
	auditRepo.events["WE-0002"] = &secondary.AuditEventRecord{
		ID: "WE-0002", Timestamp: newTime, EntityType: "task",
	}
	opsRepo.events["OE-0001"] = &secondary.OperationalEventRecord{
		ID: "OE-0001", Timestamp: oldTime, Source: "hook", Level: "info", Message: "old",
	}
	opsRepo.events["OE-0002"] = &secondary.OperationalEventRecord{
		ID: "OE-0002", Timestamp: newTime, Source: "hook", Level: "info", Message: "new",
	}

	count, err := service.PruneEvents(ctx, 30)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 pruned (1 from each repo), got %d", count)
	}
	if len(auditRepo.events) != 1 {
		t.Errorf("expected 1 audit event remaining, got %d", len(auditRepo.events))
	}
	if len(opsRepo.events) != 1 {
		t.Errorf("expected 1 ops event remaining, got %d", len(opsRepo.events))
	}
}

func TestEventService_AuditEventFields(t *testing.T) {
	service, auditRepo, _ := newTestEventService()
	ctx := context.Background()

	auditRepo.events["WE-0001"] = &secondary.AuditEventRecord{
		ID: "WE-0001", WorkshopID: "SHOP-001", Timestamp: "2024-01-01T12:00:00Z",
		ActorID: "IMP-BENCH-001", Source: "ledger", Version: "1.0",
		EntityType: "task", EntityID: "TASK-001", Action: "update",
		FieldName: "status", OldValue: "open", NewValue: "in-progress",
	}

	events, err := service.ListEvents(ctx, primary.EventFilters{EventType: "audit"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	e := events[0]
	if e.EntityType != "task" {
		t.Errorf("EntityType = %q, want %q", e.EntityType, "task")
	}
	if e.Source != "ledger" {
		t.Errorf("Source = %q, want %q", e.Source, "ledger")
	}
	// Operational fields should be zero
	if e.Level != "" {
		t.Errorf("Level should be empty for audit events, got %q", e.Level)
	}
	if e.Message != "" {
		t.Errorf("Message should be empty for audit events, got %q", e.Message)
	}
}

func TestEventService_OpsEventFields(t *testing.T) {
	service, _, opsRepo := newTestEventService()
	ctx := context.Background()

	opsRepo.events["OE-0001"] = &secondary.OperationalEventRecord{
		ID: "OE-0001", WorkshopID: "SHOP-001", Timestamp: "2024-01-01T12:00:00Z",
		ActorID: "IMP-BENCH-001", Source: "hook", Version: "1.0",
		Level: "info", Message: "hook executed", DataJSON: `{"key":"val"}`,
	}

	events, err := service.ListEvents(ctx, primary.EventFilters{EventType: "ops"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	e := events[0]
	if e.Level != "info" {
		t.Errorf("Level = %q, want %q", e.Level, "info")
	}
	if e.Message != "hook executed" {
		t.Errorf("Message = %q, want %q", e.Message, "hook executed")
	}
	if e.Data != `{"key":"val"}` {
		t.Errorf("Data = %q, want %q", e.Data, `{"key":"val"}`)
	}
	// Audit fields should be zero
	if e.EntityType != "" {
		t.Errorf("EntityType should be empty for ops events, got %q", e.EntityType)
	}
	if e.Action != "" {
		t.Errorf("Action should be empty for ops events, got %q", e.Action)
	}
}

// Suppress unused import warnings — errors is used by mockWorkshopEventRepository in log_service_test.go
var _ = errors.New
