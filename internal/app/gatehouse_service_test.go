package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockGatehouseRepository implements secondary.GatehouseRepository for testing.
type mockGatehouseRepository struct {
	gatehouses           map[string]*secondary.GatehouseRecord
	gatehousesByWorkshop map[string]*secondary.GatehouseRecord
	workshopExists       map[string]bool
	workshopHasGatehouse map[string]bool
	nextID               int
}

func newMockGatehouseRepository() *mockGatehouseRepository {
	return &mockGatehouseRepository{
		gatehouses:           make(map[string]*secondary.GatehouseRecord),
		gatehousesByWorkshop: make(map[string]*secondary.GatehouseRecord),
		workshopExists:       make(map[string]bool),
		workshopHasGatehouse: make(map[string]bool),
		nextID:               1,
	}
}

func (m *mockGatehouseRepository) Create(ctx context.Context, gatehouse *secondary.GatehouseRecord) error {
	m.gatehouses[gatehouse.ID] = gatehouse
	m.gatehousesByWorkshop[gatehouse.WorkshopID] = gatehouse
	m.workshopHasGatehouse[gatehouse.WorkshopID] = true
	return nil
}

func (m *mockGatehouseRepository) GetByID(ctx context.Context, id string) (*secondary.GatehouseRecord, error) {
	if g, ok := m.gatehouses[id]; ok {
		return g, nil
	}
	return nil, errors.New("not found")
}

func (m *mockGatehouseRepository) GetByWorkshop(ctx context.Context, workshopID string) (*secondary.GatehouseRecord, error) {
	if g, ok := m.gatehousesByWorkshop[workshopID]; ok {
		return g, nil
	}
	return nil, errors.New("not found")
}

func (m *mockGatehouseRepository) List(ctx context.Context, filters secondary.GatehouseFilters) ([]*secondary.GatehouseRecord, error) {
	var result []*secondary.GatehouseRecord
	for _, g := range m.gatehouses {
		if filters.WorkshopID != "" && g.WorkshopID != filters.WorkshopID {
			continue
		}
		if filters.Status != "" && g.Status != filters.Status {
			continue
		}
		result = append(result, g)
	}
	return result, nil
}

func (m *mockGatehouseRepository) Update(ctx context.Context, gatehouse *secondary.GatehouseRecord) error {
	if _, ok := m.gatehouses[gatehouse.ID]; !ok {
		return errors.New("not found")
	}
	return nil
}

func (m *mockGatehouseRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.gatehouses[id]; !ok {
		return errors.New("not found")
	}
	g := m.gatehouses[id]
	delete(m.gatehousesByWorkshop, g.WorkshopID)
	delete(m.gatehouses, id)
	m.workshopHasGatehouse[g.WorkshopID] = false
	return nil
}

func (m *mockGatehouseRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("GATE-%03d", id), nil
}

func (m *mockGatehouseRepository) UpdateStatus(ctx context.Context, id, status string) error {
	if g, ok := m.gatehouses[id]; ok {
		g.Status = status
		return nil
	}
	return errors.New("not found")
}

func (m *mockGatehouseRepository) WorkshopExists(ctx context.Context, workshopID string) (bool, error) {
	return m.workshopExists[workshopID], nil
}

func (m *mockGatehouseRepository) WorkshopHasGatehouse(ctx context.Context, workshopID string) (bool, error) {
	return m.workshopHasGatehouse[workshopID], nil
}

func (m *mockGatehouseRepository) UpdateFocusedID(ctx context.Context, id, focusedID string) error {
	if g, ok := m.gatehouses[id]; ok {
		g.FocusedID = focusedID
		return nil
	}
	return fmt.Errorf("gatehouse %s not found", id)
}

func newTestGatehouseService() (*GatehouseServiceImpl, *mockGatehouseRepository) {
	repo := newMockGatehouseRepository()
	service := NewGatehouseService(repo)
	return service, repo
}

func TestGatehouseService_GetGatehouse(t *testing.T) {
	service, repo := newTestGatehouseService()
	ctx := context.Background()

	repo.gatehouses["GATE-001"] = &secondary.GatehouseRecord{
		ID:         "GATE-001",
		WorkshopID: "WORK-001",
		Status:     "active",
	}

	gatehouse, err := service.GetGatehouse(ctx, "GATE-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if gatehouse.WorkshopID != "WORK-001" {
		t.Errorf("expected workshopID 'WORK-001', got %q", gatehouse.WorkshopID)
	}
}

func TestGatehouseService_GetGatehouse_NotFound(t *testing.T) {
	service, _ := newTestGatehouseService()
	ctx := context.Background()

	_, err := service.GetGatehouse(ctx, "GATE-999")
	if err == nil {
		t.Error("expected error for non-existent gatehouse")
	}
}

func TestGatehouseService_GetGatehouseByWorkshop(t *testing.T) {
	service, repo := newTestGatehouseService()
	ctx := context.Background()

	repo.gatehouses["GATE-001"] = &secondary.GatehouseRecord{
		ID:         "GATE-001",
		WorkshopID: "WORK-001",
		Status:     "active",
	}
	repo.gatehousesByWorkshop["WORK-001"] = repo.gatehouses["GATE-001"]

	gatehouse, err := service.GetGatehouseByWorkshop(ctx, "WORK-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if gatehouse.ID != "GATE-001" {
		t.Errorf("expected ID 'GATE-001', got %q", gatehouse.ID)
	}
}

func TestGatehouseService_ListGatehouses(t *testing.T) {
	service, repo := newTestGatehouseService()
	ctx := context.Background()

	repo.gatehouses["GATE-001"] = &secondary.GatehouseRecord{ID: "GATE-001", WorkshopID: "WORK-001", Status: "active"}
	repo.gatehouses["GATE-002"] = &secondary.GatehouseRecord{ID: "GATE-002", WorkshopID: "WORK-002", Status: "active"}

	gatehouses, err := service.ListGatehouses(ctx, primary.GatehouseFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(gatehouses) != 2 {
		t.Errorf("expected 2 gatehouses, got %d", len(gatehouses))
	}
}
