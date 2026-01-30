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
- If "follow" or "f": run `tail -f` as a **background task**
- If a number: use as tail count
- Otherwise: default to 30

For follow mode, use the Bash tool with `run_in_background: true`:
```bash
tail -f ~/.claude/orc-debug.log
```
This streams new entries live. User can check output with TaskOutput or stop with TaskStop.

For count mode (default):
```bash
tail -N ~/.claude/orc-debug.log
```

Display the output to the user.
