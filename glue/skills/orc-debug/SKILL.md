---
name: orc-debug
description: View recent Claude Code tool calls from the ORC debug log. Use when you want to see what tools have been called, debug hook behavior, or inspect tool usage across sessions.
---

# ORC Debug Log

Show recent tool calls from the debug log.

## Usage

`/orc-debug` - show last 30 entries
`/orc-debug 50` - show last 50 entries
`/orc-debug follow` - live stream new entries (tail -f)

## Action

Parse the argument:
- If "follow" or "f": run `tail -f ~/.claude/orc-debug.log` in background
- If a number: use as tail count
- Otherwise: default to 30

For follow mode, run in background so user can continue working:
```bash
tail -f ~/.claude/orc-debug.log
```

For count mode:
```bash
tail -N ~/.claude/orc-debug.log
```

Display the output to the user.
