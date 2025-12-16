# North Star: beads.nvim Plugin

**Created**: 2025-12-16
**Scope**: Complete beads.nvim Neovim plugin development

---

## Purpose & Scope

**What This System Is:**
- Neovim plugin providing native interface to beads issue tracker
- Read-only display of beads status, organized by ready/in-progress/blocked states
- Foundation for future interactive features (view details, mark complete, navigate dependencies)

**Non-Goals:**
- Do NOT reimplement beads CLI logic - always delegate to `bd` commands
- Do NOT provide features beads itself doesn't support (no inventing new issue states, metadata, etc.)
- Do NOT become a standalone issue tracker - always backed by `.beads/` directory
- Do NOT support other issue trackers (GitHub Issues, JIRA, etc.) - beads-only
- Do NOT handle beads installation/configuration - assume `bd` CLI exists and works

**Stakeholders/Users:**
- Developers using beads for issue tracking who want Vim/Neovim integration without leaving their editor

**What Success Looks Like:**
- Phase 1: `:Beads` command displays current status, all tests pass, TDD foundation stable
- Long-term: Fugitive-style interactive interface - view beads, navigate dependencies, mark complete, create new work - all without leaving Neovim

---

## Domain Map

### Domain: CLI Integration
- **Owns**:
  - Execution of `bd` CLI commands via `io.popen`
  - JSON parsing of CLI output
  - Error handling for CLI failures (command not found, invalid JSON, non-zero exit)
- **Inputs**:
  - Bead IDs (strings like "bd-abc123")
  - Query types (list, show, ready, complete)
- **Outputs**:
  - Lua tables with bead data (id, title, status, parent, blocks, description, etc.)
  - Error messages when CLI fails
- **Forbidden coupling**:
  - MUST NOT parse `.beads/issues.jsonl` directly (always use CLI)
  - MUST NOT create/manage Neovim buffers (that's Buffer Management's job)
  - MUST NOT format data for display (that's Status Display's job)
- **Paths**: `lua/beads/cli.lua`, `tests/beads/cli_spec.lua`

---

### Domain: Buffer Management
- **Owns**:
  - Creating scratch buffers (`nvim_create_buf`)
  - Setting buffer properties (name, lines, modifiable=false, buftype=nofile)
  - Making buffers visible (switching to buffer, splits, windows)
- **Inputs**:
  - Buffer name (string)
  - Content lines (array of strings)
  - Optional: window configuration (split direction, size)
- **Outputs**:
  - Buffer number (integer handle)
  - Active buffer in current window
- **Forbidden coupling**:
  - MUST NOT call beads CLI (that's CLI Integration's job)
  - MUST NOT format bead data (that's Status Display's job)
  - MUST NOT register vim commands (that's Command Surface's job)
- **Paths**: `lua/beads/ui.lua`, `tests/beads/ui_spec.lua`

---

### Domain: Status Display
- **Owns**:
  - Organizing beads by status (ready/in_progress/blocked/done)
  - Formatting organized data into display lines
  - Section headers, spacing, layout logic
  - Coordinating CLI + Buffer Management to show status window
- **Inputs**:
  - Raw bead data from CLI Integration
  - Display preferences (future: sort order, filters)
- **Outputs**:
  - Formatted lines array ready for buffer display
  - Mappings of line numbers to bead IDs (for future interactive features)
- **Forbidden coupling**:
  - MUST NOT call `bd` directly (use CLI Integration module)
  - MUST NOT create buffers directly (use Buffer Management module)
  - MUST NOT register vim commands (that's Command Surface's job)
- **Paths**: `lua/beads/status.lua`, `tests/beads/status_spec.lua`

---

### Domain: Command Surface
- **Owns**:
  - Registering vim commands (`:Beads`, `:Bead`, etc.)
  - Plugin initialization and configuration
  - Public API entry points (`require('beads').setup()`)
  - Routing command invocations to appropriate domain functions
- **Inputs**:
  - Vim command invocations (`:Beads`)
  - User configuration options (via `setup()`)
- **Outputs**:
  - Registered vim commands available in editor
  - Configured plugin behavior
- **Forbidden coupling**:
  - MUST NOT implement display logic (delegate to Status Display)
  - MUST NOT call CLI directly (use CLI Integration)
  - MUST NOT create buffers directly (use Buffer Management via Status Display)
  - Thin routing layer only - no business logic
- **Paths**: `lua/beads/init.lua`, `plugin/beads.vim`, `tests/beads/integration_spec.lua`

---

## Golden Journeys

### J1: View Beads Status (Happy Path)
- **Preconditions**:
  - `bd` CLI installed and in PATH
  - `.beads/` directory exists with issues
  - At least one bead exists in any status
- **Steps**:
  1. User runs `:Beads` in Neovim
  2. Plugin calls `bd list --json`
  3. Plugin organizes beads by status
  4. Plugin opens scratch buffer with formatted status
- **Expected observables**:
  - Buffer named "BeadsStatus" is active
  - Buffer contains "Beads Status" header with `===` separator
  - Section headers like "Ready Work (N)", "In Progress (N)", "Blocked (N)"
  - Each bead shows as "  bd-xxxxx  Title text"
  - Buffer is non-modifiable (`modifiable=false`)
  - Exit code: success

---

### J2: View Status with Empty Beads (Edge Case)
- **Preconditions**:
  - `bd` CLI installed
  - `.beads/` directory exists but contains zero issues
- **Steps**:
  1. User runs `:Beads`
  2. Plugin calls `bd list --json`
  3. CLI returns empty array `[]`
  4. Plugin formats empty status display
- **Expected observables**:
  - Buffer named "BeadsStatus" appears
  - Header "Beads Status" shown
  - No section headers (no "Ready Work", etc.) OR sections with "(0)" counts
  - No bead entries displayed
  - No errors or crashes

---

### J3: CLI Not Available (Failure Path)
- **Preconditions**:
  - `bd` command NOT in PATH or not installed
- **Steps**:
  1. User runs `:Beads`
  2. Plugin attempts `io.popen('bd list --json')`
  3. Command fails (command not found or non-zero exit)
- **Expected observables**:
  - Neovim error message displayed (via `vim.notify` or `echom`)
  - Message mentions "bd command not found" or similar
  - No buffer created
  - User remains in their previous buffer
  - Plugin does NOT crash Neovim

---

### J4: Invalid JSON from CLI (Failure Path)
- **Preconditions**:
  - `bd` CLI exists but returns malformed JSON (corrupted data, wrong format)
- **Steps**:
  1. User runs `:Beads`
  2. Plugin calls `bd list --json`
  3. Plugin attempts `vim.json.decode()`
  4. Decode fails with error
- **Expected observables**:
  - Neovim error message: "Failed to parse beads data" or similar
  - No buffer created
  - Plugin does NOT crash Neovim
  - Error includes hint about checking beads installation

---

### J5: Organize Beads by Multiple Statuses (Core Logic)
- **Preconditions**:
  - Mix of beads with status: `ready`, `in_progress`, `blocked`, `done`, or missing status field
- **Steps**:
  1. User runs `:Beads`
  2. Plugin fetches beads via CLI
  3. Plugin organizes into status buckets
- **Expected observables**:
  - "Ready Work" section contains beads with `status="ready"` OR no status field
  - "In Progress" section contains beads with `status="in_progress"`
  - "Blocked" section contains beads with `status="blocked"`
  - Each section header shows correct count
  - Beads appear in their respective sections only (no duplicates)

---

### J6: Test Suite Execution (Development Path)
- **Preconditions**:
  - `script/bootstrap` has been run
  - `deps/plenary.nvim` exists
  - All source files present
- **Steps**:
  1. Developer runs `script/test`
  2. Plenary loads minimal test environment
  3. All spec files execute in isolation
  4. Each test validates its module contracts
- **Expected observables**:
  - Output: "==> Running tests..."
  - All tests pass with green checkmarks
  - Output: "✅ All tests passed!"
  - Exit code: 0
  - No test pollution (each test isolated)

---

### J7: Bootstrap Development Environment (Setup Path)
- **Preconditions**:
  - Fresh clone of beads.nvim
  - Neovim installed
  - Git available
- **Steps**:
  1. Developer runs `script/bootstrap`
  2. Script creates `deps/` directory
  3. Script clones plenary.nvim
  4. Script checks for `nvim` and `bd` commands
- **Expected observables**:
  - Output: "==> Setting up beads.nvim for development..."
  - `deps/plenary.nvim/` directory exists
  - Output: "✅ Bootstrap complete!"
  - Warning shown if `bd` not found (but doesn't fail)
  - Exit code: 0 (even without bd CLI)
  - Ready to run `script/test`

---

### J8: Plugin Initialization (Startup Path)
- **Preconditions**:
  - beads.nvim installed via plugin manager
  - User's init.lua calls `require('beads').setup()`
- **Steps**:
  1. Neovim starts and loads plugins
  2. beads.nvim plugin/beads.vim executes
  3. `:Beads` command registers
  4. User setup function runs (if called)
- **Expected observables**:
  - `:Beads` command available (check with `:command Beads`)
  - No errors during startup
  - No buffers created automatically
  - Plugin loads silently (no output unless error)
  - `g:loaded_beads` variable set to prevent double-loading

---

## Contracts & Invariants

**External Contracts:**
- Always use `bd` CLI with `--json` flag - never parse `.beads/` files directly
- CLI returns: `id`, `title`, `status` (optional), `description`, `parent`, `blocks`
- Missing `status` defaults to `"ready"`

**Buffer Rules:**
- Status window always named `"BeadsStatus"`
- All plugin buffers: `buftype=nofile`, `modifiable=false`
- Running `:Beads` multiple times is idempotent (fresh data each time)

**Error Handling:**
- CLI failures show user-visible error (don't crash Neovim)
- JSON parse failures show helpful message

**Module Pattern:**
- All modules return table (`local M = {} ... return M`)
- No global namespace pollution

**Test Isolation:**
- Tests run independently (no shared state)
- `script/test` exits 0 on success, non-zero on failure

**Backwards Compatibility:**
- `:Beads` command with no arguments must always work

---

## Harnesses

**Test Suite:**
- **Command**: `script/test`
- **Validates**: All unit and integration tests pass
- **When to run**: Before every commit, in CI

**Development Setup:**
- **Command**: `script/bootstrap`
- **Validates**: Dependencies installed, ready to develop
- **When to run**: Fresh clone, new contributors

---

## Decision Rules

**MUST:**
- Write tests before implementation (TDD)
- Use `bd` CLI for all beads data (never parse `.beads/` directly)
- Run `script/test` before committing
- Keep modules pure (no side effects on require)

**SHOULD:**
- Keep functions small and focused
- Use descriptive variable names
- Follow existing code patterns in the module

**BAN:**
- Direct `.beads/issues.jsonl` file parsing
- Global namespace pollution
- Mutating beads data (read-only plugin)
- Crashing Neovim on errors

**Escape Hatch:**
- Allowed to deviate from MUST/BAN rules only if:
  - **Approved by El Presidente first**
  - Documented in code comment with reasoning
  - Tests updated to cover new behavior
  - Noted in this North Star's Open Questions section

---

## Open Questions

(None at this time)

---

## Revision History

- **2025-12-16**: Initial North Star created
  - Defined 4 domains: CLI Integration, Buffer Management, Status Display, Command Surface
  - Established 8 golden journeys covering happy path, edge cases, failures, and dev workflows
  - Set contracts for CLI integration, buffer handling, and error management
  - Defined TDD-first decision rules with clear MUST/BAN boundaries
