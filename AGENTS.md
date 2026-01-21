# AGENTS.md - Development Rules for Claude Agents

This file contains essential development workflow rules for Claude agents working on the ORC codebase.

## Build & Development

**ALWAYS use the Makefile for building and installing ORC:**

```bash
make dev        # Build local ./orc for development (preferred)
make install    # Build and install globally with local-first shim
make test       # Run all tests
make lint       # Run golangci-lint + architecture linting
make clean      # Clean build artifacts
```

### Binary Management Convention

When developing ORC itself, **always use `./orc`** (the local binary):

```bash
# In the ORC repo:
make dev                          # Build local binary
./orc status                      # Use local binary
./orc summary --mission current   # Use local binary

# The shim displays "[using local ./orc]" to confirm local usage
```

**Why this matters:**
- The local-first shim prefers `./orc` when present
- This ensures you're testing your actual changes
- Prevents confusion between global and development binaries
- `make dev && ./orc <cmd>` is the canonical development workflow

## Architecture Rules

ORC follows a hexagonal (ports & adapters) architecture with strict layer boundaries.

### Layer Hierarchy

```
┌─────────────────────────────────────────────────────────┐
│                        cmd/                             │
│                   (entry points)                        │
├─────────────────────────────────────────────────────────┤
│                      cli/                               │
│              (Cobra commands, thin)                     │
├─────────────────────────────────────────────────────────┤
│                      wire/                              │
│           (dependency injection only)                   │
├─────────────────────────────────────────────────────────┤
│    adapters/              │           app/              │
│  (implementations)        │     (orchestration)         │
├───────────────────────────┴─────────────────────────────┤
│                      ports/                             │
│               (interfaces only)                         │
├─────────────────────────────────────────────────────────┤
│                      core/                              │
│        (pure domain logic, no dependencies)             │
└─────────────────────────────────────────────────────────┘
```

### Allowed Imports

| Layer | Can Import |
|-------|------------|
| `core/` | stdlib only (no ORC imports) |
| `ports/` | stdlib only (no ORC imports) |
| `app/` | `core/`, `ports/`, `models/` |
| `adapters/` | `ports/`, `models/`, `db/` |
| `wire/` | `adapters/`, `app/`, `ports/` |
| `cli/` | `wire/`, `ports/` |
| `cmd/` | `cli/`, `wire/` |

### Architecture Principles

1. **"Core is pure"** - Domain logic in `core/` has zero external dependencies
2. **"Adapters are boring"** - No business logic in adapters; just translation
3. **"CLI is thin"** - Commands call services, they don't orchestrate
4. **"Ports are contracts"** - Interfaces define boundaries, not implementations

Run `make lint` to verify architecture compliance.

## Testing Rules

### FSM-First Development

Every workflow entity requires an FSM spec in `specs/`:

```
specs/
├── mission-provisioning.yaml
├── grove-provisioning.yaml
├── shipment-workflow.yaml
├── task-workflow.yaml
└── ... (9 specs total)
```

**Workflow for new state machines:**
1. Write the FSM spec (YAML) first
2. Generate test matrix from spec
3. Implement guards in `core/<entity>/guards.go`
4. Implement service in `app/<entity>_service.go`

### Table-Driven Tests (Default Pattern)

All tests should use the table-driven pattern:

```go
func TestCanPauseTask(t *testing.T) {
    tests := []struct {
        name        string
        ctx         StatusTransitionContext
        wantAllowed bool
        wantReason  string
    }{
        {
            name: "can pause in_progress task",
            ctx:  StatusTransitionContext{TaskID: "TASK-001", Status: "in_progress"},
            wantAllowed: true,
        },
        {
            name: "cannot pause ready task",
            ctx:  StatusTransitionContext{TaskID: "TASK-001", Status: "ready"},
            wantAllowed: false,
            wantReason:  "can only pause in_progress tasks (current status: ready)",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CanPauseTask(tt.ctx)
            if result.Allowed != tt.wantAllowed {
                t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
            }
            if !tt.wantAllowed && result.Reason != tt.wantReason {
                t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
            }
        })
    }
}
```

### Test Pyramid

```
┌─────────────────────────────┐
│     Integration Tests       │  ← Sparse: cross-repo scenarios
├─────────────────────────────┤
│     Repository Tests        │  ← Medium: SQL correctness
├─────────────────────────────┤
│      Service Tests          │  ← Most: orchestration logic
├─────────────────────────────┤
│       Guard Tests           │  ← Foundation: pure functions
└─────────────────────────────┘
```

### Test Helpers

Use `testutil_test.go` helpers in `internal/adapters/sqlite/`:
- `setupIntegrationDB(t)` - Creates in-memory DB with all tables
- `seedMission(t, db, id, title)` - Insert test mission
- `seedGrove(t, db, id, missionID, name)` - Insert test grove
- `seedTag(t, db, id, name)` - Insert test tag

## Verification Discipline

### Plans Must Include Checks

Every implementation plan should explicitly list:
- [ ] Tests to run
- [ ] Lint checks to pass
- [ ] Manual verification steps

### Completion Reports What Ran

When completing work, report:
```
✅ Ran: make test (all passing)
✅ Ran: make lint (no issues)
⏭️ Skipped: integration tests (not applicable)
```

Never claim success without actually running verification.

---

## Checklists

### Add Field to Entity

When adding a new field to an existing entity (e.g., adding `priority` to Task):

- [ ] Update model struct in `internal/models/<entity>.go`
- [ ] Update FSM spec if field affects state transitions (`specs/<entity>-workflow.yaml`)
- [ ] Update SQL schema in `internal/db/schema.sql`
- [ ] Create migration in `internal/db/migrations/`
- [ ] Update repository:
  - [ ] `internal/adapters/sqlite/<entity>_repo.go` - CRUD operations
  - [ ] `internal/adapters/sqlite/<entity>_repo_test.go` - repo tests
- [ ] Update service if field has business logic:
  - [ ] `internal/app/<entity>_service.go`
  - [ ] `internal/app/<entity>_service_test.go`
- [ ] Update CLI if field is user-facing:
  - [ ] Add flag to relevant commands
  - [ ] Update help text
- [ ] Update `testutil_test.go` if test helpers need the field
- [ ] Run: `make test && make lint`

### Add State/Transition to FSM

When adding a new state or transition to an entity's state machine:

- [ ] Update FSM spec (`specs/<entity>-workflow.yaml`):
  - [ ] Add new state to `states:` section
  - [ ] Add new event to `events:` section (if needed)
  - [ ] Add transition(s) to `transitions:` section
  - [ ] Define guards in `guards:` section
  - [ ] Add test cases to transition
- [ ] Update core guards (`internal/core/<entity>/guards.go`):
  - [ ] Add guard context struct if needed
  - [ ] Implement guard function
- [ ] Update core guards tests (`internal/core/<entity>/guards_test.go`):
  - [ ] Add table-driven tests for new guard
  - [ ] Cover all paths (allow and deny cases)
- [ ] Update service (`internal/app/<entity>_service.go`):
  - [ ] Add method for new transition
  - [ ] Wire guard into transition logic
- [ ] Update service tests (`internal/app/<entity>_service_test.go`)
- [ ] Update CLI if user-triggerable:
  - [ ] Add new subcommand or update existing
- [ ] Update workflow tests document (`specs/<entity>-workflow-tests.md`)
- [ ] Run: `make test && make lint`

### Add CLI Command

When adding a new CLI command:

- [ ] Create command file: `internal/cli/<command>.go`
- [ ] Follow existing patterns:
  - [ ] Use `cobra.Command` struct
  - [ ] Inject dependencies via wire
  - [ ] Keep command thin (delegate to services)
- [ ] Add to parent command (usually in `internal/cli/root.go` or domain command)
- [ ] Add command tests if complex flag parsing
- [ ] Update help text and examples
- [ ] Test manually: `make dev && ./orc <command> --help`
- [ ] Run: `make test && make lint`

### Add CLI Flag

When adding a flag to an existing command:

- [ ] Add flag definition in command's `init()` or builder
- [ ] Update command's `Run` function to use flag value
- [ ] Update help text/examples if needed
- [ ] If flag affects service layer, update service interface/implementation
- [ ] Test flag parsing manually
- [ ] Run: `make test && make lint`

### Add New Entity Type

When creating an entirely new entity (e.g., "Artifact"):

- [ ] **Spec first**: Create `specs/<entity>-workflow.yaml`
- [ ] **Model**: Add `internal/models/<entity>.go`
- [ ] **Core guards**: Add `internal/core/<entity>/guards.go` and tests
- [ ] **Port**: Add interface to `internal/ports/<entity>_repository.go`
- [ ] **Adapter**: Implement `internal/adapters/sqlite/<entity>_repo.go` and tests
- [ ] **Service**: Add `internal/app/<entity>_service.go` and tests
- [ ] **Wire**: Update `internal/wire/wire.go` to construct the service
- [ ] **CLI**: Add commands in `internal/cli/<entity>.go`
- [ ] **Schema**: Add table to `internal/db/schema.sql`
- [ ] **Migration**: Create migration file
- [ ] **Test helpers**: Update `testutil_test.go` with seed function
- [ ] Run: `make test && make lint`

---

## Common Mistakes to Avoid

❌ `go build -o $GOPATH/bin/orc ./cmd/orc` ← Don't do this manually
✅ `make install` ← Use this instead

❌ Using global `orc` when developing ORC
✅ Using `./orc` after `make dev`

❌ Writing business logic in adapters
✅ Keeping adapters as pure translation layers

❌ Importing `adapters/` from `core/`
✅ Core has zero internal dependencies

❌ Skipping FSM spec for new workflows
✅ Spec → Tests → Implementation

❌ Claiming tests passed without running them
✅ Actually run and report what was verified
