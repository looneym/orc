package app

import (
	"context"
	"errors"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockWorkOrderRepository implements secondary.WorkOrderRepository for testing.
type mockWorkOrderRepository struct {
	workOrders       map[string]*secondary.WorkOrderRecord
	workOrdersByShip map[string]*secondary.WorkOrderRecord
	shipmentExists   map[string]bool
	shipmentHasWO    map[string]bool
	nextID           int
	createErr        error
	getErr           error
	updateErr        error
	deleteErr        error
	updateStatusErr  error
}

func newMockWorkOrderRepository() *mockWorkOrderRepository {
	return &mockWorkOrderRepository{
		workOrders:       make(map[string]*secondary.WorkOrderRecord),
		workOrdersByShip: make(map[string]*secondary.WorkOrderRecord),
		shipmentExists:   make(map[string]bool),
		shipmentHasWO:    make(map[string]bool),
		nextID:           1,
	}
}

func (m *mockWorkOrderRepository) Create(ctx context.Context, wo *secondary.WorkOrderRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.workOrders[wo.ID] = wo
	m.workOrdersByShip[wo.ShipmentID] = wo
	m.shipmentHasWO[wo.ShipmentID] = true
	return nil
}

func (m *mockWorkOrderRepository) GetByID(ctx context.Context, id string) (*secondary.WorkOrderRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if wo, ok := m.workOrders[id]; ok {
		return wo, nil
	}
	return nil, errors.New("not found")
}

func (m *mockWorkOrderRepository) GetByShipment(ctx context.Context, shipmentID string) (*secondary.WorkOrderRecord, error) {
	if wo, ok := m.workOrdersByShip[shipmentID]; ok {
		return wo, nil
	}
	return nil, errors.New("not found")
}

func (m *mockWorkOrderRepository) List(ctx context.Context, filters secondary.WorkOrderFilters) ([]*secondary.WorkOrderRecord, error) {
	var result []*secondary.WorkOrderRecord
	for _, wo := range m.workOrders {
		if filters.ShipmentID != "" && wo.ShipmentID != filters.ShipmentID {
			continue
		}
		if filters.Status != "" && wo.Status != filters.Status {
			continue
		}
		result = append(result, wo)
	}
	return result, nil
}

func (m *mockWorkOrderRepository) Update(ctx context.Context, wo *secondary.WorkOrderRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.workOrders[wo.ID]; !ok {
		return errors.New("not found")
	}
	existing := m.workOrders[wo.ID]
	if wo.Outcome != "" {
		existing.Outcome = wo.Outcome
	}
	if wo.AcceptanceCriteria != "" {
		existing.AcceptanceCriteria = wo.AcceptanceCriteria
	}
	return nil
}

func (m *mockWorkOrderRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.workOrders[id]; !ok {
		return errors.New("not found")
	}
	wo := m.workOrders[id]
	delete(m.workOrdersByShip, wo.ShipmentID)
	delete(m.workOrders, id)
	m.shipmentHasWO[wo.ShipmentID] = false
	return nil
}

func (m *mockWorkOrderRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return "WO-" + string(rune('0'+id)), nil
}

func (m *mockWorkOrderRepository) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	return m.shipmentExists[shipmentID], nil
}

func (m *mockWorkOrderRepository) ShipmentHasWorkOrder(ctx context.Context, shipmentID string) (bool, error) {
	return m.shipmentHasWO[shipmentID], nil
}

func (m *mockWorkOrderRepository) UpdateStatus(ctx context.Context, id, status string) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	if wo, ok := m.workOrders[id]; ok {
		wo.Status = status
		return nil
	}
	return errors.New("not found")
}

func newTestWorkOrderService() (*WorkOrderServiceImpl, *mockWorkOrderRepository) {
	repo := newMockWorkOrderRepository()
	service := NewWorkOrderService(repo)
	return service, repo
}

func TestWorkOrderService_CreateWorkOrder(t *testing.T) {
	service, repo := newTestWorkOrderService()
	ctx := context.Background()

	repo.shipmentExists["SHIP-001"] = true

	resp, err := service.CreateWorkOrder(ctx, primary.CreateWorkOrderRequest{
		ShipmentID:         "SHIP-001",
		Outcome:            "Test outcome",
		AcceptanceCriteria: "Test criteria",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.WorkOrderID == "" {
		t.Error("expected work order ID to be set")
	}
	if resp.WorkOrder.Status != "draft" {
		t.Errorf("expected status 'draft', got %q", resp.WorkOrder.Status)
	}
}

func TestWorkOrderService_CreateWorkOrder_ShipmentNotFound(t *testing.T) {
	service, _ := newTestWorkOrderService()
	ctx := context.Background()

	_, err := service.CreateWorkOrder(ctx, primary.CreateWorkOrderRequest{
		ShipmentID: "SHIP-999",
		Outcome:    "Test outcome",
	})

	if err == nil {
		t.Error("expected error for non-existent shipment")
	}
}

func TestWorkOrderService_CreateWorkOrder_AlreadyExists(t *testing.T) {
	service, repo := newTestWorkOrderService()
	ctx := context.Background()

	repo.shipmentExists["SHIP-001"] = true
	repo.shipmentHasWO["SHIP-001"] = true

	_, err := service.CreateWorkOrder(ctx, primary.CreateWorkOrderRequest{
		ShipmentID: "SHIP-001",
		Outcome:    "Test outcome",
	})

	if err == nil {
		t.Error("expected error for shipment that already has WO")
	}
}

func TestWorkOrderService_GetWorkOrder(t *testing.T) {
	service, repo := newTestWorkOrderService()
	ctx := context.Background()

	repo.workOrders["WO-001"] = &secondary.WorkOrderRecord{
		ID:         "WO-001",
		ShipmentID: "SHIP-001",
		Outcome:    "Test outcome",
		Status:     "draft",
	}

	wo, err := service.GetWorkOrder(ctx, "WO-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if wo.Outcome != "Test outcome" {
		t.Errorf("expected outcome 'Test outcome', got %q", wo.Outcome)
	}
}

func TestWorkOrderService_GetWorkOrder_NotFound(t *testing.T) {
	service, _ := newTestWorkOrderService()
	ctx := context.Background()

	_, err := service.GetWorkOrder(ctx, "WO-999")
	if err == nil {
		t.Error("expected error for non-existent work order")
	}
}

func TestWorkOrderService_GetWorkOrderByShipment(t *testing.T) {
	service, repo := newTestWorkOrderService()
	ctx := context.Background()

	repo.workOrders["WO-001"] = &secondary.WorkOrderRecord{
		ID:         "WO-001",
		ShipmentID: "SHIP-001",
		Outcome:    "Test outcome",
		Status:     "draft",
	}
	repo.workOrdersByShip["SHIP-001"] = repo.workOrders["WO-001"]

	wo, err := service.GetWorkOrderByShipment(ctx, "SHIP-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if wo.ID != "WO-001" {
		t.Errorf("expected ID 'WO-001', got %q", wo.ID)
	}
}

func TestWorkOrderService_GetWorkOrderByShipment_NotFound(t *testing.T) {
	service, _ := newTestWorkOrderService()
	ctx := context.Background()

	_, err := service.GetWorkOrderByShipment(ctx, "SHIP-999")
	if err == nil {
		t.Error("expected error for non-existent shipment")
	}
}

func TestWorkOrderService_ListWorkOrders(t *testing.T) {
	service, repo := newTestWorkOrderService()
	ctx := context.Background()

	repo.workOrders["WO-001"] = &secondary.WorkOrderRecord{ID: "WO-001", ShipmentID: "SHIP-001", Status: "draft"}
	repo.workOrders["WO-002"] = &secondary.WorkOrderRecord{ID: "WO-002", ShipmentID: "SHIP-002", Status: "active"}

	workOrders, err := service.ListWorkOrders(ctx, primary.WorkOrderFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(workOrders) != 2 {
		t.Errorf("expected 2 work orders, got %d", len(workOrders))
	}
}

func TestWorkOrderService_ListWorkOrders_FilterByStatus(t *testing.T) {
	service, repo := newTestWorkOrderService()
	ctx := context.Background()

	repo.workOrders["WO-001"] = &secondary.WorkOrderRecord{ID: "WO-001", ShipmentID: "SHIP-001", Status: "draft"}
	repo.workOrders["WO-002"] = &secondary.WorkOrderRecord{ID: "WO-002", ShipmentID: "SHIP-002", Status: "active"}

	workOrders, err := service.ListWorkOrders(ctx, primary.WorkOrderFilters{Status: "draft"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(workOrders) != 1 {
		t.Errorf("expected 1 draft work order, got %d", len(workOrders))
	}
}

func TestWorkOrderService_UpdateWorkOrder(t *testing.T) {
	service, repo := newTestWorkOrderService()
	ctx := context.Background()

	repo.workOrders["WO-001"] = &secondary.WorkOrderRecord{
		ID:         "WO-001",
		ShipmentID: "SHIP-001",
		Outcome:    "Original",
		Status:     "draft",
	}

	err := service.UpdateWorkOrder(ctx, primary.UpdateWorkOrderRequest{
		WorkOrderID: "WO-001",
		Outcome:     "Updated",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	wo, _ := service.GetWorkOrder(ctx, "WO-001")
	if wo.Outcome != "Updated" {
		t.Errorf("expected outcome 'Updated', got %q", wo.Outcome)
	}
}

func TestWorkOrderService_UpdateWorkOrder_NotDraft(t *testing.T) {
	service, repo := newTestWorkOrderService()
	ctx := context.Background()

	repo.workOrders["WO-001"] = &secondary.WorkOrderRecord{
		ID:         "WO-001",
		ShipmentID: "SHIP-001",
		Status:     "complete",
	}

	err := service.UpdateWorkOrder(ctx, primary.UpdateWorkOrderRequest{
		WorkOrderID: "WO-001",
		Outcome:     "Updated",
	})
	if err == nil {
		t.Error("expected error for complete work order")
	}
}

func TestWorkOrderService_UpdateWorkOrder_NotFound(t *testing.T) {
	service, _ := newTestWorkOrderService()
	ctx := context.Background()

	err := service.UpdateWorkOrder(ctx, primary.UpdateWorkOrderRequest{
		WorkOrderID: "WO-999",
		Outcome:     "Updated",
	})
	if err == nil {
		t.Error("expected error for non-existent work order")
	}
}

func TestWorkOrderService_DeleteWorkOrder(t *testing.T) {
	service, repo := newTestWorkOrderService()
	ctx := context.Background()

	repo.workOrders["WO-001"] = &secondary.WorkOrderRecord{ID: "WO-001", ShipmentID: "SHIP-001", Status: "draft"}
	repo.workOrdersByShip["SHIP-001"] = repo.workOrders["WO-001"]

	err := service.DeleteWorkOrder(ctx, "WO-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = service.GetWorkOrder(ctx, "WO-001")
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestWorkOrderService_DeleteWorkOrder_NotFound(t *testing.T) {
	service, _ := newTestWorkOrderService()
	ctx := context.Background()

	err := service.DeleteWorkOrder(ctx, "WO-999")
	if err == nil {
		t.Error("expected error for non-existent work order")
	}
}

func TestWorkOrderService_ActivateWorkOrder(t *testing.T) {
	service, repo := newTestWorkOrderService()
	ctx := context.Background()

	repo.workOrders["WO-001"] = &secondary.WorkOrderRecord{ID: "WO-001", ShipmentID: "SHIP-001", Status: "draft"}

	err := service.ActivateWorkOrder(ctx, "WO-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	wo, _ := service.GetWorkOrder(ctx, "WO-001")
	if wo.Status != "active" {
		t.Errorf("expected status 'active', got %q", wo.Status)
	}
}

func TestWorkOrderService_ActivateWorkOrder_NotDraft(t *testing.T) {
	service, repo := newTestWorkOrderService()
	ctx := context.Background()

	repo.workOrders["WO-001"] = &secondary.WorkOrderRecord{ID: "WO-001", ShipmentID: "SHIP-001", Status: "active"}

	err := service.ActivateWorkOrder(ctx, "WO-001")
	if err == nil {
		t.Error("expected error for non-draft work order")
	}
}

func TestWorkOrderService_CompleteWorkOrder(t *testing.T) {
	service, repo := newTestWorkOrderService()
	ctx := context.Background()

	repo.workOrders["WO-001"] = &secondary.WorkOrderRecord{ID: "WO-001", ShipmentID: "SHIP-001", Status: "active"}

	err := service.CompleteWorkOrder(ctx, "WO-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	wo, _ := service.GetWorkOrder(ctx, "WO-001")
	if wo.Status != "complete" {
		t.Errorf("expected status 'complete', got %q", wo.Status)
	}
}

func TestWorkOrderService_CompleteWorkOrder_NotActive(t *testing.T) {
	service, repo := newTestWorkOrderService()
	ctx := context.Background()

	repo.workOrders["WO-001"] = &secondary.WorkOrderRecord{ID: "WO-001", ShipmentID: "SHIP-001", Status: "draft"}

	err := service.CompleteWorkOrder(ctx, "WO-001")
	if err == nil {
		t.Error("expected error for non-active work order")
	}
}
