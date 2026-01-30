#!/bin/bash
# ORC debug hook - logs all tool calls to ~/.claude/orc-debug.log
# Runs on PreToolUse, logs payload, returns immediately (non-blocking)

LOG_FILE="$HOME/.claude/orc-debug.log"
INPUT=$(cat)

{
  echo "=== $(date '+%Y-%m-%d %H:%M:%S') ==="
  echo "$INPUT" | jq -c '{tool: .tool_name, session: .session_id[0:8], input: .tool_input}' 2>/dev/null || echo "$INPUT"
  echo ""
} >> "$LOG_FILE"
