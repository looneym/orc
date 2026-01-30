---
name: imp-unblock
description: Human entered to help resolve escalation. Show context, collaborate, resolve, prompt new plan.
---

# IMP Unblock

Human has arrived to help resolve an escalation. Show context and collaborate to resolve.

## Flow

1. **Find pending escalation**
   ```bash
   orc escalation list --status pending
   ```

2. **Show full context**
   ```bash
   orc escalation show ESC-xxx
   orc plan show PLAN-xxx
   orc task show TASK-xxx
   ```

   Present to El Presidente:
   - Escalation reason
   - Original plan content
   - Task description
   - What the IMP was uncertain about

3. **Collaborate**
   "Let's figure this out together, El Presidente. The concern was: [reason]"

   Work with the human to:
   - Clarify requirements
   - Resolve architectural questions
   - Make decisions on approach

4. **Resolve escalation**
   Once aligned:
   ```bash
   orc escalation resolve ESC-xxx --outcome approved
   ```
   Or with notes:
   ```bash
   orc escalation resolve ESC-xxx --outcome approved --notes "Clarified that..."
   ```

5. **Output**
   "Escalation ESC-xxx resolved. Run /imp-plan-create to create a new plan incorporating what we discussed."

## Outcomes

- `approved` - Proceed with adjusted approach
- `rejected` - Task should not be done as specified
- `deferred` - Needs more information, pause for now

## Notes

- This skill is invoked by the human, not autonomously
- The goal is collaborative problem-solving
- After resolution, IMP creates a fresh plan with new understanding
