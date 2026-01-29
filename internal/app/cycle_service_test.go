package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockCycleRepository implements secondary.CycleRepository for testing.
type mockCycleRepository struct {
	cycles         map[string]*secondary.CycleRecord
	shipmentExists map[string]bool
	activeCycles   map[string]*secondary.CycleRecord
	nextSeq        map[string]int64
	nextID         int
	createErr      error
	getErr         error
	deleteErr      error
	updateErr      error
}

func newMockCycleRepository() *mockCycleRepository {
	return &mockCycleRepository{
		cycles:         make(map[string]*secondary.CycleRecord),
		shipmentExists: make(map[string]bool),
		activeCycles:   make(map[string]*secondary.CycleRecord),
		nextSeq:        make(map[string]int64),
		nextID:         1,
	}
}

func (m *mockCycleRepository) Create(ctx context.Context, cycle *secondary.CycleRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.cycles[cycle.ID] = cycle
	return nil
}

func (m *mockCycleRepository) GetByID(ctx context.Context, id string) (*secondary.CycleRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if c, ok := m.cycles[id]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}

func (m *mockCycleRepository) List(ctx context.Context, filters secondary.CycleFilters) ([]*secondary.CycleRecord, error) {
	var result []*secondary.CycleRecord
	for _, c := range m.cycles {
		if filters.ShipmentID != "" && c.ShipmentID != filters.ShipmentID {
			continue
		}
		if filters.Status != "" && c.Status != filters.Status {
			continue
		}
		result = append(result, c)
	}
	return result, nil
}

func (m *mockCycleRepository) Update(ctx context.Context, cycle *secondary.CycleRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.cycles[cycle.ID]; !ok {
		return errors.New("not found")
	}
	return nil
}

func (m *mockCycleRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.cycles[id]; !ok {
		return errors.New("not found")
	}
	delete(m.cycles, id)
	return nil
}

func (m *mockCycleRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("CYC-%03d", id), nil
}

func (m *mockCycleRepository) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	return m.shipmentExists[shipmentID], nil
}

func (m *mockCycleRepository) GetNextSequenceNumber(ctx context.Context, shipmentID string) (int64, error) {
	m.nextSeq[shipmentID]++
	return m.nextSeq[shipmentID], nil
}

func (m *mockCycleRepository) GetActiveCycle(ctx context.Context, shipmentID string) (*secondary.CycleRecord, error) {
	return m.activeCycles[shipmentID], nil
}

func (m *mockCycleRepository) GetByShipmentAndSequence(ctx context.Context, shipmentID string, seq int64) (*secondary.CycleRecord, error) {
	for _, c := range m.cycles {
		if c.ShipmentID == shipmentID && c.SequenceNumber == seq {
			return c, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockCycleRepository) UpdateStatus(ctx context.Context, id, status string, setStarted, setCompleted bool) error {
	if c, ok := m.cycles[id]; ok {
		c.Status = status
		if status == "active" {
			m.activeCycles[c.ShipmentID] = c
		}
		if status == "complete" || status == "failed" {
			delete(m.activeCycles, c.ShipmentID)
		}
		return nil
	}
	return errors.New("not found")
}

func newTestCycleService() (*CycleServiceImpl, *mockCycleRepository) {
	repo := newMockCycleRepository()
	service := NewCycleService(repo)
	return service, repo
}

func TestCycleService_CreateCycle(t *testing.T) {
	service, repo := newTestCycleService()
	ctx := context.Background()

	repo.shipmentExists["SHIP-001"] = true

	resp, err := service.CreateCycle(ctx, primary.CreateCycleRequest{
		ShipmentID: "SHIP-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.CycleID == "" {
		t.Error("expected cycle ID to be set")
	}
	if resp.Cycle.Status != "draft" {
		t.Errorf("expected status 'draft', got %q", resp.Cycle.Status)
	}
	if resp.Cycle.SequenceNumber != 1 {
		t.Errorf("expected sequence number 1, got %d", resp.Cycle.SequenceNumber)
	}
}

func TestCycleService_CreateCycle_ShipmentNotFound(t *testing.T) {
	service, _ := newTestCycleService()
	ctx := context.Background()

	_, err := service.CreateCycle(ctx, primary.CreateCycleRequest{
		ShipmentID: "SHIP-999",
	})

	if err == nil {
		t.Error("expected error for non-existent shipment")
	}
}

func TestCycleService_GetCycle(t *testing.T) {
	service, repo := newTestCycleService()
	ctx := context.Background()

	repo.cycles["CYC-001"] = &secondary.CycleRecord{
		ID:             "CYC-001",
		ShipmentID:     "SHIP-001",
		SequenceNumber: 1,
		Status:         "draft",
	}

	cycle, err := service.GetCycle(ctx, "CYC-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cycle.ShipmentID != "SHIP-001" {
		t.Errorf("expected shipment 'SHIP-001', got %q", cycle.ShipmentID)
	}
}

func TestCycleService_GetCycle_NotFound(t *testing.T) {
	service, _ := newTestCycleService()
	ctx := context.Background()

	_, err := service.GetCycle(ctx, "CYC-999")
	if err == nil {
		t.Error("expected error for non-existent cycle")
	}
}

func TestCycleService_ListCycles(t *testing.T) {
	service, repo := newTestCycleService()
	ctx := context.Background()

	repo.cycles["CYC-001"] = &secondary.CycleRecord{ID: "CYC-001", ShipmentID: "SHIP-001", Status: "draft"}
	repo.cycles["CYC-002"] = &secondary.CycleRecord{ID: "CYC-002", ShipmentID: "SHIP-001", Status: "active"}

	cycles, err := service.ListCycles(ctx, primary.CycleFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cycles) != 2 {
		t.Errorf("expected 2 cycles, got %d", len(cycles))
	}
}

func TestCycleService_ListCycles_FilterByShipment(t *testing.T) {
	service, repo := newTestCycleService()
	ctx := context.Background()

	repo.cycles["CYC-001"] = &secondary.CycleRecord{ID: "CYC-001", ShipmentID: "SHIP-001", Status: "draft"}
	repo.cycles["CYC-002"] = &secondary.CycleRecord{ID: "CYC-002", ShipmentID: "SHIP-002", Status: "active"}

	cycles, err := service.ListCycles(ctx, primary.CycleFilters{ShipmentID: "SHIP-001"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cycles) != 1 {
		t.Errorf("expected 1 cycle for SHIP-001, got %d", len(cycles))
	}
}

func TestCycleService_DeleteCycle(t *testing.T) {
	service, repo := newTestCycleService()
	ctx := context.Background()

	repo.cycles["CYC-001"] = &secondary.CycleRecord{ID: "CYC-001", ShipmentID: "SHIP-001", Status: "draft"}

	err := service.DeleteCycle(ctx, "CYC-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = service.GetCycle(ctx, "CYC-001")
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestCycleService_DeleteCycle_NotFound(t *testing.T) {
	service, _ := newTestCycleService()
	ctx := context.Background()

	err := service.DeleteCycle(ctx, "CYC-999")
	if err == nil {
		t.Error("expected error for non-existent cycle")
	}
}

func TestCycleService_StartCycle(t *testing.T) {
	service, repo := newTestCycleService()
	ctx := context.Background()

	repo.cycles["CYC-001"] = &secondary.CycleRecord{ID: "CYC-001", ShipmentID: "SHIP-001", Status: "queued"}

	err := service.StartCycle(ctx, "CYC-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cycle, _ := service.GetCycle(ctx, "CYC-001")
	if cycle.Status != "active" {
		t.Errorf("expected status 'active', got %q", cycle.Status)
	}
}

func TestCycleService_StartCycle_NotQueued(t *testing.T) {
	service, repo := newTestCycleService()
	ctx := context.Background()

	repo.cycles["CYC-001"] = &secondary.CycleRecord{ID: "CYC-001", ShipmentID: "SHIP-001", Status: "draft"}

	err := service.StartCycle(ctx, "CYC-001")
	if err == nil {
		t.Error("expected error for non-queued cycle")
	}
}

func TestCycleService_StartCycle_AlreadyActive(t *testing.T) {
	service, repo := newTestCycleService()
	ctx := context.Background()

	repo.cycles["CYC-001"] = &secondary.CycleRecord{ID: "CYC-001", ShipmentID: "SHIP-001", Status: "queued"}
	repo.cycles["CYC-002"] = &secondary.CycleRecord{ID: "CYC-002", ShipmentID: "SHIP-001", Status: "active"}
	repo.activeCycles["SHIP-001"] = repo.cycles["CYC-002"]

	err := service.StartCycle(ctx, "CYC-001")
	if err == nil {
		t.Error("expected error when shipment already has active cycle")
	}
}

func TestCycleService_CompleteCycle(t *testing.T) {
	service, repo := newTestCycleService()
	ctx := context.Background()

	repo.cycles["CYC-001"] = &secondary.CycleRecord{ID: "CYC-001", ShipmentID: "SHIP-001", Status: "active"}

	err := service.CompleteCycle(ctx, "CYC-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cycle, _ := service.GetCycle(ctx, "CYC-001")
	if cycle.Status != "complete" {
		t.Errorf("expected status 'complete', got %q", cycle.Status)
	}
}

func TestCycleService_CompleteCycle_NotActive(t *testing.T) {
	service, repo := newTestCycleService()
	ctx := context.Background()

	repo.cycles["CYC-001"] = &secondary.CycleRecord{ID: "CYC-001", ShipmentID: "SHIP-001", Status: "draft"}

	err := service.CompleteCycle(ctx, "CYC-001")
	if err == nil {
		t.Error("expected error for non-active cycle")
	}
}

func TestCycleService_GetActiveCycle(t *testing.T) {
	service, repo := newTestCycleService()
	ctx := context.Background()

	repo.cycles["CYC-001"] = &secondary.CycleRecord{ID: "CYC-001", ShipmentID: "SHIP-001", Status: "active"}
	repo.activeCycles["SHIP-001"] = repo.cycles["CYC-001"]

	cycle, err := service.GetActiveCycle(ctx, "SHIP-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cycle.ID != "CYC-001" {
		t.Errorf("expected cycle 'CYC-001', got %q", cycle.ID)
	}
}

func TestCycleService_GetActiveCycle_NoneActive(t *testing.T) {
	service, _ := newTestCycleService()
	ctx := context.Background()

	cycle, err := service.GetActiveCycle(ctx, "SHIP-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cycle != nil {
		t.Error("expected nil cycle when none active")
	}
}

func TestCycleService_UpdateCycleStatus(t *testing.T) {
	service, repo := newTestCycleService()
	ctx := context.Background()

	repo.cycles["CYC-001"] = &secondary.CycleRecord{ID: "CYC-001", ShipmentID: "SHIP-001", Status: "draft"}

	err := service.UpdateCycleStatus(ctx, "CYC-001", "approved")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cycle, _ := service.GetCycle(ctx, "CYC-001")
	if cycle.Status != "approved" {
		t.Errorf("expected status 'approved', got %q", cycle.Status)
	}
}

func TestCycleService_UpdateCycleStatus_InvalidStatus(t *testing.T) {
	service, repo := newTestCycleService()
	ctx := context.Background()

	repo.cycles["CYC-001"] = &secondary.CycleRecord{ID: "CYC-001", ShipmentID: "SHIP-001", Status: "draft"}

	err := service.UpdateCycleStatus(ctx, "CYC-001", "invalid")
	if err == nil {
		t.Error("expected error for invalid status")
	}
}
