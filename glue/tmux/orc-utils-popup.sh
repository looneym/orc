#!/usr/bin/env bash
# orc-utils-popup.sh — Lazy-create and attach to a workbench utils session
# Runs inside a display-popup. Creates the utils server/session on first open.
#
# Usage: orc-utils-popup.sh <workbench-name> <workbench-dir>
# Example: orc-utils-popup.sh orc-45 /Users/looneym/wb/orc-45

set -euo pipefail

BENCH_NAME="${1:?Usage: orc-utils-popup.sh <bench-name> <bench-dir>}"
BENCH_DIR="${2:?Usage: orc-utils-popup.sh <bench-name> <bench-dir>}"
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

    # Any click/double-click on status bar inside utils → detach (closes popup)
    tmux -L "$SOCKET" bind-key -T root DoubleClick1Status detach-client
    tmux -L "$SOCKET" bind-key -T root DoubleClick1StatusLeft detach-client
    tmux -L "$SOCKET" bind-key -T root MouseDown1StatusLeft detach-client
fi

# Attach — when user detaches (prefix+d), the popup closes via -E
exec tmux -L "$SOCKET" attach -t "$SESSION"
