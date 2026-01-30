---
name: imp-escalate
description: Escalate plan to gatehouse for human review.
---

# IMP Escalate

Escalate the current plan to the gatehouse for human review.

## Usage

`/imp-escalate --reason "reason for escalation"`

## Flow

1. **Get current plan**
   ```bash
   orc plan list --task TASK-xxx --status submitted
   ```

2. **Escalate plan**
   ```bash
   orc plan escalate PLAN-xxx --reason "reason for escalation"
   ```

3. **Output**
   "Escalated as ESC-xxx. Waiting for human resolution. When El Presidente arrives to help, they should run /imp-unblock."

## When to Escalate

- Self-review identified scope or architecture concerns
- Task requirements are ambiguous
- Plan requires decisions beyond IMP authority
- Significant risk or impact identified

## Notes

- Escalation fires a "flare" to the gatehouse
- The IMP should pause and wait for resolution
- This is a normal part of the workflow, not a failure
