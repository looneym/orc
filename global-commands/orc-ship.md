# ORC Ship Command

Shipment closure ceremony - closes a shipment by creating a REC with process reflection.

## Overview

This skill guides you through the structured shipment closure ceremony that:
- Verifies all cycles are complete
- Gathers process reflection from El Presidente
- Creates a REC (Receipt) documenting the shipment delivery
- Marks the shipment as complete
- Provides guidance on code merging

**Boundary:** This skill is about **closing the shipment**, not individual cycles. Use `/orc-deliver` to close cycles first.

**Workflow position:**
```
/orc-cycle → /orc-plan → implement → /orc-deliver → ... → /orc-ship
```

**Prerequisite:** All cycles for the shipment must be complete.

---

## PHASE 1: Context Gathering

### Step 1.1: Get Current State

```bash
./orc status
```

Note the focused **Shipment ID**.

### Step 1.2: Check Cycle Completion

```bash
./orc cycle list --shipment-id SHIP-XXX
```

**Guard:** ALL cycles must have status `complete` or `failed`.

If any cycles are not complete/failed, stop and report:
"Cannot close shipment. The following cycles are not complete: CYC-XXX (status), CYC-YYY (status)"

### Step 1.3: Check for Existing REC

```bash
./orc rec list --shipment-id SHIP-XXX
```

If a REC already exists, this shipment has already been closed. Show the existing REC summary and exit:
"Shipment SHIP-XXX already has REC-XXX. Showing existing receipt..."

```bash
./orc rec show REC-XXX
```

---

## PHASE 2: Process Reflection

### Step 2.1: Gather Reflection

Ask El Presidente for process reflection:

**What went well during this shipment?**
(Tooling improvements, team dynamics, technical wins, etc.)

**What was difficult or challenging?**
(Blockers, unclear requirements, technical debt, etc.)

**Is there any deferred work?**
(Features descoped, known issues, future improvements, etc.)

### Step 2.2: Collect Cycle Summary

List completed cycles and their CRECs:

```bash
./orc crec list --shipment-id SHIP-XXX
```

Note the cycle IDs and their outcomes for the REC evidence.

---

## PHASE 3: Create Receipt

### Step 3.1: Create the REC

```bash
./orc rec create "Shipment SHIP-XXX delivered" \
  --shipment-id SHIP-XXX \
  --evidence "## Process Reflection

### What Went Well
[El Presidente's response]

### What Was Difficult
[El Presidente's response]

### Deferred Work
[El Presidente's response]

## Cycles Completed
- CYC-XXX: [outcome from CREC]
- CYC-YYY: [outcome from CREC]
..."
```

Note the returned **REC ID**.

### Step 3.2: Submit and Verify REC

```bash
./orc rec submit REC-XXX
./orc rec verify REC-XXX
```

---

## PHASE 4: Complete Shipment

### Step 4.1: Mark Shipment Complete

```bash
./orc shipment complete SHIP-XXX
```

### Step 4.2: Merge Placeholder

Print the following message to El Presidente:

```
┌────────────────────────────────────────────────────────┐
│ CODE MERGING                                           │
├────────────────────────────────────────────────────────┤
│ The shipment is now complete in the ORC ledger.        │
│                                                        │
│ To merge your code changes:                            │
│ 1. Review the branch for this shipment                 │
│ 2. Create a PR to your target branch                   │
│ 3. Follow your team's merge/review process             │
│                                                        │
│ ORC does not manage git merges automatically.          │
└────────────────────────────────────────────────────────┘
```

---

## PHASE 5: Report Final State

### Step 5.1: Show Summary

```bash
./orc shipment show SHIP-XXX
./orc rec show REC-XXX
```

Report to El Presidente:
- Shipment: SHIP-XXX - status `complete`
- REC: REC-XXX - status `verified`
- Cycles completed: N

"Shipment SHIP-XXX is now complete. The REC captures your process reflection and cycle outcomes."

---

## Quick Reference

```
PREREQUISITE: All cycles complete
    │
    ▼
/orc-ship
    │
    ▼
PHASE 1: Check status, verify cycles complete, check for existing REC
    │
    ▼
PHASE 2: Gather process reflection from El Presidente
    │
    ▼
PHASE 3: Create REC with reflection and cycle summary
    │
    ▼
PHASE 4: Complete shipment, print merge placeholder
    │
    ▼
PHASE 5: Report final state
```

---

## Edge Cases

| Situation | Handling |
|-----------|----------|
| Cycles not complete | Error: list incomplete cycle IDs, refuse to proceed |
| REC already exists | Show existing REC, exit (idempotent) |
| Failed cycles exist | Include in reflection, cycles with `failed` status are acceptable |
| No cycles exist | Warn but allow: shipment may have been simple/no-cycle work |
| CLI command fails | Stop, inform El Presidente, ask how to proceed |
