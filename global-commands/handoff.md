# Handoff Command

Flush current session context to Graphiti's persistent memory store before ending session.

## Role

You are a **Session Context Archiver** that captures agent discoveries, decisions, and work state into Graphiti's temporal knowledge graph, enabling seamless session continuity without file pollution in repositories.

## Usage

```
/handoff [--worktree worktree-name] [--summary "Brief summary"]
```

**Purpose**: Preserve session context in Graphiti's global memory store:
- **Agent discoveries** and technical insights (not raw code)
- **Decisions made** and rationale
- **TODO state** from active session
- **Open questions** and blockers
- **Files/modules investigated**
- **Recommended next steps** for resume

**Perfect Companion to /g-bootstrap**: Handoff captures state, bootstrap restores it in new sessions.

## Process

<step number="1" name="detect_context">
**Detect Current Context:**
- Determine current working directory
- Detect worktree from path (~/src/worktrees/NAME)
- Check if in ORC orchestrator context
- Use --worktree flag if provided (override detection)
- Fall back to "unknown-session" if detection fails

**group_id Priority:**
1. --worktree flag (explicit override)
2. Auto-detect from ~/src/worktrees/[name] → "worktree-[name]"
3. ~/src/orc → "orc"
4. "unknown-session" (fallback)

**Check ORC Ledger:**
- Verify `orc` binary is available (run `which orc`)
- If not found, ledger handoff will fail (Graphiti-only fallback not available in handoff)
- The `orc handoff create` command will error if ledger not initialized
</step>

<step number="2" name="gather_session_state">
**Gather Session Context:**

Check if TodoWrite is active and capture:
- Current TODO items with status (pending, in_progress, completed)
- Task descriptions and active work indicators

Analyze conversation history for:
- **Key decisions**: "We chose X because Y"
- **Technical discoveries**: "Found that Z uses pattern A"
- **Architectural insights**: "System works by B"
- **Open questions**: "Need to investigate C"
- **Blockers**: "Waiting on D"

Identify investigated artifacts:
- File paths read or edited
- Modules/components explored
- APIs or services touched

Determine next steps:
- What should be resumed first?
- What's ready to work on next?
- What needs follow-up?
</step>

<step number="3" name="create_ledger_handoff">
**Create Ledger Handoff (PRIORITY: Do This First):**

**Write Narrative Note for Next Claude:**

Craft a Claude-to-Claude handoff note in markdown format:
- Write in second person ("You were working on...")
- Focus on narrative flow, not structured data
- Include what was accomplished, current state, what's next
- Add important context and gotchas
- Keep it warm but professional

**Create Ledger Handoff:**

```bash
orc handoff create \
  --note "$(cat <<'EOF'
Hey next Claude! Here's where we are:

## What We Accomplished
[List key completions from this session]

## Current State
[Describe active work, decisions made, discoveries]

## What's Next
[Clear next steps with priority]

## Important Context
[Gotchas, blockers, open questions]
EOF
)" \
  --mission [MISSION-ID if active] \
  --operation [OP-ID if active] \
  --work-order [WO-ID if active] \
  --expedition [EXP-ID if active]
```

**Result:** Ledger handoff created instantly, metadata.json updated automatically.

**Benefits:**
- Next Claude gets context in <1 second
- Structured relationships via database
- No waiting for Graphiti processing
- metadata.json points to latest handoff
</step>

<step number="4" name="create_graphiti_episode">
**Create Graphiti Episode (ASYNC: Background Memory):**

**After ledger handoff created**, also flush to Graphiti for semantic memory.

Build JSON episode with:
```json
{
  "session_summary": "<user-provided or auto-generated summary>",
  "timestamp": "<ISO 8601 timestamp>",
  "worktree": "<worktree-name>",
  "todos": [
    {"content": "Task description", "status": "in_progress"},
    {"content": "Another task", "status": "completed"}
  ],
  "decisions": [
    {"decision": "What was decided", "rationale": "Why"},
  ],
  "discoveries": [
    {"insight": "Technical discovery", "context": "Where/how found"},
  ],
  "open_questions": [
    {"question": "What needs investigation", "priority": "high/medium/low"},
  ],
  "investigated_files": [
    "path/to/file.ts",
    "path/to/module/"
  ],
  "next_steps": [
    "Recommended action to resume work",
  ]
}
```

Use MCP tool: `mcp__graphiti__add_memory()` with:
- name: "Session Handoff: [worktree-name] - [timestamp]"
- episode_body: JSON string (properly escaped)
- source: "json" (structured data)
- source_description: "ORC session handoff"
- group_id: detected/specified group_id

**Capture Graphiti Episode UUID:**
- Store the returned UUID
- Update ledger handoff with Graphiti reference:
  ```bash
  # TODO: Add this capability to orc CLI
  # For now, note UUID in handoff note or manually track
  ```

**Important:** This is background processing. Don't wait for Graphiti - ledger handoff is already complete!
</step>

<step number="5" name="confirm_flush">
**Confirm Dual Flush:**

Display confirmation to El Presidente:

```
✓ Ledger handoff created: HO-XXX
  Created: [timestamp]
  Mission: [MISSION-ID]
  Operation: [OP-ID]
  Updated: .orc/metadata.json

🧠 Graphiti episode queued: [episode-name]
   Group: [group_id]
   Processing in background (~20s)

✓ Context preserved. Safe to start new session with /g-bootstrap
```

**Two-Tier Architecture:**
- **Ledger (Factory Floor)**: Instant structured handoff, ready to query
- **Graphiti (Brain)**: Semantic memory processing in background

**Next session will:**
1. Read ledger handoff instantly (<1s)
2. Query Graphiti for deeper insights (once processing completes)
</step>

<step number="6" name="auto_clear_session">
**Automatically Clear Session (TMux Integration):**

**After handoff is complete**, automatically trigger `/clear` to start fresh session with prime injection.

**Check if in TMux Session:**
```bash
if [ -n "$TMUX" ]; then
  # Get current session and pane
  TMUX_SESSION=$(tmux display-message -p '#S' 2>/dev/null)
  TMUX_PANE=$(tmux display-message -p '#P' 2>/dev/null)

  if [ -n "$TMUX_SESSION" ] && [ -n "$TMUX_PANE" ]; then
    echo ""
    echo "🔄 Restarting Claude session (handoff → restart → prime)"
    echo "   (Session will reload in 3 seconds...)"

    sleep 3

    # Kill and respawn the pane with fresh Claude instance
    # -k flag kills existing process
    # This triggers SessionStart hook which injects orc prime context
    tmux respawn-pane -t "$TMUX_SESSION:$TMUX_PANE" -k claude

    # Don't wait or output after respawn - we've been killed!
    exit 0
  fi
fi
```

**If not in TMux:**
```
ℹ️  Not in TMux session - run /clear manually to start fresh session with prime injection
```

**What Happens Next:**
1. TMux pane is killed and respawned with fresh Claude instance
2. Claude starts up (trust prompt auto-accepted if configured)
3. SessionStart hook fires on new Claude startup
4. Hook runs `orc prime` and injects context into new session
5. New session starts with lightweight ORC context pre-loaded

**Benefits:**
- **One command**: User just runs `/handoff`
- **Automatic flow**: Handoff → Restart → Prime happens seamlessly
- **No manual steps**: Session continuity fully automated
- **Clean slate**: Fresh Claude instance with context pre-loaded
- **Reliable**: No timing issues with send-keys approach

**Note:** The SessionStart hook (configured in `~/.claude/settings.json`) handles the prime injection. The `tmux respawn-pane` creates a new Claude process which triggers the hook.
</step>

## Implementation Logic

**Context Detection Algorithm:**
```
function detectGroupId():
  if --worktree flag provided:
    return "worktree-{flag_value}"

  cwd = current_working_directory()

  if cwd matches ~/src/worktrees/[name]:
    return "worktree-{name}"

  if cwd matches ~/src/orc:
    return "orc"

  return "unknown-session"
```

**Decision Extraction:**
- Look for phrases: "decided to", "chose", "going with", "selected"
- Capture context: what was decided and why
- Avoid implementation details, focus on strategic choices

**Discovery Extraction:**
- Look for phrases: "discovered", "found that", "realized", "identified"
- Capture insights about system behavior, architecture, patterns
- NOT raw code - conceptual understanding only

## Expected Behavior

When El Presidente runs `/handoff`:

1. **"🔍 Detecting current context..."** - Analyze working directory and detect worktree
2. **"📊 Context: worktree-[name]"** - Confirm detected group_id
3. **"🧠 Gathering session state..."** - Collect TODOs, decisions, discoveries
4. **"💾 Creating ledger handoff..."** - Create handoff in ORC ledger
5. **"🧠 Flushing to Graphiti (group: worktree-[name])..."** - Queue episode via MCP
6. **"✅ Session handoff complete. Context preserved."** - Confirm successful queue
7. **"🔄 Restarting Claude session (handoff → restart → prime)"** - Auto-restart pane (if in TMux)
8. **Pane respawns, new Claude starts, SessionStart hook injects orc prime context**

**Example Output:**
```
🔍 Detecting current context...
📊 Context: worktree-ml-sidekiq-deprecation
🧠 Gathering session state...
   - 3 TODOs (1 in_progress, 2 pending)
   - 2 key decisions captured
   - 1 technical discovery recorded
   - 1 open question noted

✓ Ledger handoff created: HO-013
  Created: 2026-01-14 12:15
  Mission: MISSION-002
  Updated: .orc/metadata.json

🧠 Graphiti episode queued: Session Handoff: ml-sidekiq-deprecation - 2026-01-14
   Group: worktree-ml-sidekiq-deprecation
   Processing in background (~20s)

✓ Context preserved. Safe to start new session with /g-bootstrap

🔄 Restarting Claude session (handoff → restart → prime)
   (Session will reload in 3 seconds...)
[TMux pane respawns, new Claude instance starts with ORC context pre-loaded]
```

## Advanced Features

**Custom Summary:**
```
/handoff --summary "Made progress on worker config, blocked on Redis connection pooling"
```

**Explicit Worktree:**
```
/handoff --worktree ml-auth-refactor
```

**Graceful Error Handling:**
- If Graphiti unavailable: Display warning, suggest trying again when MCP server is up
- If no context found: Create minimal episode with timestamp and location
- If MCP tools not loaded: Display instructions to check MCP configuration

**Cross-Session Continuity:**
- Each handoff creates timestamped episode
- Episodes accumulate for temporal context
- /g-bootstrap queries recent episodes for full history
- Semantic search enables cross-investigation insights

## Integration Notes

**Works With:**
- TodoWrite: Captures active TODO state
- EnterPlanMode: Records planning decisions
- Standard investigation workflows

**Complements:**
- /bootstrap: Traditional disk-based context loading
- /g-bootstrap: Graphiti-enhanced context restoration
- /compact: Replaced by /handoff for session continuity

**Storage:**
- All data in Neo4j at localhost (no files in repos)
- Persistent across worktree switches
- Global memory store for entire dev environment
- No git-backed syncing required
