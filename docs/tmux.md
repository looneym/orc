# TMux Integration

**Status**: Living document
**Last Updated**: 2026-02-14

ORC uses TMux with [gotmux](https://github.com/GianlucaP106/gotmux) for programmatic multi-agent session management. Each workshop has a dedicated TMux session with windows for workbenches.

## Workflow Overview

The ORC tmux workflow follows this sequence:

1. **Create workbench** -- `orc workbench create` (creates DB record + worktree + config)
2. **Apply tmux session** -- `orc tmux apply WORK-xxx [--yes]` (creates/reconciles session)
3. **Connect** -- `orc tmux connect WORK-xxx` (attaches to session)

The `apply` command is the single entry point for all tmux session management. It compares
desired state (from the database) with actual tmux state and reconciles the difference.

## Gotmux Integration

ORC uses gotmux's Go API for programmatic tmux session management. Sessions are created directly via the library, not via YAML config files.

**Key operations:**
- `orc tmux apply WORK-xxx` - creates/reconciles session (plan + confirm)
- `orc tmux apply WORK-xxx --yes` - creates/reconciles session (immediate)
- `orc tmux connect WORK-xxx` - attaches to existing session
- `orc tmux enrich WORK-xxx` - re-applies enrichment (idempotent)
- `tmux kill-session -t WORK-xxx` - stops session (standard tmux)

**Benefits:**
- Direct Go API control (no YAML generation)
- Type-safe session manipulation
- Programmatic pane creation with pane options
- Plan/apply pattern for safe reconciliation
- Maintained library with stable API

## Window/Pane Layouts

### Standard Workbench Window

Each workbench window has a single pane running the goblin (coordinator agent):

```
+------------------------------------------+
|                                          |
|              goblin                      |
|         (orc connect)                    |
|                                          |
+------------------------------------------+
```

| Pane | Role (@pane_role) | Purpose |
|------|-------------------|---------|
| 0 | `goblin` | Coordinator pane (full window) |

The goblin pane is launched via `orc connect` using `respawn-pane -k`, making it the pane's
root process. The desk popup (see below) provides shell and vim access.

## Pane Identity Model

ORC uses tmux pane options for pane identity:

- **`@pane_role`** (authoritative) -- set via `pane.SetOption("@pane_role", role)` at creation. Readable by tmux format strings (`#{@pane_role}`). Used by ORC to identify panes.
- **`@bench_id`** -- workbench ID (e.g., `BENCH-001`). Provides workbench context.
- **`@workshop_id`** -- workshop ID (e.g., `WORK-001`). Provides workshop context.

| Role | Description |
|------|-------------|
| `goblin` | Coordinator pane (one per workbench window) |

**Note**: Shell environment variables (`PANE_ROLE`, etc.) are NOT used. All identity is via tmux pane options which are readable by tmux format strings.

## TMux Commands

### orc tmux apply

Reconciles tmux session state for a workshop. This is the single command for creating,
updating, and maintaining sessions.

```bash
orc tmux apply WORK-xxx          # Show plan, prompt for confirmation
orc tmux apply WORK-xxx --yes    # Apply immediately
```

**What it does:**
1. Compares desired state (workbenches from DB) with actual tmux state
2. Creates session if it doesn't exist
3. Adds windows for missing workbenches (single goblin pane each)
4. Applies ORC enrichment (bindings, pane titles)

### orc tmux connect

Attaches to an existing workshop session.

```bash
orc tmux connect WORK-xxx
```

### orc tmux enrich

Applies ORC-specific tmux bindings and enrichment to the current or specified session.

```bash
orc tmux enrich              # Enrich detected workshop session
orc tmux enrich WORK-xxx     # Enrich specific workshop
```

**What it does:**
- Configures mouse support
- Sets up ORC key bindings (session picker, summary popup, etc.)
- Applies statusline formatting
- Enables context display

Note: `orc tmux apply` runs enrichment automatically. Use `enrich` only when you need to
re-apply enrichment without full reconciliation.

## Session Management

ORC creates tmux sessions programmatically via the gotmux Go library. There are no configuration files - sessions are created directly from DB state.

To inspect or modify sessions, use standard tmux commands:
```bash
tmux list-sessions                    # Show all sessions
tmux list-windows -t WORK-xxx         # Show windows in session
tmux kill-session -t WORK-xxx         # Stop session
```

## Session Browser

### Standard Session Browser (prefix+s)

Press `prefix+s` (default: Ctrl-b, then s) to open TMux's session browser.

Format shows ORC context:
```
session1 [WORK-001] - Project Name [COMM-001]
session2 [WORK-002] - Other Project [COMM-002]
```

### ORC Session Picker (prefix+S)

Press `prefix+S` (capital S) for enhanced ORC picker:

- Vertical split with list and preview
- Shows agent type (IMP/GOBLIN) per window
- Shows current focus (-> SHIP-xxx)
- Live preview of selected pane content

Navigation:
- `j`/`k` or arrows to move
- `Enter` or `l` to select
- `q` to cancel

## Desk Popup

The desk popup is a persistent, per-workbench overlay that provides an interactive
summary TUI, a scratch shell, and a git window. It runs in its own tmux server,
separate from the main workshop session.

### Opening and Closing

| Action | How |
|--------|-----|
| Open | Double-click the status bar, or `prefix+u` |
| Close | Double-click the inner status bar, click the inner session name, or `prefix+d` (detach) |

When opened for the first time, the popup lazily creates a dedicated tmux server and session.
Subsequent opens reattach to the existing session, preserving shell history and scroll position.

### What's Inside

The desk session has three permanent windows:

| Window | Content |
|--------|---------|
| `ledger` | `orc summary --tui` -- interactive TUI dashboard with keyboard navigation |
| `shell` | Plain shell for scratch work (ad-hoc commands, grep, etc.) |
| `git` | `vim +Git` -- vim with fugitive for git operations |

The ledger window uses `orc summary --tui` as its root process, which runs an interactive
Bubble Tea TUI for browsing the commission/shipment tree. Use `r` to refresh on demand.
SIGTERM/SIGINT exit cleanly with no stack trace. Dead ledger and git panes are automatically
respawned when the popup is reopened.

### Separate Server Architecture

Each workbench gets its own tmux server via the `-L` (socket) flag:

```
Socket: {bench-name}-desk     (e.g., orc-45-desk)
Path:   /tmp/tmux-{uid}/orc-45-desk
```

**Why a separate server?**

- **No prefix collision** -- the desk session can have its own key bindings (e.g., double-click
  to detach) without conflicting with the main session's bindings.
- **No session list pollution** -- `tmux list-sessions` and `prefix+s` only show workshop
  sessions, not desk sessions.
- **Isolation** -- killing or restarting the desk server has no effect on the main workshop
  session or any running agents.

The popup script (`orc-desk-popup.sh`) handles lazy creation: if the server/session doesn't
exist, it creates it with the ledger, shell, and git windows. If it already exists, it simply
reattaches.

### Managing Desk Servers

Use `orc desk` to inspect and clean up desk servers:

```bash
orc desk list              # Show all desk servers and status
orc desk kill orc-45       # Kill a specific desk server
orc desk kill --all        # Kill all desk servers
```

Example output of `list`:

```
WORKBENCH  SOCKET       STATUS
orc-45     orc-45-desk  alive
myproj     myproj-desk  dead
```

Desk servers are also automatically killed when archiving a workshop via
`orc workshop archive`.

### Right-Click Context Menu

Right-click the statusline for a context menu:

| Option | Action |
|--------|--------|
| Show Summary | Opens the desk popup |
| Swap Left/Right | Standard TMux swap |
| Mark Pane | Standard TMux mark |
| Kill Window | Standard TMux kill |
| Respawn | Restart window with start command |
| Rename | Standard TMux rename |
| New Window | Standard TMux new window |

## Summary TUI Mode

The `--tui` flag on `orc summary` launches an interactive Bubble Tea TUI:

```bash
orc summary --tui
```

This is primarily used inside the desk popup's ledger window, but can also be run
standalone in any terminal. The TUI renders the commission/shipment tree with keyboard
navigation instead of a timer-based refresh loop.

### TUI Keybinds

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate up/down through items |
| `Enter` | Expand or collapse a section |
| `y` | Yank (copy) the ID of the selected item |
| `f` | Focus on the selected shipment |
| `o` | Open the selected item in vim |
| `c` | Complete the selected task (with y/n confirmation) |
| `n` | Create a note on the selected entity |
| `R` | Copy `/ship-run SHIP-xxx` to clipboard (shipments only) |
| `D` | Copy `/ship-deploy` to clipboard (shipments only) |
| `g` | Send a message to the goblin pane (opens editor) |
| `r` | Refresh the summary data |
| `q` | Quit the TUI |

Color output is preserved in TUI mode via `CLICOLOR_FORCE=1`, which is set by the
desk popup environment.

### Goblin Communication

When running inside a desk popup, the TUI can send messages directly to the goblin
pane in the main workshop session. Press `g` to compose a message in `$EDITOR` -- on
save, the text is sent to the goblin pane via `tmux send-keys`. This allows quick
instructions or context injection from the desk without switching windows.

## Session Click

Click on the session name in the statusline (left side) to open the session browser. This is equivalent to `prefix+s`.

## Environment Variables

ORC sets these environment variables on TMux sessions:

| Variable | Scope | Purpose |
|----------|-------|---------|
| `ORC_WORKSHOP_ID` | Session | Workshop ID (WORK-xxx) |
| `ORC_CONTEXT` | Session | Active commissions summary |
| `ORC_DESK_SESSION` | Session (desk server) | Marks desk sessions (enables TUI goblin communication) |

These enable the session browser to show ORC context.

## Window Options

ORC tracks agent state via TMux window options:

| Option | Example | Purpose |
|--------|---------|---------|
| `@orc_agent` | `IMP-main@BENCH-001` | Agent identity |
| `@orc_focus` | `SHIP-334: Docs overhaul` | Current focus |

These enable the ORC session picker to show agent details.

## Common Operations

### Connect to Workshop

```bash
orc tmux connect WORK-xxx
```

Finds and attaches to the workshop's TMux session.

### Start a New Workshop Session

```bash
# 1. Create workshop/workbench (if needed)
orc workbench create --workshop WORK-001 --repo-id REPO-001

# 2. Apply tmux session (creates + enriches in one step)
orc tmux apply WORK-001 --yes

# 3. Attach
orc tmux connect WORK-001
```

### Stop a Window

```bash
tmux kill-window -t "Workshop Name:bench-1"
```

Stops a specific workbench window while leaving the rest of the session running.

### Switch Between Workshops

1. `prefix+S` - Open ORC session picker
2. Navigate to desired workshop/window
3. Press Enter to switch

### View Workshop State

1. Double-click the status bar (opens desk popup with TUI dashboard)
2. Or: `prefix+u` to open the desk popup directly

## Next Steps

- [docs/common-workflows.md](common-workflows.md) - IMP/Goblin workflows
- [docs/dev/glue.md](dev/glue.md) - Skills that interact with TMux
- [docs/architecture.md](architecture.md) - System design
