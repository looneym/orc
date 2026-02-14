package cli

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/secondary"
)

// clearStatusMsg is sent after a timer to auto-clear transient status messages.
type clearStatusMsg struct{}

// animTickMsg advances the starfield animation by one frame.
type animTickMsg struct{}

// animNumFrames is the total number of animation frames (8 frames at 125ms = 1s).
const animNumFrames = 8

// animFrameDuration is the time between animation frames.
const animFrameDuration = 125 * time.Millisecond

// sparkleChars are the characters used in the starfield animation,
// ordered from brightest (near wave front) to dimmest.
var sparkleChars = []string{"✨", "★", "✦", "✧", "·"}

// ansiPattern matches ANSI escape sequences for stripping during entity ID parsing.
var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// entityIDPattern matches entity IDs like SHIP-123, TASK-456, NOTE-789, COMM-001, WORK-014, BENCH-051.
var entityIDPattern = regexp.MustCompile(`\b(SHIP|TASK|NOTE|COMM|WORK|BENCH|TOME|PLAN)-\d+\b`)

// parsedLine represents a single line of the rendered summary tree,
// tagged as either an entity line (navigable) or decorative (skipped by cursor).
type parsedLine struct {
	text     string // original line with ANSI codes preserved
	entityID string // extracted entity ID, empty for decorative lines
	depth    int    // tree indentation depth (0 = top-level, 1 = first-level child, etc.)
}

// summaryModel is the Bubble Tea model for the interactive summary TUI.
type summaryModel struct {
	cmd         *cobra.Command
	opts        summaryOpts
	eventWriter secondary.EventWriter

	// Parsed lines from rendered content
	lines []parsedLine

	// Indices into lines[] that are entity lines (navigable by cursor)
	entityIndices []int

	// Cursor position as index into entityIndices
	cursor int

	// Viewport for scrolling (reserves 1 line at bottom for status bar)
	viewport viewport.Model

	// Terminal dimensions
	width  int
	height int

	// Whether the viewport has been initialized with terminal dimensions
	ready bool

	// Expand/collapse state for tree nodes (keyed by entity ID)
	expanded map[string]bool

	// Status message shown briefly (e.g., "Copied SHIP-412")
	statusMsg string

	// Error from data fetch
	err error

	// Whether we're inside a utils tmux session
	isUtilsSession bool

	// Close confirmation state
	confirming      bool   // true when waiting for y/n confirmation
	confirmEntityID string // entity ID pending close confirmation

	// Starfield animation state
	animating bool       // whether the animation is currently playing
	animFrame int        // current frame index (0..animNumFrames-1)
	animRand  *rand.Rand // deterministic RNG for star positions
}

// summaryContentMsg carries the rendered summary content after async fetch.
type summaryContentMsg struct {
	content  string
	err      error
	entityID string // if set, reposition cursor on this entity after refresh
}

// yankResultMsg carries the result of a clipboard yank operation.
type yankResultMsg struct {
	entityID string
	err      error
}

// focusResultMsg carries the result of an orc focus operation.
type focusResultMsg struct {
	entityID string
	err      error
}

// goblinResultMsg carries the result of sending an entity ID to the goblin pane.
type goblinResultMsg struct {
	entityID string
	err      error
}

// closeResultMsg carries the result of closing/completing an entity.
type closeResultMsg struct {
	entityID string
	err      error
}

// cursorStyle is the visual indicator for the currently selected line.
var cursorStyle = lipgloss.NewStyle().Reverse(true)

// statusBarStyle is the style for the bottom status bar.
var statusBarStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("236")).
	Foreground(lipgloss.Color("252")).
	Padding(0, 1)

// statusKeyStyle highlights key names in the status bar.
var statusKeyStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("236")).
	Foreground(lipgloss.Color("117")).
	Bold(true)

// statusMsgStyle highlights transient messages in the status bar.
var statusMsgStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("236")).
	Foreground(lipgloss.Color("114")).
	Bold(true)

// dimKeyStyle renders unavailable key hints in dark gray (dimmed).
var dimKeyStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("236")).
	Foreground(lipgloss.Color("240"))

// entityActionMatrix maps entity type prefixes to their available context-sensitive actions.
// Actions listed here are: yank, open, focus, close, goblin.
// Global actions (refresh, expand/collapse, quit, navigate) are always available.
var entityActionMatrix = map[string]map[string]bool{
	"COMM":  {"yank": true, "open": true, "focus": true, "goblin": true},
	"SHIP":  {"yank": true, "open": true, "focus": true, "close": true, "goblin": true},
	"TASK":  {"yank": true, "open": true, "close": true, "goblin": true},
	"NOTE":  {"yank": true, "open": true, "focus": true, "goblin": true},
	"TOME":  {"yank": true, "open": true, "focus": true, "goblin": true},
	"PLAN":  {"yank": true, "open": true, "goblin": true},
	"WORK":  {"yank": true, "goblin": true},
	"BENCH": {"yank": true, "goblin": true},
}

// entityHasAction checks whether the given entity ID supports the named action.
func entityHasAction(entityID, action string) bool {
	actions, ok := entityActionMatrix[entityPrefix(entityID)]
	if !ok {
		return false
	}
	return actions[action]
}

// runSummaryTUI launches the interactive Bubble Tea TUI for summary.
func runSummaryTUI(cmd *cobra.Command, opts summaryOpts, eventWriter secondary.EventWriter) error {
	m := summaryModel{
		cmd:            cmd,
		opts:           opts,
		eventWriter:    eventWriter,
		expanded:       make(map[string]bool),
		isUtilsSession: isInsideUtilsSession(),
		animRand:       rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0)),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// emitEvent emits an operational event from the summary TUI.
// Errors are silently ignored — event emission should never disrupt the UI.
func (m summaryModel) emitEvent(level, message string, data map[string]string) {
	if m.eventWriter == nil {
		return
	}
	if err := m.eventWriter.EmitOperational(context.Background(), "summary-tui", level, message, data); err != nil {
		log.Printf("event: EmitOperational summary-tui %s: %v", level, err)
	}
}

// Init returns the initial command to fetch summary data.
func (m summaryModel) Init() tea.Cmd {
	return m.fetchSummary("")
}

// fetchSummary returns a tea.Cmd that fetches summary data.
// If entityID is non-empty, the cursor will be repositioned on that entity after refresh.
func (m summaryModel) fetchSummary(entityID string) tea.Cmd {
	cmd := m.cmd
	opts := m.opts
	return func() tea.Msg {
		content, err := runSummaryOnce(cmd, opts)
		return summaryContentMsg{content: content, err: err, entityID: entityID}
	}
}

// parseLines splits rendered content into tagged lines, identifying entity lines
// and computing tree depth from indentation.
func parseLines(content string) ([]parsedLine, []int) {
	rawLines := strings.Split(content, "\n")
	// Remove trailing empty line from Split (content typically ends with \n)
	if len(rawLines) > 0 && rawLines[len(rawLines)-1] == "" {
		rawLines = rawLines[:len(rawLines)-1]
	}

	lines := make([]parsedLine, len(rawLines))
	var entityIndices []int

	for i, line := range rawLines {
		stripped := ansiPattern.ReplaceAllString(line, "")
		entityID := ""
		if match := entityIDPattern.FindString(stripped); match != "" {
			entityID = match
		}
		lines[i] = parsedLine{text: line, entityID: entityID, depth: treeDepth(stripped)}
		if entityID != "" {
			entityIndices = append(entityIndices, i)
		}
	}

	return lines, entityIndices
}

// treeDepth computes the indentation depth from tree-drawing characters.
// Each "├── ", "└── ", "│   ", or "    " prefix segment adds one level.
func treeDepth(stripped string) int {
	depth := 0
	pos := 0
	for pos < len(stripped) {
		if strings.HasPrefix(stripped[pos:], "├── ") ||
			strings.HasPrefix(stripped[pos:], "└── ") {
			depth++
			pos += len("├── ")
		} else if strings.HasPrefix(stripped[pos:], "│   ") {
			depth++
			pos += len("│   ")
		} else if strings.HasPrefix(stripped[pos:], "    ") {
			depth++
			pos += 4
		} else {
			break
		}
	}
	return depth
}

// isExpandable returns whether an entity type can have children in the tree.
func isExpandable(entityID string) bool {
	switch entityPrefix(entityID) {
	case "COMM", "SHIP", "TOME":
		return true
	default:
		return false
	}
}

// initExpandState sets initial expand/collapse state for entities after loading data.
// Focused entities (those the summary server expanded) start expanded; others collapsed.
func (m *summaryModel) initExpandState() {
	// Detect which entities have children by checking depth relationships
	for i, idx := range m.entityIndices {
		entityID := m.lines[idx].entityID
		if !isExpandable(entityID) {
			continue
		}
		// If not yet in the map, set default: COMM entities and the focused
		// entity are expanded; others expanded if children are already visible
		// (i.e., the summary server chose to expand them).
		if _, exists := m.expanded[entityID]; !exists {
			if entityPrefix(entityID) == "COMM" {
				m.expanded[entityID] = true
			} else if entityID == m.opts.focusedEntityID {
				m.expanded[entityID] = true
			} else {
				// Check if this entity has children rendered (next entity has greater depth)
				hasChildren := false
				if i+1 < len(m.entityIndices) {
					nextIdx := m.entityIndices[i+1]
					if m.lines[nextIdx].depth > m.lines[idx].depth {
						hasChildren = true
					}
				}
				m.expanded[entityID] = hasChildren
			}
		}
	}
}

// cursorEntityID returns the entity ID under the cursor, or empty string.
func (m summaryModel) cursorEntityID() string {
	if len(m.entityIndices) == 0 || m.cursor < 0 || m.cursor >= len(m.entityIndices) {
		return ""
	}
	return m.lines[m.entityIndices[m.cursor]].entityID
}

// cursorLineIndex returns the line index (in m.lines) of the current cursor position.
func (m summaryModel) cursorLineIndex() int {
	if len(m.entityIndices) == 0 || m.cursor < 0 || m.cursor >= len(m.entityIndices) {
		return 0
	}
	return m.entityIndices[m.cursor]
}

// findEntityCursorIndex finds the cursor index for a given entity ID.
// Returns -1 if not found.
func (m summaryModel) findEntityCursorIndex(entityID string) int {
	for i, idx := range m.entityIndices {
		if m.lines[idx].entityID == entityID {
			return i
		}
	}
	return -1
}

// ensureCursorVisible adjusts the viewport offset so the cursor line is visible.
// Uses the visible line offset rather than the raw line index when collapse filtering is active.
func (m *summaryModel) ensureCursorVisible() {
	lineIdx := m.cursorLineIndex()
	// Count visible lines up to the cursor line to get the effective offset
	visibleIdx := 0
	hidden := m.hiddenLines()
	for i := 0; i < lineIdx; i++ {
		if !hidden[i] {
			visibleIdx++
		}
	}
	if visibleIdx < m.viewport.YOffset {
		m.viewport.SetYOffset(visibleIdx)
	} else if visibleIdx >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.SetYOffset(visibleIdx - m.viewport.Height + 1)
	}
}

// nextVisibleCursor finds the next visible entity index in the given direction.
// direction should be +1 (down) or -1 (up). Returns the current cursor if no
// visible entity exists in that direction.
func (m summaryModel) nextVisibleCursor(direction int) int {
	hidden := m.hiddenLines()
	candidate := m.cursor + direction
	for candidate >= 0 && candidate < len(m.entityIndices) {
		if !hidden[m.entityIndices[candidate]] {
			return candidate
		}
		candidate += direction
	}
	return m.cursor
}

// entityPrefix returns the prefix part of an entity ID (e.g., "SHIP" from "SHIP-412").
func entityPrefix(id string) string {
	if idx := strings.Index(id, "-"); idx >= 0 {
		return id[:idx]
	}
	return ""
}

// entityShowCommand returns the orc subcommand for showing an entity's details.
func entityShowCommand(entityID string) string {
	switch entityPrefix(entityID) {
	case "SHIP":
		return "shipment"
	case "TASK":
		return "task"
	case "NOTE":
		return "note"
	case "COMM":
		return "commission"
	case "TOME":
		return "tome"
	case "PLAN":
		return "plan"
	case "WORK":
		return "workshop"
	case "BENCH":
		return "workbench"
	default:
		return ""
	}
}

// yankToClipboard copies a string to the system clipboard via pbcopy.
func yankToClipboard(entityID string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(entityID)
		err := cmd.Run()
		return yankResultMsg{entityID: entityID, err: err}
	}
}

// focusEntity runs orc focus on an entity ID.
func focusEntity(entityID string) tea.Cmd {
	return func() tea.Msg {
		orcBin, err := os.Executable()
		if err != nil {
			orcBin = "orc"
		}
		cmd := exec.Command(orcBin, "focus", entityID)
		err = cmd.Run()
		return focusResultMsg{entityID: entityID, err: err}
	}
}

// clearFocus runs orc focus --clear to remove focus from the current entity.
func clearFocus() tea.Cmd {
	return func() tea.Msg {
		orcBin, err := os.Executable()
		if err != nil {
			orcBin = "orc"
		}
		cmd := exec.Command(orcBin, "focus", "--clear")
		err = cmd.Run()
		return focusResultMsg{entityID: "", err: err}
	}
}

// closeEntity runs the appropriate orc complete command for an entity.
func closeEntity(entityID string) tea.Cmd {
	return func() tea.Msg {
		orcBin, err := os.Executable()
		if err != nil {
			orcBin = "orc"
		}
		var subcmd string
		switch entityPrefix(entityID) {
		case "TASK":
			subcmd = "task"
		case "SHIP":
			subcmd = "shipment"
		default:
			return closeResultMsg{entityID: entityID, err: fmt.Errorf("cannot close %s", entityID)}
		}
		cmd := exec.Command(orcBin, subcmd, "complete", entityID)
		if out, err := cmd.CombinedOutput(); err != nil {
			return closeResultMsg{entityID: entityID, err: fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))}
		}
		return closeResultMsg{entityID: entityID}
	}
}

// sendToGoblin sends an entity ID to the goblin pane in the parent tmux server
// by finding the pane with @pane_role=goblin scoped to the current bench via @bench_id.
// The parent server is the default tmux server; we reach it by unsetting TMUX
// so that the tmux CLI doesn't target the utils server socket.
func (m summaryModel) sendToGoblin(entityID string) tea.Cmd {
	return func() tea.Msg {
		// Read ORC_BENCH_ID from the utils tmux server environment.
		// This env var was set by orc-utils-popup.sh from the parent pane's @bench_id.
		// We do NOT unset TMUX here — show-environment should target the utils server.
		benchID := ""
		if envOut, err := exec.Command("tmux", "show-environment", "ORC_BENCH_ID").Output(); err == nil {
			// Output format: "ORC_BENCH_ID=BENCH-044\n"
			line := strings.TrimSpace(string(envOut))
			if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
				benchID = parts[1]
			}
		}

		// Build the pane filter. If we have a bench ID, scope to that bench's
		// goblin pane using a compound filter on @pane_role AND @bench_id.
		// Otherwise fall back to the unscoped @pane_role=goblin filter.
		var filter string
		if benchID != "" {
			filter = fmt.Sprintf("#{&&:#{==:#{@pane_role},goblin},#{==:#{@bench_id},%s}}", benchID)
		} else {
			m.emitEvent("warn", "ORC_BENCH_ID not set, using unscoped goblin filter", map[string]string{
				"action": "goblin", "entity_id": entityID,
			})
			filter = "#{==:#{@pane_role},goblin}"
		}

		m.emitEvent("debug", "goblin pane lookup", map[string]string{
			"bench_id": benchID, "filter": filter, "entity_id": entityID,
		})

		// Find the goblin pane on the default (parent) tmux server.
		listCmd := exec.Command("tmux", "list-panes", "-a",
			"-f", filter,
			"-F", "#{pane_id}")
		// Unset TMUX so the tmux CLI targets the default server, not the utils server.
		listCmd.Env = envWithoutTMUX()
		out, err := listCmd.Output()
		if err != nil {
			m.emitEvent("error", "goblin pane lookup failed", map[string]string{
				"error": err.Error(), "bench_id": benchID, "entity_id": entityID,
			})
			return goblinResultMsg{entityID: entityID, err: fmt.Errorf("find goblin pane: %w", err)}
		}
		paneID := strings.TrimSpace(string(out))
		if paneID == "" {
			m.emitEvent("error", "no goblin pane found", map[string]string{
				"bench_id": benchID, "filter": filter, "entity_id": entityID,
			})
			return goblinResultMsg{entityID: entityID, err: fmt.Errorf("no goblin pane found")}
		}
		// If multiple goblin panes, take the first one
		if lines := strings.Split(paneID, "\n"); len(lines) > 1 {
			paneID = lines[0]
		}

		m.emitEvent("info", "sending to goblin", map[string]string{
			"pane_id": paneID, "bench_id": benchID, "entity_id": entityID,
		})

		// Send the entity ID followed by Space to separate from existing input
		sendCmd := exec.Command("tmux", "send-keys", "-t", paneID, entityID, "Space")
		sendCmd.Env = envWithoutTMUX()
		if err := sendCmd.Run(); err != nil {
			m.emitEvent("error", "send-keys to goblin failed", map[string]string{
				"error": err.Error(), "pane_id": paneID, "entity_id": entityID,
			})
			return goblinResultMsg{entityID: entityID, err: fmt.Errorf("send-keys to goblin: %w", err)}
		}
		return goblinResultMsg{entityID: entityID}
	}
}

// envWithoutTMUX returns the current environment with the TMUX variable removed,
// so that tmux CLI commands target the default server instead of the utils server.
func envWithoutTMUX() []string {
	var env []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "TMUX=") {
			env = append(env, e)
		}
	}
	return env
}

// openInVim launches vim -R with entity show output via tea.ExecProcess.
func (m summaryModel) openInVim(entityID string) tea.Cmd {
	showCmd := entityShowCommand(entityID)
	if showCmd == "" {
		return nil
	}

	return func() tea.Msg {
		// Get entity details
		orcBin, err := os.Executable()
		if err != nil {
			orcBin = "orc"
		}
		out, err := exec.Command(orcBin, showCmd, "show", entityID).CombinedOutput()
		if err != nil {
			return yankResultMsg{err: fmt.Errorf("orc %s show %s: %w", showCmd, entityID, err)}
		}

		// Write entity content to deterministic path for clean vim title
		tmpPath := fmt.Sprintf("/tmp/%s.txt", entityID)
		if err := os.WriteFile(tmpPath, out, 0o600); err != nil {
			return yankResultMsg{err: err}
		}

		// Write temp vimrc with backslash-quit mapping (avoids shell escaping issues with --cmd)
		vimrcFile, err := os.CreateTemp("", "orc-vimrc-*.vim")
		if err != nil {
			return yankResultMsg{err: err}
		}
		vimrcPath := vimrcFile.Name()
		if _, err := vimrcFile.WriteString("nnoremap \\\\ :q!<CR>\n"); err != nil {
			vimrcFile.Close()
			return yankResultMsg{err: err}
		}
		vimrcFile.Close()

		return tuiExecMsg{tmpPath: tmpPath, vimrcPath: vimrcPath}
	}
}

// tuiExecMsg triggers a tea.ExecProcess to open vim.
type tuiExecMsg struct {
	tmpPath   string
	vimrcPath string
}

// tuiExecDoneMsg is sent when vim exits.
type tuiExecDoneMsg struct {
	tmpPath   string
	vimrcPath string
	err       error
}

// setStatusMsg sets a transient status message and returns a tea.Cmd that will
// auto-clear it after 2 seconds.
func setStatusMsg(m *summaryModel, msg string) tea.Cmd {
	m.statusMsg = msg
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

// startAnimation begins the starfield refresh animation sequence.
func (m *summaryModel) startAnimation() tea.Cmd {
	m.animating = true
	m.animFrame = 0
	return tea.Tick(animFrameDuration, func(time.Time) tea.Msg {
		return animTickMsg{}
	})
}

// renderAnimationFrame overlays sparkle characters onto the current rendered content
// using the wave-front algorithm from the original starfield rain effect.
func (m summaryModel) renderAnimationFrame() string {
	base := m.renderContent()
	if m.width <= 0 || m.height <= 2 {
		return base
	}

	// Split into lines for character overlay
	displayLines := strings.Split(base, "\n")
	h := len(displayLines)
	if h == 0 {
		return base
	}

	// Strip ANSI from each line for safe rune-level overlay
	runeLines := make([][]rune, len(displayLines))
	for i, line := range displayLines {
		runeLines[i] = []rune(ansiPattern.ReplaceAllString(line, ""))
	}

	// Wave front sweeps from top to bottom
	waveFront := (m.animFrame + 1) * h / animNumFrames
	starsPerFrame := max(m.width/10, 6) + m.animFrame*2

	for range starsPerFrame {
		// Cluster stars around the wave front (+-3 rows)
		row := waveFront + m.animRand.IntN(7) - 3
		if row < 0 || row >= h {
			continue
		}
		lineWidth := len(runeLines[row])
		if lineWidth <= 0 {
			continue
		}
		col := m.animRand.IntN(lineWidth)

		// Brighter characters near wave front, dimmer further away
		dist := row - waveFront
		if dist < 0 {
			dist = -dist
		}
		charIdx := dist
		if charIdx >= len(sparkleChars) {
			charIdx = len(sparkleChars) - 1
		}

		sparkle := []rune(sparkleChars[charIdx])
		if len(sparkle) > 0 {
			runeLines[row][col] = sparkle[0]
		}
	}

	// Rebuild the display string
	var b strings.Builder
	for i, runes := range runeLines {
		b.WriteString(string(runes))
		if i < len(runeLines)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// Update handles messages and returns the updated model.
func (m summaryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case clearStatusMsg:
		m.statusMsg = ""
		return m, nil

	case animTickMsg:
		if !m.animating {
			return m, nil
		}
		m.animFrame++
		if m.animFrame >= animNumFrames {
			// Animation complete — stop animating and fetch fresh data
			m.animating = false
			m.animFrame = 0
			entityID := m.cursorEntityID()
			m.viewport.SetContent(m.renderContent())
			return m, m.fetchSummary(entityID)
		}
		// Render next animation frame
		m.viewport.SetContent(m.renderAnimationFrame())
		return m, tea.Tick(animFrameDuration, func(time.Time) tea.Msg {
			return animTickMsg{}
		})

	case summaryContentMsg:
		m.err = msg.err
		if msg.err == nil {
			m.lines, m.entityIndices = parseLines(msg.content)
			m.initExpandState()
			// Reposition cursor on target entity if specified
			if msg.entityID != "" {
				if idx := m.findEntityCursorIndex(msg.entityID); idx >= 0 {
					m.cursor = idx
				}
			}
			// Clamp cursor to valid range
			if m.cursor >= len(m.entityIndices) {
				m.cursor = len(m.entityIndices) - 1
			}
			if m.cursor < 0 {
				m.cursor = 0
			}
			// Ensure cursor is on a visible (non-hidden) entity
			if m.cursor >= 0 && m.cursor < len(m.entityIndices) {
				hidden := m.hiddenLines()
				if hidden[m.entityIndices[m.cursor]] {
					// Try moving up to find a visible entity, then down
					next := m.nextVisibleCursor(-1)
					if next == m.cursor {
						next = m.nextVisibleCursor(1)
					}
					m.cursor = next
				}
			}
			if m.ready {
				m.viewport.SetContent(m.renderContent())
				m.ensureCursorVisible()
			}
		}
		return m, nil

	case yankResultMsg:
		if msg.err != nil {
			cmd := setStatusMsg(&m, fmt.Sprintf("Error: %v", msg.err))
			return m, cmd
		}
		cmd := setStatusMsg(&m, fmt.Sprintf("Copied %s", msg.entityID))
		return m, cmd

	case focusResultMsg:
		if msg.err != nil {
			cmd := setStatusMsg(&m, fmt.Sprintf("Focus error: %v", msg.err))
			return m, cmd
		}
		// Update tracked focus for toggle behavior
		m.opts.focusedEntityID = msg.entityID
		if msg.entityID == "" {
			m.statusMsg = "Focus cleared"
		} else {
			m.statusMsg = fmt.Sprintf("Focused %s", msg.entityID)
		}
		// Refresh tree data, repositioning cursor on the focused entity
		return m, m.fetchSummary(msg.entityID)

	case goblinResultMsg:
		if msg.err != nil {
			cmd := setStatusMsg(&m, fmt.Sprintf("Goblin error: %v", msg.err))
			return m, cmd
		}
		cmd := setStatusMsg(&m, fmt.Sprintf("Sent %s to goblin", msg.entityID))
		return m, cmd

	case closeResultMsg:
		if msg.err != nil {
			cmd := setStatusMsg(&m, fmt.Sprintf("Close error: %v", msg.err))
			return m, cmd
		}
		m.statusMsg = fmt.Sprintf("Closed %s", msg.entityID)
		// Refresh tree data, repositioning cursor on the closed entity
		return m, m.fetchSummary(msg.entityID)

	case tuiExecMsg:
		// Launch vim with temp vimrc for backslash-quit binding
		c := exec.Command("vim", "-R", "-N", "-u", msg.vimrcPath, msg.tmpPath)
		tmpPath := msg.tmpPath
		vimrcPath := msg.vimrcPath
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			return tuiExecDoneMsg{tmpPath: tmpPath, vimrcPath: vimrcPath, err: err}
		})

	case tuiExecDoneMsg:
		// Clean up both temp files
		os.Remove(msg.tmpPath)
		os.Remove(msg.vimrcPath)
		if msg.err != nil {
			cmd := setStatusMsg(&m, fmt.Sprintf("vim error: %v", msg.err))
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		// Ignore key input during animation
		if m.animating {
			return m, nil
		}

		// Handle close confirmation mode: y confirms, anything else cancels
		if m.confirming {
			if msg.String() == "y" {
				entityID := m.confirmEntityID
				m.confirming = false
				m.confirmEntityID = ""
				m.statusMsg = fmt.Sprintf("Closing %s...", entityID)
				return m, closeEntity(entityID)
			}
			m.confirming = false
			m.confirmEntityID = ""
			m.statusMsg = ""
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.emitEvent("info", "tui closed", map[string]string{"action": "quit", "key": msg.String()})
			if m.isUtilsSession {
				detachUtilsSession()
			}
			return m, tea.Quit

		case "j", "down":
			next := m.nextVisibleCursor(1)
			if next != m.cursor {
				m.cursor = next
				m.viewport.SetContent(m.renderContent())
				m.ensureCursorVisible()
			}
			return m, nil

		case "k", "up":
			next := m.nextVisibleCursor(-1)
			if next != m.cursor {
				m.cursor = next
				m.viewport.SetContent(m.renderContent())
				m.ensureCursorVisible()
			}
			return m, nil

		case "enter", "l":
			entityID := m.cursorEntityID()
			if entityID != "" && isExpandable(entityID) {
				m.expanded[entityID] = !m.expanded[entityID]
				action := "collapse"
				if m.expanded[entityID] {
					action = "expand"
				}
				m.emitEvent("debug", action, map[string]string{"action": action, "entity_id": entityID})
				m.viewport.SetContent(m.renderContent())
				m.ensureCursorVisible()
			}
			return m, nil

		case "y":
			entityID := m.cursorEntityID()
			if entityID != "" && entityHasAction(entityID, "yank") {
				m.emitEvent("info", "yank", map[string]string{"action": "yank", "entity_id": entityID})
				return m, yankToClipboard(entityID)
			}
			return m, nil

		case "f":
			entityID := m.cursorEntityID()
			if entityID != "" && entityHasAction(entityID, "focus") {
				// Toggle: unfocus if already focused, focus otherwise
				if entityID == m.opts.focusedEntityID {
					m.emitEvent("info", "unfocus", map[string]string{"action": "unfocus", "entity_id": entityID})
					m.statusMsg = fmt.Sprintf("Unfocusing %s...", entityID)
					return m, clearFocus()
				}
				m.emitEvent("info", "focus", map[string]string{"action": "focus", "entity_id": entityID})
				m.statusMsg = fmt.Sprintf("Focusing %s...", entityID)
				return m, focusEntity(entityID)
			}
			return m, nil

		case "o":
			entityID := m.cursorEntityID()
			if entityID != "" && entityHasAction(entityID, "open") {
				m.emitEvent("info", "vim open", map[string]string{"action": "open", "entity_id": entityID})
				return m, m.openInVim(entityID)
			}
			return m, nil

		case "r":
			m.emitEvent("info", "refresh", map[string]string{"action": "refresh"})
			// Start starfield animation, then refresh with fresh data
			return m, m.startAnimation()

		case "c":
			entityID := m.cursorEntityID()
			if entityID != "" && entityHasAction(entityID, "close") {
				m.emitEvent("info", "close", map[string]string{"action": "close", "entity_id": entityID})
				m.confirming = true
				m.confirmEntityID = entityID
				m.statusMsg = fmt.Sprintf("Close %s? [y/n]", entityID)
			}
			return m, nil

		case "g":
			entityID := m.cursorEntityID()
			if m.isUtilsSession && entityID != "" && entityHasAction(entityID, "goblin") {
				m.emitEvent("info", "goblin send", map[string]string{"action": "goblin", "entity_id": entityID})
				return m, m.sendToGoblin(entityID)
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Reserve 1 line for status bar
		viewHeight := msg.Height - 1
		if viewHeight < 1 {
			viewHeight = 1
		}
		if !m.ready {
			m.viewport = viewport.New(msg.Width, viewHeight)
			m.viewport.SetContent(m.renderContent())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = viewHeight
		}
		return m, nil
	}

	// Forward remaining messages to the viewport for scroll handling
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// hiddenLines computes a set of line indices that should be hidden because
// their parent entity is collapsed. A line is hidden if any ancestor entity
// in the tree is collapsed.
func (m summaryModel) hiddenLines() map[int]bool {
	hidden := make(map[int]bool)

	// Build a stack of collapsed entity depths. When we encounter an entity
	// at depth D that is collapsed, all subsequent lines at depth > D are
	// hidden until we see a line at depth <= D.
	type collapseFrame struct {
		depth int
	}
	var stack []collapseFrame

	for _, idx := range m.entityIndices {
		line := m.lines[idx]
		depth := line.depth

		// Pop frames that no longer apply (we've returned to same or shallower depth)
		for len(stack) > 0 && depth <= stack[len(stack)-1].depth {
			stack = stack[:len(stack)-1]
		}

		// If currently inside a collapsed subtree, this entity is hidden
		if len(stack) > 0 {
			hidden[idx] = true
			// If this entity is also collapsed and expandable, push it
			// (its children are also hidden transitively)
			if isExpandable(line.entityID) && !m.expanded[line.entityID] {
				stack = append(stack, collapseFrame{depth: depth})
			}
			continue
		}

		// Not hidden — check if this entity is collapsed
		if isExpandable(line.entityID) && !m.expanded[line.entityID] {
			stack = append(stack, collapseFrame{depth: depth})
		}
	}

	// Now also hide decorative lines (non-entity lines) that fall within
	// collapsed ranges. Walk all lines and hide those between collapsed parent
	// and next sibling/shallower line.
	for i := range m.lines {
		if m.lines[i].entityID != "" {
			continue // entity lines are handled above
		}
		// Find the nearest preceding entity line
		parentDepth := -1
		parentCollapsed := false
		for j := i - 1; j >= 0; j-- {
			if m.lines[j].entityID != "" {
				parentDepth = m.lines[j].depth
				parentCollapsed = hidden[j]
				break
			}
		}
		// If the nearest entity parent is hidden, hide this decorative line too
		if parentCollapsed {
			hidden[i] = true
			continue
		}
		// Also hide decorative lines that are deeper than a collapsed entity above them
		if parentDepth >= 0 && m.lines[i].depth > parentDepth {
			// Check if the entity at parentDepth is collapsed
			for j := i - 1; j >= 0; j-- {
				if m.lines[j].entityID != "" && m.lines[j].depth == parentDepth {
					if isExpandable(m.lines[j].entityID) && !m.expanded[m.lines[j].entityID] {
						hidden[i] = true
					}
					break
				}
			}
		}
	}

	return hidden
}

// renderContent builds the display string with cursor indicator on the selected entity line.
// Lines belonging to collapsed subtrees are filtered out.
func (m summaryModel) renderContent() string {
	if len(m.lines) == 0 {
		return ""
	}

	cursorLine := -1
	if len(m.entityIndices) > 0 && m.cursor >= 0 && m.cursor < len(m.entityIndices) {
		cursorLine = m.entityIndices[m.cursor]
	}

	hidden := m.hiddenLines()

	var b strings.Builder
	first := true
	for i, line := range m.lines {
		if hidden[i] {
			continue
		}
		if !first {
			b.WriteByte('\n')
		}
		first = false
		if i == cursorLine {
			b.WriteString(cursorStyle.Render(line.text))
		} else {
			b.WriteString(line.text)
		}
	}
	return b.String()
}

// renderStatusBar builds the fixed-layout status bar with dimming for unavailable actions.
func (m summaryModel) renderStatusBar() string {
	// During animation, show a minimal status bar
	if m.animating {
		bar := statusMsgStyle.Render("✨ Refreshing...")
		return statusBarStyle.Width(m.width).Render(bar)
	}

	// If in close confirmation mode, show the prompt
	if m.confirming {
		bar := statusMsgStyle.Render(fmt.Sprintf("Close %s? [y/n]", m.confirmEntityID))
		return statusBarStyle.Width(m.width).Render(bar)
	}

	// If there's a transient status message, show it prominently
	if m.statusMsg != "" {
		bar := statusMsgStyle.Render(m.statusMsg)
		return statusBarStyle.Width(m.width).Render(bar)
	}

	entityID := m.cursorEntityID()

	// Fixed layout: always render all hints in stable order.
	// Context-sensitive actions are bright when available, dim when not.
	hints := formatHint("j/k", "navigate", true) + "  " +
		formatHint("y", "yank", entityHasAction(entityID, "yank")) + "  " +
		formatHint("o", "open", entityHasAction(entityID, "open")) + "  " +
		formatHint("f", "focus", entityHasAction(entityID, "focus")) + "  " +
		formatHint("c", "close", entityHasAction(entityID, "close")) + "  " +
		formatHint("g", "goblin", m.isUtilsSession && entityHasAction(entityID, "goblin")) + "  " +
		formatHint("r", "refresh", true) + "  " +
		formatHint("l", "expand", true) + "  " +
		formatHint("q", "quit", true)

	return statusBarStyle.Width(m.width).Render(hints)
}

// formatHint formats a single keybind hint for the status bar.
// When active is true, the key uses bright styling; when false, it uses dim styling.
func formatHint(key, action string, active bool) string {
	if active {
		return statusKeyStyle.Render(key) + " " + action
	}
	return dimKeyStyle.Render(key + " " + action)
}

// View renders the current model state as a string.
func (m summaryModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}
	if !m.ready {
		return "Loading..."
	}
	return m.viewport.View() + "\n" + m.renderStatusBar()
}

// isInsideUtilsSession checks if we're running inside an ORC utils tmux session
// by querying the ORC_UTILS_SESSION environment variable set by orc-utils-popup.sh.
func isInsideUtilsSession() bool {
	out, err := exec.Command("tmux", "show-environment", "ORC_UTILS_SESSION").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "ORC_UTILS_SESSION=")
}

// detachUtilsSession detaches the tmux client, which closes the utils popup.
func detachUtilsSession() {
	_ = exec.Command("tmux", "detach-client").Run()
}
