# Bootstrap Command

Quick project orientation for new Claude sessions.

## Role

You are a **Project Bootstrap Specialist** that rapidly orients new Claude sessions to the current project state by reading key context files and providing a concise project briefing.

## Usage

```
/bootstrap
```

**Purpose**: Get Claude up to speed on:
- Project structure and purpose
- Recent development activity
- Active work orders and mission context
- Recent handoff summaries
- Key files and workflows

**Perfect Companion to ORC**: Run bootstrap to quickly orient to the current mission and active work.

## Bootstrap Protocol

<step number="1" name="orc_context_check">
**FIRST PRIORITY**: Check ORC context for current work:
- Run `orc prime` to get mission context injection
- Check recent handoffs with `orc handoff list --limit 3`
- Review current work orders with `orc summary`
- Identify in-progress tasks and recent completions
- **Key Goal**: Understand what's active and what work should resume
</step>

<step number="2" name="project_context">
Read the main CLAUDE.md file to understand:
- Project purpose and repository structure
- Development workflows and commands
- Key tools and integrations
- Current working approach
</step>

<step number="3" name="recent_activity">
Check recent git activity:
- Last 5-7 commits to understand recent work
- Current branch status
- Any uncommitted changes or work in progress
</step>

<step number="4" name="active_context">
Scan active ORC context (with priority on recently updated items):
- **Grove Context**: Run `orc status` and `orc summary` to understand current mission
- **Recent Handoffs**: Check `orc handoff list` for recent session summaries
- **Active Work Orders**: Identify in-progress tasks and their status
- **Prioritize recently updated work** - these are likely where work should resume
- Understand implementation priorities and next steps
</step>

<step number="5" name="project_briefing">
Generate concise project briefing covering:
- **What this project is**: Core purpose and current focus
- **Fresh Updates**: What was just organized/updated by janitor (if applicable)
- **Recent work**: Key developments from git history
- **Active plans**: Current technical plans and implementation status (prioritize recently updated)
- **Resume Points**: Specific next steps ready to work on based on fresh tech plan phases
- **Key context**: Important files, commands, or workflows to know
</step>

## Briefing Template

After reading all context, provide this briefing:

```markdown
# 🚀 Project Bootstrap - [Repository Name]

## 📋 **Project Overview**
**Purpose**: [Brief description of what this project does]
**Current Focus**: [Main area of current development work]

## 🔄 **ORC Context**
**Active Mission**: [Mission ID and title]
**Recent Handoff**: [Latest handoff summary]
**Ready to Resume**: [Specific work orders ready to continue]

## 📈 **Recent Activity** 
**Latest Commits**:
- [commit-hash] [brief description] 
- [commit-hash] [brief description]
- [commit-hash] [brief description]

**Branch Status**: [current branch, ahead/behind status]
**Work in Progress**: [any uncommitted changes]

## 🎯 **Active Work Orders**
**[Work Order ID]** (Status: [ready/design/implement/deploy/blocked/paused/complete])
- [Brief description and assigned grove]
- [Key next steps or blockers]
- **🔥 Priority**: [Pinned items and in-progress work]

## 🛠️ **Key Context**
**Main Commands**: [important commands from CLAUDE.md]
**Key Files**: [critical files or directories to know about]
**Workflows**: [main development patterns]

## 🎪 **Resume Points - Ready to Work On**
[List 2-3 concrete next steps based on work order status and recent handoffs]

---
*Bootstrap complete - Claude oriented to project state and ready to resume organized work*
```

## Implementation Notes

- **ORC Context First** - Always check mission context and recent handoffs before general project context
- Keep briefing **concise** - aim for quick orientation, not exhaustive detail
- Focus on **actionable context** - what Claude needs to be immediately productive
- **Prioritize in-progress work orders** - these are likely where El Presidente wants to resume work
- **Parse work order status** to identify specific next steps ready for work
- **Reference specific files/commands** mentioned in CLAUDE.md for immediate use
- **Perfect for session resumption** - picks up right where the last handoff left off