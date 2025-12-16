# Tech Plan: Install and Configure Beads

**Status**: Ready to execute
**Created**: 2025-12-16
**Objective**: Install beads CLI and configure it for the ORC repository

---

## Overview

Install the beads issue tracker CLI tool and initialize it in the ORC repository to enable issue tracking with graph-based dependencies.

## Worktree Compatibility

**ORC uses true git worktrees** - verified via:
- `/worktree` command uses `git worktree add`
- Existing worktrees have `.git` file pointing to main repo (not full `.git` directory)
- Example: `ml-bot-future-planning-bot-test/.git` → `gitdir: /Users/looneym/src/intercom-bot-test/.git/worktrees/...`

**Beads has "Enhanced support for git worktrees with shared database architecture"**:
- Initialize beads once in main repo (e.g., `/Users/looneym/src/intercom/`)
- Commit `.beads/` directory to git
- All worktrees automatically share the same beads database
- SQLite cache (`beads.db`) is shared across all worktrees
- Changes in one worktree are immediately visible in others
- No per-worktree configuration needed

**What this means**:
- Initialize beads in each main repo (intercom, infrastructure, etc.)
- All worktrees from that repo see the same issues
- One unified issue tracker per project
- Perfect fit for ORC workflow

---

## Prerequisites

- ✅ macOS (Darwin detected)
- ✅ Git repository (ORC already initialized)
- ✅ Homebrew (assumed - will check during execution)

---

## Installation Plan

### Step 1: Install Beads CLI

**Method**: Homebrew (recommended for macOS)

```bash
# Add beads tap
brew tap steveyegge/beads

# Install bd CLI
brew install bd
```

**Verification**:
```bash
bd --version
which bd
```

**Alternative if Homebrew fails**: Use curl installer
```bash
curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/scripts/install.sh | bash
```

---

### Step 2: Initialize Beads in ORC Repository

**Location**: `/Users/looneym/src/orc/`

```bash
cd /Users/looneym/src/orc

# Initialize beads with quiet mode (non-interactive)
bd init --quiet
```

**What this creates**:
- `.beads/issues.jsonl` - Issue storage (commit to git)
- `.beads/deletions.jsonl` - Deletion log (commit to git)
- `.beads/config.yaml` - Configuration (commit to git)
- `.beads/metadata.json` - Metadata (commit to git)
- `.beads/README.md` - Documentation (commit to git)
- `.gitattributes` - Git merge driver config (commit to git)
- `.beads/beads.db` - SQLite cache (gitignored)
- `.beads/bd.sock` - Daemon socket (gitignored)

---

### Step 3: Update .gitignore

**Ensure local-only files are ignored**:

Add this to the global gitignore for all projects

```bash
# Add to .gitignore if not present
cat >> .gitignore <<'EOF'

# Beads local-only files
.beads/beads.db
.beads/beads.db-*
.beads/bd.sock
.beads/bd.pipe
.beads/.exclusive-lock
EOF
```

---

### Step 4: Verify Installation

```bash
# Check beads health
bd doctor

# View database info
bd info

# List issues (should be empty initially)
bd list
```

**Expected output**:
- `bd doctor` shows all green checkmarks
- `bd info` displays database metadata
- `bd list` returns empty array or "No issues found"

## Post-Installation

After successful installation:

1. **Create first bead** (optional test):
   ```bash
   bd create "Test bead creation"
   bd list
   ```

2. **Ready for vim plugin development**:
   - beads.nvim can now call `bd list --json`
   - Plugin tests will have real beads data to work with

3. **Use beads for ORC work**:
   - Create epics for major initiatives
   - Break down work into beads with dependencies
   - Use `bd ready` to find unblocked work

---

## Notes

- Beads daemon starts automatically on first command
- JSONL syncs to git every 5 seconds
- Merge conflicts in JSONL handled by custom merge driver
- All beads commands work with `--json` flag for machine-readable output

---

## References

- Beads GitHub: https://github.com/steveyegge/beads
- Installation docs: https://github.com/steveyegge/beads#installation
