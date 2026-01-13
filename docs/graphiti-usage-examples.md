# Graphiti Usage Examples for ORC

This document provides practical examples of using Graphiti integration in ORC global commands.

## Example 1: Creating a Handoff Episode in /g-handoff

### Context Detection

```python
import os
import json
from datetime import datetime

# Detect current worktree context
cwd = os.getcwd()

if "worktrees" in cwd:
    # Extract worktree name from path
    # Example: /Users/looneym/src/worktrees/ml-auth-refactor
    worktree_name = cwd.split("/worktrees/")[1].split("/")[0]
    group_id = f"worktree-{worktree_name}"
elif "orc" in cwd:
    group_id = "orc"
else:
    group_id = "unknown-session"

print(f"📊 Context: {group_id}")
```

### Gather Session State

```python
# Example: Extract from TodoWrite state and conversation analysis
session_state = {
    "session_summary": "Implemented Graphiti integration for handoff system",
    "timestamp": datetime.utcnow().isoformat() + "Z",
    "worktree": group_id.replace("worktree-", "") if "worktree" in group_id else group_id,
    "todos": [
        {
            "content": "Implement Graphiti episode creation in /g-handoff command",
            "status": "completed"
        },
        {
            "content": "Test Graphiti episode query in /g-bootstrap pattern",
            "status": "in_progress"
        },
        {
            "content": "Test complete handoff → bootstrap cycle",
            "status": "pending"
        }
    ],
    "decisions": [
        {
            "decision": "Use JSON source type for structured episode data",
            "rationale": "Enables Graphiti to automatically extract entities and relationships from structured data, providing richer semantic understanding than plain text."
        },
        {
            "decision": "Prioritize ledger handoff over Graphiti episode creation",
            "rationale": "Ledger provides instant (<1s) structured context. Graphiti processes in background (~20s). Users don't wait."
        }
    ],
    "discoveries": [
        {
            "insight": "Graphiti episodes with JSON source automatically extract entities",
            "context": "When using source='json', Graphiti's AI extracts companies, products, people, and relationships without manual tagging."
        },
        {
            "insight": "Skills are instruction documents, not code files",
            "context": "Global commands (g-handoff.md, g-bootstrap.md) are markdown instructions that Claude interprets and executes using available tools."
        }
    ],
    "open_questions": [
        {
            "question": "Should we add a CLI command to link handoffs with Graphiti UUIDs?",
            "priority": "medium"
        },
        {
            "question": "How long does Graphiti take to process large JSON episodes?",
            "priority": "low"
        }
    ],
    "investigated_files": [
        "global-commands/g-handoff.md",
        "global-commands/g-bootstrap.md",
        "docs/graphiti-integration-guide.md"
    ],
    "next_steps": [
        "Test Graphiti query patterns for bootstrap",
        "Validate episode retrieval after processing completes",
        "Document performance characteristics",
        "Test full handoff → bootstrap cycle end-to-end"
    ]
}
```

### Create Graphiti Episode

```python
# Convert to JSON string (IMPORTANT: must be string, not dict)
episode_json = json.dumps(session_state)

# Create episode via MCP tool
result = mcp__graphiti__add_memory(
    name=f"Session Handoff: {group_id} - {session_state['timestamp'][:16]}",
    episode_body=episode_json,
    source="json",
    source_description="ORC session handoff from g-handoff command",
    group_id=group_id
)

print(f"🧠 Graphiti episode queued: {result['result']['message']}")
print(f"   Group: {group_id}")
print(f"   Processing in background (~20s)")
```

### Full /g-handoff Example

```python
# Step 1: Detect context
group_id = detect_group_id()  # Returns "orc" or "worktree-X"

# Step 2: Gather session state
session_state = gather_session_state()  # Analyzes TodoWrite + conversation

# Step 3: Create ledger handoff (PRIORITY)
bash_command = f"""
./orc handoff create \\
  --note "$(cat <<'EOF'
{craft_narrative_note(session_state)}
EOF
)" \\
  --mission MISSION-001 \\
  --operation OP-001
"""
run_bash(bash_command)

# Step 4: Create Graphiti episode (BACKGROUND)
episode_json = json.dumps(session_state)
mcp__graphiti__add_memory(
    name=f"Session Handoff: {group_id} - {timestamp}",
    episode_body=episode_json,
    source="json",
    source_description="ORC session handoff",
    group_id=group_id
)

# Step 5: Confirm dual flush
print("✓ Ledger handoff created: HO-XXX")
print("🧠 Graphiti episode queued")
print("✓ Context preserved. Safe to start new session with /g-bootstrap")
```

## Example 2: Querying Episodes in /g-bootstrap

### Read Ledger Handoff (Immediate)

```python
# Get latest handoff ID from metadata
import json

metadata = json.load(open(os.path.expanduser("~/.orc/metadata.json")))
latest_handoff_id = metadata["current_handoff_id"]

# Display ledger handoff immediately
bash_command = f"./orc handoff show {latest_handoff_id}"
handoff_note = run_bash(bash_command)

print("# 📝 Ledger Handoff (from Previous Claude)")
print(handoff_note)
```

### Query Graphiti Episodes (After Display)

```python
# After displaying ledger handoff, enrich with Graphiti

# 1. Get recent episodes
episodes = mcp__graphiti__get_episodes(
    group_ids=[group_id],
    max_episodes=5
)

print("\n## 🧠 Semantic Memory (from Graphiti)\n")

# 2. Search for relevant facts
facts = mcp__graphiti__search_memory_facts(
    query="recent decisions and discoveries in handoff system",
    group_ids=[group_id],
    max_facts=10
)

# 3. Find related entities/components
nodes = mcp__graphiti__search_nodes(
    query="handoff system architecture components",
    group_ids=[group_id],
    max_nodes=10
)

# Parse and display insights
for episode in episodes:
    # Extract JSON data if source was JSON
    if episode.get("source") == "json":
        data = json.loads(episode["content"])
        display_decisions(data.get("decisions", []))
        display_discoveries(data.get("discoveries", []))
        display_open_questions(data.get("open_questions", []))
```

### Full /g-bootstrap Example

```python
# Step 1: Detect context
group_id = detect_group_id()

# Step 2: Read ledger handoff (IMMEDIATE)
metadata = json.load(open(os.path.expanduser("~/.orc/metadata.json")))
latest_handoff_id = metadata["current_handoff_id"]

handoff = run_bash(f"./orc handoff show {latest_handoff_id}")

print("# 🚀 Hybrid Bootstrap - " + group_id)
print("\n## 📝 Ledger Handoff (from Previous Claude - " + latest_handoff_id + ")")
print(handoff)

# Step 3: Query Graphiti (ENRICHMENT - after display)
print("\n## 🧠 Semantic Memory (from Graphiti)")

episodes = mcp__graphiti__get_episodes(group_ids=[group_id], max_episodes=5)
facts = mcp__graphiti__search_memory_facts(
    query="recent work decisions discoveries",
    group_ids=[group_id],
    max_facts=10
)
nodes = mcp__graphiti__search_nodes(
    query="architecture components",
    group_ids=[group_id],
    max_nodes=10
)

# Step 4: Synthesize briefing
briefing = synthesize_briefing(
    ledger_handoff=handoff,
    graphiti_episodes=episodes,
    graphiti_facts=facts,
    graphiti_nodes=nodes,
    git_history=get_recent_commits(),
    git_status=get_git_status()
)

print(briefing)
```

## Example 3: Cross-Investigation Queries (--full flag)

```python
# Query across ALL worktrees for related patterns

# Search for authentication patterns across all investigations
facts = mcp__graphiti__search_memory_facts(
    query="authentication patterns security decisions",
    # No group_ids = search everything
    max_facts=20
)

print("## 🔗 Cross-Investigation Insights\n")
for fact in facts:
    print(f"- **{fact['name']}**: {fact['fact']}")
    print(f"  - From: {fact['group_id']}")
    print(f"  - When: {fact['valid_at']}")
```

## Example 4: Error Handling Patterns

### Graceful Graphiti Unavailability

```python
try:
    episode_json = json.dumps(session_state)
    result = mcp__graphiti__add_memory(
        name=f"Session Handoff: {group_id}",
        episode_body=episode_json,
        source="json",
        group_id=group_id
    )
    print("🧠 Graphiti episode queued successfully")
except Exception as e:
    print("⚠️  Graphiti unavailable - episode not created")
    print("ℹ️  Ledger handoff still created successfully")
    print("💡 Start Graphiti: cd ~/src/graphiti/mcp_server && docker compose up")
    # Don't fail - ledger handoff already succeeded
```

### No Previous Episodes

```python
episodes = mcp__graphiti__get_episodes(group_ids=[group_id])

if not episodes or len(episodes["result"]["episodes"]) == 0:
    print("🆕 Fresh start - no previous session found in Graphiti")
    print("ℹ️  Proceeding with ledger + disk context only")
else:
    # Process episodes normally
    for episode in episodes["result"]["episodes"]:
        display_episode_insights(episode)
```

## Example 5: Testing the Integration

### Test Episode Creation

```bash
# In your Claude session:
cd ~/src/orc

# Create test episode
test_data = {
    "session_summary": "Test episode creation",
    "timestamp": "2026-01-13T20:00:00Z",
    "worktree": "orc",
    "todos": [{"content": "Test task", "status": "completed"}],
    "decisions": [{"decision": "Test decision", "rationale": "Testing"}],
    "discoveries": [],
    "open_questions": [],
    "investigated_files": [],
    "next_steps": ["Verify episode was created"]
}

mcp__graphiti__add_memory(
    name="Test Episode: orc - 2026-01-13T20:00",
    episode_body=json.dumps(test_data),
    source="json",
    source_description="Test episode",
    group_id="orc"
)

# Wait ~20 seconds for processing

# Verify creation
episodes = mcp__graphiti__get_episodes(group_ids=["orc"], max_episodes=1)
print(episodes)
```

### Test Episode Query

```python
# Query recent episodes
episodes = mcp__graphiti__get_episodes(group_ids=["orc"], max_episodes=3)

print("Recent episodes in 'orc' group:")
for ep in episodes["result"]["episodes"]:
    print(f"- {ep['name']}")
    print(f"  Created: {ep['created_at']}")
    print(f"  Source: {ep.get('source', 'text')}")

    # If JSON source, parse and display structure
    if ep.get("source") == "json":
        try:
            data = json.loads(ep["content"])
            print(f"  Todos: {len(data.get('todos', []))}")
            print(f"  Decisions: {len(data.get('decisions', []))}")
            print(f"  Discoveries: {len(data.get('discoveries', []))}")
        except:
            print("  (JSON parsing failed)")
    print()
```

## Example 6: Performance Measurement

```python
import time

# Measure ledger handoff speed
start = time.time()
bash_result = run_bash("./orc handoff show HO-003")
ledger_time = time.time() - start
print(f"Ledger handoff read: {ledger_time*1000:.0f}ms")

# Measure Graphiti query speed
start = time.time()
episodes = mcp__graphiti__get_episodes(group_ids=["orc"], max_episodes=5)
graphiti_time = time.time() - start
print(f"Graphiti episodes query: {graphiti_time*1000:.0f}ms")

# Typical results:
# Ledger handoff read: 50-100ms
# Graphiti episodes query: 1000-2000ms
#
# Bootstrap shows ledger immediately, then enriches with Graphiti
```

## Key Takeaways

1. **Always prioritize ledger handoff** - instant context (<1s)
2. **Use JSON source for structured data** - enables entity extraction
3. **Query Graphiti after displaying ledger** - enrichment, not blocking
4. **Consistent group_id detection** - same logic in handoff and bootstrap
5. **Graceful error handling** - never block on Graphiti failures
6. **Wait ~20 seconds** - episode processing is async
7. **Test with small data first** - validate integration before complex episodes

## Next Steps

After understanding these examples:
1. Implement the patterns in /g-handoff skill
2. Implement the query patterns in /g-bootstrap skill
3. Test full cycle: handoff → wait → bootstrap
4. Measure performance with real workloads
5. Document any edge cases or gotchas discovered
