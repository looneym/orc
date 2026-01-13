# Graphiti Integration Guide for ORC Skills

## Overview

This guide explains how ORC global commands (`/g-handoff` and `/g-bootstrap`) integrate with Graphiti's semantic memory system to provide persistent, cross-session context for Claude.

## Architecture: Two-Tier Memory System

**Ledger (SQLite)**:
- **Purpose**: Instant structured queries (<1 second)
- **Storage**: `~/.orc/orc.db`
- **Data**: Handoffs with narrative notes, active mission/operation/work-order links
- **Use case**: "What was I working on?" - immediate context restoration

**Graphiti (Neo4j)**:
- **Purpose**: Semantic memory with temporal reasoning
- **Storage**: Neo4j graph database via MCP server
- **Data**: Decisions, discoveries, relationships, "why" not just "what"
- **Use case**: "Why did we choose X?" - deeper insights and cross-investigation patterns

## /g-handoff: Dual Flush Pattern

When `/g-handoff` runs, it performs a **two-tier flush**:

### Step 1: Ledger Handoff (Synchronous - Priority)

```bash
# Create structured handoff in SQLite
./orc handoff create \
  --note "$(cat <<'EOF'
[Claude-to-Claude narrative note in markdown]
EOF
)" \
  --mission MISSION-001 \
  --operation OP-001
```

**Result**: Instant handoff created (HO-XXX), metadata.json updated

### Step 2: Graphiti Episode (Asynchronous - Background)

```python
# Queue episode for background processing
mcp__graphiti__add_memory(
    name="Session Handoff: orc - 2026-01-13T19:30",
    episode_body=json.dumps({
        "session_summary": "Brief summary",
        "timestamp": "ISO 8601",
        "worktree": "orc",
        "todos": [...],
        "decisions": [...],
        "discoveries": [...],
        "open_questions": [...],
        "investigated_files": [...],
        "next_steps": [...]
    }),
    source="json",
    source_description="ORC session handoff",
    group_id="orc"  # or "worktree-{name}"
)
```

**Result**: Episode queued (~20 seconds processing time)

### Key Insight: Don't Wait for Graphiti

The ledger handoff completes immediately. Graphiti processes in background. Users can start new sessions instantly without waiting.

## /g-bootstrap: Hybrid Restoration Pattern

When `/g-bootstrap` runs, it performs a **three-source synthesis**:

### Step 1: Ledger Handoff (Immediate Display)

```bash
# Read latest handoff from ledger
./orc handoff show HO-003
```

**Result**: Full narrative note displayed instantly (<1 second)

### Step 2: Graphiti Semantic Memory (After Display)

```python
# Query recent episodes
mcp__graphiti__get_episodes(
    group_ids=["orc"],
    max_episodes=5
)

# Search for relevant facts
mcp__graphiti__search_memory_facts(
    query="recent work on handoff system",
    group_ids=["orc"],
    max_facts=10
)

# Find related entities
mcp__graphiti__search_nodes(
    query="handoff system architecture",
    group_ids=["orc"],
    max_nodes=10
)
```

**Result**: Deeper insights, cross-references, temporal context

### Step 3: Disk Context (Traditional Bootstrap)

```bash
# Read CLAUDE.md
# Check git history
# Review uncommitted changes
```

**Result**: Current project state

### Synthesis: Best of All Three

Bootstrap combines:
1. **Ledger**: What was being worked on (immediate)
2. **Graphiti**: Why decisions were made (semantic)
3. **Disk**: What changed since then (current)

## Episode Schema for /g-handoff

When creating Graphiti episodes, use this JSON structure:

```json
{
  "session_summary": "High-level summary of session work",
  "timestamp": "2026-01-13T19:30:00Z",
  "worktree": "orc" or "worktree-name",
  "todos": [
    {
      "content": "Task description",
      "status": "pending|in_progress|completed"
    }
  ],
  "decisions": [
    {
      "decision": "What was decided",
      "rationale": "Why it was decided"
    }
  ],
  "discoveries": [
    {
      "insight": "What was discovered",
      "context": "Where/how it was found"
    }
  ],
  "open_questions": [
    {
      "question": "What needs investigation",
      "priority": "high|medium|low"
    }
  ],
  "investigated_files": [
    "path/to/file.ts",
    "path/to/module/"
  ],
  "next_steps": [
    "Recommended action 1",
    "Recommended action 2"
  ]
}
```

**Important**: When using `source="json"`, the `episode_body` must be a JSON string, not a Python dict. Use `json.dumps()` or equivalent.

## Implementation in Skills

### Pattern for /g-handoff Skill

The `/g-handoff` skill file (`global-commands/g-handoff.md`) contains instructions for Claude. When executed:

1. **Claude reads the skill markdown** (instructions)
2. **Claude executes the steps** using available tools
3. **Step 3**: Create ledger handoff via Bash (`./orc handoff create`)
4. **Step 4**: Create Graphiti episode via MCP (`mcp__graphiti__add_memory`)
5. **Step 5**: Display confirmation to user

**Key Point**: Skills are instructions, not code. Claude interprets and executes them.

### Pattern for /g-bootstrap Skill

The `/g-bootstrap` skill file (`global-commands/g-bootstrap.md`) contains instructions for Claude. When executed:

1. **Claude reads the skill markdown** (instructions)
2. **Step 2**: Read ledger handoff via Bash (`./orc handoff show HO-XXX`)
3. **Step 3**: Query Graphiti via MCP (get_episodes, search_memory_facts, search_nodes)
4. **Step 4**: Load disk context via Read tool
5. **Step 5**: Synthesize all three sources into briefing
6. **Step 6**: Display comprehensive briefing to user

## Group ID Strategy

**Consistent group_id across handoff and bootstrap is critical** for context continuity.

### Detection Priority (same for both commands):

1. **--worktree flag**: Explicit override → `worktree-{flag_value}`
2. **Auto-detect from path**:
   - `~/src/worktrees/ml-auth` → `worktree-ml-auth`
   - `~/src/orc` → `orc`
3. **Fallback**: `unknown-session`

### Example:

```bash
# In ~/src/worktrees/ml-auth/
/g-handoff  # group_id: "worktree-ml-auth"

# Later, in new session:
cd ~/src/worktrees/ml-auth/
/g-bootstrap  # group_id: "worktree-ml-auth" (matches!)
```

## Error Handling

### Graphiti Unavailable

```python
try:
    mcp__graphiti__add_memory(...)
except GraphitiUnavailable:
    display("⚠️  Graphiti unavailable - episode not created")
    display("ℹ️  Ledger handoff still created successfully")
    display("💡 Start Graphiti: cd ~/src/graphiti/mcp_server && docker compose up")
```

**Key**: Ledger handoff succeeds even if Graphiti fails. Never block on Graphiti.

### No Previous Episodes

```python
episodes = mcp__graphiti__get_episodes(group_ids=["orc"])
if not episodes:
    display("🆕 Fresh start - no previous session found")
    display("ℹ️  Proceeding with ledger + disk context only")
```

**Key**: Bootstrap gracefully degrades to ledger + disk if Graphiti has no history.

## Testing the Integration

### Test /g-handoff:

```python
# 1. Detect context
pwd = os.getcwd()  # /Users/looneym/src/orc
group_id = "orc"

# 2. Create ledger handoff
./orc handoff create --note "Test handoff" --mission MISSION-001

# 3. Create Graphiti episode
mcp__graphiti__add_memory(
    name=f"Session Handoff: {group_id} - {timestamp}",
    episode_body=json.dumps({...}),
    source="json",
    group_id=group_id
)

# 4. Verify both succeeded
./orc handoff list  # Should show new HO-XXX
# Wait ~20 seconds
mcp__graphiti__get_episodes(group_ids=[group_id])  # Should include new episode
```

### Test /g-bootstrap:

```bash
# In new Claude session:
/g-bootstrap

# Should display:
# 1. Ledger handoff (immediate - HO-XXX narrative)
# 2. Graphiti insights (semantic memory from episodes)
# 3. Git/disk context (current state)
```

## Performance Characteristics

| Operation | Tool | Time | Blocking? |
|-----------|------|------|-----------|
| Create ledger handoff | `orc handoff create` | <100ms | Yes (but fast) |
| Create Graphiti episode | `mcp__graphiti__add_memory` | ~20s | No (background) |
| Read ledger handoff | `orc handoff show` | <100ms | Yes (but fast) |
| Query Graphiti episodes | `mcp__graphiti__get_episodes` | ~1-2s | Yes (after display) |
| Query Graphiti facts | `mcp__graphiti__search_memory_facts` | ~1-2s | Yes (enrichment) |
| Query Graphiti nodes | `mcp__graphiti__search_nodes` | ~1-2s | Yes (enrichment) |

**Total bootstrap time**: <1 second for initial context (ledger), +3-5 seconds for Graphiti enrichment (non-blocking for user)

## Best Practices

### DO:
- ✅ Create ledger handoff first (priority)
- ✅ Queue Graphiti episode in background (async)
- ✅ Display ledger context immediately in bootstrap
- ✅ Enrich with Graphiti after user has initial context
- ✅ Use consistent group_id detection logic
- ✅ Handle Graphiti errors gracefully (don't block)

### DON'T:
- ❌ Wait for Graphiti before showing context
- ❌ Block handoff creation on Graphiti success
- ❌ Skip ledger handoff if Graphiti fails
- ❌ Use different group_id logic in handoff vs bootstrap
- ❌ Store raw code in Graphiti (concepts only)

## Future Enhancements

### Linking Handoffs to Episodes

Consider adding CLI command:

```bash
orc handoff link HO-003 --graphiti-uuid <episode-uuid>
```

This would update the `graphiti_episode_uuid` column in the handoffs table, enabling bidirectional references.

### Cross-Investigation Insights

With `--full` flag on /g-bootstrap:

```python
# Query across ALL group_ids for related work
mcp__graphiti__search_memory_facts(
    query="authentication patterns",
    # No group_ids filter - search everything
    max_facts=20
)
```

This surfaces discoveries from other investigations that might be relevant.

## Conclusion

The two-tier memory architecture combines the best of both worlds:
- **Ledger**: Fast, structured, deterministic queries
- **Graphiti**: Semantic understanding, temporal reasoning, cross-investigation insights

By prioritizing ledger handoffs and using Graphiti for enrichment, we achieve instant session continuity with deep contextual understanding.
