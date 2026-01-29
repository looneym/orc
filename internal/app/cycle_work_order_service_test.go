package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockCycleWorkOrderRepository implements secondary.CycleWorkOrderRepository for testing.
type mockCycleWorkOrderRepository struct {
	cwos            map[string]*secondary.CycleWorkOrderRecord
	cwosByCycle     map[string]*secondary.CycleWorkOrderRecord
	cycleExists     map[string]bool
	cycleHasCWO     map[string]bool
	cycleShipmentID map[string]string
	cycleStatus     map[string]string
	nextID          int
	createErr       error
	getErr          error
	updateErr       error
	deleteErr       error
}

func newMockCycleWorkOrderRepository() *mockCycleWorkOrderRepository {
	return &mockCycleWorkOrderRepository{
		cwos:            make(map[string]*secondary.CycleWorkOrderRecord),
		cwosByCycle:     make(map[string]*secondary.CycleWorkOrderRecord),
		cycleExists:     make(map[string]bool),
		cycleHasCWO:     make(map[string]bool),
		cycleShipmentID: make(map[string]string),
		cycleStatus:     make(map[string]string),
		nextID:          1,
	}
}

func (m *mockCycleWorkOrderRepository) Create(ctx context.Context, cwo *secondary.CycleWorkOrderRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.cwos[cwo.ID] = cwo
	m.cwosByCycle[cwo.CycleID] = cwo
	m.cycleHasCWO[cwo.CycleID] = true
	return nil
}

func (m *mockCycleWorkOrderRepository) GetByID(ctx context.Context, id string) (*secondary.CycleWorkOrderRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if cwo, ok := m.cwos[id]; ok {
		return cwo, nil
	}
	return nil, errors.New("not found")
}

func (m *mockCycleWorkOrderRepository) GetByCycle(ctx context.Context, cycleID string) (*secondary.CycleWorkOrderRecord, error) {
	if cwo, ok := m.cwosByCycle[cycleID]; ok {
		return cwo, nil
	}
	return nil, errors.New("not found")
}

func (m *mockCycleWorkOrderRepository) List(ctx context.Context, filters secondary.CycleWorkOrderFilters) ([]*secondary.CycleWorkOrderRecord, error) {
	var result []*secondary.CycleWorkOrderRecord
	for _, cwo := range m.cwos {
		if filters.CycleID != "" && cwo.CycleID != filters.CycleID {
			continue
		}
		if filters.ShipmentID != "" && cwo.ShipmentID != filters.ShipmentID {
			continue
		}
		if filters.Status != "" && cwo.Status != filters.Status {
			continue
		}
		result = append(result, cwo)
	}
	return result, nil
}

func (m *mockCycleWorkOrderRepository) Update(ctx context.Context, cwo *secondary.CycleWorkOrderRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.cwos[cwo.ID]; !ok {
		return errors.New("not found")
	}
	existing := m.cwos[cwo.ID]
	if cwo.Outcome != "" {
		existing.Outcome = cwo.Outcome
	}
	if cwo.AcceptanceCriteria != "" {
		existing.AcceptanceCriteria = cwo.AcceptanceCriteria
	}
	return nil
}

func (m *mockCycleWorkOrderRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.cwos[id]; !ok {
		return errors.New("not found")
	}
	cwo := m.cwos[id]
	delete(m.cwosByCycle, cwo.CycleID)
	delete(m.cwos, id)
	m.cycleHasCWO[cwo.CycleID] = false
	return nil
}

func (m *mockCycleWorkOrderRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("CWO-%03d", id), nil
}

func (m *mockCycleWorkOrderRepository) UpdateStatus(ctx context.Context, id, status string) error {
	if cwo, ok := m.cwos[id]; ok {
		cwo.Status = status
		return nil
	}
	return errors.New("not found")
}

func (m *mockCycleWorkOrderRepository) CycleExists(ctx context.Context, cycleID string) (bool, error) {
	return m.cycleExists[cycleID], nil
}

func (m *mockCycleWorkOrderRepository) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	return true, nil
}

func (m *mockCycleWorkOrderRepository) CycleHasCWO(ctx context.Context, cycleID string) (bool, error) {
	return m.cycleHasCWO[cycleID], nil
}

func (m *mockCycleWorkOrderRepository) GetCycleShipmentID(ctx context.Context, cycleID string) (string, error) {
	if sid, ok := m.cycleShipmentID[cycleID]; ok {
		return sid, nil
	}
	return "", errors.New("cycle not found")
}

func (m *mockCycleWorkOrderRepository) GetCycleStatus(ctx context.Context, cycleID string) (string, error) {
	if status, ok := m.cycleStatus[cycleID]; ok {
		return status, nil
	}
	return "", errors.New("cycle not found")
}

// mockCycleServiceForCWO implements primary.CycleService for testing cascade updates.
type mockCycleServiceForCWO struct {
	updateStatusCalls []struct {
		CycleID string
		Status  string
	}
	updateStatusErr error
}

func newMockCycleServiceForCWO() *mockCycleServiceForCWO {
	return &mockCycleServiceForCWO{}
}

func (m *mockCycleServiceForCWO) CreateCycle(ctx context.Context, req primary.CreateCycleRequest) (*primary.CreateCycleResponse, error) {
	return nil, nil
}

func (m *mockCycleServiceForCWO) GetCycle(ctx context.Context, cycleID string) (*primary.Cycle, error) {
	return nil, nil
}

func (m *mockCycleServiceForCWO) ListCycles(ctx context.Context, filters primary.CycleFilters) ([]*primary.Cycle, error) {
	return nil, nil
}

func (m *mockCycleServiceForCWO) DeleteCycle(ctx context.Context, cycleID string) error {
	return nil
}

func (m *mockCycleServiceForCWO) StartCycle(ctx context.Context, cycleID string) error {
	return nil
}

func (m *mockCycleServiceForCWO) CompleteCycle(ctx context.Context, cycleID string) error {
	return nil
}

func (m *mockCycleServiceForCWO) GetActiveCycle(ctx context.Context, shipmentID string) (*primary.Cycle, error) {
	return nil, nil
}

func (m *mockCycleServiceForCWO) UpdateCycleStatus(ctx context.Context, cycleID string, status string) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	m.updateStatusCalls = append(m.updateStatusCalls, struct {
		CycleID string
		Status  string
	}{cycleID, status})
	return nil
}

func newTestCycleWorkOrderService() (*CycleWorkOrderServiceImpl, *mockCycleWorkOrderRepository, *mockCycleServiceForCWO) {
	repo := newMockCycleWorkOrderRepository()
	cycleService := newMockCycleServiceForCWO()
	service := NewCycleWorkOrderService(repo, cycleService)
	return service, repo, cycleService
}

func TestCycleWorkOrderService_CreateCycleWorkOrder(t *testing.T) {
	service, repo, _ := newTestCycleWorkOrderService()
	ctx := context.Background()

	repo.cycleExists["CYC-001"] = true
	repo.cycleShipmentID["CYC-001"] = "SHIP-001"

	resp, err := service.CreateCycleWorkOrder(ctx, primary.CreateCycleWorkOrderRequest{
		CycleID:            "CYC-001",
		Outcome:            "Test outcome",
		AcceptanceCriteria: "Test criteria",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.CycleWorkOrderID == "" {
		t.Error("expected CWO ID to be set")
	}
	if resp.CycleWorkOrder.Status != "draft" {
		t.Errorf("expected status 'draft', got %q", resp.CycleWorkOrder.Status)
	}
}

func TestCycleWorkOrderService_CreateCycleWorkOrder_CycleNotFound(t *testing.T) {
	service, _, _ := newTestCycleWorkOrderService()
	ctx := context.Background()

	_, err := service.CreateCycleWorkOrder(ctx, primary.CreateCycleWorkOrderRequest{
		CycleID: "CYC-999",
		Outcome: "Test",
	})

	if err == nil {
		t.Error("expected error for non-existent cycle")
	}
}

func TestCycleWorkOrderService_CreateCycleWorkOrder_AlreadyExists(t *testing.T) {
	service, repo, _ := newTestCycleWorkOrderService()
	ctx := context.Background()

	repo.cycleExists["CYC-001"] = true
	repo.cycleHasCWO["CYC-001"] = true

	_, err := service.CreateCycleWorkOrder(ctx, primary.CreateCycleWorkOrderRequest{
		CycleID: "CYC-001",
		Outcome: "Test",
	})

	if err == nil {
		t.Error("expected error for cycle that already has CWO")
	}
}

func TestCycleWorkOrderService_GetCycleWorkOrder(t *testing.T) {
	service, repo, _ := newTestCycleWorkOrderService()
	ctx := context.Background()

	repo.cwos["CWO-001"] = &secondary.CycleWorkOrderRecord{
		ID:         "CWO-001",
		CycleID:    "CYC-001",
		ShipmentID: "SHIP-001",
		Outcome:    "Test outcome",
		Status:     "draft",
	}

	cwo, err := service.GetCycleWorkOrder(ctx, "CWO-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cwo.Outcome != "Test outcome" {
		t.Errorf("expected outcome 'Test outcome', got %q", cwo.Outcome)
	}
}

func TestCycleWorkOrderService_GetCycleWorkOrder_NotFound(t *testing.T) {
	service, _, _ := newTestCycleWorkOrderService()
	ctx := context.Background()

	_, err := service.GetCycleWorkOrder(ctx, "CWO-999")
	if err == nil {
		t.Error("expected error for non-existent CWO")
	}
}

func TestCycleWorkOrderService_GetCycleWorkOrderByCycle(t *testing.T) {
	service, repo, _ := newTestCycleWorkOrderService()
	ctx := context.Background()

	repo.cwos["CWO-001"] = &secondary.CycleWorkOrderRecord{
		ID:         "CWO-001",
		CycleID:    "CYC-001",
		ShipmentID: "SHIP-001",
		Status:     "draft",
	}
	repo.cwosByCycle["CYC-001"] = repo.cwos["CWO-001"]

	cwo, err := service.GetCycleWorkOrderByCycle(ctx, "CYC-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cwo.ID != "CWO-001" {
		t.Errorf("expected ID 'CWO-001', got %q", cwo.ID)
	}
}

func TestCycleWorkOrderService_GetCycleWorkOrderByCycle_NotFound(t *testing.T) {
	service, _, _ := newTestCycleWorkOrderService()
	ctx := context.Background()

	_, err := service.GetCycleWorkOrderByCycle(ctx, "CYC-999")
	if err == nil {
		t.Error("expected error for non-existent cycle")
	}
}

func TestCycleWorkOrderService_ListCycleWorkOrders(t *testing.T) {
	service, repo, _ := newTestCycleWorkOrderService()
	ctx := context.Background()

	repo.cwos["CWO-001"] = &secondary.CycleWorkOrderRecord{ID: "CWO-001", CycleID: "CYC-001", Status: "draft"}
	repo.cwos["CWO-002"] = &secondary.CycleWorkOrderRecord{ID: "CWO-002", CycleID: "CYC-002", Status: "active"}

	cwos, err := service.ListCycleWorkOrders(ctx, primary.CycleWorkOrderFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cwos) != 2 {
		t.Errorf("expected 2 CWOs, got %d", len(cwos))
	}
}

func TestCycleWorkOrderService_UpdateCycleWorkOrder(t *testing.T) {
	service, repo, _ := newTestCycleWorkOrderService()
	ctx := context.Background()

	repo.cwos["CWO-001"] = &secondary.CycleWorkOrderRecord{
		ID:      "CWO-001",
		CycleID: "CYC-001",
		Outcome: "Original",
		Status:  "draft",
	}

	err := service.UpdateCycleWorkOrder(ctx, primary.UpdateCycleWorkOrderRequest{
		CycleWorkOrderID: "CWO-001",
		Outcome:          "Updated",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cwo, _ := service.GetCycleWorkOrder(ctx, "CWO-001")
	if cwo.Outcome != "Updated" {
		t.Errorf("expected outcome 'Updated', got %q", cwo.Outcome)
	}
}

func TestCycleWorkOrderService_UpdateCycleWorkOrder_NotDraft(t *testing.T) {
	service, repo, _ := newTestCycleWorkOrderService()
	ctx := context.Background()

	repo.cwos["CWO-001"] = &secondary.CycleWorkOrderRecord{
		ID:      "CWO-001",
		CycleID: "CYC-001",
		Status:  "active",
	}

	err := service.UpdateCycleWorkOrder(ctx, primary.UpdateCycleWorkOrderRequest{
		CycleWorkOrderID: "CWO-001",
		Outcome:          "Updated",
	})
	if err == nil {
		t.Error("expected error for non-draft CWO")
	}
}

func TestCycleWorkOrderService_UpdateCycleWorkOrder_NotFound(t *testing.T) {
	service, _, _ := newTestCycleWorkOrderService()
	ctx := context.Background()

	err := service.UpdateCycleWorkOrder(ctx, primary.UpdateCycleWorkOrderRequest{
		CycleWorkOrderID: "CWO-999",
		Outcome:          "Updated",
	})
	if err == nil {
		t.Error("expected error for non-existent CWO")
	}
}

func TestCycleWorkOrderService_DeleteCycleWorkOrder(t *testing.T) {
	service, repo, _ := newTestCycleWorkOrderService()
	ctx := context.Background()

	repo.cwos["CWO-001"] = &secondary.CycleWorkOrderRecord{ID: "CWO-001", CycleID: "CYC-001", Status: "draft"}
	repo.cwosByCycle["CYC-001"] = repo.cwos["CWO-001"]

	err := service.DeleteCycleWorkOrder(ctx, "CWO-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = service.GetCycleWorkOrder(ctx, "CWO-001")
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestCycleWorkOrderService_DeleteCycleWorkOrder_NotFound(t *testing.T) {
	service, _, _ := newTestCycleWorkOrderService()
	ctx := context.Background()

	err := service.DeleteCycleWorkOrder(ctx, "CWO-999")
	if err == nil {
		t.Error("expected error for non-existent CWO")
	}
}

func TestCycleWorkOrderService_ApproveCycleWorkOrder(t *testing.T) {
	service, repo, cycleService := newTestCycleWorkOrderService()
	ctx := context.Background()

	repo.cwos["CWO-001"] = &secondary.CycleWorkOrderRecord{
		ID:      "CWO-001",
		CycleID: "CYC-001",
		Outcome: "Test outcome",
		Status:  "draft",
	}
	repo.cycleExists["CYC-001"] = true

	err := service.ApproveCycleWorkOrder(ctx, "CWO-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cwo, _ := service.GetCycleWorkOrder(ctx, "CWO-001")
	if cwo.Status != "active" {
		t.Errorf("expected status 'active', got %q", cwo.Status)
	}

	// Verify cascade to cycle service
	if len(cycleService.updateStatusCalls) != 1 {
		t.Errorf("expected 1 cascade call, got %d", len(cycleService.updateStatusCalls))
	}
	if cycleService.updateStatusCalls[0].Status != "approved" {
		t.Errorf("expected cascade status 'approved', got %q", cycleService.updateStatusCalls[0].Status)
	}
}

func TestCycleWorkOrderService_ApproveCycleWorkOrder_NotDraft(t *testing.T) {
	service, repo, _ := newTestCycleWorkOrderService()
	ctx := context.Background()

	repo.cwos["CWO-001"] = &secondary.CycleWorkOrderRecord{
		ID:      "CWO-001",
		CycleID: "CYC-001",
		Status:  "active",
	}
	repo.cycleExists["CYC-001"] = true

	err := service.ApproveCycleWorkOrder(ctx, "CWO-001")
	if err == nil {
		t.Error("expected error for non-draft CWO")
	}
}

func TestCycleWorkOrderService_CompleteCycleWorkOrder(t *testing.T) {
	service, repo, _ := newTestCycleWorkOrderService()
	ctx := context.Background()

	repo.cwos["CWO-001"] = &secondary.CycleWorkOrderRecord{
		ID:      "CWO-001",
		CycleID: "CYC-001",
		Outcome: "Test outcome",
		Status:  "active",
	}
	repo.cycleExists["CYC-001"] = true
	repo.cycleStatus["CYC-001"] = "implementing"

	err := service.CompleteCycleWorkOrder(ctx, "CWO-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cwo, _ := service.GetCycleWorkOrder(ctx, "CWO-001")
	if cwo.Status != "complete" {
		t.Errorf("expected status 'complete', got %q", cwo.Status)
	}
}

func TestCycleWorkOrderService_CompleteCycleWorkOrder_NotActive(t *testing.T) {
	service, repo, _ := newTestCycleWorkOrderService()
	ctx := context.Background()

	repo.cwos["CWO-001"] = &secondary.CycleWorkOrderRecord{
		ID:      "CWO-001",
		CycleID: "CYC-001",
		Status:  "draft",
	}
	repo.cycleExists["CYC-001"] = true

	err := service.CompleteCycleWorkOrder(ctx, "CWO-001")
	if err == nil {
		t.Error("expected error for non-active CWO")
	}
}
