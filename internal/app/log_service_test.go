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

// mockWorkshopEventRepository implements secondary.WorkshopEventRepository for testing.
type mockWorkshopEventRepository struct {
	events         map[string]*secondary.AuditEventRecord
	workshopExists map[string]bool
	nextID         int
}

func newMockWorkshopEventRepository() *mockWorkshopEventRepository {
	return &mockWorkshopEventRepository{
		events:         make(map[string]*secondary.AuditEventRecord),
		workshopExists: make(map[string]bool),
		nextID:         1,
	}
}

func (m *mockWorkshopEventRepository) Create(ctx context.Context, event *secondary.AuditEventRecord) error {
	m.events[event.ID] = event
	return nil
}

func (m *mockWorkshopEventRepository) GetByID(ctx context.Context, id string) (*secondary.AuditEventRecord, error) {
	if e, ok := m.events[id]; ok {
		return e, nil
	}
	return nil, errors.New("not found")
}

func (m *mockWorkshopEventRepository) List(ctx context.Context, filters secondary.AuditEventFilters) ([]*secondary.AuditEventRecord, error) {
	var result []*secondary.AuditEventRecord
	for _, e := range m.events {
		if filters.WorkshopID != "" && e.WorkshopID != filters.WorkshopID {
			continue
		}
		if filters.EntityType != "" && e.EntityType != filters.EntityType {
			continue
		}
		if filters.EntityID != "" && e.EntityID != filters.EntityID {
			continue
		}
		if filters.ActorID != "" && e.ActorID != filters.ActorID {
			continue
		}
		if filters.Action != "" && e.Action != filters.Action {
			continue
		}
		if filters.Source != "" && e.Source != filters.Source {
			continue
		}
		result = append(result, e)
	}

	// Apply limit
	if filters.Limit > 0 && len(result) > filters.Limit {
		result = result[:filters.Limit]
	}

	return result, nil
}

func (m *mockWorkshopEventRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("WE-%04d", id), nil
}

func (m *mockWorkshopEventRepository) WorkshopExists(ctx context.Context, workshopID string) (bool, error) {
	return m.workshopExists[workshopID], nil
}

func (m *mockWorkshopEventRepository) PruneOlderThan(ctx context.Context, days int) (int, error) {
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

func newTestLogService() (*LogServiceImpl, *mockWorkshopEventRepository) {
	repo := newMockWorkshopEventRepository()
	service := NewLogService(repo)
	return service, repo
}

func TestLogService_GetLog(t *testing.T) {
	service, repo := newTestLogService()
	ctx := context.Background()

	repo.events["WE-0001"] = &secondary.AuditEventRecord{
		ID:         "WE-0001",
		WorkshopID: "SHOP-001",
		Timestamp:  "2024-01-01T12:00:00Z",
		ActorID:    "IMP-BENCH-001",
		EntityType: "task",
		EntityID:   "TASK-001",
		Action:     "create",
	}

	log, err := service.GetLog(ctx, "WE-0001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if log.EntityID != "TASK-001" {
		t.Errorf("expected entityID 'TASK-001', got %q", log.EntityID)
	}
}

func TestLogService_GetLog_NotFound(t *testing.T) {
	service, _ := newTestLogService()
	ctx := context.Background()

	_, err := service.GetLog(ctx, "WE-9999")
	if err == nil {
		t.Error("expected error for non-existent log")
	}
}

func TestLogService_ListLogs(t *testing.T) {
	service, repo := newTestLogService()
	ctx := context.Background()

	repo.events["WE-0001"] = &secondary.AuditEventRecord{ID: "WE-0001", WorkshopID: "SHOP-001", EntityType: "task", EntityID: "TASK-001", Action: "create", Timestamp: "2024-01-01T12:00:00Z"}
	repo.events["WE-0002"] = &secondary.AuditEventRecord{ID: "WE-0002", WorkshopID: "SHOP-001", EntityType: "task", EntityID: "TASK-002", Action: "create", Timestamp: "2024-01-01T12:01:00Z"}

	logs, err := service.ListLogs(ctx, primary.LogFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("expected 2 logs, got %d", len(logs))
	}
}

func TestLogService_ListLogs_WithFilters(t *testing.T) {
	service, repo := newTestLogService()
	ctx := context.Background()

	repo.events["WE-0001"] = &secondary.AuditEventRecord{ID: "WE-0001", WorkshopID: "SHOP-001", EntityType: "task", EntityID: "TASK-001", Action: "create", ActorID: "IMP-BENCH-001", Timestamp: "2024-01-01T12:00:00Z"}
	repo.events["WE-0002"] = &secondary.AuditEventRecord{ID: "WE-0002", WorkshopID: "SHOP-001", EntityType: "task", EntityID: "TASK-002", Action: "create", ActorID: "IMP-BENCH-002", Timestamp: "2024-01-01T12:01:00Z"}

	logs, err := service.ListLogs(ctx, primary.LogFilters{ActorID: "IMP-BENCH-001"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}
	if logs[0].ActorID != "IMP-BENCH-001" {
		t.Errorf("expected actorID 'IMP-BENCH-001', got %q", logs[0].ActorID)
	}
}

func TestLogService_PruneLogs(t *testing.T) {
	service, repo := newTestLogService()
	ctx := context.Background()

	// Add old and new events
	oldTime := time.Now().AddDate(0, 0, -60).Format(time.RFC3339) // 60 days old
	newTime := time.Now().Format(time.RFC3339)

	repo.events["WE-0001"] = &secondary.AuditEventRecord{ID: "WE-0001", WorkshopID: "SHOP-001", EntityType: "task", Timestamp: oldTime}
	repo.events["WE-0002"] = &secondary.AuditEventRecord{ID: "WE-0002", WorkshopID: "SHOP-001", EntityType: "task", Timestamp: newTime}

	count, err := service.PruneLogs(ctx, 30)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 pruned, got %d", count)
	}
	if len(repo.events) != 1 {
		t.Errorf("expected 1 event remaining, got %d", len(repo.events))
	}
}
