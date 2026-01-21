package app

import (
	"context"
	"errors"
	"testing"

	"github.com/example/orc/internal/core/effects"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// ============================================================================
// Mock Implementations
// ============================================================================

// mockCommissionRepository implements secondary.CommissionRepository for testing.
type mockCommissionRepository struct {
	commissions   map[string]*secondary.CommissionRecord
	shipmentCount map[string]int
	createErr     error
	getErr        error
	updateErr     error
	deleteErr     error
	listErr       error
}

func newMockCommissionRepository() *mockCommissionRepository {
	return &mockCommissionRepository{
		commissions:   make(map[string]*secondary.CommissionRecord),
		shipmentCount: make(map[string]int),
	}
}

func (m *mockCommissionRepository) Create(ctx context.Context, commission *secondary.CommissionRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	commission.ID = "COMM-001"
	m.commissions[commission.ID] = commission
	return nil
}

func (m *mockCommissionRepository) GetByID(ctx context.Context, id string) (*secondary.CommissionRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if commission, ok := m.commissions[id]; ok {
		return commission, nil
	}
	return nil, errors.New("commission not found")
}

func (m *mockCommissionRepository) Update(ctx context.Context, commission *secondary.CommissionRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.commissions[commission.ID] = commission
	return nil
}

func (m *mockCommissionRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.commissions, id)
	return nil
}

func (m *mockCommissionRepository) List(ctx context.Context, filters secondary.CommissionFilters) ([]*secondary.CommissionRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.CommissionRecord
	for _, commission := range m.commissions {
		if filters.Status == "" || commission.Status == filters.Status {
			result = append(result, commission)
		}
	}
	return result, nil
}

func (m *mockCommissionRepository) GetNextID(ctx context.Context) (string, error) {
	return "COMM-001", nil
}

func (m *mockCommissionRepository) CountShipments(ctx context.Context, commissionID string) (int, error) {
	return m.shipmentCount[commissionID], nil
}

func (m *mockCommissionRepository) Pin(ctx context.Context, id string) error {
	if commission, ok := m.commissions[id]; ok {
		commission.Pinned = true
	}
	return nil
}

func (m *mockCommissionRepository) Unpin(ctx context.Context, id string) error {
	if commission, ok := m.commissions[id]; ok {
		commission.Pinned = false
	}
	return nil
}

// mockGroveRepository implements secondary.GroveRepository for testing.
type mockGroveRepository struct {
	groves map[string][]*secondary.GroveRecord
}

func newMockGroveRepository() *mockGroveRepository {
	return &mockGroveRepository{
		groves: make(map[string][]*secondary.GroveRecord),
	}
}

func (m *mockGroveRepository) Create(ctx context.Context, grove *secondary.GroveRecord) error {
	grove.ID = "GROVE-001"
	m.groves[grove.CommissionID] = append(m.groves[grove.CommissionID], grove)
	return nil
}

func (m *mockGroveRepository) GetByID(ctx context.Context, id string) (*secondary.GroveRecord, error) {
	for _, groves := range m.groves {
		for _, g := range groves {
			if g.ID == id {
				return g, nil
			}
		}
	}
	return nil, errors.New("grove not found")
}

func (m *mockGroveRepository) GetByCommission(ctx context.Context, commissionID string) ([]*secondary.GroveRecord, error) {
	return m.groves[commissionID], nil
}

func (m *mockGroveRepository) Update(ctx context.Context, grove *secondary.GroveRecord) error {
	return nil
}

func (m *mockGroveRepository) Delete(ctx context.Context, id string) error {
	for commissionID, groves := range m.groves {
		for i, g := range groves {
			if g.ID == id {
				m.groves[commissionID] = append(groves[:i], groves[i+1:]...)
				return nil
			}
		}
	}
	return nil
}

func (m *mockGroveRepository) GetNextID(ctx context.Context) (string, error) {
	return "GROVE-001", nil
}

func (m *mockGroveRepository) GetByPath(ctx context.Context, path string) (*secondary.GroveRecord, error) {
	for _, groves := range m.groves {
		for _, g := range groves {
			if g.WorktreePath == path {
				return g, nil
			}
		}
	}
	return nil, errors.New("grove not found")
}

func (m *mockGroveRepository) List(ctx context.Context, commissionID string) ([]*secondary.GroveRecord, error) {
	if commissionID == "" {
		var all []*secondary.GroveRecord
		for _, groves := range m.groves {
			all = append(all, groves...)
		}
		return all, nil
	}
	return m.groves[commissionID], nil
}

func (m *mockGroveRepository) Rename(ctx context.Context, id, newName string) error {
	return nil
}

func (m *mockGroveRepository) UpdatePath(ctx context.Context, id, newPath string) error {
	return nil
}

// mockAgentProvider implements secondary.AgentIdentityProvider for testing.
type mockAgentProvider struct {
	identity *secondary.AgentIdentity
	err      error
}

func newMockAgentProvider(agentType secondary.AgentType) *mockAgentProvider {
	return &mockAgentProvider{
		identity: &secondary.AgentIdentity{
			Type:         agentType,
			ID:           "001",
			FullID:       string(agentType) + "-001",
			CommissionID: "",
		},
	}
}

func (m *mockAgentProvider) GetCurrentIdentity(ctx context.Context) (*secondary.AgentIdentity, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.identity, nil
}

// mockEffectExecutor implements EffectExecutor for testing.
type mockEffectExecutor struct {
	executedEffects []effects.Effect
	executeErr      error
}

func newMockEffectExecutor() *mockEffectExecutor {
	return &mockEffectExecutor{
		executedEffects: []effects.Effect{},
	}
}

func (m *mockEffectExecutor) Execute(ctx context.Context, effs []effects.Effect) error {
	if m.executeErr != nil {
		return m.executeErr
	}
	m.executedEffects = append(m.executedEffects, effs...)
	return nil
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestService(agentType secondary.AgentType) (*CommissionServiceImpl, *mockCommissionRepository, *mockGroveRepository, *mockEffectExecutor) {
	commissionRepo := newMockCommissionRepository()
	groveRepo := newMockGroveRepository()
	agentProvider := newMockAgentProvider(agentType)
	executor := newMockEffectExecutor()

	service := NewCommissionService(commissionRepo, groveRepo, agentProvider, executor)
	return service, commissionRepo, groveRepo, executor
}

// ============================================================================
// CreateCommission Tests
// ============================================================================

func TestCreateCommission_ORCCanCreate(t *testing.T) {
	service, _, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	resp, err := service.CreateCommission(ctx, primary.CreateCommissionRequest{
		Title:       "Test Commission",
		Description: "A test commission",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.CommissionID == "" {
		t.Error("expected commission ID to be set")
	}
	if resp.Commission.Title != "Test Commission" {
		t.Errorf("expected title 'Test Commission', got '%s'", resp.Commission.Title)
	}
}

func TestCreateCommission_IMPCannotCreate(t *testing.T) {
	service, _, _, _ := newTestService(secondary.AgentTypeIMP)
	ctx := context.Background()

	_, err := service.CreateCommission(ctx, primary.CreateCommissionRequest{
		Title:       "Test Commission",
		Description: "A test commission",
	})

	if err == nil {
		t.Fatal("expected error for IMP creating commission, got nil")
	}
}

// Note: Only ORC and IMP agent types are defined. ORC can create, IMP cannot.
// Additional agent types could be added in the future.

// ============================================================================
// CompleteCommission Tests
// ============================================================================

func TestCompleteCommission_UnpinnedAllowed(t *testing.T) {
	service, commissionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup: create an unpinned commission
	commissionRepo.commissions["COMM-001"] = &secondary.CommissionRecord{
		ID:     "COMM-001",
		Title:  "Test Commission",
		Status: "active",
		Pinned: false,
	}

	err := service.CompleteCommission(ctx, "COMM-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if commissionRepo.commissions["COMM-001"].Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", commissionRepo.commissions["COMM-001"].Status)
	}
}

func TestCompleteCommission_PinnedBlocked(t *testing.T) {
	service, commissionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup: create a pinned commission
	commissionRepo.commissions["COMM-001"] = &secondary.CommissionRecord{
		ID:     "COMM-001",
		Title:  "Pinned Commission",
		Status: "active",
		Pinned: true,
	}

	err := service.CompleteCommission(ctx, "COMM-001")

	if err == nil {
		t.Fatal("expected error for completing pinned commission, got nil")
	}
}

func TestCompleteCommission_NotFound(t *testing.T) {
	service, _, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	err := service.CompleteCommission(ctx, "COMM-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent commission, got nil")
	}
}

// ============================================================================
// ArchiveCommission Tests
// ============================================================================

func TestArchiveCommission_UnpinnedAllowed(t *testing.T) {
	service, commissionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup: create an unpinned commission
	commissionRepo.commissions["COMM-001"] = &secondary.CommissionRecord{
		ID:     "COMM-001",
		Title:  "Test Commission",
		Status: "complete",
		Pinned: false,
	}

	err := service.ArchiveCommission(ctx, "COMM-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if commissionRepo.commissions["COMM-001"].Status != "archived" {
		t.Errorf("expected status 'archived', got '%s'", commissionRepo.commissions["COMM-001"].Status)
	}
}

func TestArchiveCommission_PinnedBlocked(t *testing.T) {
	service, commissionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup: create a pinned commission
	commissionRepo.commissions["COMM-001"] = &secondary.CommissionRecord{
		ID:     "COMM-001",
		Title:  "Pinned Commission",
		Status: "complete",
		Pinned: true,
	}

	err := service.ArchiveCommission(ctx, "COMM-001")

	if err == nil {
		t.Fatal("expected error for archiving pinned commission, got nil")
	}
}

// ============================================================================
// DeleteCommission Tests
// ============================================================================

func TestDeleteCommission_NoDependents(t *testing.T) {
	service, commissionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup: create a commission with no dependents
	commissionRepo.commissions["COMM-001"] = &secondary.CommissionRecord{
		ID:     "COMM-001",
		Title:  "Empty Commission",
		Status: "active",
	}
	commissionRepo.shipmentCount["COMM-001"] = 0

	err := service.DeleteCommission(ctx, primary.DeleteCommissionRequest{
		CommissionID: "COMM-001",
		Force:        false,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := commissionRepo.commissions["COMM-001"]; exists {
		t.Error("expected commission to be deleted")
	}
}

func TestDeleteCommission_HasDependentsNoForce(t *testing.T) {
	service, commissionRepo, groveRepo, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup: create a commission with shipments
	commissionRepo.commissions["COMM-001"] = &secondary.CommissionRecord{
		ID:     "COMM-001",
		Title:  "Commission with Shipments",
		Status: "active",
	}
	commissionRepo.shipmentCount["COMM-001"] = 3
	groveRepo.groves["COMM-001"] = []*secondary.GroveRecord{
		{ID: "GROVE-001", CommissionID: "COMM-001", Name: "main-grove"},
	}

	err := service.DeleteCommission(ctx, primary.DeleteCommissionRequest{
		CommissionID: "COMM-001",
		Force:        false,
	})

	if err == nil {
		t.Fatal("expected error for deleting commission with dependents without force, got nil")
	}
	if _, exists := commissionRepo.commissions["COMM-001"]; !exists {
		t.Error("commission should not have been deleted")
	}
}

func TestDeleteCommission_HasDependentsWithForce(t *testing.T) {
	service, commissionRepo, groveRepo, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup: create a commission with shipments
	commissionRepo.commissions["COMM-001"] = &secondary.CommissionRecord{
		ID:     "COMM-001",
		Title:  "Commission with Shipments",
		Status: "active",
	}
	commissionRepo.shipmentCount["COMM-001"] = 3
	groveRepo.groves["COMM-001"] = []*secondary.GroveRecord{
		{ID: "GROVE-001", CommissionID: "COMM-001", Name: "main-grove"},
	}

	err := service.DeleteCommission(ctx, primary.DeleteCommissionRequest{
		CommissionID: "COMM-001",
		Force:        true,
	})

	if err != nil {
		t.Fatalf("expected no error with force flag, got %v", err)
	}
	if _, exists := commissionRepo.commissions["COMM-001"]; exists {
		t.Error("expected commission to be deleted with force")
	}
}

// ============================================================================
// StartCommission Tests
// ============================================================================

func TestStartCommission_ORCCanStart(t *testing.T) {
	service, commissionRepo, groveRepo, executor := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup: create a commission with groves
	commissionRepo.commissions["COMM-001"] = &secondary.CommissionRecord{
		ID:     "COMM-001",
		Title:  "Test Commission",
		Status: "active",
	}
	groveRepo.groves["COMM-001"] = []*secondary.GroveRecord{
		{ID: "GROVE-001", CommissionID: "COMM-001", Name: "main-grove", WorktreePath: "/tmp/worktree"},
	}

	resp, err := service.StartCommission(ctx, primary.StartCommissionRequest{
		CommissionID: "COMM-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Commission.ID != "COMM-001" {
		t.Errorf("expected commission ID 'COMM-001', got '%s'", resp.Commission.ID)
	}
	// Verify effects were generated
	if len(executor.executedEffects) == 0 {
		t.Error("expected effects to be executed")
	}
}

func TestStartCommission_IMPCannotStart(t *testing.T) {
	service, commissionRepo, _, _ := newTestService(secondary.AgentTypeIMP)
	ctx := context.Background()

	// Setup: create a commission
	commissionRepo.commissions["COMM-001"] = &secondary.CommissionRecord{
		ID:     "COMM-001",
		Title:  "Test Commission",
		Status: "active",
	}

	_, err := service.StartCommission(ctx, primary.StartCommissionRequest{
		CommissionID: "COMM-001",
	})

	if err == nil {
		t.Fatal("expected error for IMP starting commission, got nil")
	}
}

func TestStartCommission_GeneratesTMuxEffects(t *testing.T) {
	service, commissionRepo, groveRepo, executor := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup
	commissionRepo.commissions["COMM-001"] = &secondary.CommissionRecord{
		ID:     "COMM-001",
		Title:  "Test Commission",
		Status: "active",
	}
	groveRepo.groves["COMM-001"] = []*secondary.GroveRecord{
		{ID: "GROVE-001", CommissionID: "COMM-001", Name: "grove1", WorktreePath: "/path/to/worktree"},
	}

	_, err := service.StartCommission(ctx, primary.StartCommissionRequest{
		CommissionID: "COMM-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check that TMux effects were generated
	hasTMuxEffect := false
	for _, eff := range executor.executedEffects {
		if _, ok := eff.(effects.TMuxEffect); ok {
			hasTMuxEffect = true
			break
		}
	}
	if !hasTMuxEffect {
		t.Error("expected TMux effects to be generated for start commission")
	}
}

// ============================================================================
// LaunchCommission Tests
// ============================================================================

func TestLaunchCommission_ORCCanLaunch(t *testing.T) {
	service, _, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	resp, err := service.LaunchCommission(ctx, primary.LaunchCommissionRequest{
		Title:       "New Commission",
		Description: "A launched commission",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.CommissionID == "" {
		t.Error("expected commission ID to be set")
	}
}

func TestLaunchCommission_IMPCannotLaunch(t *testing.T) {
	service, _, _, _ := newTestService(secondary.AgentTypeIMP)
	ctx := context.Background()

	_, err := service.LaunchCommission(ctx, primary.LaunchCommissionRequest{
		Title:       "New Commission",
		Description: "A launched commission",
	})

	if err == nil {
		t.Fatal("expected error for IMP launching commission, got nil")
	}
}

// ============================================================================
// GetCommission / ListCommissions Tests
// ============================================================================

func TestGetCommission_Found(t *testing.T) {
	service, commissionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	commissionRepo.commissions["COMM-001"] = &secondary.CommissionRecord{
		ID:     "COMM-001",
		Title:  "Test Commission",
		Status: "active",
	}

	commission, err := service.GetCommission(ctx, "COMM-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if commission.Title != "Test Commission" {
		t.Errorf("expected title 'Test Commission', got '%s'", commission.Title)
	}
}

func TestGetCommission_NotFound(t *testing.T) {
	service, _, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	_, err := service.GetCommission(ctx, "COMM-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent commission, got nil")
	}
}

func TestListCommissions_FilterByStatus(t *testing.T) {
	service, commissionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	commissionRepo.commissions["COMM-001"] = &secondary.CommissionRecord{
		ID:     "COMM-001",
		Title:  "Active Commission",
		Status: "active",
	}
	commissionRepo.commissions["COMM-002"] = &secondary.CommissionRecord{
		ID:     "COMM-002",
		Title:  "Complete Commission",
		Status: "complete",
	}

	commissions, err := service.ListCommissions(ctx, primary.CommissionFilters{Status: "active"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(commissions) != 1 {
		t.Errorf("expected 1 active commission, got %d", len(commissions))
	}
}

// ============================================================================
// Pin/Unpin Tests
// ============================================================================

func TestPinCommission(t *testing.T) {
	service, commissionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	commissionRepo.commissions["COMM-001"] = &secondary.CommissionRecord{
		ID:     "COMM-001",
		Title:  "Test Commission",
		Status: "active",
		Pinned: false,
	}

	err := service.PinCommission(ctx, "COMM-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !commissionRepo.commissions["COMM-001"].Pinned {
		t.Error("expected commission to be pinned")
	}
}

func TestUnpinCommission(t *testing.T) {
	service, commissionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	commissionRepo.commissions["COMM-001"] = &secondary.CommissionRecord{
		ID:     "COMM-001",
		Title:  "Test Commission",
		Status: "active",
		Pinned: true,
	}

	err := service.UnpinCommission(ctx, "COMM-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if commissionRepo.commissions["COMM-001"].Pinned {
		t.Error("expected commission to be unpinned")
	}
}

// ============================================================================
// UpdateCommission Tests
// ============================================================================

func TestUpdateCommission_Title(t *testing.T) {
	service, commissionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	commissionRepo.commissions["COMM-001"] = &secondary.CommissionRecord{
		ID:          "COMM-001",
		Title:       "Old Title",
		Description: "Original description",
		Status:      "active",
	}

	err := service.UpdateCommission(ctx, primary.UpdateCommissionRequest{
		CommissionID: "COMM-001",
		Title:        "New Title",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if commissionRepo.commissions["COMM-001"].Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", commissionRepo.commissions["COMM-001"].Title)
	}
}

func TestUpdateCommission_Description(t *testing.T) {
	service, commissionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	commissionRepo.commissions["COMM-001"] = &secondary.CommissionRecord{
		ID:          "COMM-001",
		Title:       "Test Commission",
		Description: "Old description",
		Status:      "active",
	}

	err := service.UpdateCommission(ctx, primary.UpdateCommissionRequest{
		CommissionID: "COMM-001",
		Description:  "New description",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if commissionRepo.commissions["COMM-001"].Description != "New description" {
		t.Errorf("expected description 'New description', got '%s'", commissionRepo.commissions["COMM-001"].Description)
	}
}
