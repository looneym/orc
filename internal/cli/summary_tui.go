package cli

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// ansiPattern matches ANSI escape sequences for stripping during entity ID parsing.
var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// entityIDPattern matches entity IDs like SHIP-123, TASK-456, NOTE-789, COMM-001, WORK-014, BENCH-051.
var entityIDPattern = regexp.MustCompile(`\b(SHIP|TASK|NOTE|COMM|WORK|BENCH|TOME|PLAN)-\d+\b`)

// parsedLine represents a single line of the rendered summary tree,
// tagged as either an entity line (navigable) or decorative (skipped by cursor).
type parsedLine struct {
	text     string // original line with ANSI codes preserved
	entityID string // extracted entity ID, empty for decorative lines
}

// summaryModel is the Bubble Tea model for the interactive summary TUI.
type summaryModel struct {
	cmd  *cobra.Command
	opts summaryOpts

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

// runSummaryTUI launches the interactive Bubble Tea TUI for summary.
func runSummaryTUI(cmd *cobra.Command, opts summaryOpts) error {
	m := summaryModel{
		cmd:            cmd,
		opts:           opts,
		expanded:       make(map[string]bool),
		isUtilsSession: isInsideUtilsSession(),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
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

// parseLines splits rendered content into tagged lines, identifying entity lines.
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
		lines[i] = parsedLine{text: line, entityID: entityID}
		if entityID != "" {
			entityIndices = append(entityIndices, i)
		}
	}

	return lines, entityIndices
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
func (m *summaryModel) ensureCursorVisible() {
	lineIdx := m.cursorLineIndex()
	if lineIdx < m.viewport.YOffset {
		m.viewport.SetYOffset(lineIdx)
	} else if lineIdx >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.SetYOffset(lineIdx - m.viewport.Height + 1)
	}
}

// entityPrefix returns the prefix part of an entity ID (e.g., "SHIP" from "SHIP-412").
func entityPrefix(id string) string {
	if idx := strings.Index(id, "-"); idx >= 0 {
		return id[:idx]
	}
	return ""
}

// isFocusable returns whether an entity type can be focused with orc focus.
func isFocusable(entityID string) bool {
	switch entityPrefix(entityID) {
	case "COMM", "SHIP", "TOME", "NOTE":
		return true
	default:
		return false
	}
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

		// Write to temp file
		tmpFile, err := os.CreateTemp("", fmt.Sprintf("orc-%s-*.txt", entityID))
		if err != nil {
			return yankResultMsg{err: err}
		}
		tmpPath := tmpFile.Name()
		if _, err := tmpFile.Write(out); err != nil {
			tmpFile.Close()
			return yankResultMsg{err: err}
		}
		tmpFile.Close()

		// Return an exec command that will be handled by the tea runtime
		return tuiExecMsg{tmpPath: tmpPath}
	}
}

// tuiExecMsg triggers a tea.ExecProcess to open vim.
type tuiExecMsg struct {
	tmpPath string
}

// tuiExecDoneMsg is sent when vim exits.
type tuiExecDoneMsg struct {
	tmpPath string
	err     error
}

// Update handles messages and returns the updated model.
func (m summaryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case summaryContentMsg:
		m.err = msg.err
		if msg.err == nil {
			m.lines, m.entityIndices = parseLines(msg.content)
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
			if m.ready {
				m.viewport.SetContent(m.renderContent())
				m.ensureCursorVisible()
			}
		}
		return m, nil

	case yankResultMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.statusMsg = fmt.Sprintf("Copied %s", msg.entityID)
		}
		return m, nil

	case focusResultMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Focus error: %v", msg.err)
			return m, nil
		}
		m.statusMsg = fmt.Sprintf("Focused %s", msg.entityID)
		// Refresh tree data, repositioning cursor on the focused entity
		return m, m.fetchSummary(msg.entityID)

	case tuiExecMsg:
		// Launch vim with the temp file
		c := exec.Command("vim", "-R",
			"--cmd", `nnoremap \\ :q!<CR>`,
			msg.tmpPath)
		tmpPath := msg.tmpPath
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			return tuiExecDoneMsg{tmpPath: tmpPath, err: err}
		})

	case tuiExecDoneMsg:
		// Clean up temp file
		os.Remove(msg.tmpPath)
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("vim error: %v", msg.err)
		}
		return m, nil

	case tea.KeyMsg:
		// Clear transient status message on any keypress
		m.statusMsg = ""

		switch msg.String() {
		case "q", "ctrl+c", "esc":
			if m.isUtilsSession {
				detachUtilsSession()
			}
			return m, tea.Quit

		case "j", "down":
			if m.cursor < len(m.entityIndices)-1 {
				m.cursor++
				m.viewport.SetContent(m.renderContent())
				m.ensureCursorVisible()
			}
			return m, nil

		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				m.viewport.SetContent(m.renderContent())
				m.ensureCursorVisible()
			}
			return m, nil

		case "y":
			entityID := m.cursorEntityID()
			if entityID != "" {
				return m, yankToClipboard(entityID)
			}
			return m, nil

		case "f":
			entityID := m.cursorEntityID()
			if entityID != "" && isFocusable(entityID) {
				m.statusMsg = fmt.Sprintf("Focusing %s...", entityID)
				return m, focusEntity(entityID)
			}
			return m, nil

		case "o":
			entityID := m.cursorEntityID()
			if entityID != "" {
				return m, m.openInVim(entityID)
			}
			return m, nil

		case "r":
			// Manual refresh, preserving cursor on current entity
			entityID := m.cursorEntityID()
			m.statusMsg = "Refreshing..."
			return m, m.fetchSummary(entityID)
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

// renderContent builds the display string with cursor indicator on the selected entity line.
func (m summaryModel) renderContent() string {
	if len(m.lines) == 0 {
		return ""
	}

	cursorLine := -1
	if len(m.entityIndices) > 0 && m.cursor >= 0 && m.cursor < len(m.entityIndices) {
		cursorLine = m.entityIndices[m.cursor]
	}

	var b strings.Builder
	for i, line := range m.lines {
		if i == cursorLine {
			b.WriteString(cursorStyle.Render(line.text))
		} else {
			b.WriteString(line.text)
		}
		if i < len(m.lines)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// renderStatusBar builds the context-sensitive status bar based on the entity under cursor.
func (m summaryModel) renderStatusBar() string {
	// If there's a transient status message, show it prominently
	if m.statusMsg != "" {
		bar := statusMsgStyle.Render(m.statusMsg)
		return statusBarStyle.Width(m.width).Render(bar)
	}

	entityID := m.cursorEntityID()
	if entityID == "" {
		return statusBarStyle.Width(m.width).Render(
			formatHint("j/k", "navigate") + "  " +
				formatHint("r", "refresh") + "  " +
				formatHint("q", "quit"),
		)
	}

	hints := formatHint("j/k", "navigate") + "  " +
		formatHint("y", "yank ID") + "  " +
		formatHint("o", "open in vim") + "  " +
		formatHint("r", "refresh")

	if isFocusable(entityID) {
		hints += "  " + formatHint("f", "focus")
	}

	hints += "  " + formatHint("q", "quit")

	return statusBarStyle.Width(m.width).Render(hints)
}

// formatHint formats a single keybind hint for the status bar.
func formatHint(key, action string) string {
	return statusKeyStyle.Render(key) + " " + action
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
