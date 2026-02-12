# Shipment & Task Lifecycles

**Status**: Living document
**Last Updated**: 2026-02-11

Shipments and tasks use simple, manual lifecycles. All transitions are decided by the Goblin (coordinator). There are no auto-transitions.

---

## Shipment Lifecycle

### State Diagram

```mermaid
stateDiagram-v2
    [*] --> draft: create shipment
    draft --> ready: Goblin marks ready
    ready --> in_progress: Goblin starts work
    in_progress --> closed: Goblin closes
    closed --> [*]

    in_progress: in-progress
```

### States

| State | Description | Next Step |
|-------|-------------|-----------|
| `draft` | Shipment created, not yet ready for work | Mark ready when scoped |
| `ready` | Scoped and ready for implementation | Start work |
| `in-progress` | Active implementation | Close when complete |
| `closed` | Terminal state | -- |

All transitions are manual. The Goblin decides when to advance.

---

## Task Lifecycle

### State Diagram

```mermaid
stateDiagram-v2
    [*] --> open: create task
    open --> in_progress: start work
    in_progress --> blocked: pause
    blocked --> in_progress: resume
    in_progress --> closed: complete work
    closed --> [*]

    in_progress: in-progress
```

### States

| State | Description | Next Step |
|-------|-------------|-----------|
| `open` | Task created, available for work | Start work |
| `in-progress` | Actively being worked on | Close when done |
| `blocked` | Work paused due to external dependency | Resume when unblocked |
| `closed` | Terminal state | -- |

```bash
orc task pause TASK-xxx     # Transition to blocked
orc task resume TASK-xxx   # Transition back to in-progress
```

---

## See Also

- [docs/schema.md](schema.md) - Database schema with all valid states
- [internal/core/shipment/guards.go](../../internal/core/shipment/guards.go) - Guard implementations
