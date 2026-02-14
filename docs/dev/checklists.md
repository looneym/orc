# Development Checklists

Step-by-step checklists for common ORC development tasks.

## Add Field to Entity

When adding a new field to an existing entity (e.g., adding `priority` to Task):

- [ ] Create workbench DB: `make setup-workbench`
- [ ] Update model struct in `internal/models/<entity>.go`
- [ ] Update SQL schema in `internal/db/schema.sql`
- [ ] Preview with Atlas: `make schema-diff-workbench`
- [ ] Apply to workbench: `make schema-apply-workbench`
- [ ] Update repository:
  - [ ] `internal/adapters/sqlite/<entity>_repo.go`
  - [ ] `internal/adapters/sqlite/<entity>_repo_test.go`
- [ ] Update service if field has business logic:
  - [ ] `internal/app/<entity>_service.go`
  - [ ] `internal/app/<entity>_service_test.go`
- [ ] Update CLI if field is user-facing
- [ ] Run: `make test && make lint`

## Add State/Transition

When adding a new state or transition to an entity's state machine:

- [ ] Update core guards + tests:
  - [ ] `internal/core/<entity>/guards.go`
  - [ ] `internal/core/<entity>/guards_test.go`
- [ ] Update service + tests:
  - [ ] `internal/app/<entity>_service.go`
  - [ ] `internal/app/<entity>_service_test.go`
- [ ] Update CLI if user-triggerable
- [ ] Run: `make test && make lint`

## Add CLI Command

- [ ] Create: `internal/cli/<command>.go`
- [ ] Keep it thin: parse args/flags, call services, render output
- [ ] Inject dependencies via wire (no globals)
- [ ] Manual smoke: `make dev && orc-dev <command> --help`
- [ ] Run: `make test && make lint`

## Add New Entity (with Repository)

When adding a new entity that requires persistence (e.g., CycleWorkOrder):

- [ ] Create workbench DB: `make setup-workbench`
- [ ] Guards in `internal/core/<entity>/guards.go`
- [ ] Guard tests in `internal/core/<entity>/guards_test.go`
- [ ] Schema in `internal/db/schema.sql`
- [ ] Preview with Atlas: `make schema-diff-workbench`
- [ ] Apply to workbench: `make schema-apply-workbench`
- [ ] Secondary port interface in `internal/ports/secondary/persistence.go`
- [ ] Primary port interface in `internal/ports/primary/<entity>.go`
- [ ] **Repository implementation + tests** (REQUIRED):
  - [ ] `internal/adapters/sqlite/<entity>_repo.go`
  - [ ] `internal/adapters/sqlite/<entity>_repo_test.go`
- [ ] Service implementation in `internal/app/<entity>_service.go`
- [ ] CLI commands in `internal/cli/<entity>.go`
- [ ] Wire registration in `internal/wire/`
- [ ] Run: `make test && make lint`

**Hard rule:** Repository tests are NOT optional. Every `*_repo.go` MUST have a corresponding `*_repo_test.go`.

## Add TUI Action

When adding a new action to the summary TUI (e.g., adding "archive" for shipments):

- [ ] Add action to `entityActionMatrix` in `summary_tui.go` for each eligible entity type
- [ ] Add key handler in the `Update` switch block — guard with `entityHasAction(id, "action")`
- [ ] Add status bar hint in `renderStatusBar` — use `formatHint("key", "label", entityHasAction(...))`
- [ ] Add action to `allActions` slice in `TestEntityActionMatrixCompleteness`
- [ ] Update the matrix table in `docs/dev/tui.md`
- [ ] Run: `make test && make lint`

See [docs/dev/tui.md](tui.md) for full entity-action matrix and key reference.

## Emit Operational Events

When adding operational event emission to a new subsystem:

- [ ] Choose source constant in internal/core/event/sources.go (or add new one)
- [ ] Inject EventWriter into your app service via wire
- [ ] Call eventWriter.EmitOperational(ctx, source, level, message, data)
- [ ] Test with: make dev && orc-dev events tail --source <your-source> --level debug
- [ ] Run: make test && make lint
