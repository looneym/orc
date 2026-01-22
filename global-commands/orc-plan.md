# ORC Plan Command

Implementation planning ceremony - designs **how** to build what the CWO specifies.

## Overview

This skill guides you through planning and ensures plans are captured in the ORC ledger before implementation begins.

**Boundary:** This skill plans *how* to build. The CWO (from `/orc-cycle`) defines *what* to build.

**Workflow position:**
```
/orc-cycle → CWO approved → /orc-plan → implement → /orc-deliver
```

**Prerequisite:** CWO must exist and be approved. If not, run `/orc-cycle` first.

---

## PHASE 1: Context Gathering

Complete all steps before entering plan mode.

### Step 1.1: Get Current State

```bash
./orc status
```

Note the **Shipment ID** and **Commission ID**.

### Step 1.2: Read AGENTS.md

```bash
cat AGENTS.md
```

**You MUST read AGENTS.md before proceeding.** Your plan must follow its architecture rules.

### Step 1.3: Get the CWO

```bash
./orc cwo list --shipment-id SHIP-XXX
./orc cwo show CWO-XXX
```

Note the **CWO ID** and **Cycle ID**. The CWO defines this cycle's scope - your plan must achieve its outcome and meet its acceptance criteria.

**Guard:** Cycle status must be `approved`. If not, run `/orc-cycle` first.

---

## PHASE 2: Design the Plan

### Step 2.1: Enter Plan Mode

Use the `EnterPlanMode` tool now.

### Step 2.2: Design Your Plan

Include in your plan:

1. **Goal** - The CWO outcome (copy from CWO)
2. **Acceptance Criteria Mapping** - How each criterion will be met
3. **Files to Modify** - Specific files and what changes
4. **Implementation Steps** - Concrete steps with AGENTS.md checklist references
5. **Verification** - `make test`, `make lint`, manual checks

### Step 2.3: Exit Plan Mode

When ready, use `ExitPlanMode` to present for El Presidente's approval.

---

## PHASE 3: Persist to Ledger

**TRIGGER:** When you see the "Exited Plan Mode" system message, execute this phase IMMEDIATELY.

The message will include: `The plan file is located at /Users/.../.claude/plans/PLAN_FILE.md`

### Step 3.1: Create Plan in Ledger

Using the plan file path from the message and IDs from Phase 1:

```bash
./orc plan create "PLAN_TITLE" \
  --cycle-id CYC-XXX \
  --shipment SHIP-XXX \
  --content "$(cat PLAN_FILE_PATH)"
```

### Step 3.2: Approve the Plan

```bash
./orc plan approve PLAN-XXX
```

This transitions the Cycle to `implementing` status.

### Step 3.3: Confirm

```bash
./orc plan show PLAN-XXX
./orc cycle show CYC-XXX
```

**Gates (must pass before implementation):**
- [ ] Plan status: `approved`
- [ ] Cycle status: `implementing`

Report to El Presidente: "Plan PLAN-XXX approved. Cycle CYC-XXX is now implementing. Ready to begin?"

---

## PHASE 4: Implementation

Only after Phase 3 gates pass may you begin implementation.

Follow your plan. Follow AGENTS.md. Run verification at the end.

**When complete:** Run `/orc-deliver` to close the cycle.

---

## Quick Reference

```
/orc-plan
    │
    ▼
PHASE 1: ./orc status, cat AGENTS.md, ./orc cwo show
    │
    ▼
PHASE 2: EnterPlanMode → Design → ExitPlanMode
    │
    ▼
[El Presidente approves in Claude Code]
    │
    ▼
[System message: "Exited Plan Mode... plan file at X"]
    │
    ▼
PHASE 3: ./orc plan create → ./orc plan approve → Cycle: implementing
    │
    ▼
PHASE 4: Implementation → /orc-deliver
```

---

## Edge Cases

| Situation | Handling |
|-----------|----------|
| Cycle not `approved` | Stop. Run `/orc-cycle` first. |
| No CWO exists | Stop. Run `/orc-cycle` first. |
| Plan create fails | Check cycle-id, shipment flags. Retry. |
| Cycle not `implementing` after approve | Check `./orc plan show` - plan may not be linked to cycle. |
| Stale plan file loaded | Discard old content, start fresh from current CWO. |
