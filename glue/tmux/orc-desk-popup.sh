#!/usr/bin/env bash
# orc-desk-popup.sh — Lazy-create and attach to a workbench desk session
# Runs inside a display-popup. Creates the desk server/session on first open.
#
# Desk windows:
#   ledger  — Bubble Tea TUI dashboard (orc summary --tui)
#   shell   — scratch terminal for ad-hoc commands
#   git     — vim with fugitive (vim +Git)
#
# Note: display-popup does NOT expand #{} format variables in shell-command,
# so we query the main tmux server directly via TMUX env var.

set -euo pipefail

# Query the main tmux server for the current window name and pane path.
# Inside a display-popup, TMUX is set and display-message targets the popup's parent pane.
BENCH_NAME="$(tmux display-message -p '#{window_name}')"
BENCH_DIR="$(tmux display-message -p '#{pane_current_path}')"

if [[ -z "$BENCH_NAME" || -z "$BENCH_DIR" ]]; then
    echo "ERROR: could not resolve window_name or pane_current_path from tmux"
    sleep 3
    exit 1
fi

SOCKET="${BENCH_NAME}-desk"
SESSION="desk"

# Lazy create: if session doesn't exist on this server, build it
if ! tmux -L "$SOCKET" has-session -t "$SESSION" 2>/dev/null; then
    # ledger: interactive TUI dashboard (orc summary --tui is the root process)
    # CLICOLOR_FORCE=1 ensures colored output even though tmux pipe isn't a real tty
    tmux -L "$SOCKET" new-session -d -s "$SESSION" -c "$BENCH_DIR" -n ledger \
        -e CLICOLOR_FORCE=1 \
        "orc summary --tui"

    # shell: plain shell for scratch work
    tmux -L "$SOCKET" new-window -t "$SESSION" -n shell -c "$BENCH_DIR"

    # git: vim with fugitive for git operations
    tmux -L "$SOCKET" new-window -t "$SESSION" -n git -c "$BENCH_DIR" \
        "vim +Git"

    # Select ledger as the default window
    tmux -L "$SOCKET" select-window -t "$SESSION:ledger"

    # Mark this as a desk session so orc summary --tui can detect it
    tmux -L "$SOCKET" set-environment -t "$SESSION" ORC_DESK_SESSION 1

    # Any click/double-click on status bar inside desk → detach (closes popup)
    tmux -L "$SOCKET" bind-key -T root DoubleClick1Status detach-client
    tmux -L "$SOCKET" bind-key -T root DoubleClick1StatusLeft detach-client
    tmux -L "$SOCKET" bind-key -T root MouseDown1StatusLeft detach-client
fi

# Respawn ledger pane if it died (e.g., after q-quit or crash)
if tmux -L "$SOCKET" list-panes -t "$SESSION:ledger" -F '#{pane_dead}' 2>/dev/null | grep -q 1; then
    tmux -L "$SOCKET" respawn-pane -k -t "$SESSION:ledger" "orc summary --tui"
fi

# Respawn git pane if it died (e.g., after :q in vim)
if tmux -L "$SOCKET" list-panes -t "$SESSION:git" -F '#{pane_dead}' 2>/dev/null | grep -q 1; then
    tmux -L "$SOCKET" respawn-pane -k -t "$SESSION:git" "vim +Git"
fi

# Attach — when user detaches (prefix+d), the popup closes via -E
exec tmux -L "$SOCKET" attach -t "$SESSION"
