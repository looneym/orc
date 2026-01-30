---
name: imp-plan-submit
description: Submit plan for review. Uses Haiku subagent for self-review, auto-approves or prompts escalation.
---

# IMP Plan Submit

Submit the current plan for review. Performs self-review using a Haiku subagent.

## Flow

1. **Get draft plan for current task**
   ```bash
   orc plan list --task TASK-xxx --status draft
   orc plan show PLAN-xxx
   ```

2. **Submit plan**
   ```bash
   orc plan submit PLAN-xxx
   ```

3. **Launch Haiku review subagent**
   Use the Task tool with:
   - `subagent_type: "general-purpose"`
   - `model: "haiku"`
   - Prompt (see below)

   **Review Prompt:**
   ```
   You are reviewing an IMP's implementation plan.

   ## Task Description
   [paste task description from orc task show]

   ## Plan Content
   [paste plan content from orc plan show]

   ## Review Criteria (STRICT)
   - SCOPE MATCH: Plan addresses all task requirements, nothing extra
   - BOUNDED: No scope creep, no "while we're here" additions
   - VERIFICATION CONCRETE: Specific test/lint commands included
   - CHANGES CLEAR: File paths and modifications are explicit

   Respond with EXACTLY one of:
   - APPROVE
   - ESCALATE: [specific concern in one sentence]
   ```

4. **Handle review result**

   **If APPROVE:**
   ```bash
   orc plan approve PLAN-xxx
   ```
   Output: "Plan PLAN-xxx approved. Implement the plan, then run /imp-rec when complete."

   **If ESCALATE:**
   Output: "Self-review concern: [reason]. Run `/imp-escalate --reason '[reason]'` to escalate for human review."

## Notes

- Haiku is used for cost efficiency on simple approve/reject decisions
- Escalation is not failure - it's the correct response to uncertainty
- The review is deliberately strict to catch scope creep early
