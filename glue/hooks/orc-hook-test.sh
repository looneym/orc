#!/bin/bash
# ORC test hook - logs metadata about tool calls

LOG_FILE="$HOME/.claude/orc-hook.log"

# Read stdin (Claude Code passes JSON context)
INPUT=$(cat)

# Log timestamp + available context
{
  echo "=== $(date '+%Y-%m-%d %H:%M:%S') ==="
  echo "PWD: $PWD"
  echo "CLAUDE_*: $(env | grep CLAUDE || echo 'none')"
  echo "INPUT: $INPUT"
  echo ""
} >> "$LOG_FILE"
