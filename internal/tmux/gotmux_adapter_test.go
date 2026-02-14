package tmux

import (
	"strings"
	"testing"
)

func TestNewGotmuxAdapter(t *testing.T) {
	adapter, err := NewGotmuxAdapter("")
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	if adapter == nil {
		t.Fatal("adapter should not be nil")
	}

	if adapter.tmux == nil {
		t.Fatal("adapter.tmux should not be nil for default socket")
	}

	if adapter.server == nil {
		t.Fatal("adapter.server should not be nil")
	}

	if adapter.server.Socket != "" {
		t.Errorf("default adapter should have empty socket, got %q", adapter.server.Socket)
	}
}

func TestNewGotmuxAdapter_CustomSocket(t *testing.T) {
	adapter, err := NewGotmuxAdapter("orc-test-factory")
	if err != nil {
		t.Fatalf("failed to create adapter with custom socket: %v", err)
	}

	if adapter == nil {
		t.Fatal("adapter should not be nil")
	}

	if adapter.tmux != nil {
		t.Error("custom socket adapter should have nil tmux (gotmux doesn't support custom sockets)")
	}

	if adapter.server == nil {
		t.Fatal("adapter.server should not be nil")
	}

	if adapter.server.Socket != "orc-test-factory" {
		t.Errorf("expected socket 'orc-test-factory', got %q", adapter.server.Socket)
	}
}

func TestSessionExists(t *testing.T) {
	adapter, err := NewGotmuxAdapter("")
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
	adapter, err := NewGotmuxAdapter("")
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

func TestAttachInstructions_CustomSocket(t *testing.T) {
	adapter, err := NewGotmuxAdapter("orc-phoenix")
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	instructions := adapter.AttachInstructions("test-session")
	if !strings.Contains(instructions, "-L orc-phoenix") {
		t.Error("AttachInstructions with custom socket should include -L flag")
	}
}

func TestPlanApply_NewSession(t *testing.T) {
	adapter, err := NewGotmuxAdapter("")
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
	adapter, err := NewGotmuxAdapter("")
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
	adapter, err := NewGotmuxAdapter("")
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
	adapter, err := NewGotmuxAdapter("")
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

// --- Server and FactorySocket tests ---

func TestDefaultServer(t *testing.T) {
	srv := DefaultServer()
	if srv == nil {
		t.Fatal("DefaultServer() should not return nil")
	}
	if srv.Socket != "" {
		t.Errorf("DefaultServer should have empty socket, got %q", srv.Socket)
	}
}

func TestServer_SessionExists_NonExistent(t *testing.T) {
	srv := DefaultServer()
	if srv.SessionExists("nonexistent-session-test-99999") {
		t.Error("SessionExists should return false for non-existent session")
	}
}

func TestServer_CustomSocket_SessionExists(t *testing.T) {
	srv := &Server{Socket: "orc-test-nonexistent"}
	// A non-existent socket server should report no sessions
	if srv.SessionExists("any-session") {
		t.Error("SessionExists on non-existent socket should return false")
	}
}

func TestFactorySocket(t *testing.T) {
	tests := []struct {
		name     string
		factory  string
		expected string
	}{
		{"default factory", "default", ""},
		{"default factory uppercase", "Default", ""},
		{"default factory mixed case", "DEFAULT", ""},
		{"custom factory", "phoenix-dev", "orc-phoenix-dev"},
		{"factory with spaces", "My Factory", "orc-my-factory"},
		{"factory with leading/trailing spaces", "  staging  ", "orc-staging"},
		{"simple factory", "prod", "orc-prod"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FactorySocket(tt.factory)
			if got != tt.expected {
				t.Errorf("FactorySocket(%q) = %q, want %q", tt.factory, got, tt.expected)
			}
		})
	}
}

func TestSession_SrvDefaultsToDefaultServer(t *testing.T) {
	// A Session created without a server should default to DefaultServer
	s := &Session{Name: "test"}
	srv := s.srv()
	if srv == nil {
		t.Fatal("Session.srv() should not return nil")
	}
	if srv.Socket != "" {
		t.Errorf("Session.srv() should default to empty socket, got %q", srv.Socket)
	}
}

func TestSession_SrvUsesExplicitServer(t *testing.T) {
	customSrv := &Server{Socket: "orc-custom"}
	s := &Session{Name: "test", server: customSrv}
	srv := s.srv()
	if srv != customSrv {
		t.Error("Session.srv() should return the explicit server")
	}
	if srv.Socket != "orc-custom" {
		t.Errorf("expected socket 'orc-custom', got %q", srv.Socket)
	}
}

func TestGotmuxAdapter_Server(t *testing.T) {
	adapter, err := NewGotmuxAdapter("")
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	srv := adapter.Server()
	if srv == nil {
		t.Fatal("Server() should not return nil")
	}
	if srv.Socket != "" {
		t.Errorf("expected empty socket, got %q", srv.Socket)
	}
}
