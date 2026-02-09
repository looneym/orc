package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockPatrolRepository implements secondary.PatrolRepository for testing.
type mockPatrolRepository struct {
	patrols        map[string]*secondary.PatrolRecord
	activeByKennel map[string]*secondary.PatrolRecord
	nextID         int
}

func newMockPatrolRepository() *mockPatrolRepository {
	return &mockPatrolRepository{
		patrols:        make(map[string]*secondary.PatrolRecord),
		activeByKennel: make(map[string]*secondary.PatrolRecord),
		nextID:         1,
	}
}

func (m *mockPatrolRepository) Create(ctx context.Context, patrol *secondary.PatrolRecord) error {
	m.patrols[patrol.ID] = patrol
	if patrol.Status == primary.PatrolStatusActive {
		m.activeByKennel[patrol.KennelID] = patrol
	}
	return nil
}

func (m *mockPatrolRepository) GetByID(ctx context.Context, id string) (*secondary.PatrolRecord, error) {
	if p, ok := m.patrols[id]; ok {
		return p, nil
	}
	return nil, errors.New("not found")
}

func (m *mockPatrolRepository) GetByKennel(ctx context.Context, kennelID string) ([]*secondary.PatrolRecord, error) {
	var result []*secondary.PatrolRecord
	for _, p := range m.patrols {
		if p.KennelID == kennelID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockPatrolRepository) GetActiveByKennel(ctx context.Context, kennelID string) (*secondary.PatrolRecord, error) {
	if p, ok := m.activeByKennel[kennelID]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *mockPatrolRepository) List(ctx context.Context, filters secondary.PatrolFilters) ([]*secondary.PatrolRecord, error) {
	var result []*secondary.PatrolRecord
	for _, p := range m.patrols {
		if filters.KennelID != "" && p.KennelID != filters.KennelID {
			continue
		}
		if filters.Status != "" && p.Status != filters.Status {
			continue
		}
		result = append(result, p)
	}
	return result, nil
}

func (m *mockPatrolRepository) Update(ctx context.Context, patrol *secondary.PatrolRecord) error {
	if _, ok := m.patrols[patrol.ID]; !ok {
		return errors.New("not found")
	}
	m.patrols[patrol.ID] = patrol
	return nil
}

func (m *mockPatrolRepository) UpdateStatus(ctx context.Context, id, status string) error {
	if p, ok := m.patrols[id]; ok {
		p.Status = status
		if status != primary.PatrolStatusActive {
			delete(m.activeByKennel, p.KennelID)
		}
		return nil
	}
	return errors.New("not found")
}

func (m *mockPatrolRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("PATROL-%03d", id), nil
}

func (m *mockPatrolRepository) KennelExists(ctx context.Context, kennelID string) (bool, error) {
	return true, nil
}

// mockWorkbenchRepoForPatrol implements secondary.WorkbenchRepository for testing.
type mockWorkbenchRepoForPatrol struct {
	workbenches map[string]*secondary.WorkbenchRecord
}

func newMockWorkbenchRepoForPatrol() *mockWorkbenchRepoForPatrol {
	return &mockWorkbenchRepoForPatrol{
		workbenches: make(map[string]*secondary.WorkbenchRecord),
	}
}

func (m *mockWorkbenchRepoForPatrol) GetByID(ctx context.Context, id string) (*secondary.WorkbenchRecord, error) {
	if wb, ok := m.workbenches[id]; ok {
		return wb, nil
	}
	return nil, errors.New("not found")
}

func (m *mockWorkbenchRepoForPatrol) Create(ctx context.Context, workbench *secondary.WorkbenchRecord) error {
	return nil
}

func (m *mockWorkbenchRepoForPatrol) Update(ctx context.Context, workbench *secondary.WorkbenchRecord) error {
	return nil
}

func (m *mockWorkbenchRepoForPatrol) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockWorkbenchRepoForPatrol) List(ctx context.Context, workshopID string) ([]*secondary.WorkbenchRecord, error) {
	return nil, nil
}

func (m *mockWorkbenchRepoForPatrol) GetByName(ctx context.Context, workshopID, name string) (*secondary.WorkbenchRecord, error) {
	return nil, nil
}

func (m *mockWorkbenchRepoForPatrol) GetByPath(ctx context.Context, path string) (*secondary.WorkbenchRecord, error) {
	return nil, nil
}

func (m *mockWorkbenchRepoForPatrol) GetByWorkshop(ctx context.Context, workshopID string) ([]*secondary.WorkbenchRecord, error) {
	return nil, nil
}

func (m *mockWorkbenchRepoForPatrol) GetNextID(ctx context.Context) (string, error) {
	return "BENCH-001", nil
}

func (m *mockWorkbenchRepoForPatrol) UpdateStatus(ctx context.Context, id, status string) error {
	return nil
}

func (m *mockWorkbenchRepoForPatrol) UpdateFocus(ctx context.Context, id, focusedID string) error {
	return nil
}

func (m *mockWorkbenchRepoForPatrol) ClearFocus(ctx context.Context, focusedID string) error {
	return nil
}

func (m *mockWorkbenchRepoForPatrol) WorkshopExists(ctx context.Context, workshopID string) (bool, error) {
	return true, nil
}

func (m *mockWorkbenchRepoForPatrol) Rename(ctx context.Context, id, newName string) error {
	return nil
}

func (m *mockWorkbenchRepoForPatrol) UpdatePath(ctx context.Context, id, newPath string) error {
	return nil
}

func (m *mockWorkbenchRepoForPatrol) UpdateFocusedID(ctx context.Context, id, focusedID string) error {
	return nil
}

func (m *mockWorkbenchRepoForPatrol) GetByFocusedID(ctx context.Context, focusedID string) ([]*secondary.WorkbenchRecord, error) {
	return nil, nil
}

// mockTMuxAdapterForPatrol implements secondary.TMuxAdapter for testing.
type mockTMuxAdapterForPatrol struct {
	sessionsByWorkshop map[string]string
}

func newMockTMuxAdapterForPatrol() *mockTMuxAdapterForPatrol {
	return &mockTMuxAdapterForPatrol{
		sessionsByWorkshop: make(map[string]string),
	}
}

func (m *mockTMuxAdapterForPatrol) FindSessionByWorkshopID(ctx context.Context, workshopID string) string {
	if session, ok := m.sessionsByWorkshop[workshopID]; ok {
		return session
	}
	return ""
}

// Stub methods to satisfy interface
func (m *mockTMuxAdapterForPatrol) CreateSession(ctx context.Context, name, workingDir string) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) SessionExists(ctx context.Context, name string) bool { return false }
func (m *mockTMuxAdapterForPatrol) KillSession(ctx context.Context, name string) error  { return nil }
func (m *mockTMuxAdapterForPatrol) GetSessionInfo(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (m *mockTMuxAdapterForPatrol) CreateOrcWindow(ctx context.Context, sessionName string, workingDir string) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) CreateWorkbenchWindow(ctx context.Context, sessionName string, windowIndex int, windowName string, workingDir string) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) CreateWorkbenchWindowShell(ctx context.Context, sessionName string, windowIndex int, windowName string, workingDir string) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) WindowExists(ctx context.Context, sessionName string, windowName string) bool {
	return false
}
func (m *mockTMuxAdapterForPatrol) KillWindow(ctx context.Context, sessionName string, windowName string) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) SendKeys(ctx context.Context, target, keys string) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) GetPaneCount(ctx context.Context, sessionName, windowName string) int {
	return 0
}
func (m *mockTMuxAdapterForPatrol) GetPaneCommand(ctx context.Context, sessionName, windowName string, paneNum int) string {
	return ""
}
func (m *mockTMuxAdapterForPatrol) GetPaneStartPath(ctx context.Context, sessionName, windowName string, paneNum int) string {
	return ""
}
func (m *mockTMuxAdapterForPatrol) GetPaneStartCommand(ctx context.Context, sessionName, windowName string, paneNum int) string {
	return ""
}
func (m *mockTMuxAdapterForPatrol) CapturePaneContent(ctx context.Context, target string, lines int) (string, error) {
	return "", nil
}
func (m *mockTMuxAdapterForPatrol) SplitVertical(ctx context.Context, target, workingDir string) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) SplitHorizontal(ctx context.Context, target, workingDir string) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) JoinPane(ctx context.Context, source, target string, vertical bool, size int) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) NudgeSession(ctx context.Context, target, message string) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) AttachInstructions(sessionName string) string { return "" }
func (m *mockTMuxAdapterForPatrol) SelectWindow(ctx context.Context, sessionName string, index int) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) RenameWindow(ctx context.Context, target, newName string) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) RespawnPane(ctx context.Context, target string, command ...string) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) RenameSession(ctx context.Context, session, newName string) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) ConfigureStatusBar(ctx context.Context, session string, config secondary.StatusBarConfig) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) DisplayPopup(ctx context.Context, session, command string, config secondary.PopupConfig) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) ConfigureSessionBindings(ctx context.Context, session string, bindings []secondary.KeyBinding) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) ConfigureSessionPopupBindings(ctx context.Context, session string, bindings []secondary.PopupKeyBinding) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) GetCurrentSessionName(ctx context.Context) string { return "" }
func (m *mockTMuxAdapterForPatrol) SetEnvironment(ctx context.Context, sessionName, key, value string) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) GetEnvironment(ctx context.Context, sessionName, key string) (string, error) {
	return "", nil
}
func (m *mockTMuxAdapterForPatrol) ListSessions(ctx context.Context) ([]string, error) {
	return nil, nil
}
func (m *mockTMuxAdapterForPatrol) ListWindows(ctx context.Context, sessionName string) ([]string, error) {
	return nil, nil
}
func (m *mockTMuxAdapterForPatrol) GetWindowOption(ctx context.Context, target, option string) string {
	return ""
}
func (m *mockTMuxAdapterForPatrol) SetWindowOption(ctx context.Context, target, option, value string) error {
	return nil
}
func (m *mockTMuxAdapterForPatrol) SetupGoblinPane(ctx context.Context, sessionName, windowName string) error {
	return nil
}

// mockKennelRepoForPatrol implements secondary.KennelRepository for testing.
type mockKennelRepoForPatrol struct {
	kennels            map[string]*secondary.KennelRecord
	kennelsByWorkbench map[string]*secondary.KennelRecord
}

func newMockKennelRepoForPatrol() *mockKennelRepoForPatrol {
	return &mockKennelRepoForPatrol{
		kennels:            make(map[string]*secondary.KennelRecord),
		kennelsByWorkbench: make(map[string]*secondary.KennelRecord),
	}
}

func (m *mockKennelRepoForPatrol) Create(ctx context.Context, kennel *secondary.KennelRecord) error {
	return nil
}

func (m *mockKennelRepoForPatrol) GetByID(ctx context.Context, id string) (*secondary.KennelRecord, error) {
	if k, ok := m.kennels[id]; ok {
		return k, nil
	}
	return nil, errors.New("not found")
}

func (m *mockKennelRepoForPatrol) GetByWorkbench(ctx context.Context, workbenchID string) (*secondary.KennelRecord, error) {
	if k, ok := m.kennelsByWorkbench[workbenchID]; ok {
		return k, nil
	}
	return nil, errors.New("not found")
}

func (m *mockKennelRepoForPatrol) List(ctx context.Context, filters secondary.KennelFilters) ([]*secondary.KennelRecord, error) {
	return nil, nil
}

func (m *mockKennelRepoForPatrol) Update(ctx context.Context, kennel *secondary.KennelRecord) error {
	return nil
}

func (m *mockKennelRepoForPatrol) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockKennelRepoForPatrol) GetNextID(ctx context.Context) (string, error) {
	return "KENNEL-001", nil
}

func (m *mockKennelRepoForPatrol) UpdateStatus(ctx context.Context, id, status string) error {
	return nil
}

func (m *mockKennelRepoForPatrol) WorkbenchExists(ctx context.Context, workbenchID string) (bool, error) {
	return true, nil
}

func (m *mockKennelRepoForPatrol) WorkbenchHasKennel(ctx context.Context, workbenchID string) (bool, error) {
	_, has := m.kennelsByWorkbench[workbenchID]
	return has, nil
}

func newTestPatrolService() (*PatrolServiceImpl, *mockPatrolRepository, *mockKennelRepoForPatrol, *mockWorkbenchRepoForPatrol, *mockTMuxAdapterForPatrol) {
	patrolRepo := newMockPatrolRepository()
	kennelRepo := newMockKennelRepoForPatrol()
	workbenchRepo := newMockWorkbenchRepoForPatrol()
	tmuxAdapter := newMockTMuxAdapterForPatrol()
	service := NewPatrolService(patrolRepo, kennelRepo, workbenchRepo, tmuxAdapter)
	return service, patrolRepo, kennelRepo, workbenchRepo, tmuxAdapter
}

func TestPatrolService_StartPatrol(t *testing.T) {
	service, patrolRepo, kennelRepo, workbenchRepo, tmuxAdapter := newTestPatrolService()
	ctx := context.Background()

	// Setup: workbench and kennel exist
	workbenchRepo.workbenches["BENCH-001"] = &secondary.WorkbenchRecord{
		ID:         "BENCH-001",
		Name:       "test-bench",
		WorkshopID: "WORK-001",
	}
	kennelRepo.kennels["KENNEL-001"] = &secondary.KennelRecord{
		ID:          "KENNEL-001",
		WorkbenchID: "BENCH-001",
	}
	kennelRepo.kennelsByWorkbench["BENCH-001"] = kennelRepo.kennels["KENNEL-001"]
	// Setup tmux session mapping
	tmuxAdapter.sessionsByWorkshop["WORK-001"] = "test-session"

	patrol, err := service.StartPatrol(ctx, "BENCH-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if patrol.ID != "PATROL-001" {
		t.Errorf("expected ID 'PATROL-001', got %q", patrol.ID)
	}
	if patrol.KennelID != "KENNEL-001" {
		t.Errorf("expected KennelID 'KENNEL-001', got %q", patrol.KennelID)
	}
	// Target format: session:window.2 (pane 2 is IMP)
	if patrol.Target != "test-session:test-bench.2" {
		t.Errorf("expected Target 'test-session:test-bench.2', got %q", patrol.Target)
	}
	if patrol.Status != primary.PatrolStatusActive {
		t.Errorf("expected status 'active', got %q", patrol.Status)
	}

	// Verify stored
	if _, ok := patrolRepo.patrols["PATROL-001"]; !ok {
		t.Error("patrol not stored in repository")
	}
}

func TestPatrolService_StartPatrol_WorkbenchNotFound(t *testing.T) {
	service, _, _, _, _ := newTestPatrolService()
	ctx := context.Background()

	_, err := service.StartPatrol(ctx, "BENCH-999")
	if err == nil {
		t.Error("expected error for non-existent workbench")
	}
}

func TestPatrolService_StartPatrol_NoKennel(t *testing.T) {
	service, _, _, workbenchRepo, _ := newTestPatrolService()
	ctx := context.Background()

	// Setup: workbench exists but no kennel
	workbenchRepo.workbenches["BENCH-001"] = &secondary.WorkbenchRecord{
		ID:   "BENCH-001",
		Name: "test-bench",
	}

	_, err := service.StartPatrol(ctx, "BENCH-001")
	if err == nil {
		t.Error("expected error when no kennel")
	}
}

func TestPatrolService_StartPatrol_AlreadyActive(t *testing.T) {
	service, patrolRepo, kennelRepo, workbenchRepo, _ := newTestPatrolService()
	ctx := context.Background()

	// Setup: workbench and kennel exist, with active patrol
	workbenchRepo.workbenches["BENCH-001"] = &secondary.WorkbenchRecord{
		ID:   "BENCH-001",
		Name: "test-bench",
	}
	kennelRepo.kennels["KENNEL-001"] = &secondary.KennelRecord{
		ID:          "KENNEL-001",
		WorkbenchID: "BENCH-001",
	}
	kennelRepo.kennelsByWorkbench["BENCH-001"] = kennelRepo.kennels["KENNEL-001"]

	// Add existing active patrol
	patrolRepo.patrols["PATROL-001"] = &secondary.PatrolRecord{
		ID:       "PATROL-001",
		KennelID: "KENNEL-001",
		Status:   primary.PatrolStatusActive,
	}
	patrolRepo.activeByKennel["KENNEL-001"] = patrolRepo.patrols["PATROL-001"]

	_, err := service.StartPatrol(ctx, "BENCH-001")
	if err == nil {
		t.Error("expected error when patrol already active")
	}
}

func TestPatrolService_EndPatrol(t *testing.T) {
	service, patrolRepo, _, _, _ := newTestPatrolService()
	ctx := context.Background()

	// Setup: active patrol
	patrolRepo.patrols["PATROL-001"] = &secondary.PatrolRecord{
		ID:       "PATROL-001",
		KennelID: "KENNEL-001",
		Status:   primary.PatrolStatusActive,
	}
	patrolRepo.activeByKennel["KENNEL-001"] = patrolRepo.patrols["PATROL-001"]

	err := service.EndPatrol(ctx, "PATROL-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if patrolRepo.patrols["PATROL-001"].Status != primary.PatrolStatusCompleted {
		t.Errorf("expected status 'completed', got %q", patrolRepo.patrols["PATROL-001"].Status)
	}
}

func TestPatrolService_EndPatrol_NotFound(t *testing.T) {
	service, _, _, _, _ := newTestPatrolService()
	ctx := context.Background()

	err := service.EndPatrol(ctx, "PATROL-999")
	if err == nil {
		t.Error("expected error for non-existent patrol")
	}
}

func TestPatrolService_EndPatrol_NotActive(t *testing.T) {
	service, patrolRepo, _, _, _ := newTestPatrolService()
	ctx := context.Background()

	// Setup: already completed patrol
	patrolRepo.patrols["PATROL-001"] = &secondary.PatrolRecord{
		ID:       "PATROL-001",
		KennelID: "KENNEL-001",
		Status:   primary.PatrolStatusCompleted,
	}

	err := service.EndPatrol(ctx, "PATROL-001")
	if err == nil {
		t.Error("expected error when patrol not active")
	}
}

func TestPatrolService_GetPatrol(t *testing.T) {
	service, patrolRepo, _, _, _ := newTestPatrolService()
	ctx := context.Background()

	patrolRepo.patrols["PATROL-001"] = &secondary.PatrolRecord{
		ID:       "PATROL-001",
		KennelID: "KENNEL-001",
		Target:   "test-session:test.2",
		Status:   "active",
	}

	patrol, err := service.GetPatrol(ctx, "PATROL-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if patrol.Target != "test-session:test.2" {
		t.Errorf("expected Target 'test-session:test.2', got %q", patrol.Target)
	}
}

func TestPatrolService_ListPatrols(t *testing.T) {
	service, patrolRepo, _, _, _ := newTestPatrolService()
	ctx := context.Background()

	patrolRepo.patrols["PATROL-001"] = &secondary.PatrolRecord{ID: "PATROL-001", KennelID: "KENNEL-001", Status: "active"}
	patrolRepo.patrols["PATROL-002"] = &secondary.PatrolRecord{ID: "PATROL-002", KennelID: "KENNEL-001", Status: "completed"}

	patrols, err := service.ListPatrols(ctx, primary.PatrolFilters{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(patrols) != 2 {
		t.Errorf("expected 2 patrols, got %d", len(patrols))
	}
}

func TestPatrolService_ListPatrols_FilterByStatus(t *testing.T) {
	service, patrolRepo, _, _, _ := newTestPatrolService()
	ctx := context.Background()

	patrolRepo.patrols["PATROL-001"] = &secondary.PatrolRecord{ID: "PATROL-001", KennelID: "KENNEL-001", Status: "active"}
	patrolRepo.patrols["PATROL-002"] = &secondary.PatrolRecord{ID: "PATROL-002", KennelID: "KENNEL-001", Status: "completed"}

	patrols, err := service.ListPatrols(ctx, primary.PatrolFilters{Status: "active"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(patrols) != 1 {
		t.Errorf("expected 1 patrol, got %d", len(patrols))
	}
	if patrols[0].ID != "PATROL-001" {
		t.Errorf("expected PATROL-001, got %q", patrols[0].ID)
	}
}

func TestPatrolService_GetActivePatrolForKennel(t *testing.T) {
	service, patrolRepo, _, _, _ := newTestPatrolService()
	ctx := context.Background()

	patrolRepo.patrols["PATROL-001"] = &secondary.PatrolRecord{
		ID:       "PATROL-001",
		KennelID: "KENNEL-001",
		Status:   "active",
	}
	patrolRepo.activeByKennel["KENNEL-001"] = patrolRepo.patrols["PATROL-001"]

	patrol, err := service.GetActivePatrolForKennel(ctx, "KENNEL-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if patrol.ID != "PATROL-001" {
		t.Errorf("expected PATROL-001, got %q", patrol.ID)
	}
}

func TestPatrolService_GetActivePatrolForKennel_NoActive(t *testing.T) {
	service, _, _, _, _ := newTestPatrolService()
	ctx := context.Background()

	patrol, err := service.GetActivePatrolForKennel(ctx, "KENNEL-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if patrol != nil {
		t.Error("expected nil for no active patrol")
	}
}
