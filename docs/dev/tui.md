# TUI Development Guide

The summary TUI (`orc summary --tui`) is an interactive Bubble Tea interface for navigating
the commission/shipment/tome hierarchy. This document covers the entity-action matrix,
key bindings, and how to extend the TUI.

## Entity-Action Matrix

The `entityActionMatrix` in `internal/cli/summary_tui.go` is the **single source of truth**
for which actions each entity type supports. Every keybind guard and status bar hint must
use `entityHasAction(id, "action")` — no standalone boolean functions.

| Entity | yank | open | focus | close | goblin | note | review | run | deploy | expand |
|--------|------|------|-------|-------|--------|------|--------|-----|--------|--------|
| COMM   | T    | T    | T     |       | T      |      |        |     |        | T      |
| SHIP   | T    | T    | T     | T     | T      | T    |        | T   | T      | T      |
| TASK   | T    | T    |       | T     | T      |      |        |     |        |        |
| NOTE   | T    | T    | T     |       | T      |      | T      |     |        |        |
| TOME   | T    | T    | T     |       | T      | T    |        |     |        | T      |
| PLAN   | T    | T    |       |       | T      |      |        |     |        |        |
| WORK   | T    |      |       |       | T      |      |        |     |        |        |
| BENCH  | T    |      |       |       | T      |      |        |     |        |        |

### Action Semantics

- **yank** (y): Copy entity ID to clipboard
- **open** (o): Open entity details in vim (read-only)
- **focus** (f): Set/toggle focus on entity (controls tree expansion and context)
- **close** (c): Complete/close entity (with y/n confirmation)
- **goblin** (g): Send entity to goblin pane (requires runtime context: workshop session or desk session)
- **note** (n): Create child note attached to entity (opens $EDITOR for title)
- **review** (d): Open desk review window for entity
- **run** (R): Copy `/ship-run` command to clipboard
- **deploy** (D): Copy `/ship-deploy` command to clipboard
- **expand** (enter/l): Toggle expand/collapse of entity's children in tree

### Runtime Context for Goblin

The goblin action has a compound condition: the entity must have `goblin: true` in the matrix
AND at least one of these runtime conditions must be true:
- `m.workshopSession != ""` (has a workshop tmux session)
- `m.isDeskSession` (running inside a desk popup)

## Key Reference

| Key | Action | Condition |
|-----|--------|-----------|
| j/k, arrows | Navigate cursor | Always |
| enter, l | Expand/collapse | `entityHasAction(id, "expand")` |
| y | Yank to clipboard | `entityHasAction(id, "yank")` |
| o | Open in vim | `entityHasAction(id, "open")` |
| f | Focus/unfocus | `entityHasAction(id, "focus")` |
| c | Close (confirm) | `entityHasAction(id, "close")` |
| n | Create note | `entityHasAction(id, "note")` |
| d | Desk review | `entityHasAction(id, "review")` |
| R | Ship run | `entityHasAction(id, "run")` |
| D | Ship deploy | `entityHasAction(id, "deploy")` |
| g | Goblin send | `entityHasAction(id, "goblin")` + runtime |
| r | Refresh | Always |
| q/esc | Quit (or detach in desk) | Always |

## Desk Mode Behavior

When running inside a desk tmux session (`isDeskSession = true`):
- `q`/`esc` detaches the tmux client (closes the popup) instead of quitting
- `ctrl+c` still quits the process
- Review (`d`) opens an ephemeral tmux window instead of running inline
- Goblin (`g`) uses `sendToGoblinByBenchID` to find the parent server's goblin pane

## How to Add a New TUI Action

Follow this checklist (also in [checklists.md](checklists.md)):

1. Add action to `entityActionMatrix` in `summary_tui.go` for each eligible entity type
2. Add key handler in the `Update` switch block — guard with `entityHasAction(id, "action")`
3. Add status bar hint in `renderStatusBar` — use `formatHint("key", "label", entityHasAction(...))`
4. Add action to `allActions` slice in `TestEntityActionMatrixCompleteness`
5. Update the matrix table in this document (`docs/dev/tui.md`)
6. Run `make test && make lint`

## How to Add a New Entity Type

1. Add entity type regex to `entityIDPattern` in `summary_tui.go`
2. Add row to `entityActionMatrix` with all supported actions
3. Add entity to `entityShowCommand` switch
4. Add entity type to `allEntityTypes` slice in `TestEntityActionMatrixCompleteness`
5. Update the matrix table in this document (`docs/dev/tui.md`)
6. Run `make test && make lint`
