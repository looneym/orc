# Events & Operational Diagnostics

ORC's unified event system provides two distinct event types: **audit events** (entity CRUD trail) and **operational events** (runtime diagnostics and lifecycle).

## Event Types

### Audit Events
Automatically track all entity CRUD operations. Emitted by adapters during persistence operations.

**Source**: Always `ledger` (see `internal/core/event/sources.go`)

**When used**: Entity create, update, delete operations (shipments, tasks, tomes, etc.)

**Storage**: `workshop_events` table in SQLite

**Format**:
- EntityType, EntityID (e.g., `shipment`, `SHIP-042`)
- Action: `create`, `update`, `delete`
- FieldName, OldValue, NewValue (for updates)

### Operational Events
Runtime diagnostics, lifecycle events, debug traces. Emitted from app layer services.

**Source**: Variable (poll, tmux-apply, deploy-glue, workbench)

**When used**: Poll mode diagnostics, tmux session lifecycle, glue deployment traces, workbench operations

**Storage**: `operational_events` table in SQLite

**Format**:
- Level: `debug`, `info`, `warn`, `error`
- Message: Human-readable description
- Data: `map[string]string` for structured context

## BaseEvent (Shared Foundation)

All events inherit these fields:

| Field | Description |
|-------|-------------|
| ID | Unique identifier (WE-xxx for audit, OE-xxx for ops) |
| Timestamp | RFC3339 timestamp |
| Actor | Actor ID (IMP-BENCH-014, GOBLIN, etc.) |
| Workshop | Workshop ID (nullable for non-workshop operations) |
| Source | Origin constant (see below) |
| Version | ORC version string (for forward compatibility) |

Workshop resolution: For `BENCH-xxx` actors, EventWriter looks up the workbench's workshop. For non-workshop actors or workbench-not-found cases, Workshop is empty (operational events support this).

## Source Constants

Defined in `internal/core/event/sources.go`:

| Constant | Value | When to Use |
|----------|-------|-------------|
| `SourceLedger` | `ledger` | Audit events (automatic) |
| `SourcePoll` | `poll` | Poll mode diagnostics |
| `SourceTmuxApply` | `tmux-apply` | TMux session lifecycle |
| `SourceDeployGlue` | `deploy-glue` | Glue deployment traces |
| `SourceWorkbench` | `workbench` | Workbench operations |
| `SourceSummaryTUI` | `summary-tui` | Interactive summary TUI key actions |

**When to add new sources**: If you're implementing a new subsystem or mode that warrants isolated filtering, add a new source constant. Sources enable targeted debugging (e.g., `orc events tail --source poll`).

## Level Constants

Defined in `internal/core/event/sources.go`:

| Level | When to Use |
|-------|-------------|
| `LevelDebug` | Verbose traces (hidden by default) |
| `LevelInfo` | Normal operation milestones |
| `LevelWarn` | Recoverable issues |
| `LevelError` | Failure conditions |

**Default filtering**: `orc events tail` shows `info+` by default. Use `--level debug` for full firehose.

## How to Emit Operational Events

**1. Inject EventWriter via wire**

```go
type MyService struct {
    eventWriter secondary.EventWriter
}

func NewMyService(eventWriter secondary.EventWriter) *MyService {
    return &MyService{eventWriter: eventWriter}
}
```

**2. Call EmitOperational**

```go
data := map[string]string{
    "shipment_id": "SHIP-042",
    "poll_count":  "3",
}

err := s.eventWriter.EmitOperational(
    ctx,
    event.SourcePoll,
    event.LevelInfo,
    "poll completed successfully",
    data,
)
```

**Actor extraction**: EventWriter automatically extracts actor from context (`ctxutil.ActorFromContext(ctx)`). Ensure actor is set via `ctxutil.WithActor(ctx, actorID)` before service calls.

## How to Query Events

### Tail recent events
```bash
# Default: last 50 events, info+ only
orc events tail

# Show debug events (firehose)
orc events tail --level debug

# Filter by source
orc events tail --source poll

# Filter by type
orc events tail --type ops
orc events tail --type audit

# Filter by actor
orc events tail --actor IMP-BENCH-014

# Follow mode (live tail)
orc events tail --follow
```

### Show events for an entity
```bash
# Entity-specific history
orc events show SHIP-042
orc events show TASK-001
```

### Prune old events
```bash
# Delete events older than 30 days (default)
orc events prune

# Custom retention period
orc events prune --days 7
```

## Practical Examples

### Debugging a Poll Mode Failure

**Scenario**: Poll mode isn't detecting changes in a tome.

**Steps**:
1. `orc events tail --source poll --level debug`
2. Look for poll cycle start/end messages
3. Check Data field for tome paths and detection results
4. Correlate timestamps with expected file changes

### Tracing a TMux Apply

**Scenario**: TMux session creation failed.

**Steps**:
1. `orc events tail --source tmux-apply`
2. Look for session creation, window setup, and error events
3. Check Data field for session IDs and command output
4. Filter by shipment: `orc events tail --source tmux-apply | grep SHIP-042`

### Diagnosing Deploy-Glue

**Scenario**: Glue deployment silently failed.

**Steps**:
1. `orc events tail --source deploy-glue`
2. Look for glue sync start/end events
3. Check for warn/error level events with deployment details
4. Verify deployment timestamps match expected deployment window

## Storage & Retention

**Tables**:
- `workshop_events`: Audit events
- `operational_events`: Operational events

**Retention**: Both tables share unified retention via `orc events prune --days N`.

**Performance**: Events are indexed by timestamp, source, and actor for fast queries. No pagination in CLI (yet) — use `--limit` to control result size.

## Architecture Notes

**Port**: `internal/ports/secondary/event.go` — EventWriter interface

**Adapter**: `internal/adapters/sqlite/event_writer.go` — SQLite persistence

**Service**: `internal/app/event_service.go` — Query and pruning logic

**CLI**: `internal/cli/events.go` — User-facing commands

**Core types**: `internal/core/event/sources.go` — BaseEvent, AuditEvent, OperationalEvent

**Why two event types?** Audit events track *what changed* in the ledger (entity state). Operational events track *how the system behaves* at runtime (diagnostics). Different query patterns, different retention needs, unified interface.
