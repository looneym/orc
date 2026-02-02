---
name: imp-poll
description: Check shipyard queue and claim work. IMP uses this to find available shipments when idle.
---

# IMP Poll

Check for available shipments (ready_for_imp status) and claim work.

## When to Use

- IMP has no current shipment assigned
- IMP completed a shipment and needs new work
- IMP wants to see what's available

## Flow

### Step 1: Check Current State

```bash
orc status
```

Verify:
- Is there already a shipment focused for this workbench? If yes, suggest `/imp-start` instead.
- Get the commission context for filtering.

### Step 2: Display Available Shipments

```bash
orc shipment list --available
```

Show available shipments with this format:
```
Checking for available shipments...

ID        TITLE                           TASKS   COMMISSION
--        -----                           -----   ----------
SHIP-240  Critical hotfix                 3       COMM-001
SHIP-237  Plan/Apply Refactor             11      COMM-001

Options:
[c] Claim #1 (top of list)
[n] Claim specific shipment
[r] Refresh
[q] Quit
```

### Step 3: Handle Selection

**[c] Claim top shipment:**
```bash
orc focus SHIP-xxx
```

This focuses the shipment and transitions it to `implementing` status.

Output: "Claimed SHIP-xxx. Run `/imp-start` to begin work."

**[n] Claim specific:**
Prompt: "Enter shipment number (1-N) or ID (SHIP-xxx):"

If number: Map to shipment ID from displayed list
If ID: Use directly

```bash
orc focus SHIP-xxx
```

**[r] Refresh:**
Re-run `orc shipment list --available` and display updated list.

**[q] Quit:**
Exit without claiming.

## After Claiming

After successful claim:
```
✓ Claimed SHIP-xxx: [title]
  Status: implementing
  Tasks: X ready

Run /imp-start to begin work on the first task.
```

## Error Handling

- No available shipments → "No shipments available. Queue is empty."
- Already has shipment → "Already focused on SHIP-xxx. Run /imp-start or /imp-nudge."
- Claim fails → "Failed to focus shipment: [error]"

## Example Session

```
> /imp-poll

Checking for available shipments...

ID        TITLE                    TASKS   COMMISSION
--        -----                    -----   ----------
SHIP-240  Critical hotfix          3       COMM-001
SHIP-237  Plan/Apply Refactor      11      COMM-001

[c] Claim #1 (SHIP-240)
[n] Claim specific
[r] Refresh
[q] Quit

> c

[runs orc focus SHIP-240]

✓ Claimed SHIP-240: Critical hotfix
  Status: implementing
  Tasks: 3 ready

Run /imp-start to begin work on the first task.
```
