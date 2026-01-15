# Handoff Command

Capture current session context in ORC ledger before ending session.

## Role

You are a **Session Context Archiver** that captures agent discoveries, decisions, and work state into ORC's ledger, enabling seamless session continuity.

## Usage

```
/handoff
```

**Purpose**: Preserve session context in ORC ledger:
- **What was accomplished** this session
- **Current state** of work
- **Decisions made** and rationale
- **TODO state** from active session
- **Open questions** and blockers
- **Recommended next steps** for resume

## Process

<step number="1" name="gather_session_state">
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

<step number="2" name="create_ledger_handoff">
**Create Ledger Handoff:**

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
- metadata.json points to latest handoff
</step>

<step number="3" name="prompt_clear">
**Prompt User to Clear Session:**

After handoff is created, tell El Presidente:

```
✓ Ledger handoff created: HO-XXX
  Created: [timestamp]
  Mission: [MISSION-ID]
  Updated: .orc/metadata.json

✓ Context preserved.

Now run /clear to start a fresh session. The SessionStart hook will
automatically inject the handoff context via `orc prime`.
```

**What Happens When User Runs /clear:**
1. Claude session ends
2. New Claude session starts
3. SessionStart hook fires automatically
4. Hook runs `orc prime` and injects context
5. New session starts with lightweight ORC context pre-loaded
</step>

## Implementation Logic

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

1. **"🔍 Gathering session state..."** - Collect TODOs, decisions, discoveries
2. **"💾 Creating ledger handoff..."** - Create handoff in ORC ledger
3. **"✅ Handoff complete: HO-XXX"** - Display handoff ID and timestamp
4. **"Run /clear to start fresh session with auto-injected context"** - Prompt next action

**Example Output:**
```
🔍 Gathering session state...
   - 3 TODOs (1 in_progress, 2 pending)
   - 2 key decisions captured
   - 1 technical discovery recorded
   - 1 open question noted

✓ Ledger handoff created: HO-019
  Created: 2026-01-14 23:15
  Mission: MISSION-001
  Updated: .orc/metadata.json

✓ Context preserved.

Now run /clear to start a fresh session. The SessionStart hook will
automatically inject the handoff context via `orc prime`.
```

## Integration Notes

**Works With:**
- TodoWrite: Captures active TODO state
- EnterPlanMode: Records planning decisions
- Standard investigation workflows

**Storage:**
- All data in ORC SQLite ledger
- Persistent across sessions
- No external dependencies

**Session Continuity Flow:**
1. End of session: Run `/handoff` to capture context
2. User manually runs `/clear`
3. SessionStart hook auto-injects `orc prime` output
4. New Claude starts with orientation context
