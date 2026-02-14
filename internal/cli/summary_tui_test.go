package cli

import (
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTreeDepth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "top level", input: "COMM-001 - My Commission", want: 0},
		{name: "first level child", input: "├── SHIP-412 - My Shipment", want: 1},
		{name: "first level last child", input: "└── SHIP-413 - Last Shipment", want: 1},
		{name: "second level child", input: "│   ├── TASK-100 - My Task", want: 2},
		{name: "second level last child", input: "│   └── TASK-101 - Last Task", want: 2},
		{name: "third level child", input: "│   │   ├── PLAN-001 approved", want: 3},
		{name: "decorative pipe", input: "│", want: 0},
		{name: "empty string", input: "", want: 0},
		{name: "spaces then child", input: "    └── NOTE-001 - A note", want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := treeDepth(tt.input)
			if got != tt.want {
				t.Errorf("treeDepth(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsExpandable(t *testing.T) {
	tests := []struct {
		entityID string
		want     bool
	}{
		{"COMM-001", true},
		{"SHIP-412", true},
		{"TOME-001", true},
		{"TASK-100", false},
		{"NOTE-001", false},
		{"PLAN-001", false},
		{"BENCH-001", false},
		{"WORK-001", false},
	}

	for _, tt := range tests {
		t.Run(tt.entityID, func(t *testing.T) {
			got := isExpandable(tt.entityID)
			if got != tt.want {
				t.Errorf("isExpandable(%q) = %v, want %v", tt.entityID, got, tt.want)
			}
		})
	}
}

func TestParseLinesDepth(t *testing.T) {
	content := strings.Join([]string{
		"COMM-001 - Commission",
		"│",
		"├── SHIP-412 - Shipment",
		"│   ├── TASK-100 - Task One",
		"│   └── TASK-101 - Task Two",
		"└── TOME-001 - Tome One",
	}, "\n")

	lines, entityIndices := parseLines(content)

	// Verify entity count
	if len(entityIndices) != 5 {
		t.Fatalf("expected 5 entity lines, got %d", len(entityIndices))
	}

	// Verify depths of entity lines
	wantDepths := map[string]int{
		"COMM-001": 0,
		"SHIP-412": 1,
		"TASK-100": 2,
		"TASK-101": 2,
		"TOME-001": 1,
	}
	for _, idx := range entityIndices {
		id := lines[idx].entityID
		wantDepth, ok := wantDepths[id]
		if !ok {
			t.Errorf("unexpected entity %q at line %d", id, idx)
			continue
		}
		if lines[idx].depth != wantDepth {
			t.Errorf("entity %q depth = %d, want %d", id, lines[idx].depth, wantDepth)
		}
	}
}

func TestHiddenLines(t *testing.T) {
	content := strings.Join([]string{
		"COMM-001 - Commission",
		"│",
		"├── SHIP-412 - Shipment",
		"│   ├── TASK-100 - Task One",
		"│   └── TASK-101 - Task Two",
		"└── SHIP-413 - Another",
	}, "\n")

	lines, entityIndices := parseLines(content)

	t.Run("collapse SHIP-412 hides its children", func(t *testing.T) {
		m := summaryModel{
			lines:         lines,
			entityIndices: entityIndices,
			expanded: map[string]bool{
				"COMM-001": true,
				"SHIP-412": false, // collapsed
				"SHIP-413": true,
			},
		}

		hidden := m.hiddenLines()

		// TASK-100 and TASK-101 should be hidden
		for _, idx := range entityIndices {
			id := m.lines[idx].entityID
			switch id {
			case "TASK-100", "TASK-101":
				if !hidden[idx] {
					t.Errorf("expected %s (line %d) to be hidden", id, idx)
				}
			case "COMM-001", "SHIP-412", "SHIP-413":
				if hidden[idx] {
					t.Errorf("expected %s (line %d) to be visible", id, idx)
				}
			}
		}
	})

	t.Run("all expanded shows everything", func(t *testing.T) {
		m := summaryModel{
			lines:         lines,
			entityIndices: entityIndices,
			expanded: map[string]bool{
				"COMM-001": true,
				"SHIP-412": true,
				"SHIP-413": true,
			},
		}

		hidden := m.hiddenLines()

		for _, idx := range entityIndices {
			if hidden[idx] {
				t.Errorf("expected %s (line %d) to be visible when all expanded",
					m.lines[idx].entityID, idx)
			}
		}
	})

	t.Run("collapse COMM hides all children", func(t *testing.T) {
		m := summaryModel{
			lines:         lines,
			entityIndices: entityIndices,
			expanded: map[string]bool{
				"COMM-001": false, // collapsed
				"SHIP-412": true,
				"SHIP-413": true,
			},
		}

		hidden := m.hiddenLines()

		// Everything except COMM-001 should be hidden
		for _, idx := range entityIndices {
			id := m.lines[idx].entityID
			if id == "COMM-001" {
				if hidden[idx] {
					t.Errorf("expected COMM-001 to be visible")
				}
			} else {
				if !hidden[idx] {
					t.Errorf("expected %s to be hidden when COMM is collapsed", id)
				}
			}
		}
	})
}

func TestClearStatusMsg(t *testing.T) {
	m := summaryModel{
		statusMsg: "Some message",
		expanded:  make(map[string]bool),
	}

	// Simulate receiving a clearStatusMsg
	result, _ := m.Update(clearStatusMsg{})
	updated := result.(summaryModel)

	if updated.statusMsg != "" {
		t.Errorf("expected statusMsg to be cleared, got %q", updated.statusMsg)
	}
}

func TestSetStatusMsg(t *testing.T) {
	m := summaryModel{
		expanded: make(map[string]bool),
	}

	cmd := setStatusMsg(&m, "Hello")
	if m.statusMsg != "Hello" {
		t.Errorf("expected statusMsg to be set, got %q", m.statusMsg)
	}
	if cmd == nil {
		t.Fatal("expected a tea.Cmd to be returned for the timer")
	}
}

func TestEnterKeyToggle(t *testing.T) {
	content := strings.Join([]string{
		"COMM-001 - Commission",
		"│",
		"├── SHIP-412 - Shipment",
		"│   └── TASK-100 - Task",
		"└── SHIP-413 - Another",
	}, "\n")

	lines, entityIndices := parseLines(content)

	m := summaryModel{
		lines:         lines,
		entityIndices: entityIndices,
		cursor:        1, // pointing at SHIP-412 (entityIndices[1])
		expanded: map[string]bool{
			"COMM-001": true,
			"SHIP-412": true,
			"SHIP-413": true,
		},
		ready: false,
	}

	// Verify cursor points to SHIP-412
	if id := m.cursorEntityID(); id != "SHIP-412" {
		t.Fatalf("expected cursor on SHIP-412, got %s", id)
	}

	// Verify SHIP-412 is expanded
	if !m.expanded["SHIP-412"] {
		t.Fatal("SHIP-412 should start expanded")
	}

	// Send enter key
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(summaryModel)

	if updated.expanded["SHIP-412"] {
		t.Error("SHIP-412 should be collapsed after Enter")
	}

	// Toggle again
	result, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated = result.(summaryModel)

	if !updated.expanded["SHIP-412"] {
		t.Error("SHIP-412 should be expanded after second Enter")
	}
}

func TestEnterKeyNonExpandable(t *testing.T) {
	content := strings.Join([]string{
		"COMM-001 - Commission",
		"├── SHIP-412 - Shipment",
		"│   └── TASK-100 - Task",
	}, "\n")

	lines, entityIndices := parseLines(content)

	m := summaryModel{
		lines:         lines,
		entityIndices: entityIndices,
		cursor:        2, // pointing at TASK-100
		expanded: map[string]bool{
			"COMM-001": true,
			"SHIP-412": true,
		},
		ready: false,
	}

	// Enter on a TASK should be a no-op (TASK is not expandable)
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(summaryModel)

	// expanded map should be unchanged
	if !updated.expanded["COMM-001"] || !updated.expanded["SHIP-412"] {
		t.Error("expanded state should not change when Enter pressed on non-expandable entity")
	}
}

func TestStatusBarFixedLayout(t *testing.T) {
	content := strings.Join([]string{
		"COMM-001 - Commission",
		"├── SHIP-412 - Shipment",
		"│   └── TASK-100 - Task",
	}, "\n")

	lines, entityIndices := parseLines(content)

	t.Run("always shows all hints in fixed layout", func(t *testing.T) {
		m := summaryModel{
			lines:         lines,
			entityIndices: entityIndices,
			cursor:        1, // SHIP-412
			expanded:      map[string]bool{"COMM-001": true, "SHIP-412": true},
			width:         120,
		}

		bar := m.renderStatusBar()
		// Fixed layout: all hints always present regardless of entity type
		for _, hint := range []string{"yank", "open", "focus", "close", "expand", "refresh", "quit"} {
			if !strings.Contains(bar, hint) {
				t.Errorf("status bar should always contain %q hint", hint)
			}
		}
	})

	t.Run("fixed layout for TASK too", func(t *testing.T) {
		m := summaryModel{
			lines:         lines,
			entityIndices: entityIndices,
			cursor:        2, // TASK-100
			expanded:      map[string]bool{"COMM-001": true, "SHIP-412": true},
			width:         120,
		}

		bar := m.renderStatusBar()
		// All hints present even for TASK (unavailable ones are dimmed, not hidden)
		for _, hint := range []string{"yank", "open", "focus", "close", "expand", "refresh", "quit"} {
			if !strings.Contains(bar, hint) {
				t.Errorf("status bar should always contain %q hint even for TASK", hint)
			}
		}
	})
}

// TestKeyMsgNoClearStatus verifies that keypress no longer clears statusMsg
// (the old behavior). Instead, statusMsg is cleared by clearStatusMsg timer.
func TestKeyMsgNoClearStatus(t *testing.T) {
	m := summaryModel{
		statusMsg:     "Copied SHIP-412",
		expanded:      make(map[string]bool),
		entityIndices: []int{},
	}

	// Send a j key — should NOT clear status msg anymore
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updated := result.(summaryModel)

	if updated.statusMsg == "" {
		t.Error("keypress should NOT clear statusMsg; auto-clear timer should handle it")
	}
}

func TestAnimTickMsg(t *testing.T) {
	content := strings.Join([]string{
		"COMM-001 - Commission",
		"├── SHIP-412 - Shipment",
	}, "\n")

	lines, entityIndices := parseLines(content)

	t.Run("advances frame and continues", func(t *testing.T) {
		m := summaryModel{
			lines:         lines,
			entityIndices: entityIndices,
			expanded:      map[string]bool{"COMM-001": true, "SHIP-412": true},
			animating:     true,
			animFrame:     0,
			animRand:      newTestRand(),
			width:         80,
			height:        24,
		}

		result, cmd := m.Update(animTickMsg{})
		updated := result.(summaryModel)

		if updated.animFrame != 1 {
			t.Errorf("expected animFrame=1, got %d", updated.animFrame)
		}
		if !updated.animating {
			t.Error("should still be animating")
		}
		if cmd == nil {
			t.Error("expected a tick cmd for next frame")
		}
	})

	t.Run("last frame stops animation", func(t *testing.T) {
		m := summaryModel{
			lines:         lines,
			entityIndices: entityIndices,
			expanded:      map[string]bool{"COMM-001": true, "SHIP-412": true},
			animating:     true,
			animFrame:     animNumFrames - 1, // last frame
			animRand:      newTestRand(),
			width:         80,
			height:        24,
		}

		result, cmd := m.Update(animTickMsg{})
		updated := result.(summaryModel)

		if updated.animating {
			t.Error("animation should stop after last frame")
		}
		if updated.animFrame != 0 {
			t.Errorf("animFrame should be reset to 0, got %d", updated.animFrame)
		}
		if cmd == nil {
			t.Error("expected a fetchSummary cmd after animation")
		}
	})

	t.Run("ignored when not animating", func(t *testing.T) {
		m := summaryModel{
			lines:         lines,
			entityIndices: entityIndices,
			expanded:      map[string]bool{"COMM-001": true},
			animating:     false,
			animRand:      newTestRand(),
		}

		result, cmd := m.Update(animTickMsg{})
		updated := result.(summaryModel)

		if updated.animating {
			t.Error("should remain not animating")
		}
		if cmd != nil {
			t.Error("no cmd expected when not animating")
		}
	})
}

func TestKeyIgnoredDuringAnimation(t *testing.T) {
	m := summaryModel{
		expanded:      make(map[string]bool),
		entityIndices: []int{},
		animating:     true,
		animRand:      newTestRand(),
	}

	// Any key should be ignored during animation
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updated := result.(summaryModel)

	if cmd != nil {
		t.Error("no cmd expected when keys are blocked during animation")
	}
	if !updated.animating {
		t.Error("animation state should not change from keypress")
	}
}

func TestRenderAnimationFrame(t *testing.T) {
	content := strings.Join([]string{
		"COMM-001 - Commission",
		"├── SHIP-412 - Shipment",
		"│   └── TASK-100 - Task",
	}, "\n")

	lines, entityIndices := parseLines(content)

	m := summaryModel{
		lines:         lines,
		entityIndices: entityIndices,
		expanded:      map[string]bool{"COMM-001": true, "SHIP-412": true},
		animating:     true,
		animFrame:     3,
		animRand:      newTestRand(),
		width:         80,
		height:        24,
	}

	frame := m.renderAnimationFrame()

	// Should produce non-empty output
	if frame == "" {
		t.Error("animation frame should not be empty")
	}

	// Should contain some sparkle characters (with deterministic RNG this is reliable)
	hasSparkle := false
	for _, ch := range sparkleChars {
		if strings.Contains(frame, ch) {
			hasSparkle = true
			break
		}
	}
	if !hasSparkle {
		t.Error("animation frame should contain at least one sparkle character")
	}
}

func TestStatusBarDuringAnimation(t *testing.T) {
	m := summaryModel{
		expanded:  make(map[string]bool),
		animating: true,
		width:     80,
	}

	bar := m.renderStatusBar()
	if !strings.Contains(bar, "Refreshing") {
		t.Error("status bar should show refreshing message during animation")
	}
}

func TestDeskModeQuitBehavior(t *testing.T) {
	m := summaryModel{
		expanded:      make(map[string]bool),
		entityIndices: []int{},
		isDeskSession: true,
	}

	t.Run("q in desk mode returns nil cmd (detach)", func(t *testing.T) {
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if cmd != nil {
			t.Error("q in desk mode should return nil cmd (detach), not tea.Quit")
		}
	})

	t.Run("esc in desk mode returns nil cmd (detach)", func(t *testing.T) {
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cmd != nil {
			t.Error("esc in desk mode should return nil cmd (detach), not tea.Quit")
		}
	})

	t.Run("ctrl+c in desk mode returns tea.Quit", func(t *testing.T) {
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		if cmd == nil {
			t.Error("ctrl+c in desk mode should still return tea.Quit")
		}
	})

	t.Run("q in standalone mode returns tea.Quit", func(t *testing.T) {
		standalone := summaryModel{
			expanded:      make(map[string]bool),
			entityIndices: []int{},
			isDeskSession: false,
		}
		_, cmd := standalone.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if cmd == nil {
			t.Error("q in standalone mode should return tea.Quit")
		}
	})
}

func TestQuitHintLabel(t *testing.T) {
	t.Run("desk mode shows close", func(t *testing.T) {
		m := summaryModel{isDeskSession: true}
		if got := m.quitHintLabel(); got != "close" {
			t.Errorf("expected 'close', got %q", got)
		}
	})

	t.Run("standalone mode shows quit", func(t *testing.T) {
		m := summaryModel{isDeskSession: false}
		if got := m.quitHintLabel(); got != "quit" {
			t.Errorf("expected 'quit', got %q", got)
		}
	})
}

func TestStatusBarNoteHintLabel(t *testing.T) {
	content := "SHIP-412 - Shipment\n"
	lines, entityIndices := parseLines(content)

	m := summaryModel{
		lines:         lines,
		entityIndices: entityIndices,
		cursor:        0,
		expanded:      map[string]bool{},
		width:         120,
	}

	bar := m.renderStatusBar()
	if !strings.Contains(bar, "+note") {
		t.Error("status bar should show '+note' instead of 'note'")
	}
}

// TestEntityActionMatrixCompleteness verifies the entity-action matrix has an entry
// for every known entity type and every known action is represented.
func TestEntityActionMatrixCompleteness(t *testing.T) {
	allEntityTypes := []string{"COMM", "SHIP", "TASK", "NOTE", "TOME", "PLAN", "WORK", "BENCH"}
	allActions := []string{"yank", "open", "focus", "close", "goblin", "note", "review", "run", "deploy", "expand"}

	// Every entity type must have a matrix entry
	for _, etype := range allEntityTypes {
		if _, ok := entityActionMatrix[etype]; !ok {
			t.Errorf("entityActionMatrix missing entry for entity type %q", etype)
		}
	}

	// Every action must appear at least once across all entity types
	for _, action := range allActions {
		found := false
		for _, actions := range entityActionMatrix {
			if actions[action] {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("action %q not found in any entity type in entityActionMatrix", action)
		}
	}

	// No unknown entity types in the matrix
	for etype := range entityActionMatrix {
		known := false
		for _, et := range allEntityTypes {
			if etype == et {
				known = true
				break
			}
		}
		if !known {
			t.Errorf("entityActionMatrix contains unknown entity type %q", etype)
		}
	}
}

// newTestRand creates a deterministic RNG for tests.
func newTestRand() *rand.Rand {
	return rand.New(rand.NewPCG(42, 0))
}

// Verify time import is used (compile-time check)
var _ = time.Second
