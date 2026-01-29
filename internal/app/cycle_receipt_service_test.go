package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockCycleReceiptRepository implements secondary.CycleReceiptRepository for testing.
type mockCycleReceiptRepository struct {
	crecs         map[string]*secondary.CycleReceiptRecord
	crecsByCWO    map[string]*secondary.CycleReceiptRecord
	cwoExists     map[string]bool
	cwoHasCREC    map[string]bool
	cwoShipmentID map[string]string
	cwoStatus     map[string]string
	cwoCycleID    map[string]string
	nextID        int
	createErr     error
	getErr        error
	updateErr     error
	deleteErr     error
}

func newMockCycleReceiptRepository() *mockCycleReceiptRepository {
	return &mockCycleReceiptRepository{
		crecs:         make(map[string]*secondary.CycleReceiptRecord),
		crecsByCWO:    make(map[string]*secondary.CycleReceiptRecord),
		cwoExists:     make(map[string]bool),
		cwoHasCREC:    make(map[string]bool),
		cwoShipmentID: make(map[string]string),
		cwoStatus:     make(map[string]string),
		cwoCycleID:    make(map[string]string),
		nextID:        1,
	}
}

func (m *mockCycleReceiptRepository) Create(ctx context.Context, crec *secondary.CycleReceiptRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.crecs[crec.ID] = crec
	m.crecsByCWO[crec.CWOID] = crec
	m.cwoHasCREC[crec.CWOID] = true
	return nil
}

func (m *mockCycleReceiptRepository) GetByID(ctx context.Context, id string) (*secondary.CycleReceiptRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if crec, ok := m.crecs[id]; ok {
		return crec, nil
	}
	return nil, errors.New("not found")
}

func (m *mockCycleReceiptRepository) GetByCWO(ctx context.Context, cwoID string) (*secondary.CycleReceiptRecord, error) {
	if crec, ok := m.crecsByCWO[cwoID]; ok {
		return crec, nil
	}
	return nil, errors.New("not found")
}

func (m *mockCycleReceiptRepository) List(ctx context.Context, filters secondary.CycleReceiptFilters) ([]*secondary.CycleReceiptRecord, error) {
	var result []*secondary.CycleReceiptRecord
	for _, crec := range m.crecs {
		if filters.CWOID != "" && crec.CWOID != filters.CWOID {
			continue
		}
		if filters.ShipmentID != "" && crec.ShipmentID != filters.ShipmentID {
			continue
		}
		if filters.Status != "" && crec.Status != filters.Status {
			continue
		}
		result = append(result, crec)
	}
	return result, nil
}

func (m *mockCycleReceiptRepository) Update(ctx context.Context, crec *secondary.CycleReceiptRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.crecs[crec.ID]; !ok {
		return errors.New("not found")
	}
	existing := m.crecs[crec.ID]
	if crec.DeliveredOutcome != "" {
		existing.DeliveredOutcome = crec.DeliveredOutcome
	}
	if crec.Evidence != "" {
		existing.Evidence = crec.Evidence
	}
	if crec.VerificationNotes != "" {
		existing.VerificationNotes = crec.VerificationNotes
	}
	return nil
}

func (m *mockCycleReceiptRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.crecs[id]; !ok {
		return errors.New("not found")
	}
	crec := m.crecs[id]
	delete(m.crecsByCWO, crec.CWOID)
	delete(m.crecs, id)
	m.cwoHasCREC[crec.CWOID] = false
	return nil
}

func (m *mockCycleReceiptRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("CREC-%03d", id), nil
}

func (m *mockCycleReceiptRepository) UpdateStatus(ctx context.Context, id, status string) error {
	if crec, ok := m.crecs[id]; ok {
		crec.Status = status
		return nil
	}
	return errors.New("not found")
}

func (m *mockCycleReceiptRepository) CWOExists(ctx context.Context, cwoID string) (bool, error) {
	return m.cwoExists[cwoID], nil
}

func (m *mockCycleReceiptRepository) CWOHasCREC(ctx context.Context, cwoID string) (bool, error) {
	return m.cwoHasCREC[cwoID], nil
}

func (m *mockCycleReceiptRepository) GetCWOShipmentID(ctx context.Context, cwoID string) (string, error) {
	if sid, ok := m.cwoShipmentID[cwoID]; ok {
		return sid, nil
	}
	return "", errors.New("CWO not found")
}

func (m *mockCycleReceiptRepository) GetCWOStatus(ctx context.Context, cwoID string) (string, error) {
	if status, ok := m.cwoStatus[cwoID]; ok {
		return status, nil
	}
	return "", errors.New("CWO not found")
}

func (m *mockCycleReceiptRepository) GetCWOCycleID(ctx context.Context, cwoID string) (string, error) {
	if cycleID, ok := m.cwoCycleID[cwoID]; ok {
		return cycleID, nil
	}
	return "", errors.New("CWO not found")
}

// mockCycleServiceForCREC implements primary.CycleService for testing cascade updates.
type mockCycleServiceForCREC struct {
	updateStatusCalls []struct {
		CycleID string
		Status  string
	}
	updateStatusErr error
}

func newMockCycleServiceForCREC() *mockCycleServiceForCREC {
	return &mockCycleServiceForCREC{}
}

func (m *mockCycleServiceForCREC) CreateCycle(ctx context.Context, req primary.CreateCycleRequest) (*primary.CreateCycleResponse, error) {
	return nil, nil
}

func (m *mockCycleServiceForCREC) GetCycle(ctx context.Context, cycleID string) (*primary.Cycle, error) {
	return nil, nil
}

func (m *mockCycleServiceForCREC) ListCycles(ctx context.Context, filters primary.CycleFilters) ([]*primary.Cycle, error) {
	return nil, nil
}

func (m *mockCycleServiceForCREC) DeleteCycle(ctx context.Context, cycleID string) error {
	return nil
}

func (m *mockCycleServiceForCREC) StartCycle(ctx context.Context, cycleID string) error {
	return nil
}

func (m *mockCycleServiceForCREC) CompleteCycle(ctx context.Context, cycleID string) error {
	return nil
}

func (m *mockCycleServiceForCREC) GetActiveCycle(ctx context.Context, shipmentID string) (*primary.Cycle, error) {
	return nil, nil
}

func (m *mockCycleServiceForCREC) UpdateCycleStatus(ctx context.Context, cycleID string, status string) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	m.updateStatusCalls = append(m.updateStatusCalls, struct {
		CycleID string
		Status  string
	}{cycleID, status})
	return nil
}

func newTestCycleReceiptService() (*CycleReceiptServiceImpl, *mockCycleReceiptRepository, *mockCycleServiceForCREC) {
	repo := newMockCycleReceiptRepository()
	cycleService := newMockCycleServiceForCREC()
	service := NewCycleReceiptService(repo, cycleService)
	return service, repo, cycleService
}

func TestCycleReceiptService_CreateCycleReceipt(t *testing.T) {
	service, repo, _ := newTestCycleReceiptService()
	ctx := context.Background()

	repo.cwoExists["CWO-001"] = true
	repo.cwoShipmentID["CWO-001"] = "SHIP-001"

	resp, err := service.CreateCycleReceipt(ctx, primary.CreateCycleReceiptRequest{
		CWOID:            "CWO-001",
		DeliveredOutcome: "Completed the task",
		Evidence:         "Screenshot attached",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.CycleReceiptID == "" {
		t.Error("expected CREC ID to be set")
	}
	if resp.CycleReceipt.Status != "draft" {
		t.Errorf("expected status 'draft', got %q", resp.CycleReceipt.Status)
	}
}

func TestCycleReceiptService_CreateCycleReceipt_CWONotFound(t *testing.T) {
	service, _, _ := newTestCycleReceiptService()
	ctx := context.Background()

	_, err := service.CreateCycleReceipt(ctx, primary.CreateCycleReceiptRequest{
		CWOID:            "CWO-999",
		DeliveredOutcome: "Test",
	})

	if err == nil {
		t.Error("expected error for non-existent CWO")
	}
}

func TestCycleReceiptService_CreateCycleReceipt_AlreadyExists(t *testing.T) {
	service, repo, _ := newTestCycleReceiptService()
	ctx := context.Background()

	repo.cwoExists["CWO-001"] = true
	repo.cwoHasCREC["CWO-001"] = true

	_, err := service.CreateCycleReceipt(ctx, primary.CreateCycleReceiptRequest{
		CWOID:            "CWO-001",
		DeliveredOutcome: "Test",
	})

	if err == nil {
		t.Error("expected error for CWO that already has CREC")
	}
}

func TestCycleReceiptService_GetCycleReceipt(t *testing.T) {
	service, repo, _ := newTestCycleReceiptService()
	ctx := context.Background()

	repo.crecs["CREC-001"] = &secondary.CycleReceiptRecord{
		ID:               "CREC-001",
		CWOID:            "CWO-001",
		ShipmentID:       "SHIP-001",
		DeliveredOutcome: "Test outcome",
		Status:           "draft",
	}

	crec, err := service.GetCycleReceipt(ctx, "CREC-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if crec.DeliveredOutcome != "Test outcome" {
		t.Errorf("expected outcome 'Test outcome', got %q", crec.DeliveredOutcome)
	}
}

func TestCycleReceiptService_GetCycleReceipt_NotFound(t *testing.T) {
	service, _, _ := newTestCycleReceiptService()
	ctx := context.Background()

	_, err := service.GetCycleReceipt(ctx, "CREC-999")
	if err == nil {
		t.Error("expected error for non-existent CREC")
	}
}

func TestCycleReceiptService_GetCycleReceiptByCWO(t *testing.T) {
	service, repo, _ := newTestCycleReceiptService()
	ctx := context.Background()

	repo.crecs["CREC-001"] = &secondary.CycleReceiptRecord{
		ID:         "CREC-001",
		CWOID:      "CWO-001",
		ShipmentID: "SHIP-001",
		Status:     "draft",
	}
	repo.crecsByCWO["CWO-001"] = repo.crecs["CREC-001"]

	crec, err := service.GetCycleReceiptByCWO(ctx, "CWO-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if crec.ID != "CREC-001" {
		t.Errorf("expected ID 'CREC-001', got %q", crec.ID)
	}
}

func TestCycleReceiptService_GetCycleReceiptByCWO_NotFound(t *testing.T) {
	service, _, _ := newTestCycleReceiptService()
	ctx := context.Background()

	_, err := service.GetCycleReceiptByCWO(ctx, "CWO-999")
	if err == nil {
		t.Error("expected error for non-existent CWO")
	}
}

func TestCycleReceiptService_ListCycleReceipts(t *testing.T) {
	service, repo, _ := newTestCycleReceiptService()
	ctx := context.Background()

	repo.crecs["CREC-001"] = &secondary.CycleReceiptRecord{ID: "CREC-001", CWOID: "CWO-001", Status: "draft"}
	repo.crecs["CREC-002"] = &secondary.CycleReceiptRecord{ID: "CREC-002", CWOID: "CWO-002", Status: "submitted"}

	crecs, err := service.ListCycleReceipts(ctx, primary.CycleReceiptFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(crecs) != 2 {
		t.Errorf("expected 2 CRECs, got %d", len(crecs))
	}
}

func TestCycleReceiptService_UpdateCycleReceipt(t *testing.T) {
	service, repo, _ := newTestCycleReceiptService()
	ctx := context.Background()

	repo.crecs["CREC-001"] = &secondary.CycleReceiptRecord{
		ID:               "CREC-001",
		CWOID:            "CWO-001",
		DeliveredOutcome: "Original",
		Status:           "draft",
	}

	err := service.UpdateCycleReceipt(ctx, primary.UpdateCycleReceiptRequest{
		CycleReceiptID:   "CREC-001",
		DeliveredOutcome: "Updated",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	crec, _ := service.GetCycleReceipt(ctx, "CREC-001")
	if crec.DeliveredOutcome != "Updated" {
		t.Errorf("expected outcome 'Updated', got %q", crec.DeliveredOutcome)
	}
}

func TestCycleReceiptService_UpdateCycleReceipt_NotDraft(t *testing.T) {
	service, repo, _ := newTestCycleReceiptService()
	ctx := context.Background()

	repo.crecs["CREC-001"] = &secondary.CycleReceiptRecord{
		ID:     "CREC-001",
		CWOID:  "CWO-001",
		Status: "submitted",
	}

	err := service.UpdateCycleReceipt(ctx, primary.UpdateCycleReceiptRequest{
		CycleReceiptID:   "CREC-001",
		DeliveredOutcome: "Updated",
	})
	if err == nil {
		t.Error("expected error for non-draft CREC")
	}
}

func TestCycleReceiptService_UpdateCycleReceipt_NotFound(t *testing.T) {
	service, _, _ := newTestCycleReceiptService()
	ctx := context.Background()

	err := service.UpdateCycleReceipt(ctx, primary.UpdateCycleReceiptRequest{
		CycleReceiptID:   "CREC-999",
		DeliveredOutcome: "Updated",
	})
	if err == nil {
		t.Error("expected error for non-existent CREC")
	}
}

func TestCycleReceiptService_DeleteCycleReceipt(t *testing.T) {
	service, repo, _ := newTestCycleReceiptService()
	ctx := context.Background()

	repo.crecs["CREC-001"] = &secondary.CycleReceiptRecord{ID: "CREC-001", CWOID: "CWO-001", Status: "draft"}
	repo.crecsByCWO["CWO-001"] = repo.crecs["CREC-001"]

	err := service.DeleteCycleReceipt(ctx, "CREC-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = service.GetCycleReceipt(ctx, "CREC-001")
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestCycleReceiptService_DeleteCycleReceipt_NotFound(t *testing.T) {
	service, _, _ := newTestCycleReceiptService()
	ctx := context.Background()

	err := service.DeleteCycleReceipt(ctx, "CREC-999")
	if err == nil {
		t.Error("expected error for non-existent CREC")
	}
}

func TestCycleReceiptService_SubmitCycleReceipt(t *testing.T) {
	service, repo, cycleService := newTestCycleReceiptService()
	ctx := context.Background()

	repo.crecs["CREC-001"] = &secondary.CycleReceiptRecord{
		ID:     "CREC-001",
		CWOID:  "CWO-001",
		Status: "draft",
	}
	repo.cwoExists["CWO-001"] = true
	repo.cwoStatus["CWO-001"] = "complete"
	repo.cwoCycleID["CWO-001"] = "CYC-001"

	err := service.SubmitCycleReceipt(ctx, "CREC-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	crec, _ := service.GetCycleReceipt(ctx, "CREC-001")
	if crec.Status != "submitted" {
		t.Errorf("expected status 'submitted', got %q", crec.Status)
	}

	// Verify cascade to cycle service
	if len(cycleService.updateStatusCalls) != 1 {
		t.Errorf("expected 1 cascade call, got %d", len(cycleService.updateStatusCalls))
	}
	if cycleService.updateStatusCalls[0].Status != "review" {
		t.Errorf("expected cascade status 'review', got %q", cycleService.updateStatusCalls[0].Status)
	}
}

func TestCycleReceiptService_SubmitCycleReceipt_NotDraft(t *testing.T) {
	service, repo, _ := newTestCycleReceiptService()
	ctx := context.Background()

	repo.crecs["CREC-001"] = &secondary.CycleReceiptRecord{
		ID:     "CREC-001",
		CWOID:  "CWO-001",
		Status: "verified",
	}
	repo.cwoExists["CWO-001"] = true

	err := service.SubmitCycleReceipt(ctx, "CREC-001")
	if err == nil {
		t.Error("expected error for non-draft CREC")
	}
}

func TestCycleReceiptService_VerifyCycleReceipt(t *testing.T) {
	service, repo, cycleService := newTestCycleReceiptService()
	ctx := context.Background()

	repo.crecs["CREC-001"] = &secondary.CycleReceiptRecord{
		ID:               "CREC-001",
		CWOID:            "CWO-001",
		DeliveredOutcome: "Success",
		Status:           "submitted",
	}
	repo.cwoCycleID["CWO-001"] = "CYC-001"

	err := service.VerifyCycleReceipt(ctx, "CREC-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	crec, _ := service.GetCycleReceipt(ctx, "CREC-001")
	if crec.Status != "verified" {
		t.Errorf("expected status 'verified', got %q", crec.Status)
	}

	// Verify cascade to cycle service with "complete" status
	if len(cycleService.updateStatusCalls) != 1 {
		t.Errorf("expected 1 cascade call, got %d", len(cycleService.updateStatusCalls))
	}
	if cycleService.updateStatusCalls[0].Status != "complete" {
		t.Errorf("expected cascade status 'complete', got %q", cycleService.updateStatusCalls[0].Status)
	}
}

func TestCycleReceiptService_VerifyCycleReceipt_Failed(t *testing.T) {
	service, repo, cycleService := newTestCycleReceiptService()
	ctx := context.Background()

	repo.crecs["CREC-001"] = &secondary.CycleReceiptRecord{
		ID:               "CREC-001",
		CWOID:            "CWO-001",
		DeliveredOutcome: "FAILED: Tests did not pass",
		Status:           "submitted",
	}
	repo.cwoCycleID["CWO-001"] = "CYC-001"

	err := service.VerifyCycleReceipt(ctx, "CREC-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify cascade to cycle service with "failed" status
	if len(cycleService.updateStatusCalls) != 1 {
		t.Errorf("expected 1 cascade call, got %d", len(cycleService.updateStatusCalls))
	}
	if cycleService.updateStatusCalls[0].Status != "failed" {
		t.Errorf("expected cascade status 'failed', got %q", cycleService.updateStatusCalls[0].Status)
	}
}

func TestCycleReceiptService_VerifyCycleReceipt_NotSubmitted(t *testing.T) {
	service, repo, _ := newTestCycleReceiptService()
	ctx := context.Background()

	repo.crecs["CREC-001"] = &secondary.CycleReceiptRecord{
		ID:     "CREC-001",
		CWOID:  "CWO-001",
		Status: "draft",
	}

	err := service.VerifyCycleReceipt(ctx, "CREC-001")
	if err == nil {
		t.Error("expected error for non-submitted CREC")
	}
}
