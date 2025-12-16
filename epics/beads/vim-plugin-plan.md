# beads.nvim Plugin Plan

**Status**: Planning
**Created**: 2025-12-16
**Approach**: Test-driven development, small increments, agent-friendly
**Location**: `/Users/looneym/src/orc/beads.nvim/` (lives inside ORC repository)

## Goals

Create a Neovim plugin for the [beads issue tracker](https://github.com/steveyegge/beads) with a fugitive-style interactive interface.

**Phase 1 (This Plan)**: Basic foundation with `:Beads` command
- No complex UI interactions
- Stable, tested foundation
- CLI integration working
- Simple display of beads status

**Future Phases**: Interactive keybindings, dependency trees, bead creation, etc.

## Architecture Decisions

### Technology Stack
- **Language**: Lua (Neovim native, first-class support)
- **Testing**: plenary.nvim (standard for Neovim plugins)
- **Integration**: beads CLI (`bd` commands with `--json` output)
- **Pattern**: Scripts to Rule Them All (bootstrap, update, test)

### Why This Approach
- **TDD First**: Write tests before implementation to avoid mess
- **Small Increments**: Each milestone is independently testable
- **Agent-Friendly**: Clear requirements, discrete tasks, automated verification
- **Standard Tools**: Follow Neovim plugin ecosystem conventions

## Project Structure

**Note**: Plugin lives at `/Users/looneym/src/orc/beads.nvim/`

```
/Users/looneym/src/orc/beads.nvim/
├── lua/
│   └── beads/
│       ├── init.lua          ← Main entry point, public API
│       ├── cli.lua           ← BD CLI interface (run commands, parse JSON)
│       ├── ui.lua            ← Buffer/window management
│       └── status.lua        ← Status window logic
├── tests/
│   ├── minimal_init.lua      ← Bootstrap for test environment
│   └── beads/
│       ├── cli_spec.lua      ← Tests for CLI interface
│       ├── ui_spec.lua       ← Tests for UI helpers
│       ├── status_spec.lua   ← Tests for status formatter
│       └── integration_spec.lua  ← End-to-end tests
├── script/
│   ├── bootstrap             ← Set up for the first time
│   ├── update                ← Update dependencies
│   └── test                  ← Run test suite
├── plugin/
│   └── beads.vim             ← Command registration
├── deps/                     ← Dev dependencies (gitignored)
│   └── plenary.nvim/         ← Testing framework
├── Makefile                  ← Convenience targets
└── README.md                 ← Documentation
```

## Development Milestones

### Milestone 1: CLI Interface (Pure Logic, No UI)

**Goal**: Call `bd` CLI commands and parse JSON output

**Tests First** (`tests/beads/cli_spec.lua`):
```lua
describe("beads.cli", function()
  local cli = require("beads.cli")

  it("should parse bd list --json output", function()
    local beads = cli.list()
    assert.is_not_nil(beads)
    assert.is_table(beads)
  end)

  it("should return bead with id, title, status", function()
    local beads = cli.list()
    if #beads > 0 then
      local bead = beads[1]
      assert.is_string(bead.id)
      assert.is_string(bead.title)
    end
  end)

  it("should show specific bead by id", function()
    local bead = cli.show("bd-test1")
    assert.is_table(bead)
    assert.equals("bd-test1", bead.id)
  end)

  it("should return ready work", function()
    local ready = cli.ready()
    assert.is_table(ready)
  end)
end)
```

**Implementation** (`lua/beads/cli.lua`):
```lua
local M = {}

function M.list()
  local handle = io.popen('bd list --json')
  local result = handle:read('*a')
  handle:close()
  return vim.json.decode(result)
end

function M.show(bead_id)
  local handle = io.popen('bd show ' .. bead_id .. ' --json')
  local result = handle:read('*a')
  handle:close()
  return vim.json.decode(result)
end

function M.ready()
  local handle = io.popen('bd ready --json')
  local result = handle:read('*a')
  handle:close()
  return vim.json.decode(result)
end

function M.complete(bead_id)
  os.execute('bd complete ' .. bead_id)
end

return M
```

**Success Criteria**: `script/test tests/beads/cli_spec.lua` passes

---

### Milestone 2: UI Helpers (Buffer Management)

**Goal**: Create and manage scratch buffers for displaying content

**Tests First** (`tests/beads/ui_spec.lua`):
```lua
describe("beads.ui", function()
  local ui = require("beads.ui")

  it("should create scratch buffer with content", function()
    local lines = {"line 1", "line 2", "line 3"}
    local bufnr = ui.create_scratch_buffer("TestBuffer", lines)

    assert.is_number(bufnr)
    assert.is_true(vim.api.nvim_buf_is_valid(bufnr))

    local content = vim.api.nvim_buf_get_lines(bufnr, 0, -1, false)
    assert.equals(3, #content)
    assert.equals("line 1", content[1])
  end)

  it("should create non-modifiable buffer", function()
    local bufnr = ui.create_scratch_buffer("Test", {"content"})
    local modifiable = vim.api.nvim_buf_get_option(bufnr, "modifiable")
    assert.is_false(modifiable)
  end)

  it("should set buffer type to nofile", function()
    local bufnr = ui.create_scratch_buffer("Test", {"content"})
    local buftype = vim.api.nvim_buf_get_option(bufnr, "buftype")
    assert.equals("nofile", buftype)
  end)
end)
```

**Implementation** (`lua/beads/ui.lua`):
```lua
local M = {}

function M.create_scratch_buffer(name, lines)
  local bufnr = vim.api.nvim_create_buf(false, true)

  vim.api.nvim_buf_set_name(bufnr, name)
  vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, lines)
  vim.api.nvim_buf_set_option(bufnr, 'modifiable', false)
  vim.api.nvim_buf_set_option(bufnr, 'buftype', 'nofile')

  vim.api.nvim_set_current_buf(bufnr)

  return bufnr
end

return M
```

**Success Criteria**: `script/test tests/beads/ui_spec.lua` passes

---

### Milestone 3: Status Formatter (Pure Functions)

**Goal**: Format beads data into displayable lines

**Tests First** (`tests/beads/status_spec.lua`):
```lua
describe("beads.status", function()
  local status = require("beads.status")

  describe("organize", function()
    it("should organize beads by status", function()
      local beads = {
        {id = "bd-1", title = "Ready work", status = "ready"},
        {id = "bd-2", title = "In progress", status = "in_progress"},
        {id = "bd-3", title = "Also ready", status = "ready"}
      }

      local organized = status.organize(beads)

      assert.equals(2, #organized.ready)
      assert.equals(1, #organized.in_progress)
      assert.equals(0, #organized.blocked)
    end)

    it("should default to ready if no status field", function()
      local beads = {
        {id = "bd-1", title = "No status field"}
      }

      local organized = status.organize(beads)

      assert.equals(1, #organized.ready)
    end)
  end)

  describe("format", function()
    it("should format organized beads as lines", function()
      local organized = {
        ready = {{id = "bd-1", title = "Test bead"}},
        in_progress = {},
        blocked = {}
      }

      local lines = status.format(organized)

      assert.is_table(lines)
      assert.is_true(#lines > 0)
    end)

    it("should include section headers", function()
      local organized = {
        ready = {{id = "bd-1", title = "Test"}},
        in_progress = {{id = "bd-2", title = "Working"}},
        blocked = {}
      }

      local lines = status.format(organized)
      local text = table.concat(lines, "\n")

      assert.is_true(string.match(text, "Ready Work"))
      assert.is_true(string.match(text, "In Progress"))
    end)

    it("should include bead ids and titles", function()
      local organized = {
        ready = {{id = "bd-abc1", title = "My test bead"}},
        in_progress = {},
        blocked = {}
      }

      local lines = status.format(organized)
      local text = table.concat(lines, "\n")

      assert.is_true(string.match(text, "bd%-abc1"))
      assert.is_true(string.match(text, "My test bead"))
    end)
  end)
end)
```

**Implementation** (`lua/beads/status.lua`):
```lua
local M = {}

function M.organize(beads)
  local organized = {
    ready = {},
    in_progress = {},
    blocked = {},
    done = {}
  }

  for _, bead in ipairs(beads) do
    local status = bead.status or 'ready'
    if organized[status] then
      table.insert(organized[status], bead)
    else
      table.insert(organized.ready, bead)
    end
  end

  return organized
end

function M.format(organized)
  local lines = {}

  table.insert(lines, 'Beads Status')
  table.insert(lines, string.rep('=', 80))
  table.insert(lines, '')

  -- Ready work section
  if #organized.ready > 0 then
    table.insert(lines, 'Ready Work (' .. #organized.ready .. ')')
    for _, bead in ipairs(organized.ready) do
      local line = string.format('  %s  %s', bead.id, bead.title)
      table.insert(lines, line)
    end
    table.insert(lines, '')
  end

  -- In progress section
  if #organized.in_progress > 0 then
    table.insert(lines, 'In Progress (' .. #organized.in_progress .. ')')
    for _, bead in ipairs(organized.in_progress) do
      local line = string.format('  %s  %s', bead.id, bead.title)
      table.insert(lines, line)
    end
    table.insert(lines, '')
  end

  -- Blocked section
  if #organized.blocked > 0 then
    table.insert(lines, 'Blocked (' .. #organized.blocked .. ')')
    for _, bead in ipairs(organized.blocked) do
      local line = string.format('  %s  %s', bead.id, bead.title)
      table.insert(lines, line)
    end
    table.insert(lines, '')
  end

  return lines
end

function M.open()
  local cli = require('beads.cli')
  local ui = require('beads.ui')

  local beads = cli.list()
  local organized = M.organize(beads)
  local lines = M.format(organized)

  ui.create_scratch_buffer('BeadsStatus', lines)
end

return M
```

**Success Criteria**: `script/test tests/beads/status_spec.lua` passes

---

### Milestone 4: Integration (Wire Everything Together)

**Goal**: Connect all pieces and expose `:Beads` command

**Tests First** (`tests/beads/integration_spec.lua`):
```lua
describe("beads integration", function()
  it("should provide :Beads command", function()
    -- Check command exists
    local commands = vim.api.nvim_get_commands({})
    assert.is_not_nil(commands.Beads)
  end)

  it("should open status window with :Beads", function()
    vim.cmd("Beads")

    local bufname = vim.api.nvim_buf_get_name(0)
    assert.is_true(string.match(bufname, "BeadsStatus"))
  end)

  it("should display beads in status buffer", function()
    vim.cmd("Beads")

    local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)
    assert.is_true(#lines > 0)

    local text = table.concat(lines, "\n")
    assert.is_true(string.match(text, "Beads Status"))
  end)
end)
```

**Implementation** (`lua/beads/init.lua`):
```lua
local M = {}

function M.setup(opts)
  opts = opts or {}
  -- Future: configuration handling
end

function M.status()
  require('beads.status').open()
end

return M
```

**Command Registration** (`plugin/beads.vim`):
```vim
if exists('g:loaded_beads')
  finish
endif
let g:loaded_beads = 1

command! Beads lua require('beads').status()
```

**Success Criteria**: `script/test` passes all tests

---

## Scripts to Rule Them All

### `script/bootstrap`
```bash
#!/bin/bash
# script/bootstrap: Set up the plugin for development and testing

set -e

cd "$(dirname "$0")/.."

echo "==> Setting up beads.nvim for development..."

# Create deps directory if it doesn't exist
mkdir -p deps

# Install plenary.nvim for testing
if [ ! -d "deps/plenary.nvim" ]; then
  echo "==> Installing plenary.nvim..."
  git clone --depth 1 https://github.com/nvim-lua/plenary.nvim deps/plenary.nvim
else
  echo "==> plenary.nvim already installed"
fi

# Check for Neovim
if ! command -v nvim &> /dev/null; then
  echo "❌ Neovim not found. Please install Neovim first."
  exit 1
fi

echo "==> Checking Neovim version..."
nvim --version | head -n 1

# Check for beads CLI
if ! command -v bd &> /dev/null; then
  echo "⚠️  Warning: 'bd' command not found."
  echo "    The beads CLI is required for this plugin to work."
  echo "    Install from: https://github.com/steveyegge/beads"
else
  echo "==> beads CLI found: $(which bd)"
fi

echo ""
echo "✅ Bootstrap complete!"
echo ""
echo "Next steps:"
echo "  - Run 'script/test' to run the test suite"
echo "  - Run 'script/update' to update dependencies"
```

### `script/update`
```bash
#!/bin/bash
# script/update: Update dependencies to latest versions

set -e

cd "$(dirname "$0")/.."

echo "==> Updating dependencies..."

# Update plenary.nvim
if [ -d "deps/plenary.nvim" ]; then
  echo "==> Updating plenary.nvim..."
  cd deps/plenary.nvim
  git pull --ff-only
  cd ../..
else
  echo "❌ plenary.nvim not found. Run script/bootstrap first."
  exit 1
fi

echo ""
echo "✅ Dependencies updated!"
```

### `script/test`
```bash
#!/bin/bash
# script/test: Run test suite

set -e

cd "$(dirname "$0")/.."

echo "==> Running tests..."

# Check if plenary is installed
if [ ! -d "deps/plenary.nvim" ]; then
  echo "❌ Test dependencies not found. Run script/bootstrap first."
  exit 1
fi

# Run tests with plenary
nvim --headless --noplugin \
  -u tests/minimal_init.lua \
  -c "PlenaryBustedDirectory tests/ { minimal_init = 'tests/minimal_init.lua' }"

exit_code=$?

if [ $exit_code -eq 0 ]; then
  echo ""
  echo "✅ All tests passed!"
else
  echo ""
  echo "❌ Tests failed!"
  exit $exit_code
fi
```

**Make executable**: `chmod +x script/bootstrap script/update script/test`

---

## Test Infrastructure

### `tests/minimal_init.lua`
```lua
-- Minimal init for running tests

local root = vim.fn.fnamemodify(vim.fn.getcwd(), ":p")
local plenary_dir = root .. "deps/plenary.nvim"

-- Check if plenary exists
if vim.fn.isdirectory(plenary_dir) == 0 then
  error(
    "plenary.nvim not found. Run 'script/bootstrap' to set up dependencies."
  )
end

-- Add to runtimepath
vim.opt.rtp:append(root)
vim.opt.rtp:append(plenary_dir)

-- Set up test environment
vim.cmd("runtime plugin/plenary.vim")

-- Expose test helpers
require("plenary.busted")
```

---

## Agent Workflow

Each milestone is a discrete, testable task:

### Example Task: Milestone 1
```markdown
**Task**: Implement CLI Interface
**Test File**: tests/beads/cli_spec.lua (provided)
**Implementation File**: lua/beads/cli.lua
**Success Criteria**: Run `script/test` - cli_spec.lua must pass
**Dependencies**: None (first milestone)

Steps:
1. Read tests/beads/cli_spec.lua to understand requirements
2. Implement lua/beads/cli.lua to satisfy tests
3. Run script/test to verify
4. Iterate until all cli_spec.lua tests pass
```

### Build Order
1. ✅ Write `tests/beads/cli_spec.lua` → agent implements `lua/beads/cli.lua`
2. ✅ Write `tests/beads/ui_spec.lua` → agent implements `lua/beads/ui.lua`
3. ✅ Write `tests/beads/status_spec.lua` → agent implements `lua/beads/status.lua`
4. ✅ Write `tests/beads/integration_spec.lua` → agent wires everything together

Each step is independent, testable, automated.

---

## Success Criteria

**Milestone 1 Complete**: CLI tests pass
**Milestone 2 Complete**: UI tests pass
**Milestone 3 Complete**: Status tests pass
**Milestone 4 Complete**: Integration tests pass + `:Beads` command works

**Final Success**:
- User can run `:Beads` in Neovim
- Status window displays with organized beads
- All tests pass: `script/test` exits 0
- Clean foundation ready for future features

---

## Future Phases (Not In This Plan)

- Interactive keybindings (Enter to view, dd to complete)
- Bead detail view in split
- Dependency tree visualization
- Bead creation from Neovim
- Search/filter capabilities
- Integration with ORC commands

**This plan focuses exclusively on the foundation**: TDD, clean architecture, basic display.

---

## References

- Beads: https://github.com/steveyegge/beads
- Plenary.nvim: https://github.com/nvim-lua/plenary.nvim
- Scripts to Rule Them All: https://github.blog/engineering/scripts-to-rule-them-all/
- Fugitive: https://github.com/tpope/vim-fugitive (inspiration)
