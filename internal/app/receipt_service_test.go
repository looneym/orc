package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockReceiptRepository implements secondary.ReceiptRepository for testing.
type mockReceiptRepository struct {
	receipts         map[string]*secondary.ReceiptRecord
	receiptsByShip   map[string]*secondary.ReceiptRecord
	shipmentExists   map[string]bool
	shipmentHasREC   map[string]bool
	woStatus         map[string]string
	allCRECsVerified map[string]bool
	nextID           int
	createErr        error
	getErr           error
	updateErr        error
	deleteErr        error
}

func newMockReceiptRepository() *mockReceiptRepository {
	return &mockReceiptRepository{
		receipts:         make(map[string]*secondary.ReceiptRecord),
		receiptsByShip:   make(map[string]*secondary.ReceiptRecord),
		shipmentExists:   make(map[string]bool),
		shipmentHasREC:   make(map[string]bool),
		woStatus:         make(map[string]string),
		allCRECsVerified: make(map[string]bool),
		nextID:           1,
	}
}

func (m *mockReceiptRepository) Create(ctx context.Context, rec *secondary.ReceiptRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.receipts[rec.ID] = rec
	m.receiptsByShip[rec.ShipmentID] = rec
	m.shipmentHasREC[rec.ShipmentID] = true
	return nil
}

func (m *mockReceiptRepository) GetByID(ctx context.Context, id string) (*secondary.ReceiptRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if r, ok := m.receipts[id]; ok {
		return r, nil
	}
	return nil, errors.New("not found")
}

func (m *mockReceiptRepository) GetByShipment(ctx context.Context, shipmentID string) (*secondary.ReceiptRecord, error) {
	if r, ok := m.receiptsByShip[shipmentID]; ok {
		return r, nil
	}
	return nil, errors.New("not found")
}

func (m *mockReceiptRepository) List(ctx context.Context, filters secondary.ReceiptFilters) ([]*secondary.ReceiptRecord, error) {
	var result []*secondary.ReceiptRecord
	for _, r := range m.receipts {
		if filters.ShipmentID != "" && r.ShipmentID != filters.ShipmentID {
			continue
		}
		if filters.Status != "" && r.Status != filters.Status {
			continue
		}
		result = append(result, r)
	}
	return result, nil
}

func (m *mockReceiptRepository) Update(ctx context.Context, rec *secondary.ReceiptRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.receipts[rec.ID]; !ok {
		return errors.New("not found")
	}
	existing := m.receipts[rec.ID]
	if rec.DeliveredOutcome != "" {
		existing.DeliveredOutcome = rec.DeliveredOutcome
	}
	if rec.Evidence != "" {
		existing.Evidence = rec.Evidence
	}
	if rec.VerificationNotes != "" {
		existing.VerificationNotes = rec.VerificationNotes
	}
	return nil
}

func (m *mockReceiptRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.receipts[id]; !ok {
		return errors.New("not found")
	}
	r := m.receipts[id]
	delete(m.receiptsByShip, r.ShipmentID)
	delete(m.receipts, id)
	m.shipmentHasREC[r.ShipmentID] = false
	return nil
}

func (m *mockReceiptRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("REC-%03d", id), nil
}

func (m *mockReceiptRepository) UpdateStatus(ctx context.Context, id, status string) error {
	if r, ok := m.receipts[id]; ok {
		r.Status = status
		return nil
	}
	return errors.New("not found")
}

func (m *mockReceiptRepository) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	return m.shipmentExists[shipmentID], nil
}

func (m *mockReceiptRepository) ShipmentHasREC(ctx context.Context, shipmentID string) (bool, error) {
	return m.shipmentHasREC[shipmentID], nil
}

func (m *mockReceiptRepository) GetWOStatus(ctx context.Context, shipmentID string) (string, error) {
	if status, ok := m.woStatus[shipmentID]; ok {
		return status, nil
	}
	return "", fmt.Errorf("Work Order for shipment %s not found", shipmentID)
}

func (m *mockReceiptRepository) AllCRECsVerified(ctx context.Context, shipmentID string) (bool, error) {
	return m.allCRECsVerified[shipmentID], nil
}

func newTestReceiptService() (*ReceiptServiceImpl, *mockReceiptRepository) {
	repo := newMockReceiptRepository()
	service := NewReceiptService(repo)
	return service, repo
}

func TestReceiptService_CreateReceipt(t *testing.T) {
	service, repo := newTestReceiptService()
	ctx := context.Background()

	repo.shipmentExists["SHIP-001"] = true

	resp, err := service.CreateReceipt(ctx, primary.CreateReceiptRequest{
		ShipmentID:       "SHIP-001",
		DeliveredOutcome: "Completed the task",
		Evidence:         "Screenshot attached",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.ReceiptID == "" {
		t.Error("expected receipt ID to be set")
	}
	if resp.Receipt.Status != "draft" {
		t.Errorf("expected status 'draft', got %q", resp.Receipt.Status)
	}
}

func TestReceiptService_CreateReceipt_ShipmentNotFound(t *testing.T) {
	service, _ := newTestReceiptService()
	ctx := context.Background()

	_, err := service.CreateReceipt(ctx, primary.CreateReceiptRequest{
		ShipmentID:       "SHIP-999",
		DeliveredOutcome: "Test",
	})

	if err == nil {
		t.Error("expected error for non-existent shipment")
	}
}

func TestReceiptService_CreateReceipt_AlreadyExists(t *testing.T) {
	service, repo := newTestReceiptService()
	ctx := context.Background()

	repo.shipmentExists["SHIP-001"] = true
	repo.shipmentHasREC["SHIP-001"] = true

	_, err := service.CreateReceipt(ctx, primary.CreateReceiptRequest{
		ShipmentID:       "SHIP-001",
		DeliveredOutcome: "Test",
	})

	if err == nil {
		t.Error("expected error for shipment that already has REC")
	}
}

func TestReceiptService_GetReceipt(t *testing.T) {
	service, repo := newTestReceiptService()
	ctx := context.Background()

	repo.receipts["REC-001"] = &secondary.ReceiptRecord{
		ID:               "REC-001",
		ShipmentID:       "SHIP-001",
		DeliveredOutcome: "Test outcome",
		Status:           "draft",
	}

	rec, err := service.GetReceipt(ctx, "REC-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.DeliveredOutcome != "Test outcome" {
		t.Errorf("expected outcome 'Test outcome', got %q", rec.DeliveredOutcome)
	}
}

func TestReceiptService_GetReceipt_NotFound(t *testing.T) {
	service, _ := newTestReceiptService()
	ctx := context.Background()

	_, err := service.GetReceipt(ctx, "REC-999")
	if err == nil {
		t.Error("expected error for non-existent receipt")
	}
}

func TestReceiptService_GetReceiptByShipment(t *testing.T) {
	service, repo := newTestReceiptService()
	ctx := context.Background()

	repo.receipts["REC-001"] = &secondary.ReceiptRecord{
		ID:               "REC-001",
		ShipmentID:       "SHIP-001",
		DeliveredOutcome: "Test",
		Status:           "draft",
	}
	repo.receiptsByShip["SHIP-001"] = repo.receipts["REC-001"]

	rec, err := service.GetReceiptByShipment(ctx, "SHIP-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.ID != "REC-001" {
		t.Errorf("expected ID 'REC-001', got %q", rec.ID)
	}
}

func TestReceiptService_GetReceiptByShipment_NotFound(t *testing.T) {
	service, _ := newTestReceiptService()
	ctx := context.Background()

	_, err := service.GetReceiptByShipment(ctx, "SHIP-999")
	if err == nil {
		t.Error("expected error for non-existent shipment")
	}
}

func TestReceiptService_ListReceipts(t *testing.T) {
	service, repo := newTestReceiptService()
	ctx := context.Background()

	repo.receipts["REC-001"] = &secondary.ReceiptRecord{ID: "REC-001", ShipmentID: "SHIP-001", Status: "draft"}
	repo.receipts["REC-002"] = &secondary.ReceiptRecord{ID: "REC-002", ShipmentID: "SHIP-002", Status: "submitted"}

	receipts, err := service.ListReceipts(ctx, primary.ReceiptFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(receipts) != 2 {
		t.Errorf("expected 2 receipts, got %d", len(receipts))
	}
}

func TestReceiptService_UpdateReceipt(t *testing.T) {
	service, repo := newTestReceiptService()
	ctx := context.Background()

	repo.receipts["REC-001"] = &secondary.ReceiptRecord{
		ID:               "REC-001",
		ShipmentID:       "SHIP-001",
		DeliveredOutcome: "Original",
		Status:           "draft",
	}

	err := service.UpdateReceipt(ctx, primary.UpdateReceiptRequest{
		ReceiptID:        "REC-001",
		DeliveredOutcome: "Updated",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	rec, _ := service.GetReceipt(ctx, "REC-001")
	if rec.DeliveredOutcome != "Updated" {
		t.Errorf("expected outcome 'Updated', got %q", rec.DeliveredOutcome)
	}
}

func TestReceiptService_UpdateReceipt_NotDraft(t *testing.T) {
	service, repo := newTestReceiptService()
	ctx := context.Background()

	repo.receipts["REC-001"] = &secondary.ReceiptRecord{
		ID:         "REC-001",
		ShipmentID: "SHIP-001",
		Status:     "submitted",
	}

	err := service.UpdateReceipt(ctx, primary.UpdateReceiptRequest{
		ReceiptID:        "REC-001",
		DeliveredOutcome: "Updated",
	})
	if err == nil {
		t.Error("expected error for non-draft receipt")
	}
}

func TestReceiptService_UpdateReceipt_NotFound(t *testing.T) {
	service, _ := newTestReceiptService()
	ctx := context.Background()

	err := service.UpdateReceipt(ctx, primary.UpdateReceiptRequest{
		ReceiptID:        "REC-999",
		DeliveredOutcome: "Updated",
	})
	if err == nil {
		t.Error("expected error for non-existent receipt")
	}
}

func TestReceiptService_DeleteReceipt(t *testing.T) {
	service, repo := newTestReceiptService()
	ctx := context.Background()

	repo.receipts["REC-001"] = &secondary.ReceiptRecord{ID: "REC-001", ShipmentID: "SHIP-001", Status: "draft"}
	repo.receiptsByShip["SHIP-001"] = repo.receipts["REC-001"]

	err := service.DeleteReceipt(ctx, "REC-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = service.GetReceipt(ctx, "REC-001")
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestReceiptService_DeleteReceipt_NotFound(t *testing.T) {
	service, _ := newTestReceiptService()
	ctx := context.Background()

	err := service.DeleteReceipt(ctx, "REC-999")
	if err == nil {
		t.Error("expected error for non-existent receipt")
	}
}

func TestReceiptService_SubmitReceipt(t *testing.T) {
	service, repo := newTestReceiptService()
	ctx := context.Background()

	repo.receipts["REC-001"] = &secondary.ReceiptRecord{
		ID:         "REC-001",
		ShipmentID: "SHIP-001",
		Status:     "draft",
	}
	repo.woStatus["SHIP-001"] = "complete"
	repo.allCRECsVerified["SHIP-001"] = true

	err := service.SubmitReceipt(ctx, "REC-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	rec, _ := service.GetReceipt(ctx, "REC-001")
	if rec.Status != "submitted" {
		t.Errorf("expected status 'submitted', got %q", rec.Status)
	}
}

func TestReceiptService_SubmitReceipt_NotDraft(t *testing.T) {
	service, repo := newTestReceiptService()
	ctx := context.Background()

	repo.receipts["REC-001"] = &secondary.ReceiptRecord{
		ID:         "REC-001",
		ShipmentID: "SHIP-001",
		Status:     "verified",
	}

	err := service.SubmitReceipt(ctx, "REC-001")
	if err == nil {
		t.Error("expected error for non-draft receipt")
	}
}

func TestReceiptService_VerifyReceipt(t *testing.T) {
	service, repo := newTestReceiptService()
	ctx := context.Background()

	repo.receipts["REC-001"] = &secondary.ReceiptRecord{
		ID:         "REC-001",
		ShipmentID: "SHIP-001",
		Status:     "submitted",
	}

	err := service.VerifyReceipt(ctx, "REC-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	rec, _ := service.GetReceipt(ctx, "REC-001")
	if rec.Status != "verified" {
		t.Errorf("expected status 'verified', got %q", rec.Status)
	}
}

func TestReceiptService_VerifyReceipt_NotSubmitted(t *testing.T) {
	service, repo := newTestReceiptService()
	ctx := context.Background()

	repo.receipts["REC-001"] = &secondary.ReceiptRecord{
		ID:         "REC-001",
		ShipmentID: "SHIP-001",
		Status:     "draft",
	}

	err := service.VerifyReceipt(ctx, "REC-001")
	if err == nil {
		t.Error("expected error for non-submitted receipt")
	}
}
