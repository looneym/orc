---
name: orc-self-test
description: Integration self-testing for ORC. Tests the gotmux-based tmux workflow by creating test entities, running apply, and verifying infrastructure. Use when you want to verify ORC is working correctly end-to-end.
---

# ORC Self-Test

Run integration tests for the gotmux-based tmux workflow and workbench infrastructure.

## Safety: Default TMux Server

**CRITICAL**: The `orc tmux apply` command operates on the default tmux server. The test
creates a uniquely named session (`orc-self-test`) to avoid collisions. Always use exact
session matching (`-t "=$SESSION_NAME"`) for all verification commands.

**Rule**: Always qualify tmux targets as `session:window.pane`. Always use exact matching
with `=` prefix. Always run cleanup even on failure.

## Prerequisites

- `orc-dev` must be available (use `make dev` to build, `make install` for the shim)
- tmux must be installed
- Test repository `orc-canary` must exist (check `orc repo list`)

## Variables

Throughout the test, track these variables. Print them after creation for debugging:

```
TEST_REPO_ID=<from orc repo list>
FACTORY_ID=<created>
WORKSHOP_ID=<created>
WORKBENCH_ID=<created>
WORKBENCH_NAME=<created>
WORKBENCH_PATH=<created>
SESSION_NAME=<workshop name>
```

## Flow

### 1. Detect Test Repository

```bash
TEST_REPO_ID=$(orc-dev repo list | grep orc-canary | awk '{print $1}')
if [ -z "$TEST_REPO_ID" ]; then
  echo "[FAIL] orc-canary repo not registered. Run: orc repo create orc-canary --path ~/src/orc-canary --default-branch master"
  exit 1
fi
echo "[PASS] Test repo found: $TEST_REPO_ID"
```

### 2. Create Test Entities

```bash
FACTORY_ID=$(orc-dev factory create "orc-self-test" 2>&1 | grep -o 'FACT-[0-9]*')
echo "Created factory: $FACTORY_ID"

WORKSHOP_ID=$(orc-dev workshop create --factory "$FACTORY_ID" --name "orc-self-test" 2>&1 | grep -o 'WORK-[0-9]*')
echo "Created workshop: $WORKSHOP_ID"

orc-dev workbench create --workshop "$WORKSHOP_ID" --repo-id "$TEST_REPO_ID"
# Parse output for BENCH ID, name, path
```

### 3. Verify Filesystem Created (Immediate)

**Key test**: Workbench creation is atomic (DB + worktree + config in one operation).

```bash
# Check worktree exists
[ -d "$WORKBENCH_PATH" ] || { echo "[FAIL] Worktree not created"; exit 1; }
echo "[PASS] Worktree exists: $WORKBENCH_PATH"

# Check config file exists and has correct ID
[ -f "$WORKBENCH_PATH/.orc/config.json" ] || { echo "[FAIL] Config not created"; exit 1; }
echo "[PASS] Config file exists"

grep -q "$WORKBENCH_ID" "$WORKBENCH_PATH/.orc/config.json" || { echo "[FAIL] Config has wrong ID"; exit 1; }
echo "[PASS] Config contains correct workbench ID"
```

### 4. Apply TMux Session

Use `orc tmux apply --yes` to create the session. This replaces the old `orc tmux start`.

```bash
# Apply session via ORC (creates session + windows + enrichment in one command)
orc-dev tmux apply "$WORKSHOP_ID" --yes

SESSION_NAME="orc-self-test"

# Verify session exists
if ! tmux has-session -t "=$SESSION_NAME" 2>/dev/null; then
  echo "[FAIL] TMux session not created: $SESSION_NAME"
  # Run cleanup
  exit 1
fi
echo "[PASS] TMux session created: $SESSION_NAME"
```

**Important**: Use `-t "=$SESSION_NAME"` (with `=` prefix) for exact session matching.
This prevents tmux from matching partial names against other sessions.

### 5. Verify Pane Structure

Each workbench window should have exactly 1 pane (the goblin pane).

```bash
# Use exact session targeting to avoid cross-session interference
# List panes with session:window qualification
PANES=$(tmux list-panes -t "=$SESSION_NAME:$WORKBENCH_NAME" -F "#{pane_index}:#{pane_current_path}")

echo "Panes found:"
echo "$PANES"

PANE_COUNT=$(echo "$PANES" | wc -l | tr -d ' ')
if [ "$PANE_COUNT" -ne 1 ]; then
  echo "[FAIL] Expected 1 pane (goblin), found $PANE_COUNT"
  exit 1
fi
echo "[PASS] Found 1 pane (single goblin pane)"

# Verify pane is in workbench directory
PANE_PATH=$(echo "$PANES" | head -1 | cut -d: -f2)
if [ "$PANE_PATH" != "$WORKBENCH_PATH" ]; then
  echo "[FAIL] Pane in wrong directory: $PANE_PATH (expected: $WORKBENCH_PATH)"
  exit 1
fi
echo "[PASS] Pane in correct directory"
```

### 6. Verify Pane Options (@pane_role, @bench_id, @workshop_id)

Pane identity uses tmux pane options (NOT shell env vars -- those can't be
read by tmux format strings). Verify using `#{@pane_role}`, `#{@bench_id}`,
and `#{@workshop_id}` format strings.

```bash
# Get pane index for the single goblin pane
PANE_IDX=$(tmux list-panes -t "=$SESSION_NAME:$WORKBENCH_NAME" -F "#{pane_index}" | head -1)
TARGET="=$SESSION_NAME:$WORKBENCH_NAME.$PANE_IDX"

# Check @pane_role is "goblin"
ROLE=$(tmux display-message -t "$TARGET" -p '#{@pane_role}')
if [ "$ROLE" != "goblin" ]; then
  echo "[FAIL] Pane: expected @pane_role=goblin, got '$ROLE'"
  exit 1
fi
echo "[PASS] Pane: @pane_role=$ROLE"

# Check @bench_id
BENCH=$(tmux display-message -t "$TARGET" -p '#{@bench_id}')
if [ -z "$BENCH" ]; then
  echo "[FAIL] Pane: @bench_id not set"
  exit 1
fi
if [ "$BENCH" != "$WORKBENCH_ID" ]; then
  echo "[FAIL] Pane: expected @bench_id=$WORKBENCH_ID, got '$BENCH'"
  exit 1
fi
echo "[PASS] Pane: @bench_id=$BENCH"

# Check @workshop_id
WS=$(tmux display-message -t "$TARGET" -p '#{@workshop_id}')
if [ -z "$WS" ]; then
  echo "[FAIL] Pane: @workshop_id not set"
  exit 1
fi
if [ "$WS" != "$WORKSHOP_ID" ]; then
  echo "[FAIL] Pane: expected @workshop_id=$WORKSHOP_ID, got '$WS'"
  exit 1
fi
echo "[PASS] Pane: @workshop_id=$WS"

echo "[PASS] All pane options set correctly (@pane_role=goblin, @bench_id, @workshop_id)"
```

### 7. Test Idempotency

```bash
# Run apply again -- should be a no-op (except enrichment)
orc-dev tmux apply "$WORKSHOP_ID" --yes

# Verify session still exists and has correct pane count
if ! tmux has-session -t "=$SESSION_NAME" 2>/dev/null; then
  echo "[FAIL] Session disappeared after idempotent apply"
  exit 1
fi

PANE_COUNT=$(tmux list-panes -t "=$SESSION_NAME:$WORKBENCH_NAME" | wc -l | tr -d ' ')
if [ "$PANE_COUNT" -ne 1 ]; then
  echo "[FAIL] Expected 1 pane after idempotent apply, found $PANE_COUNT"
  exit 1
fi

# Re-verify pane role survived
ROLE=$(tmux list-panes -t "=$SESSION_NAME:$WORKBENCH_NAME" -F '#{@pane_role}' | head -1)
if [ "$ROLE" != "goblin" ]; then
  echo "[FAIL] Pane role not preserved after idempotent apply (got '$ROLE')"
  exit 1
fi
echo "[PASS] Idempotent apply preserved session, pane, and goblin role"
```

### 8. Verify Auto-Enrichment

`orc tmux apply` applies enrichment automatically. Verify the `@orc_enriched`
window option was set on the workbench window.

```bash
# Verify @orc_enriched window option was set by auto-enrichment
ENRICHED=$(tmux show-options -t "=$SESSION_NAME:$WORKBENCH_NAME" -wqv @orc_enriched 2>/dev/null)
if [ "$ENRICHED" = "1" ]; then
  echo "[PASS] Auto-enrichment applied (@orc_enriched=1)"
else
  echo "[FAIL] Auto-enrichment not applied (expected @orc_enriched=1, got '$ENRICHED')"
  exit 1
fi
```

### 9. Cleanup

**IMPORTANT**: Run cleanup even if tests fail. Always use exact session matching.

```bash
# Kill the tmux session (exact match)
tmux kill-session -t "=$SESSION_NAME" 2>/dev/null

if tmux has-session -t "=$SESSION_NAME" 2>/dev/null; then
  echo "[FAIL] Session still exists after kill"
else
  echo "[PASS] TMux session killed"
fi

# Archive workbench and clean up worktree
orc-dev workbench archive "$WORKBENCH_ID"
rm -rf "$WORKBENCH_PATH"

# Prune git worktrees
(cd ~/src/orc-canary && git worktree prune 2>/dev/null)

if [ -d "$WORKBENCH_PATH" ]; then
  echo "[FAIL] Worktree still exists: $WORKBENCH_PATH"
else
  echo "[PASS] Worktree cleaned up"
fi
echo "[PASS] Test entities archived"
```

## Success Criteria

```
ORC Self-Test Results
---------------------
[PASS] Test repo found
[PASS] Factory created
[PASS] Workshop created
[PASS] Workbench created (immediate: DB + worktree + config)
[PASS] Worktree exists
[PASS] Config file exists
[PASS] Config correct
[PASS] TMux session created (via apply)
[PASS] Found 1 pane (single goblin pane)
[PASS] Pane in correct directory
[PASS] @pane_role=goblin set correctly
[PASS] @bench_id set correctly
[PASS] @workshop_id set correctly
[PASS] Idempotent apply preserved session, pane, and goblin role
[PASS] Auto-enrichment applied (@orc_enriched=1)
[PASS] TMux session killed
[PASS] Worktree cleaned up
[PASS] Test entities archived

All tests passed!
```

## On Failure

If ANY step fails, you MUST still run cleanup:

1. Report which step failed and the error
2. **Always** attempt cleanup:
   ```bash
   tmux kill-session -t "=orc-self-test" 2>/dev/null
   orc-dev workbench archive $WORKBENCH_ID 2>/dev/null
   rm -rf $WORKBENCH_PATH 2>/dev/null
   (cd ~/src/orc-canary && git worktree prune 2>/dev/null)
   ```
3. Suggest running `orc-dev doctor` for diagnostics

## Safety Rules for Agents

1. **Exact session matching**: Always use `-t "=$SESSION_NAME"` (with `=`) to prevent partial matching
2. **Never bare `tmux` commands**: Always qualify with session:window.pane targets
3. **Never `orc tmux enrich`**: It mutates global tmux bindings on the user's live server
4. **Never `cd` in main shell**: Use subshells `(cd ... && cmd)` to avoid changing agent working directory
5. **Always cleanup**: Even on failure. Kill session, archive workbench, remove worktree
6. **Use `orc-dev`**: The dev shim runs the local binary with workbench DB

## Key Architecture Notes

**Removed (vs old architecture):**
- Gatehouse entity (absorbed into workbench)
- `orc infra plan/apply` for tmux (replaced by gotmux)
- Deferred filesystem creation (now immediate)
- Smug YAML config generation (replaced by gotmux programmatic API)
- Separate `orc tmux start` and `orc tmux refresh` commands (unified into `apply`)

**Current patterns:**
- Workbench creation is atomic (DB + worktree + config)
- Gotmux creates tmux sessions programmatically
- `orc tmux apply WORK-xxx --yes` is the single command for session lifecycle
- `apply` uses plan/execute pattern: compares desired (DB) vs actual (tmux) state
- Each workbench window has a single goblin pane (no multi-pane layout)
- `respawn-pane -k` sets pane root process at creation time (no SendKeys)
- `@pane_role` tmux pane option is authoritative for pane identity (only "goblin" role)
- `@bench_id` and `@workshop_id` tmux pane options provide workbench context
- Shell env vars (`PANE_ROLE`, `BENCH_ID`, `WORKSHOP_ID`) are NOT used -- all identity is via tmux pane options
- `apply` auto-applies enrichment (no separate `orc tmux enrich` step needed)
- Desk popup provides shell, vim+fugitive, and TUI dashboard (separate tmux server per workbench)
