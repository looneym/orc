---
name: orc-workbench
description: Guide through creating a workbench. Use when user says /orc-workbench or wants to create a new workbench for implementation work.
---

# Workbench Creation Skill

Guide users through creating a workbench for implementation work.

## Usage

```
/orc-workbench <name>
/orc-workbench              (will prompt for name)
```

## Flow

### Step 1: Gather Required Info

If name not provided, ask:
- "What should this workbench be called?" (slug format: lowercase, hyphens)

### Step 2: Detect Context

```bash
orc status
orc workshop list
```

Identify:
- Current workshop (if in workshop context)
- Available workshops
- If no workshop detected, ask which workshop

### Step 3: Optional Repo Link

Ask if workbench should be linked to a repo:
```bash
orc repo list
```

If yes, get repo ID.

### Step 4: Create Workbench Record

```bash
orc workbench create <name> --workshop WORK-xxx [--repo-id REPO-xxx]
```

Capture the created BENCH-xxx ID.

### Step 5: Apply Infrastructure

```bash
orc infra apply WORK-xxx
```

This creates:
- Git worktree at ~/wb/<name>
- .orc/config.json with place_id

### Step 6: Confirm Ready

Output:
```
Workbench created:
  BENCH-xxx: <name>
  Path: ~/wb/<name>
  Workshop: WORK-xxx

To start working:
  cd ~/wb/<name>
  # Or attach to TMux: orc tmux connect WORK-xxx
```

## Example Session

```
User: /orc-workbench auth-refactor

Agent: [runs orc status, detects WORK-003]
       [runs orc workbench create auth-refactor --workshop WORK-003]
       [runs orc infra apply WORK-003]

Agent: Workbench created:
         BENCH-xxx: auth-refactor
         Path: ~/wb/auth-refactor
         Workshop: WORK-003

       To start working:
         cd ~/wb/auth-refactor
```
