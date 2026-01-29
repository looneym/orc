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
	receipts       map[string]*secondary.ReceiptRecord
	receiptsByTask map[string]*secondary.ReceiptRecord
	taskExists     map[string]bool
	taskHasReceipt map[string]bool
	nextID         int
	createErr      error
	getErr         error
	updateErr      error
	deleteErr      error
}

func newMockReceiptRepository() *mockReceiptRepository {
	return &mockReceiptRepository{
		receipts:       make(map[string]*secondary.ReceiptRecord),
		receiptsByTask: make(map[string]*secondary.ReceiptRecord),
		taskExists:     make(map[string]bool),
		taskHasReceipt: make(map[string]bool),
		nextID:         1,
	}
}

func (m *mockReceiptRepository) Create(ctx context.Context, rec *secondary.ReceiptRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.receipts[rec.ID] = rec
	m.receiptsByTask[rec.TaskID] = rec
	m.taskHasReceipt[rec.TaskID] = true
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

func (m *mockReceiptRepository) GetByTask(ctx context.Context, taskID string) (*secondary.ReceiptRecord, error) {
	if r, ok := m.receiptsByTask[taskID]; ok {
		return r, nil
	}
	return nil, errors.New("not found")
}

func (m *mockReceiptRepository) List(ctx context.Context, filters secondary.ReceiptFilters) ([]*secondary.ReceiptRecord, error) {
	var result []*secondary.ReceiptRecord
	for _, r := range m.receipts {
		if filters.TaskID != "" && r.TaskID != filters.TaskID {
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
	delete(m.receiptsByTask, r.TaskID)
	delete(m.receipts, id)
	m.taskHasReceipt[r.TaskID] = false
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

func (m *mockReceiptRepository) TaskExists(ctx context.Context, taskID string) (bool, error) {
	return m.taskExists[taskID], nil
}

func (m *mockReceiptRepository) TaskHasReceipt(ctx context.Context, taskID string) (bool, error) {
	return m.taskHasReceipt[taskID], nil
}

func newTestReceiptService() (*ReceiptServiceImpl, *mockReceiptRepository) {
	repo := newMockReceiptRepository()
	service := NewReceiptService(repo)
	return service, repo
}

func TestReceiptService_CreateReceipt(t *testing.T) {
	service, repo := newTestReceiptService()
	ctx := context.Background()

	repo.taskExists["TASK-001"] = true

	resp, err := service.CreateReceipt(ctx, primary.CreateReceiptRequest{
		TaskID:           "TASK-001",
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

func TestReceiptService_CreateReceipt_TaskNotFound(t *testing.T) {
	service, _ := newTestReceiptService()
	ctx := context.Background()

	_, err := service.CreateReceipt(ctx, primary.CreateReceiptRequest{
		TaskID:           "TASK-999",
		DeliveredOutcome: "Test",
	})

	if err == nil {
		t.Error("expected error for non-existent task")
	}
}

func TestReceiptService_CreateReceipt_AlreadyExists(t *testing.T) {
	service, repo := newTestReceiptService()
	ctx := context.Background()

	repo.taskExists["TASK-001"] = true
	repo.taskHasReceipt["TASK-001"] = true

	_, err := service.CreateReceipt(ctx, primary.CreateReceiptRequest{
		TaskID:           "TASK-001",
		DeliveredOutcome: "Test",
	})

	if err == nil {
		t.Error("expected error for task that already has receipt")
	}
}

func TestReceiptService_GetReceipt(t *testing.T) {
	service, repo := newTestReceiptService()
	ctx := context.Background()

	repo.receipts["REC-001"] = &secondary.ReceiptRecord{
		ID:               "REC-001",
		TaskID:           "TASK-001",
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

func TestReceiptService_GetReceiptByTask(t *testing.T) {
	service, repo := newTestReceiptService()
	ctx := context.Background()

	repo.receipts["REC-001"] = &secondary.ReceiptRecord{
		ID:               "REC-001",
		TaskID:           "TASK-001",
		DeliveredOutcome: "Test",
		Status:           "draft",
	}
	repo.receiptsByTask["TASK-001"] = repo.receipts["REC-001"]

	rec, err := service.GetReceiptByTask(ctx, "TASK-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.ID != "REC-001" {
		t.Errorf("expected ID 'REC-001', got %q", rec.ID)
	}
}

func TestReceiptService_GetReceiptByTask_NotFound(t *testing.T) {
	service, _ := newTestReceiptService()
	ctx := context.Background()

	_, err := service.GetReceiptByTask(ctx, "TASK-999")
	if err == nil {
		t.Error("expected error for non-existent task")
	}
}

func TestReceiptService_ListReceipts(t *testing.T) {
	service, repo := newTestReceiptService()
	ctx := context.Background()

	repo.receipts["REC-001"] = &secondary.ReceiptRecord{ID: "REC-001", TaskID: "TASK-001", Status: "draft"}
	repo.receipts["REC-002"] = &secondary.ReceiptRecord{ID: "REC-002", TaskID: "TASK-002", Status: "submitted"}

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
		TaskID:           "TASK-001",
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
		ID:     "REC-001",
		TaskID: "TASK-001",
		Status: "submitted",
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

	repo.receipts["REC-001"] = &secondary.ReceiptRecord{ID: "REC-001", TaskID: "TASK-001", Status: "draft"}
	repo.receiptsByTask["TASK-001"] = repo.receipts["REC-001"]

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
		ID:     "REC-001",
		TaskID: "TASK-001",
		Status: "draft",
	}

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
		ID:     "REC-001",
		TaskID: "TASK-001",
		Status: "verified",
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
		ID:     "REC-001",
		TaskID: "TASK-001",
		Status: "submitted",
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
		ID:     "REC-001",
		TaskID: "TASK-001",
		Status: "draft",
	}

	err := service.VerifyReceipt(ctx, "REC-001")
	if err == nil {
		t.Error("expected error for non-submitted receipt")
	}
}
