package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Server represents a tmux server, optionally on a custom socket.
// An empty Socket means the default server.
type Server struct {
	Socket string
}

// DefaultServer returns a Server targeting the default tmux socket.
func DefaultServer() *Server {
	return &Server{}
}

// cmd builds an exec.Cmd targeting this server's socket.
func (srv *Server) cmd(args ...string) *exec.Cmd {
	if srv.Socket != "" {
		args = append([]string{"-L", srv.Socket}, args...)
	}
	return exec.Command("tmux", args...)
}

// FactorySocket derives a tmux socket name from a factory name.
// Returns "" for the default factory (maps to default tmux server).
func FactorySocket(factoryName string) string {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(factoryName), " ", "-"))
	if normalized == "default" {
		return ""
	}
	return "orc-" + normalized
}

// exactSession returns a tmux target string that matches the session name exactly.
// Prefixing with "=" prevents tmux from partial-matching against other sessions.
// Use this for all session-targeted commands to avoid cross-session interference.
func exactSession(sessionName string) string {
	return "=" + sessionName
}

// exactTarget returns a tmux target string for session:window with exact session matching.
func exactTarget(sessionName, windowName string) string {
	return exactSession(sessionName) + ":" + windowName
}

// Session represents a TMux session
type Session struct {
	Name   string
	server *Server
}

// srv returns the Server for this session, defaulting to the default server.
func (s *Session) srv() *Server {
	if s.server != nil {
		return s.server
	}
	return DefaultServer()
}

// Window represents a TMux window
type Window struct {
	Session *Session
	Index   int
	Name    string
}

// NewSession creates a new TMux session on this server.
func (srv *Server) NewSession(name, workingDir string) (*Session, error) {
	// Create session with first window, start numbering from 1
	if err := srv.cmd("new-session", "-d", "-s", name, "-c", workingDir).Run(); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Set base-index to 1 for this session (windows start at 1)
	srv.cmd("set-option", "-t", name, "base-index", "1").Run()
	// Set pane-base-index to 1 (panes start at 1)
	srv.cmd("set-option", "-t", name, "pane-base-index", "1").Run()

	// Rename the auto-created first window to a placeholder
	// The apply logic will rename it to the proper name (e.g., "goblin")
	srv.cmd("rename-window", "-t", name+":^", "__init__").Run()

	return &Session{Name: name, server: srv}, nil
}

// NewSession creates a new TMux session on the default server.
func NewSession(name, workingDir string) (*Session, error) {
	return DefaultServer().NewSession(name, workingDir)
}

// KillSession terminates a TMux session on this server.
func (srv *Server) KillSession(name string) error {
	return srv.cmd("kill-session", "-t", exactSession(name)).Run()
}

// KillSession terminates a TMux session on the default server.
func KillSession(name string) error {
	return DefaultServer().KillSession(name)
}

// WindowExists checks if a window exists in a session on this server.
func (srv *Server) WindowExists(sessionName, windowName string) bool {
	output, err := srv.cmd("list-windows", "-t", exactSession(sessionName), "-F", "#{window_name}").Output()
	if err != nil {
		return false
	}
	windows := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, w := range windows {
		if w == windowName {
			return true
		}
	}
	return false
}

// WindowExists checks if a window exists in a session on the default server.
func WindowExists(sessionName, windowName string) bool {
	return DefaultServer().WindowExists(sessionName, windowName)
}

// KillWindow kills a window in a session on this server.
func (srv *Server) KillWindow(sessionName, windowName string) error {
	return srv.cmd("kill-window", "-t", exactTarget(sessionName, windowName)).Run()
}

// KillWindow kills a window in a session on the default server.
func KillWindow(sessionName, windowName string) error {
	return DefaultServer().KillWindow(sessionName, windowName)
}

// GetPaneCount returns the number of panes in a window on this server.
func (srv *Server) GetPaneCount(sessionName, windowName string) int {
	target := exactTarget(sessionName, windowName)
	output, err := srv.cmd("list-panes", "-t", target).Output()
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	return len(lines)
}

// GetPaneCount returns the number of panes in a window on the default server.
func GetPaneCount(sessionName, windowName string) int {
	return DefaultServer().GetPaneCount(sessionName, windowName)
}

// GetPaneCommand returns the current command running in a specific pane on this server.
// Returns empty string if pane doesn't exist or error occurs.
func (srv *Server) GetPaneCommand(sessionName, windowName string, paneNum int) string {
	target := fmt.Sprintf("=%s:%s.%d", sessionName, windowName, paneNum)
	output, err := srv.cmd("display-message", "-t", target, "-p", "#{pane_current_command}").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// GetPaneCommand returns the current command running in a specific pane on the default server.
func GetPaneCommand(sessionName, windowName string, paneNum int) string {
	return DefaultServer().GetPaneCommand(sessionName, windowName, paneNum)
}

// GetPaneStartPath returns the initial directory for a pane (pane_start_path) on this server.
// This is set when the pane is created and does not change.
// Returns empty string if pane doesn't exist or error occurs.
func (srv *Server) GetPaneStartPath(sessionName, windowName string, paneNum int) string {
	target := fmt.Sprintf("=%s:%s.%d", sessionName, windowName, paneNum)
	output, err := srv.cmd("display-message", "-t", target, "-p", "#{pane_start_path}").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// GetPaneStartPath returns the initial directory for a pane on the default server.
func GetPaneStartPath(sessionName, windowName string, paneNum int) string {
	return DefaultServer().GetPaneStartPath(sessionName, windowName, paneNum)
}

// GetPaneStartCommand returns the initial command for a pane (pane_start_command) on this server.
// This is only set when the pane is created with respawn-pane or similar.
// Returns empty string if not set, pane doesn't exist, or error occurs.
func (srv *Server) GetPaneStartCommand(sessionName, windowName string, paneNum int) string {
	target := fmt.Sprintf("=%s:%s.%d", sessionName, windowName, paneNum)
	output, err := srv.cmd("display-message", "-t", target, "-p", "#{pane_start_command}").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// GetPaneStartCommand returns the initial command for a pane on the default server.
func GetPaneStartCommand(sessionName, windowName string, paneNum int) string {
	return DefaultServer().GetPaneStartCommand(sessionName, windowName, paneNum)
}

// CapturePaneContent captures visible content from a pane on this server.
// target is in format "session:window.pane" (e.g., "workshop:bench.2")
// lines specifies how many lines to capture (0 for all visible)
func (srv *Server) CapturePaneContent(target string, lines int) (string, error) {
	args := []string{"capture-pane", "-t", target, "-p"}
	if lines > 0 {
		args = append(args, "-S", fmt.Sprintf("-%d", lines))
	}
	output, err := srv.cmd(args...).Output()
	if err != nil {
		return "", fmt.Errorf("failed to capture pane content: %w", err)
	}
	return string(output), nil
}

// CapturePaneContent captures visible content from a pane on the default server.
func CapturePaneContent(target string, lines int) (string, error) {
	return DefaultServer().CapturePaneContent(target, lines)
}

// CreateOrcWindow creates the ORC orchestrator window with layout:
// Layout:
//
//	+---------------------+--------------+
//	|                     |   vim (top)  |
//	|      claude         |--------------+
//	|    (full height)    |  shell (bot) |
//	|                     |              |
//	+---------------------+--------------+
func (s *Session) CreateOrcWindow(workingDir string) error {
	// First window is already created (window 1), rename it
	target := fmt.Sprintf("%s:1", s.Name)

	if err := s.srv().cmd("rename-window", "-t", target, "goblin").Run(); err != nil {
		return fmt.Errorf("failed to rename goblin window: %w", err)
	}

	// Split vertically (creates pane on the right)
	if err := s.SplitVertical(target, workingDir); err != nil {
		return err
	}

	// Now split the right pane horizontally
	// Target the right pane (pane 2)
	rightPane := fmt.Sprintf("%s.2", target)
	if err := s.SplitHorizontal(rightPane, workingDir); err != nil {
		return err
	}

	// Now we have 3 panes:
	// Pane 1 (left): claude (orchestrator via orc connect --role goblin)
	// Pane 2 (top right): vim
	// Pane 3 (bottom right): shell

	// Launch orc connect --role goblin in pane 1 (left) - uses respawn-pane so it's the root command
	pane1 := fmt.Sprintf("%s.1", target)
	if err := s.srv().cmd("respawn-pane", "-t", pane1, "-k", "orc", "connect", "--role", "goblin").Run(); err != nil {
		return fmt.Errorf("failed to launch orc connect: %w", err)
	}

	// Launch vim in pane 2 (top right)
	pane2 := fmt.Sprintf("%s.2", target)
	if err := s.SendKeys(pane2, "vim"); err != nil {
		return fmt.Errorf("failed to launch vim: %w", err)
	}

	// Pane 3 (bottom right) is just a shell, already there

	return nil
}

// CreateWorkbenchWindowShell creates a workbench window with layout but NO app launching
// Layout:
//
//	+-----------------+-----------------+
//	|                 | (top right)     |
//	| (left pane)     +-----------------+
//	|                 | (bottom right)  |
//	+-----------------+-----------------+
//
// Apps (vim, claude) can be launched later
func (s *Session) CreateWorkbenchWindowShell(index int, name, workingDir string) (*Window, error) {
	// Create new window
	if err := s.srv().cmd("new-window", "-t", s.Name, "-n", name, "-c", workingDir).Run(); err != nil {
		return nil, fmt.Errorf("failed to create workbench window: %w", err)
	}

	target := fmt.Sprintf("%s:%s", s.Name, name)

	// Split vertically (creates pane on the right)
	if err := s.SplitVertical(target, workingDir); err != nil {
		return nil, err
	}

	// Split the right pane horizontally
	rightPane := fmt.Sprintf("%s.2", target)
	if err := s.SplitHorizontal(rightPane, workingDir); err != nil {
		return nil, err
	}

	// Now we have 3 panes ready:
	// Pane 1 (left): shell (for vim)
	// Pane 2 (top right): will become IMP (orc connect)
	// Pane 3 (bottom right): shell

	// Launch orc connect in top-right pane (pane 2)
	// Using respawn-pane makes "orc connect" the root command
	// This means if the pane exits or is respawned, it runs orc connect again
	topRightPane := fmt.Sprintf("%s.2", target)
	if err := s.srv().cmd("respawn-pane", "-t", topRightPane, "-k", "orc", "connect").Run(); err != nil {
		return nil, fmt.Errorf("failed to launch orc connect in top-right pane: %w", err)
	}

	return &Window{Session: s, Index: index, Name: name}, nil
}

// CreateWorkbenchWindow creates a workbench window with sophisticated layout:
// Layout:
//
//	+-----------------+-----------------+
//	|                 | claude (IMP)    |
//	| vim             +-----------------+
//	|                 | shell           |
//	+-----------------+-----------------+
func (s *Session) CreateWorkbenchWindow(index int, name, workingDir string) (*Window, error) {
	// Create new window
	if err := s.srv().cmd("new-window", "-t", s.Name, "-n", name, "-c", workingDir).Run(); err != nil {
		return nil, fmt.Errorf("failed to create workbench window: %w", err)
	}

	target := fmt.Sprintf("%s:%s", s.Name, name)

	// Get the pane ID for the first pane (will be vim)
	// Split vertically (creates pane on the right)
	if err := s.SplitVertical(target, workingDir); err != nil {
		return nil, err
	}

	// Now split the right pane horizontally
	// Target the right pane (pane 2)
	rightPane := fmt.Sprintf("%s.2", target)
	if err := s.SplitHorizontal(rightPane, workingDir); err != nil {
		return nil, err
	}

	// Now we have 3 panes:
	// Pane 1 (left): vim
	// Pane 2 (top right): claude (IMP via orc connect)
	// Pane 3 (bottom right): shell

	// Launch vim in pane 1 (left) - use respawn-pane so pane_start_command is set
	pane1 := fmt.Sprintf("%s.1", target)
	if err := s.srv().cmd("respawn-pane", "-t", pane1, "-k", "vim").Run(); err != nil {
		return nil, fmt.Errorf("failed to launch vim: %w", err)
	}

	// Launch orc connect in pane 2 (top right - IMP) - uses respawn-pane so it's the root command
	pane2 := fmt.Sprintf("%s.2", target)
	if err := s.srv().cmd("respawn-pane", "-t", pane2, "-k", "orc", "connect").Run(); err != nil {
		return nil, fmt.Errorf("failed to launch orc connect: %w", err)
	}

	// Pane 3 (bottom right) is just a shell, already there

	return &Window{Session: s, Index: index, Name: name}, nil
}

// SplitVertical splits a pane vertically (creates pane on the right)
func (s *Session) SplitVertical(target, workingDir string) error {
	return s.srv().cmd("split-window", "-h", "-t", target, "-c", workingDir).Run()
}

// SplitHorizontal splits a pane horizontally (creates pane below)
func (s *Session) SplitHorizontal(target, workingDir string) error {
	return s.srv().cmd("split-window", "-v", "-t", target, "-c", workingDir).Run()
}

// JoinPane moves a pane from source to target on this server.
// If vertical is true, joins vertically (-v); otherwise horizontally (-h).
// Size specifies the target pane size in lines (if vertical) or columns (if horizontal).
func (srv *Server) JoinPane(source, target string, vertical bool, size int) error {
	args := []string{"join-pane"}
	if vertical {
		args = append(args, "-v")
	} else {
		args = append(args, "-h")
	}
	if size > 0 {
		args = append(args, "-l", strconv.Itoa(size))
	}
	args = append(args, "-s", source, "-t", target)
	return srv.cmd(args...).Run()
}

// JoinPane moves a pane from source to target on the default server.
func JoinPane(source, target string, vertical bool, size int) error {
	return DefaultServer().JoinPane(source, target, vertical, size)
}

// SendKeys sends keystrokes to a pane (with Enter)
func (s *Session) SendKeys(target, keys string) error {
	return s.srv().cmd("send-keys", "-t", target, keys, "C-m").Run()
}

// SelectWindow switches to a specific window
func (s *Session) SelectWindow(windowIndex int) error {
	target := fmt.Sprintf("%s:%d", s.Name, windowIndex)
	return s.srv().cmd("select-window", "-t", target).Run()
}

// RenameWindow renames a window on this server.
func (srv *Server) RenameWindow(target, newName string) error {
	return srv.cmd("rename-window", "-t", target, newName).Run()
}

// RenameWindow renames a window on the default server.
func RenameWindow(target, newName string) error {
	return DefaultServer().RenameWindow(target, newName)
}

// RespawnPane respawns a pane with optional command on this server.
func (srv *Server) RespawnPane(target string, command ...string) error {
	args := []string{"respawn-pane", "-t", target, "-k"}
	args = append(args, command...)
	return srv.cmd(args...).Run()
}

// RespawnPane respawns a pane with optional command on the default server.
func RespawnPane(target string, command ...string) error {
	return DefaultServer().RespawnPane(target, command...)
}

// SetupGoblinPane launches orc connect --role goblin in pane 1 of an existing window on this server.
// Target format: "session:window" (e.g., "WORK-005:goblin")
func (srv *Server) SetupGoblinPane(target string) error {
	pane1 := fmt.Sprintf("%s.1", target)
	if err := srv.cmd("respawn-pane", "-t", pane1, "-k", "orc", "connect", "--role", "goblin").Run(); err != nil {
		return fmt.Errorf("failed to launch orc connect in goblin pane: %w", err)
	}
	return nil
}

// SetupGoblinPane launches orc connect --role goblin in pane 1 on the default server.
func SetupGoblinPane(target string) error {
	return DefaultServer().SetupGoblinPane(target)
}

// GetSessionInfo returns formatted information about the session on this server.
func (srv *Server) GetSessionInfo(name string) (string, error) {
	output, err := srv.cmd("list-windows", "-t", exactSession(name)).Output()
	if err != nil {
		return "", fmt.Errorf("failed to get session info: %w", err)
	}
	return string(output), nil
}

// GetSessionInfo returns formatted information about the session on the default server.
func GetSessionInfo(name string) (string, error) {
	return DefaultServer().GetSessionInfo(name)
}

// SessionExists checks if a TMux session exists on this server.
func (srv *Server) SessionExists(name string) bool {
	return srv.cmd("has-session", "-t", exactSession(name)).Run() == nil
}

// SessionExists checks if a TMux session exists on the default server.
func SessionExists(name string) bool {
	return DefaultServer().SessionExists(name)
}

// AttachInstructions returns user-friendly instructions for attaching to session
func AttachInstructions(sessionName string) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Attach to session: tmux attach -t %s\n", sessionName))
	b.WriteString("\n")
	b.WriteString("Window Layout:\n")
	b.WriteString("  Each window has a single goblin pane (orc connect)\n")
	b.WriteString("\n")
	b.WriteString("TMux Commands:\n")
	b.WriteString("  Switch windows: Ctrl+b then window number (1, 2, 3...)\n")
	b.WriteString("  Detach session: Ctrl+b then d\n")
	b.WriteString("  Open desk: Double-click status bar or Ctrl+b then u\n")
	b.WriteString("  List windows: Ctrl+b then w\n")

	return b.String()
}

// SendKeysLiteral sends text literally without interpretation
func (s *Session) SendKeysLiteral(target, text string) error {
	return s.srv().cmd("send-keys", "-t", target, "-l", text).Run()
}

// SendEscape sends the Escape key
func (s *Session) SendEscape(target string) error {
	return s.srv().cmd("send-keys", "-t", target, "Escape").Run()
}

// SendEnter sends the Enter key
func (s *Session) SendEnter(target string) error {
	return s.srv().cmd("send-keys", "-t", target, "Enter").Run()
}

// RenameSession renames a tmux session on this server.
func (srv *Server) RenameSession(oldName, newName string) error {
	return srv.cmd("rename-session", "-t", exactSession(oldName), newName).Run()
}

// RenameSession renames a tmux session on the default server.
func RenameSession(oldName, newName string) error {
	return DefaultServer().RenameSession(oldName, newName)
}

// GetCurrentSessionName returns the name of the current tmux session on this server.
// Returns empty string if not in tmux or on error.
func (srv *Server) GetCurrentSessionName() string {
	output, err := srv.cmd("display-message", "-p", "#{session_name}").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// GetCurrentSessionName returns the name of the current tmux session on the default server.
func GetCurrentSessionName() string {
	return DefaultServer().GetCurrentSessionName()
}

// IsOrcSession returns true if the current tmux session has ORC_WORKSHOP_ID set,
// indicating this is an ORC-managed workshop session.
// Returns false if not in tmux, session name unavailable, or env var not set.
func (srv *Server) IsOrcSession() bool {
	session := srv.GetCurrentSessionName()
	if session == "" {
		return false
	}
	val, err := srv.GetEnvironment(session, "ORC_WORKSHOP_ID")
	return err == nil && val != ""
}

// IsOrcSession returns true if the current tmux session has ORC_WORKSHOP_ID set on the default server.
func IsOrcSession() bool {
	return DefaultServer().IsOrcSession()
}

// SetOption sets a tmux option for a session on this server.
func (srv *Server) SetOption(session, option, value string) error {
	return srv.cmd("set-option", "-t", session, option, value).Run()
}

// SetOption sets a tmux option for a session on the default server.
func SetOption(session, option, value string) error {
	return DefaultServer().SetOption(session, option, value)
}

// DisplayPopup shows a popup window with a command on this server.
func (srv *Server) DisplayPopup(session, command string, width, height int, title string) error {
	args := []string{"display-popup", "-t", session, "-E"}
	if width > 0 {
		args = append(args, "-w", strconv.Itoa(width))
	}
	if height > 0 {
		args = append(args, "-h", strconv.Itoa(height))
	}
	if title != "" {
		args = append(args, "-T", title)
	}
	args = append(args, command)
	return srv.cmd(args...).Run()
}

// DisplayPopup shows a popup window with a command on the default server.
func DisplayPopup(session, command string, width, height int, title string) error {
	return DefaultServer().DisplayPopup(session, command, width, height, title)
}

// BindKey binds a key to a command for a session on this server.
func (srv *Server) BindKey(session, key, command string) error {
	// Use bind-key with -T root for global bindings (like mouse events)
	return srv.cmd("bind-key", "-T", "root", key, "run-shell", command).Run()
}

// BindKey binds a key to a command on the default server.
func BindKey(session, key, command string) error {
	return DefaultServer().BindKey(session, key, command)
}

// BindKeyPopup binds a key to display a command in a popup on this server.
func (srv *Server) BindKeyPopup(session, key, command string, width, height int, title, workingDir string) error {
	args := []string{"bind-key", "-T", "root", key, "display-popup", "-E"}
	if workingDir != "" {
		args = append(args, "-d", workingDir)
	}
	if width > 0 {
		args = append(args, "-w", strconv.Itoa(width))
	}
	if height > 0 {
		args = append(args, "-h", strconv.Itoa(height))
	}
	if title != "" {
		args = append(args, "-T", title)
	}
	args = append(args, command)
	return srv.cmd(args...).Run()
}

// BindKeyPopup binds a key to display a command in a popup on the default server.
func BindKeyPopup(session, key, command string, width, height int, title, workingDir string) error {
	return DefaultServer().BindKeyPopup(session, key, command, width, height, title, workingDir)
}

// MenuItem represents an item in a tmux context menu.
type MenuItem struct {
	Label   string // Display text
	Key     string // Shortcut key (single char, or "" for none)
	Command string // tmux command to execute
}

// BindContextMenu binds a key to display a context menu on this server.
// Uses -x M -y M to position at mouse coordinates, -O to keep menu open.
func (srv *Server) BindContextMenu(key, title string, items []MenuItem) error {
	args := []string{"bind-key", "-T", "root", key, "display-menu", "-O", "-T", title, "-x", "M", "-y", "M"}
	for _, item := range items {
		args = append(args, item.Label, item.Key, item.Command)
	}
	return srv.cmd(args...).Run()
}

// BindContextMenu binds a key to display a context menu on the default server.
func BindContextMenu(key, title string, items []MenuItem) error {
	return DefaultServer().BindContextMenu(key, title, items)
}

// ApplyGlobalBindings sets up ORC's global tmux key bindings on this server.
// Safe to call repeatedly (idempotent). Silently ignores errors (tmux may not be running).
func (srv *Server) ApplyGlobalBindings() {
	// Desk popup command -- shared by double-click, prefix+u, and context menu
	// Note: display-popup does NOT expand #{} formats in shell-command,
	// so the script queries the main tmux server directly via TMUX env var.
	deskPopup := "$HOME/.orc/tmux/orc-desk-popup.sh"
	deskPopupArgs := []string{
		"display-popup", "-E", "-w", "80%", "-h", "80%",
		"-T", " ORC Desk ", deskPopup,
	}

	// Double-click status bar -> desk popup
	_ = srv.cmd(append([]string{"bind-key", "-T", "root", "DoubleClick1Status"}, deskPopupArgs...)...).Run()

	// prefix+u -> desk popup
	_ = srv.cmd(append([]string{"bind-key", "-T", "prefix", "u"}, deskPopupArgs...)...).Run()

	// Right-click status bar -> context menu
	_ = srv.BindContextMenu("MouseDown3Status", " ORC ", []MenuItem{
		// ORC custom options
		{Label: "Show Summary", Key: "s", Command: "display-popup -E -w 80% -h 80% -T ' ORC Desk ' '" + deskPopup + "'"},
		{Label: "Archive Workbench", Key: "a", Command: "display-popup -E -w 80 -h 20 -T 'Archive Workbench' 'cd #{pane_current_path} && orc tmux archive-workbench'"},
		// Separator
		{Label: "", Key: "", Command: ""},
		// Default tmux window options
		{Label: "Swap Left", Key: "<", Command: "swap-window -t :-1"},
		{Label: "Swap Right", Key: ">", Command: "swap-window -t :+1"},
		{Label: "#{?pane_marked,Unmark,Mark}", Key: "m", Command: "select-pane -m"},
		{Label: "Kill", Key: "X", Command: "kill-window"},
		{Label: "Respawn", Key: "R", Command: "respawn-window -k"},
		{Label: "Rename", Key: "r", Command: "command-prompt -I \"#W\" \"rename-window -- '%%'\""},
		{Label: "New Window", Key: "c", Command: "new-window"},
	})
}

// ApplyGlobalBindings sets up ORC's global tmux key bindings on the default server.
func ApplyGlobalBindings() {
	DefaultServer().ApplyGlobalBindings()
}

// SetEnvironment sets an environment variable for a tmux session on this server.
func (srv *Server) SetEnvironment(sessionName, key, value string) error {
	return srv.cmd("set-environment", "-t", exactSession(sessionName), key, value).Run()
}

// SetEnvironment sets an environment variable for a tmux session on the default server.
func SetEnvironment(sessionName, key, value string) error {
	return DefaultServer().SetEnvironment(sessionName, key, value)
}

// GetEnvironment gets an environment variable from a tmux session on this server.
// Returns the value, or error if not found.
func (srv *Server) GetEnvironment(sessionName, key string) (string, error) {
	output, err := srv.cmd("show-environment", "-t", exactSession(sessionName), key).Output()
	if err != nil {
		return "", err
	}
	// Output format: "KEY=value\n"
	line := strings.TrimSpace(string(output))
	if strings.HasPrefix(line, key+"=") {
		return strings.TrimPrefix(line, key+"="), nil
	}
	return "", fmt.Errorf("env var %s not found", key)
}

// GetEnvironment gets an environment variable from a tmux session on the default server.
func GetEnvironment(sessionName, key string) (string, error) {
	return DefaultServer().GetEnvironment(sessionName, key)
}

// ListSessions returns all tmux session names on this server.
func (srv *Server) ListSessions() ([]string, error) {
	output, err := srv.cmd("list-sessions", "-F", "#{session_name}").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var sessions []string
	for _, line := range lines {
		if line != "" {
			sessions = append(sessions, line)
		}
	}
	return sessions, nil
}

// ListSessions returns all tmux session names on the default server.
func ListSessions() ([]string, error) {
	return DefaultServer().ListSessions()
}

// FindSessionByWorkshopID finds the session with ORC_WORKSHOP_ID=workshopID on this server.
// Returns session name, or empty string if not found.
func (srv *Server) FindSessionByWorkshopID(workshopID string) string {
	sessions, err := srv.ListSessions()
	if err != nil {
		return ""
	}
	for _, session := range sessions {
		val, err := srv.GetEnvironment(session, "ORC_WORKSHOP_ID")
		if err == nil && val == workshopID {
			return session
		}
	}
	return ""
}

// FindSessionByWorkshopID finds the session with ORC_WORKSHOP_ID=workshopID on the default server.
func FindSessionByWorkshopID(workshopID string) string {
	return DefaultServer().FindSessionByWorkshopID(workshopID)
}

// GetWindowOption gets a window option value on this server.
// target format: "session:window" (e.g., "mysession:1" or "mysession:mywindow")
func (srv *Server) GetWindowOption(target, option string) string {
	output, err := srv.cmd("show-options", "-t", target, "-wqv", option).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// GetWindowOption gets a window option value on the default server.
func GetWindowOption(target, option string) string {
	return DefaultServer().GetWindowOption(target, option)
}

// SetWindowOption sets a window option value on this server.
// target format: "session:window" (e.g., "mysession:1" or "mysession:mywindow")
func (srv *Server) SetWindowOption(target, option, value string) error {
	return srv.cmd("set-option", "-t", target, "-w", option, value).Run()
}

// SetWindowOption sets a window option value on the default server.
func SetWindowOption(target, option, value string) error {
	return DefaultServer().SetWindowOption(target, option, value)
}

// ListWindows returns window names in a session on this server.
func (srv *Server) ListWindows(sessionName string) ([]string, error) {
	output, err := srv.cmd("list-windows", "-t", exactSession(sessionName), "-F", "#{window_name}").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var windows []string
	for _, line := range lines {
		if line != "" {
			windows = append(windows, line)
		}
	}
	return windows, nil
}

// ListWindows returns window names in a session on the default server.
func ListWindows(sessionName string) ([]string, error) {
	return DefaultServer().ListWindows(sessionName)
}

// PaneInfo contains information about a tmux pane
type PaneInfo struct {
	ID        string // e.g., "%0", "%1"
	Index     int    // pane index in window
	HasRole   bool   // whether @pane_role tmux option is set
	RoleValue string // value of @pane_role if set
}

// ListPanes returns information about all panes in a window on this server.
func (srv *Server) ListPanes(sessionName, windowName string) ([]PaneInfo, error) {
	target := exactTarget(sessionName, windowName)

	// Get pane IDs and indices
	output, err := srv.cmd("list-panes", "-t", target, "-F", "#{pane_id}:#{pane_index}").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list panes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	panes := make([]PaneInfo, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}

		paneID := parts[0]
		paneIndex, _ := strconv.Atoi(parts[1])

		// Check if @pane_role option is set (tmux pane option, set by gotmux adapter)
		// Note: we use @pane_role (tmux option) NOT PANE_ROLE (shell env var) because
		// tmux format strings cannot read shell environment variables.
		roleOutput, _ := srv.cmd("display-message", "-t", paneID, "-p", "#{@pane_role}").Output()
		role := strings.TrimSpace(string(roleOutput))

		paneInfo := PaneInfo{
			ID:        paneID,
			Index:     paneIndex,
			HasRole:   role != "",
			RoleValue: role,
		}

		panes = append(panes, paneInfo)
	}

	return panes, nil
}

// ListPanes returns information about all panes in a window on the default server.
func ListPanes(sessionName, windowName string) ([]PaneInfo, error) {
	return DefaultServer().ListPanes(sessionName, windowName)
}

// BreakPane breaks a pane into a new window on this server.
func (srv *Server) BreakPane(paneID, targetWindow string) error {
	return srv.cmd("break-pane", "-s", paneID, "-t", targetWindow).Run()
}

// BreakPane breaks a pane into a new window on the default server.
func BreakPane(paneID, targetWindow string) error {
	return DefaultServer().BreakPane(paneID, targetWindow)
}

// MoveWindowAfter moves a window to be positioned after another window on this server.
// If the window is already at afterIndex+1, the move is skipped.
// If the target index is occupied, the error is returned to the caller.
func (srv *Server) MoveWindowAfter(sessionName, windowName, afterWindow string) error {
	// Get the index of the afterWindow
	afterTarget := exactTarget(sessionName, afterWindow)
	output, err := srv.cmd("display-message", "-t", afterTarget, "-p", "#{window_index}").Output()
	if err != nil {
		return fmt.Errorf("failed to get window index for %s: %w", afterWindow, err)
	}

	afterIndex, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return fmt.Errorf("failed to parse window index: %w", err)
	}

	// Get the current index of the window being moved
	moveTarget := exactTarget(sessionName, windowName)
	output, err = srv.cmd("display-message", "-t", moveTarget, "-p", "#{window_index}").Output()
	if err != nil {
		return fmt.Errorf("failed to get window index for %s: %w", windowName, err)
	}

	currentIndex, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return fmt.Errorf("failed to parse current window index: %w", err)
	}

	// Already in the correct position -- nothing to do
	newIndex := afterIndex + 1
	if currentIndex == newIndex {
		return nil
	}

	return srv.cmd("move-window", "-s", moveTarget, "-t", fmt.Sprintf("=%s:%d", sessionName, newIndex)).Run()
}

// MoveWindowAfter moves a window to be positioned after another window on the default server.
func MoveWindowAfter(sessionName, windowName, afterWindow string) error {
	return DefaultServer().MoveWindowAfter(sessionName, windowName, afterWindow)
}

// SetPaneTitle sets the title of a pane using select-pane -T on this server.
func (srv *Server) SetPaneTitle(paneID, title string) error {
	return srv.cmd("select-pane", "-t", paneID, "-T", title).Run()
}

// SetPaneTitle sets the title of a pane on the default server.
func SetPaneTitle(paneID, title string) error {
	return DefaultServer().SetPaneTitle(paneID, title)
}

// EnrichSession applies ORC enrichment to all windows in a session on this server.
// This includes setting pane titles and window options (NOT PANE_ROLE - that must be set at pane creation)
func (srv *Server) EnrichSession(sessionName string) error {
	// Get all windows in the session
	windows, err := srv.ListWindows(sessionName)
	if err != nil {
		return fmt.Errorf("failed to list windows: %w", err)
	}

	// Process each window
	for _, window := range windows {
		if err := srv.enrichWindow(sessionName, window); err != nil {
			// Log warning but continue with other windows
			fmt.Printf("Warning: failed to enrich window %s: %v\n", window, err)
		}
	}

	return nil
}

// EnrichSession applies ORC enrichment to all windows in a session on the default server.
func EnrichSession(sessionName string) error {
	return DefaultServer().EnrichSession(sessionName)
}

// enrichWindow applies enrichment to a single window on this server.
func (srv *Server) enrichWindow(sessionName, windowName string) error {
	// Get all panes in the window
	panes, err := srv.ListPanes(sessionName, windowName)
	if err != nil {
		return fmt.Errorf("failed to list panes: %w", err)
	}

	for _, pane := range panes {
		if pane.HasRole {
			title := fmt.Sprintf("%s [%s]", pane.RoleValue, windowName)
			_ = srv.SetPaneTitle(pane.ID, title)
		}
	}

	// Set window option @orc_enriched=1 to mark as enriched
	target := exactTarget(sessionName, windowName)
	_ = srv.SetWindowOption(target, "@orc_enriched", "1")

	return nil
}

// DeskServerInfo describes a discovered desk tmux server.
type DeskServerInfo struct {
	Socket    string // socket name (e.g., "orc-45-desk")
	BenchName string // workbench name derived from socket (e.g., "orc-45")
	Alive     bool   // whether the server responds to commands
}

// ListDeskServers scans the tmux socket directory for *-desk sockets
// and probes each to determine if it's alive.
func ListDeskServers() ([]DeskServerInfo, error) {
	uid := os.Getuid()
	socketDir := fmt.Sprintf("/tmp/tmux-%d", uid)

	entries, err := os.ReadDir(socketDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read socket directory: %w", err)
	}

	var servers []DeskServerInfo
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, "-desk") {
			continue
		}

		benchName := strings.TrimSuffix(name, "-desk")
		alive := isDeskServerAlive(name)

		servers = append(servers, DeskServerInfo{
			Socket:    name,
			BenchName: benchName,
			Alive:     alive,
		})
	}

	return servers, nil
}

// isDeskServerAlive probes a desk server socket to check if it responds.
func isDeskServerAlive(socket string) bool {
	cmd := exec.Command("tmux", "-L", socket, "list-sessions")
	return cmd.Run() == nil
}

// KillDeskServer kills a specific desk server by workbench name.
func KillDeskServer(benchName string) error {
	socket := benchName + "-desk"

	// Verify socket exists
	uid := os.Getuid()
	socketPath := filepath.Join(fmt.Sprintf("/tmp/tmux-%d", uid), socket)
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return fmt.Errorf("desk server not found for workbench %q", benchName)
	}

	cmd := exec.Command("tmux", "-L", socket, "kill-server")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill desk server %s: %w", socket, err)
	}
	return nil
}

// KillAllDeskServers kills all discoverable desk servers.
func KillAllDeskServers() (int, error) {
	servers, err := ListDeskServers()
	if err != nil {
		return 0, err
	}

	killed := 0
	for _, s := range servers {
		if !s.Alive {
			continue
		}
		cmd := exec.Command("tmux", "-L", s.Socket, "kill-server")
		if err := cmd.Run(); err == nil {
			killed++
		}
	}
	return killed, nil
}
