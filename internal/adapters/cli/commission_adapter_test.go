package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/example/orc/internal/ports/primary"
)

// mockCommissionService implements primary.CommissionService for testing
type mockCommissionService struct {
	createCommissionFn   func(ctx context.Context, req primary.CreateCommissionRequest) (*primary.CreateCommissionResponse, error)
	listCommissionsFn    func(ctx context.Context, filters primary.CommissionFilters) ([]*primary.Commission, error)
	getCommissionFn      func(ctx context.Context, commissionID string) (*primary.Commission, error)
	updateCommissionFn   func(ctx context.Context, req primary.UpdateCommissionRequest) error
	completeCommissionFn func(ctx context.Context, commissionID string) error
	archiveCommissionFn  func(ctx context.Context, commissionID string) error
	deleteCommissionFn   func(ctx context.Context, req primary.DeleteCommissionRequest) error
	pinCommissionFn      func(ctx context.Context, commissionID string) error
	unpinCommissionFn    func(ctx context.Context, commissionID string) error

	// Track calls for verification
	lastCreateReq primary.CreateCommissionRequest
	lastUpdateReq primary.UpdateCommissionRequest
	lastDeleteReq primary.DeleteCommissionRequest
}

func (m *mockCommissionService) CreateCommission(ctx context.Context, req primary.CreateCommissionRequest) (*primary.CreateCommissionResponse, error) {
	m.lastCreateReq = req
	if m.createCommissionFn != nil {
		return m.createCommissionFn(ctx, req)
	}
	return &primary.CreateCommissionResponse{
		CommissionID: "COMM-001",
		Commission:   &primary.Commission{ID: "COMM-001", Title: req.Title},
	}, nil
}

func (m *mockCommissionService) ListCommissions(ctx context.Context, filters primary.CommissionFilters) ([]*primary.Commission, error) {
	if m.listCommissionsFn != nil {
		return m.listCommissionsFn(ctx, filters)
	}
	return []*primary.Commission{}, nil
}

func (m *mockCommissionService) GetCommission(ctx context.Context, commissionID string) (*primary.Commission, error) {
	if m.getCommissionFn != nil {
		return m.getCommissionFn(ctx, commissionID)
	}
	return &primary.Commission{ID: commissionID, Title: "Test Commission", Status: "active"}, nil
}

func (m *mockCommissionService) StartCommission(ctx context.Context, req primary.StartCommissionRequest) (*primary.StartCommissionResponse, error) {
	return nil, errors.New("not implemented in adapter")
}

func (m *mockCommissionService) LaunchCommission(ctx context.Context, req primary.LaunchCommissionRequest) (*primary.LaunchCommissionResponse, error) {
	return nil, errors.New("not implemented in adapter")
}

func (m *mockCommissionService) UpdateCommission(ctx context.Context, req primary.UpdateCommissionRequest) error {
	m.lastUpdateReq = req
	if m.updateCommissionFn != nil {
		return m.updateCommissionFn(ctx, req)
	}
	return nil
}

func (m *mockCommissionService) CompleteCommission(ctx context.Context, commissionID string) error {
	if m.completeCommissionFn != nil {
		return m.completeCommissionFn(ctx, commissionID)
	}
	return nil
}

func (m *mockCommissionService) ArchiveCommission(ctx context.Context, commissionID string) error {
	if m.archiveCommissionFn != nil {
		return m.archiveCommissionFn(ctx, commissionID)
	}
	return nil
}

func (m *mockCommissionService) DeleteCommission(ctx context.Context, req primary.DeleteCommissionRequest) error {
	m.lastDeleteReq = req
	if m.deleteCommissionFn != nil {
		return m.deleteCommissionFn(ctx, req)
	}
	return nil
}

func (m *mockCommissionService) PinCommission(ctx context.Context, commissionID string) error {
	if m.pinCommissionFn != nil {
		return m.pinCommissionFn(ctx, commissionID)
	}
	return nil
}

func (m *mockCommissionService) UnpinCommission(ctx context.Context, commissionID string) error {
	if m.unpinCommissionFn != nil {
		return m.unpinCommissionFn(ctx, commissionID)
	}
	return nil
}

// ============================================================================
// Create Tests
// ============================================================================

func TestCommissionAdapter_Create_Success(t *testing.T) {
	mock := &mockCommissionService{}
	var buf bytes.Buffer
	adapter := NewCommissionAdapter(mock, &buf)

	err := adapter.Create(context.Background(), "Test Commission", "A description")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastCreateReq.Title != "Test Commission" {
		t.Errorf("expected title 'Test Commission', got '%s'", mock.lastCreateReq.Title)
	}
	if mock.lastCreateReq.Description != "A description" {
		t.Errorf("expected description 'A description', got '%s'", mock.lastCreateReq.Description)
	}
	if !strings.Contains(buf.String(), "Created commission COMM-001") {
		t.Errorf("expected output to contain 'Created commission COMM-001', got '%s'", buf.String())
	}
}

func TestCommissionAdapter_Create_ServiceError(t *testing.T) {
	mock := &mockCommissionService{
		createCommissionFn: func(ctx context.Context, req primary.CreateCommissionRequest) (*primary.CreateCommissionResponse, error) {
			return nil, errors.New("IMPs cannot create commissions")
		},
	}
	var buf bytes.Buffer
	adapter := NewCommissionAdapter(mock, &buf)

	err := adapter.Create(context.Background(), "Test", "")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "IMPs cannot create commissions") {
		t.Errorf("expected error to contain guard message, got '%s'", err.Error())
	}
}

// ============================================================================
// List Tests
// ============================================================================

func TestCommissionAdapter_List_WithResults(t *testing.T) {
	mock := &mockCommissionService{
		listCommissionsFn: func(ctx context.Context, filters primary.CommissionFilters) ([]*primary.Commission, error) {
			return []*primary.Commission{
				{ID: "COMM-001", Title: "First", Status: "active"},
				{ID: "COMM-002", Title: "Second", Status: "complete"},
			}, nil
		},
	}
	var buf bytes.Buffer
	adapter := NewCommissionAdapter(mock, &buf)

	err := adapter.List(context.Background(), "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "COMM-001") {
		t.Errorf("expected output to contain 'COMM-001', got '%s'", output)
	}
	if !strings.Contains(output, "COMM-002") {
		t.Errorf("expected output to contain 'COMM-002', got '%s'", output)
	}
}

func TestCommissionAdapter_List_Empty(t *testing.T) {
	mock := &mockCommissionService{
		listCommissionsFn: func(ctx context.Context, filters primary.CommissionFilters) ([]*primary.Commission, error) {
			return []*primary.Commission{}, nil
		},
	}
	var buf bytes.Buffer
	adapter := NewCommissionAdapter(mock, &buf)

	err := adapter.List(context.Background(), "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(buf.String(), "No commissions found") {
		t.Errorf("expected 'No commissions found', got '%s'", buf.String())
	}
}

func TestCommissionAdapter_List_FilterByStatus(t *testing.T) {
	var capturedStatus string
	mock := &mockCommissionService{
		listCommissionsFn: func(ctx context.Context, filters primary.CommissionFilters) ([]*primary.Commission, error) {
			capturedStatus = filters.Status
			return []*primary.Commission{}, nil
		},
	}
	var buf bytes.Buffer
	adapter := NewCommissionAdapter(mock, &buf)

	_ = adapter.List(context.Background(), "active")

	if capturedStatus != "active" {
		t.Errorf("expected status filter 'active', got '%s'", capturedStatus)
	}
}

// ============================================================================
// Show Tests
// ============================================================================

func TestCommissionAdapter_Show_Success(t *testing.T) {
	mock := &mockCommissionService{
		getCommissionFn: func(ctx context.Context, commissionID string) (*primary.Commission, error) {
			return &primary.Commission{
				ID:          commissionID,
				Title:       "Test Commission",
				Description: "Detailed description",
				Status:      "active",
				CreatedAt:   "2026-01-19",
			}, nil
		},
	}
	var buf bytes.Buffer
	adapter := NewCommissionAdapter(mock, &buf)

	commission, err := adapter.Show(context.Background(), "COMM-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if commission.ID != "COMM-001" {
		t.Errorf("expected commission ID 'COMM-001', got '%s'", commission.ID)
	}
	output := buf.String()
	if !strings.Contains(output, "Test Commission") {
		t.Errorf("expected output to contain title, got '%s'", output)
	}
	if !strings.Contains(output, "Detailed description") {
		t.Errorf("expected output to contain description, got '%s'", output)
	}
}

func TestCommissionAdapter_Show_NotFound(t *testing.T) {
	mock := &mockCommissionService{
		getCommissionFn: func(ctx context.Context, commissionID string) (*primary.Commission, error) {
			return nil, errors.New("commission not found")
		},
	}
	var buf bytes.Buffer
	adapter := NewCommissionAdapter(mock, &buf)

	_, err := adapter.Show(context.Background(), "COMM-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ============================================================================
// Update Tests
// ============================================================================

func TestCommissionAdapter_Update_TitleOnly(t *testing.T) {
	mock := &mockCommissionService{}
	var buf bytes.Buffer
	adapter := NewCommissionAdapter(mock, &buf)

	err := adapter.Update(context.Background(), "COMM-001", "New Title", "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastUpdateReq.Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", mock.lastUpdateReq.Title)
	}
}

func TestCommissionAdapter_Update_NoFieldsError(t *testing.T) {
	mock := &mockCommissionService{}
	var buf bytes.Buffer
	adapter := NewCommissionAdapter(mock, &buf)

	err := adapter.Update(context.Background(), "COMM-001", "", "")

	if err == nil {
		t.Fatal("expected error when no fields specified, got nil")
	}
	if !strings.Contains(err.Error(), "must specify") {
		t.Errorf("expected validation error, got '%s'", err.Error())
	}
}

// ============================================================================
// Complete Tests
// ============================================================================

func TestCommissionAdapter_Complete_Success(t *testing.T) {
	mock := &mockCommissionService{}
	var buf bytes.Buffer
	adapter := NewCommissionAdapter(mock, &buf)

	err := adapter.Complete(context.Background(), "COMM-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(buf.String(), "marked as complete") {
		t.Errorf("expected completion message, got '%s'", buf.String())
	}
}

func TestCommissionAdapter_Complete_GuardError(t *testing.T) {
	mock := &mockCommissionService{
		completeCommissionFn: func(ctx context.Context, commissionID string) error {
			return errors.New("cannot complete pinned commission")
		},
	}
	var buf bytes.Buffer
	adapter := NewCommissionAdapter(mock, &buf)

	err := adapter.Complete(context.Background(), "COMM-001")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "pinned") {
		t.Errorf("expected guard error, got '%s'", err.Error())
	}
}

// ============================================================================
// Archive Tests
// ============================================================================

func TestCommissionAdapter_Archive_Success(t *testing.T) {
	mock := &mockCommissionService{}
	var buf bytes.Buffer
	adapter := NewCommissionAdapter(mock, &buf)

	err := adapter.Archive(context.Background(), "COMM-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(buf.String(), "archived") {
		t.Errorf("expected archive message, got '%s'", buf.String())
	}
}

// ============================================================================
// Delete Tests
// ============================================================================

func TestCommissionAdapter_Delete_Success(t *testing.T) {
	mock := &mockCommissionService{}
	var buf bytes.Buffer
	adapter := NewCommissionAdapter(mock, &buf)

	err := adapter.Delete(context.Background(), "COMM-001", false)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(buf.String(), "Deleted commission") {
		t.Errorf("expected delete message, got '%s'", buf.String())
	}
}

func TestCommissionAdapter_Delete_WithForce(t *testing.T) {
	mock := &mockCommissionService{}
	var buf bytes.Buffer
	adapter := NewCommissionAdapter(mock, &buf)

	_ = adapter.Delete(context.Background(), "COMM-001", true)

	if !mock.lastDeleteReq.Force {
		t.Error("expected force flag to be true")
	}
}

// ============================================================================
// Pin/Unpin Tests
// ============================================================================

func TestCommissionAdapter_Pin_Success(t *testing.T) {
	mock := &mockCommissionService{}
	var buf bytes.Buffer
	adapter := NewCommissionAdapter(mock, &buf)

	err := adapter.Pin(context.Background(), "COMM-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(buf.String(), "pinned") {
		t.Errorf("expected pin message, got '%s'", buf.String())
	}
}

func TestCommissionAdapter_Unpin_Success(t *testing.T) {
	mock := &mockCommissionService{}
	var buf bytes.Buffer
	adapter := NewCommissionAdapter(mock, &buf)

	err := adapter.Unpin(context.Background(), "COMM-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(buf.String(), "unpinned") {
		t.Errorf("expected unpin message, got '%s'", buf.String())
	}
}
