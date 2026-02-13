#!/bin/bash
# SessionStart hook: detect ORC workbench and instruct agent to run /orc-prime
set -euo pipefail

# Read hook input from stdin
HOOK_INPUT=$(cat)

# Parse source field (startup, resume, clear, compact)
SOURCE=$(echo "$HOOK_INPUT" | jq -r '.source // "unknown"' 2>/dev/null || echo "unknown")

# Skip on resume and compact (those preserve session state)
if [ "$SOURCE" = "resume" ] || [ "$SOURCE" = "compact" ]; then
    exit 0
fi

# Check if cwd is an ORC workbench
CONFIG_FILE=".orc/config.json"
if [ ! -f "$CONFIG_FILE" ]; then
    exit 0
fi

PLACE_ID=$(jq -r '.place_id // ""' "$CONFIG_FILE" 2>/dev/null || echo "")
if [[ ! "$PLACE_ID" =~ ^BENCH- ]]; then
    exit 0
fi

# Workbench detected — instruct agent to bootstrap context
cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "You are in ORC workbench $PLACE_ID. Run /orc-prime to bootstrap project context."
  }
}
EOF
