package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockApprovalRepository implements secondary.ApprovalRepository for testing.
type mockApprovalRepository struct {
	approvals       map[string]*secondary.ApprovalRecord
	approvalsByPlan map[string]*secondary.ApprovalRecord
	planExists      map[string]bool
	taskExists      map[string]bool
	planHasApproval map[string]bool
	nextID          int
}

func newMockApprovalRepository() *mockApprovalRepository {
	return &mockApprovalRepository{
		approvals:       make(map[string]*secondary.ApprovalRecord),
		approvalsByPlan: make(map[string]*secondary.ApprovalRecord),
		planExists:      make(map[string]bool),
		taskExists:      make(map[string]bool),
		planHasApproval: make(map[string]bool),
		nextID:          1,
	}
}

func (m *mockApprovalRepository) Create(ctx context.Context, approval *secondary.ApprovalRecord) error {
	m.approvals[approval.ID] = approval
	m.approvalsByPlan[approval.PlanID] = approval
	m.planHasApproval[approval.PlanID] = true
	return nil
}

func (m *mockApprovalRepository) GetByID(ctx context.Context, id string) (*secondary.ApprovalRecord, error) {
	if a, ok := m.approvals[id]; ok {
		return a, nil
	}
	return nil, errors.New("not found")
}

func (m *mockApprovalRepository) GetByPlan(ctx context.Context, planID string) (*secondary.ApprovalRecord, error) {
	if a, ok := m.approvalsByPlan[planID]; ok {
		return a, nil
	}
	return nil, errors.New("not found")
}

func (m *mockApprovalRepository) List(ctx context.Context, filters secondary.ApprovalFilters) ([]*secondary.ApprovalRecord, error) {
	var result []*secondary.ApprovalRecord
	for _, a := range m.approvals {
		if filters.TaskID != "" && a.TaskID != filters.TaskID {
			continue
		}
		if filters.Outcome != "" && a.Outcome != filters.Outcome {
			continue
		}
		result = append(result, a)
	}
	return result, nil
}

func (m *mockApprovalRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.approvals[id]; !ok {
		return errors.New("not found")
	}
	a := m.approvals[id]
	delete(m.approvalsByPlan, a.PlanID)
	delete(m.approvals, id)
	m.planHasApproval[a.PlanID] = false
	return nil
}

func (m *mockApprovalRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("APPR-%03d", id), nil
}

func (m *mockApprovalRepository) PlanExists(ctx context.Context, planID string) (bool, error) {
	return m.planExists[planID], nil
}

func (m *mockApprovalRepository) TaskExists(ctx context.Context, taskID string) (bool, error) {
	return m.taskExists[taskID], nil
}

func (m *mockApprovalRepository) PlanHasApproval(ctx context.Context, planID string) (bool, error) {
	return m.planHasApproval[planID], nil
}

func newTestApprovalService() (*ApprovalServiceImpl, *mockApprovalRepository) {
	repo := newMockApprovalRepository()
	service := NewApprovalService(repo)
	return service, repo
}

func TestApprovalService_GetApproval(t *testing.T) {
	service, repo := newTestApprovalService()
	ctx := context.Background()

	repo.approvals["APPR-001"] = &secondary.ApprovalRecord{
		ID:        "APPR-001",
		PlanID:    "PLAN-001",
		TaskID:    "TASK-001",
		Mechanism: "subagent",
		Outcome:   "approved",
	}

	approval, err := service.GetApproval(ctx, "APPR-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if approval.PlanID != "PLAN-001" {
		t.Errorf("expected planID 'PLAN-001', got %q", approval.PlanID)
	}
}

func TestApprovalService_GetApproval_NotFound(t *testing.T) {
	service, _ := newTestApprovalService()
	ctx := context.Background()

	_, err := service.GetApproval(ctx, "APPR-999")
	if err == nil {
		t.Error("expected error for non-existent approval")
	}
}

func TestApprovalService_GetApprovalByPlan(t *testing.T) {
	service, repo := newTestApprovalService()
	ctx := context.Background()

	repo.approvals["APPR-001"] = &secondary.ApprovalRecord{
		ID:        "APPR-001",
		PlanID:    "PLAN-001",
		TaskID:    "TASK-001",
		Mechanism: "subagent",
		Outcome:   "approved",
	}
	repo.approvalsByPlan["PLAN-001"] = repo.approvals["APPR-001"]

	approval, err := service.GetApprovalByPlan(ctx, "PLAN-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if approval.ID != "APPR-001" {
		t.Errorf("expected ID 'APPR-001', got %q", approval.ID)
	}
}

func TestApprovalService_ListApprovals(t *testing.T) {
	service, repo := newTestApprovalService()
	ctx := context.Background()

	repo.approvals["APPR-001"] = &secondary.ApprovalRecord{ID: "APPR-001", PlanID: "PLAN-001", TaskID: "TASK-001", Mechanism: "subagent", Outcome: "approved"}
	repo.approvals["APPR-002"] = &secondary.ApprovalRecord{ID: "APPR-002", PlanID: "PLAN-002", TaskID: "TASK-002", Mechanism: "manual", Outcome: "escalated"}

	approvals, err := service.ListApprovals(ctx, primary.ApprovalFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(approvals) != 2 {
		t.Errorf("expected 2 approvals, got %d", len(approvals))
	}
}
