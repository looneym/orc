---
name: exorcism
description: Ledger maintenance skill for cleaning entropy and converging exploration into work. Use when user says /exorcism or wants to tidy a conclave, consolidate notes, or synthesize exploration into a spec/shipment.
---

# Exorcism Skill

Ledger maintenance for cleaning up entropy, consolidating ideas, and maintaining semantic health.

## Objectives

| Objective | Purpose | Creates |
|-----------|---------|---------|
| **Clean** | Tidy existing container | Nothing new (maybe exorcism note) |
| **Ship** | Converge exploration into work | Draft shipment + spec |

### Selection

Explicit upfront or agent-proposed after survey:

```bash
/exorcism --clean CON-018
/exorcism --ship CON-018
```

Or let the agent propose after survey:
```
Agent: Surveyed CON-018. State: Chaotic.
       Recommend: Clean first, then ship?
       [c]lean / [s]hip / [b]oth sequentially
```

## Survey Flow

When `/exorcism` is invoked, follow these steps:

### Step 1: Identify Target

If argument provided (e.g., `/exorcism CON-018`):
- Use the provided container ID

If no argument:
```bash
orc status
```
- Use the focused container from output
- If no focus: "No target specified. Which container should I survey? (CON-xxx, SHIP-xxx, or TOME-xxx)"

### Step 2: Collect Data

For conclave:
```bash
orc conclave show CON-xxx
orc note list --conclave CON-xxx
```

For shipment:
```bash
orc shipment show SHIP-xxx
orc note list --shipment SHIP-xxx
```

For tome:
```bash
orc tome show TOME-xxx
orc note list --tome TOME-xxx
```

### Step 3: Analyze

From the collected data, compute:
- Total note count
- Notes grouped by type (idea, question, concern, etc.)
- Notes grouped by status (open vs closed)
- Open questions (type=question, status=open)
- Unaddressed concerns (type=concern, status=open)
- Potential duplicates (notes with very similar titles)
- Stale notes (status=open but old created_at, no recent updated_at)

### Step 4: Assess State

**Chaotic** if any of:
- 3+ open questions
- 2+ unaddressed concerns
- Multiple potential duplicates
- High ratio of ideas to decisions/specs

**Orderly** if:
- Questions mostly answered
- Concerns addressed
- Ideas synthesized into decisions/specs

### Step 5: Present Survey

Output format:
```
## Survey: [CONTAINER-ID] ([Container Title])

### Summary
- X tomes (if conclave), Y notes total
- Z open questions, W unaddressed concerns

### Notes by Type
| Type     | Open | Closed |
|----------|------|--------|
| idea     | X    | Y      |
| question | X    | Y      |
| concern  | X    | Y      |
| spec     | X    | Y      |
| decision | X    | Y      |

### State: [Chaotic/Orderly]
Signals: [list specific signals that led to assessment]

### Recommendation
**[Clean/Ship]** - [rationale based on state]

Select objective: [c]lean / [s]hip / [b]oth sequentially
```

### Step 6: Await Selection

Wait for user to choose c, s, or b.
- If clean: proceed to theme selection for clean patterns
- If ship: proceed to theme selection for ship patterns
- If both: clean first, then ship

## Theme Selection & Interview

After objective is selected:

### Theme Selection

- Identify 3-5 themes from the survey data (e.g., "Scattered ideas about X", "Unanswered questions about Y")
- Present themes to human for selection

### Interview (per theme)

- Max 5 questions
- Progress indicator: "2/3 questions remaining"
- Each question: context + why it matters + choices
- Choices map to patterns

## Patterns

| Pattern | When | Move | Objective |
|---------|------|------|-----------|
| SYNTHESIZE | Multiple notes → one conclusion | Combine into decision/spec | Ship |
| EXTRACT-LAYER | Note mixes C4 levels | Split by layer | Both |
| CLOSE-SUPERSEDED | Content now in better artifact | Close original | Both |
| CONSOLIDATE-DUPLICATES | Same concept, different words | Merge | Clean |
| PROMOTE-TO-DECISION | Implicit decision buried | Extract to decision note | Ship |
| BRIDGE-CONTEXT | Orphan L3/L4 detail | Add reference to L1/L2 | Clean |
| DEFER-TO-LIBRARY | Valid but not now | Park | Clean |
| SPLIT-SCOPE | Kitchen-sink container | Split into focused pieces | Clean |

## Commands

Pattern execution uses CLI operations:

```bash
orc note merge <source> <target>
orc note close <id> --reason <reason> [--by <note-id>]
```

**Reason vocabulary:** superseded, synthesized, resolved, deferred, duplicate, stale

## Exorcism Note

Each maintenance session can produce an `exorcism` note as a record:

```markdown
# Exorcism: CON-018 Consolidation

## Before
- 4 tomes, 17 notes scattered
- Multiple overlapping specs

## After
- All tomes closed
- Unified spec (NOTE-311)

## Key decisions
- Single command: /exorcism
- Two objectives: clean vs ship
```

## Reference

See NOTE-311 for full specification.
