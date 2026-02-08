---
name: orc-help
description: Orientation to ORC skills. Shows categories (ship, imp, orc, goblin) with examples. Use when user asks for help with ORC or wants to discover available skills.
---

# ORC Help

Orientation skill that shows available ORC skill categories with examples.

## Usage

```
/orc-help                 (show category overview)
/orc-help ship            (show all ship-* skills)
/orc-help imp             (show all imp-* skills)
```

## Behavior

### Step 1: Read Skills

Read the deployed skills directory:
```bash
ls ~/.claude/skills/
```

Filter to ORC prefixes only:
- `ship-*` - Shipment workflow
- `imp-*` - Implementation workflow
- `orc-*` - Utilities
- `goblin-*` - Review & escalation

### Step 2: Check for Argument

If a category argument is provided (ship, imp, orc, goblin):
- List all skills in that category
- For each, read frontmatter and show name + description
- Skip to Step 4

If no argument, continue to Step 3.

### Step 3: Show Category Overview

Display the category overview:

```
📋 ORC Skill Categories

**ship-*** - Shipment Workflow
  Create and manage shipments from exploration to deployment
  Examples: /ship-new, /ship-plan, /ship-deploy

**imp-*** - Implementation
  IMP agent workflow for executing tasks
  Examples: /imp-start, /imp-plan-create, /imp-rec

**orc-*** - Utilities
  General ORC commands and maintenance
  Examples: /orc-interview, /orc-debug, /orc-help

**goblin-*** - Review & Escalation
  Gatehouse review workflows
  Examples: /goblin-escalation-receive

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Want details on a category? Say "tell me about ship" or run /orc-help ship
```

### Step 4: Check for Missing Frontmatter

For any skills that couldn't be parsed (missing or invalid frontmatter), add a warning:

```
⚠️ Skills with missing frontmatter: orc-foo, orc-bar
```

## Category Drill-Down

When user asks about a specific category (e.g., "tell me about ship" or `/orc-help ship`):

1. List all skills with that prefix from ~/.claude/skills/
2. For each skill, read SKILL.md and extract frontmatter
3. Display as:

```
📦 Ship Skills (Shipment Workflow)

/ship-new
  Create a new shipment for implementation work

/ship-synthesize
  Knowledge compaction for shipments

/ship-plan
  C2/C3 engineering review that creates tasks

/ship-queue
  View and manage the shipyard queue

/ship-deploy
  Merge shipment branch to master

/ship-verify
  Post-deploy verification

/ship-complete
  Mark shipment as complete
```

## Notes

- Only shows ORC-related skills (filtered by prefix)
- Reads from deployed location (~/.claude/skills/), not source
- Gracefully handles missing frontmatter with warnings
