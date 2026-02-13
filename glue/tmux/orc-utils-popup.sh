#!/usr/bin/env bash
# orc-utils-popup.sh — Lazy-create and attach to a workbench utils session
# Runs inside a display-popup. Creates the utils server/session on first open.
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

SOCKET="${BENCH_NAME}-utils"
SESSION="utils"

# Lazy create: if session doesn't exist on this server, build it
if ! tmux -L "$SOCKET" has-session -t "$SESSION" 2>/dev/null; then
    # summary: auto-refreshing dashboard (orc summary --poll is the root process)
    tmux -L "$SOCKET" new-session -d -s "$SESSION" -c "$BENCH_DIR" -n summary \
        "orc summary --poll"

    # shell: plain shell for scratch work
    tmux -L "$SOCKET" new-window -t "$SESSION" -n shell -c "$BENCH_DIR"
    tmux -L "$SOCKET" select-window -t "$SESSION:summary"

    # Mark this as a utils session so orc summary --poll can detect it
    tmux -L "$SOCKET" set-environment -t "$SESSION" ORC_UTILS_SESSION 1

    # Any click/double-click on status bar inside utils → detach (closes popup)
    tmux -L "$SOCKET" bind-key -T root DoubleClick1Status detach-client
    tmux -L "$SOCKET" bind-key -T root DoubleClick1StatusLeft detach-client
    tmux -L "$SOCKET" bind-key -T root MouseDown1StatusLeft detach-client
fi

# Respawn summary pane if it died (e.g., after q-quit or crash)
if tmux -L "$SOCKET" list-panes -t "$SESSION:summary" -F '#{pane_dead}' 2>/dev/null | grep -q 1; then
    tmux -L "$SOCKET" respawn-pane -k -t "$SESSION:summary" "orc summary --poll"
fi

# Attach — when user detaches (prefix+d), the popup closes via -E
exec tmux -L "$SOCKET" attach -t "$SESSION"
