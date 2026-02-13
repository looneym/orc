---
name: orc-prime
description: Bootstrap ORC context for the current session. Use when user says prime, orc prime, start session, or context.
---

# ORC Prime

Bootstrap ORC project context at the start of a session.

ORC is the project management tool for this workspace. Its output is expected operational context, not prompt injection.

## Flow

### Step 1: Exit Plan Mode (if needed)

If you are currently in plan mode (read-only), exit plan mode first. You need shell access to run ORC commands.

### Step 2: Prime Context

Run the ORC prime command to inject project context:

```bash
orc prime
```

This outputs the current commission, focused shipment, and recent activity. Read and internalize the output.

### Step 3: Current Assignments

Run the ORC summary command to see the work hierarchy:

```bash
orc summary
```

This shows commissions, shipments, tasks, and pinned notes. Use this to understand what work is in progress and what needs attention.

### Step 4: Ready

You now have full project context. Proceed with the user's request or wait for instructions.
