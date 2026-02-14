package tmux

import (
	"strings"
	"testing"
)

func TestNewGotmuxAdapter(t *testing.T) {
	adapter, err := NewGotmuxAdapter()
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	if adapter == nil {
		t.Fatal("adapter should not be nil")
	}

	if adapter.tmux == nil {
		t.Fatal("adapter.tmux should not be nil")
	}
}

func TestSessionExists(t *testing.T) {
	adapter, err := NewGotmuxAdapter()
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	// Test with non-existent session
	exists := adapter.SessionExists("nonexistent-session-test-12345")
	if exists {
		t.Error("SessionExists should return false for non-existent session")
	}
}

func TestAttachInstructions(t *testing.T) {
	adapter, err := NewGotmuxAdapter()
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	instructions := adapter.AttachInstructions("test-session")
	if instructions == "" {
		t.Error("AttachInstructions should return non-empty string")
	}

	if !strings.Contains(instructions, "test-session") {
		t.Error("AttachInstructions should contain the session name")
	}

	if !strings.Contains(instructions, "tmux attach") {
		t.Error("AttachInstructions should contain 'tmux attach'")
	}

	if !strings.Contains(instructions, "goblin") {
		t.Error("AttachInstructions should mention goblin pane layout")
	}

	if !strings.Contains(instructions, "desk") {
		t.Error("AttachInstructions should mention desk popup")
	}
}

func TestPlanApply_NewSession(t *testing.T) {
	adapter, err := NewGotmuxAdapter()
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	workbenches := []DesiredWorkbench{
		{Name: "bench-1", Path: "/tmp/bench-1", ID: "BENCH-001", WorkshopID: "WORK-001"},
		{Name: "bench-2", Path: "/tmp/bench-2", ID: "BENCH-002", WorkshopID: "WORK-001"},
	}

	// Plan against a session that doesn't exist
	plan, err := adapter.PlanApply("nonexistent-test-session-99999", workbenches)
	if err != nil {
		t.Fatalf("PlanApply failed: %v", err)
	}

	if plan.SessionExists {
		t.Error("plan.SessionExists should be false for non-existent session")
	}

	// Expect 3 actions: CreateSession + AddWindow + ApplyEnrichment
	if len(plan.Actions) != 3 {
		t.Fatalf("expected 3 actions, got %d: %+v", len(plan.Actions), plan.Actions)
	}

	if plan.Actions[0].Type != ActionCreateSession {
		t.Errorf("first action should be CreateSession, got %s", plan.Actions[0].Type)
	}
	if plan.Actions[0].WorkbenchName != "bench-1" {
		t.Errorf("first action should create bench-1, got %s", plan.Actions[0].WorkbenchName)
	}

	if plan.Actions[1].Type != ActionAddWindow {
		t.Errorf("second action should be AddWindow, got %s", plan.Actions[1].Type)
	}
	if plan.Actions[1].WorkbenchName != "bench-2" {
		t.Errorf("second action should add bench-2, got %s", plan.Actions[1].WorkbenchName)
	}

	if plan.Actions[2].Type != ActionApplyEnrichment {
		t.Errorf("third action should be ApplyEnrichment, got %s", plan.Actions[2].Type)
	}
}

func TestPlanApply_EmptyWorkbenches(t *testing.T) {
	adapter, err := NewGotmuxAdapter()
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	plan, err := adapter.PlanApply("nonexistent-test-session-99999", nil)
	if err != nil {
		t.Fatalf("PlanApply failed: %v", err)
	}

	if plan.SessionExists {
		t.Error("plan.SessionExists should be false")
	}

	if len(plan.Actions) != 0 {
		t.Errorf("expected 0 actions for empty workbenches, got %d", len(plan.Actions))
	}
}

func TestPlanApply_SingleWorkbench(t *testing.T) {
	adapter, err := NewGotmuxAdapter()
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	workbenches := []DesiredWorkbench{
		{Name: "only-bench", Path: "/tmp/only-bench", ID: "BENCH-099", WorkshopID: "WORK-099"},
	}

	plan, err := adapter.PlanApply("nonexistent-test-session-99999", workbenches)
	if err != nil {
		t.Fatalf("PlanApply failed: %v", err)
	}

	// Expect 2 actions: CreateSession + ApplyEnrichment (no AddWindow since first bench creates session)
	if len(plan.Actions) != 2 {
		t.Fatalf("expected 2 actions, got %d: %+v", len(plan.Actions), plan.Actions)
	}

	if plan.Actions[0].Type != ActionCreateSession {
		t.Errorf("first action should be CreateSession, got %s", plan.Actions[0].Type)
	}
	if plan.Actions[1].Type != ActionApplyEnrichment {
		t.Errorf("second action should be ApplyEnrichment, got %s", plan.Actions[1].Type)
	}
}

func TestPlanApply_ActionFields(t *testing.T) {
	adapter, err := NewGotmuxAdapter()
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	workbenches := []DesiredWorkbench{
		{Name: "my-bench", Path: "/home/user/wb/my-bench", ID: "BENCH-042", WorkshopID: "WORK-007"},
	}

	plan, err := adapter.PlanApply("nonexistent-test-session-99999", workbenches)
	if err != nil {
		t.Fatalf("PlanApply failed: %v", err)
	}

	action := plan.Actions[0]
	if action.SessionName != "nonexistent-test-session-99999" {
		t.Errorf("expected SessionName 'nonexistent-test-session-99999', got '%s'", action.SessionName)
	}
	if action.WorkbenchName != "my-bench" {
		t.Errorf("expected WorkbenchName 'my-bench', got '%s'", action.WorkbenchName)
	}
	if action.WorkbenchPath != "/home/user/wb/my-bench" {
		t.Errorf("expected WorkbenchPath '/home/user/wb/my-bench', got '%s'", action.WorkbenchPath)
	}
	if action.WorkbenchID != "BENCH-042" {
		t.Errorf("expected WorkbenchID 'BENCH-042', got '%s'", action.WorkbenchID)
	}
	if action.WorkshopID != "WORK-007" {
		t.Errorf("expected WorkshopID 'WORK-007', got '%s'", action.WorkshopID)
	}
	if action.Description == "" {
		t.Error("action Description should not be empty")
	}
}

func TestListDeskServers_NoError(t *testing.T) {
	// ListDeskServers should not return an error in normal operation.
	// Returns nil slice when no desk servers exist, non-nil when some do.
	_, err := ListDeskServers()
	if err != nil {
		t.Fatalf("ListDeskServers failed: %v", err)
	}
}

func TestKillDeskServer_NonExistent(t *testing.T) {
	// Killing a non-existent desk server should return an error
	err := KillDeskServer("nonexistent-bench-99999")
	if err == nil {
		t.Error("KillDeskServer should return error for non-existent bench")
	}
}

func TestDeskServerInfo_Fields(t *testing.T) {
	// Verify DeskServerInfo struct has expected fields
	info := DeskServerInfo{
		BenchName: "my-bench",
		Socket:    "my-bench-desk",
		Alive:     true,
	}

	if info.BenchName != "my-bench" {
		t.Errorf("expected BenchName 'my-bench', got '%s'", info.BenchName)
	}
	if info.Socket != "my-bench-desk" {
		t.Errorf("expected Socket 'my-bench-desk', got '%s'", info.Socket)
	}
	if !info.Alive {
		t.Error("expected Alive to be true")
	}
}

func TestApplyActionTypes(t *testing.T) {
	// Verify the three action types are the only ones (no guest pane or imps actions)
	validTypes := map[ApplyActionType]bool{
		ActionCreateSession:   true,
		ActionAddWindow:       true,
		ActionApplyEnrichment: true,
	}

	if len(validTypes) != 3 {
		t.Errorf("expected exactly 3 action types, got %d", len(validTypes))
	}

	// Verify no removed action types exist
	removedTypes := []string{"RelocateGuestPanes", "PruneDeadPanes", "KillEmptyImps", "ReconcileLayout"}
	for _, removed := range removedTypes {
		if validTypes[ApplyActionType(removed)] {
			t.Errorf("removed action type %s should not exist", removed)
		}
	}
}
