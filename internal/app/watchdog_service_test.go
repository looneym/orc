package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockWatchdogRepository implements secondary.WatchdogRepository for testing.
type mockWatchdogRepository struct {
	watchdogs            map[string]*secondary.WatchdogRecord
	watchdogsByWorkbench map[string]*secondary.WatchdogRecord
	workbenchExists      map[string]bool
	workbenchHasWatchdog map[string]bool
	nextID               int
}

func newMockWatchdogRepository() *mockWatchdogRepository {
	return &mockWatchdogRepository{
		watchdogs:            make(map[string]*secondary.WatchdogRecord),
		watchdogsByWorkbench: make(map[string]*secondary.WatchdogRecord),
		workbenchExists:      make(map[string]bool),
		workbenchHasWatchdog: make(map[string]bool),
		nextID:               1,
	}
}

func (m *mockWatchdogRepository) Create(ctx context.Context, watchdog *secondary.WatchdogRecord) error {
	m.watchdogs[watchdog.ID] = watchdog
	m.watchdogsByWorkbench[watchdog.WorkbenchID] = watchdog
	m.workbenchHasWatchdog[watchdog.WorkbenchID] = true
	return nil
}

func (m *mockWatchdogRepository) GetByID(ctx context.Context, id string) (*secondary.WatchdogRecord, error) {
	if w, ok := m.watchdogs[id]; ok {
		return w, nil
	}
	return nil, errors.New("not found")
}

func (m *mockWatchdogRepository) GetByWorkbench(ctx context.Context, workbenchID string) (*secondary.WatchdogRecord, error) {
	if w, ok := m.watchdogsByWorkbench[workbenchID]; ok {
		return w, nil
	}
	return nil, errors.New("not found")
}

func (m *mockWatchdogRepository) List(ctx context.Context, filters secondary.WatchdogFilters) ([]*secondary.WatchdogRecord, error) {
	var result []*secondary.WatchdogRecord
	for _, w := range m.watchdogs {
		if filters.WorkbenchID != "" && w.WorkbenchID != filters.WorkbenchID {
			continue
		}
		if filters.Status != "" && w.Status != filters.Status {
			continue
		}
		result = append(result, w)
	}
	return result, nil
}

func (m *mockWatchdogRepository) Update(ctx context.Context, watchdog *secondary.WatchdogRecord) error {
	if _, ok := m.watchdogs[watchdog.ID]; !ok {
		return errors.New("not found")
	}
	return nil
}

func (m *mockWatchdogRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.watchdogs[id]; !ok {
		return errors.New("not found")
	}
	w := m.watchdogs[id]
	delete(m.watchdogsByWorkbench, w.WorkbenchID)
	delete(m.watchdogs, id)
	m.workbenchHasWatchdog[w.WorkbenchID] = false
	return nil
}

func (m *mockWatchdogRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("WATCH-%03d", id), nil
}

func (m *mockWatchdogRepository) UpdateStatus(ctx context.Context, id, status string) error {
	if w, ok := m.watchdogs[id]; ok {
		w.Status = status
		return nil
	}
	return errors.New("not found")
}

func (m *mockWatchdogRepository) WorkbenchExists(ctx context.Context, workbenchID string) (bool, error) {
	return m.workbenchExists[workbenchID], nil
}

func (m *mockWatchdogRepository) WorkbenchHasWatchdog(ctx context.Context, workbenchID string) (bool, error) {
	return m.workbenchHasWatchdog[workbenchID], nil
}

func newTestWatchdogService() (*WatchdogServiceImpl, *mockWatchdogRepository) {
	repo := newMockWatchdogRepository()
	service := NewWatchdogService(repo)
	return service, repo
}

func TestWatchdogService_GetWatchdog(t *testing.T) {
	service, repo := newTestWatchdogService()
	ctx := context.Background()

	repo.watchdogs["WATCH-001"] = &secondary.WatchdogRecord{
		ID:          "WATCH-001",
		WorkbenchID: "BENCH-001",
		Status:      "inactive",
	}

	watchdog, err := service.GetWatchdog(ctx, "WATCH-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if watchdog.WorkbenchID != "BENCH-001" {
		t.Errorf("expected workbenchID 'BENCH-001', got %q", watchdog.WorkbenchID)
	}
}

func TestWatchdogService_GetWatchdog_NotFound(t *testing.T) {
	service, _ := newTestWatchdogService()
	ctx := context.Background()

	_, err := service.GetWatchdog(ctx, "WATCH-999")
	if err == nil {
		t.Error("expected error for non-existent watchdog")
	}
}

func TestWatchdogService_GetWatchdogByWorkbench(t *testing.T) {
	service, repo := newTestWatchdogService()
	ctx := context.Background()

	repo.watchdogs["WATCH-001"] = &secondary.WatchdogRecord{
		ID:          "WATCH-001",
		WorkbenchID: "BENCH-001",
		Status:      "inactive",
	}
	repo.watchdogsByWorkbench["BENCH-001"] = repo.watchdogs["WATCH-001"]

	watchdog, err := service.GetWatchdogByWorkbench(ctx, "BENCH-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if watchdog.ID != "WATCH-001" {
		t.Errorf("expected ID 'WATCH-001', got %q", watchdog.ID)
	}
}

func TestWatchdogService_ListWatchdogs(t *testing.T) {
	service, repo := newTestWatchdogService()
	ctx := context.Background()

	repo.watchdogs["WATCH-001"] = &secondary.WatchdogRecord{ID: "WATCH-001", WorkbenchID: "BENCH-001", Status: "inactive"}
	repo.watchdogs["WATCH-002"] = &secondary.WatchdogRecord{ID: "WATCH-002", WorkbenchID: "BENCH-002", Status: "active"}

	watchdogs, err := service.ListWatchdogs(ctx, primary.WatchdogFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(watchdogs) != 2 {
		t.Errorf("expected 2 watchdogs, got %d", len(watchdogs))
	}
}
