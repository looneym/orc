# Orchestrator Workflow

**Coordination Patterns for Single-Repository Worktree Management**

## Overview

The orchestrator operates as the command and coordination layer, managing worktree creation, status reporting, and cross-investigation awareness while maintaining clean separation from individual investigation work.

## Core Responsibilities

### Orchestrator Claude ONLY Does:
- **Worktree Setup**: Creates single-repo worktrees with symlinked tech plans
- **Status Reporting**: Cross-worktree visibility and coordination
- **TMux Management**: Window creation and workspace organization  
- **CLAUDE.md Coordination**: Updates investigation context files
- **Tech Plan Lifecycle**: Manages central storage and archiving

### Orchestrator Claude NEVER Does:
- **Code Changes**: No direct file edits within investigation worktrees
- **Investigation Work**: No debugging or technical implementation
- **Deep Context**: Stays at coordination level, not implementation details

## Single-Repository Worktree Creation

### GitHub Issue Workflow

When El Presidente provides a GitHub issue URL:

1. **Fetch Complete Context**:
   ```bash
   gh issue view <number> --repo intercom/intercom
   ```

2. **Create Descriptive Worktree**:
   ```bash
   # Fetch latest master (preserve current work)
   cd /Users/looneym/src/intercom && git fetch origin
   
   # Create single-repo worktree
   git worktree add /Users/looneym/src/worktrees/ml-descriptive-name-intercom -b ml/descriptive-name origin/master
   ```

3. **Setup Tech Plans Architecture**:
   ```bash
   cd /Users/looneym/src/worktrees/ml-descriptive-name-intercom
   mkdir -p /Users/looneym/src/orc/tech-plans/in-progress/ml-descriptive-name-intercom
   ln -sf /Users/looneym/src/orc/tech-plans/in-progress/ml-descriptive-name-intercom .tech-plans
   ```

4. **Create Investigation Context**:
   - Generate comprehensive CLAUDE.md with issue context
   - Include all links and resources from issue/comments
   - Add progress tracking structure

5. **Launch TMux Environment**:
   ```bash
   tmux new-window -n "descriptive-name" -c "/Users/looneym/src/worktrees/ml-descriptive-name-intercom" \; send-keys "muxup" Enter
   ```

### Manual Investigation Setup

For investigations not tied to GitHub issues:

```bash
# 1. Create worktree from fresh master
cd /Users/looneym/src/intercom && git fetch origin
git worktree add /Users/looneym/src/worktrees/ml-investigation-name-intercom -b ml/investigation-name origin/master

# 2. Setup tech plans symlink
cd /Users/looneym/src/worktrees/ml-investigation-name-intercom
mkdir -p /Users/looneym/src/orc/tech-plans/in-progress/ml-investigation-name-intercom
ln -sf /Users/looneym/src/orc/tech-plans/in-progress/ml-investigation-name-intercom .tech-plans

# 3. Launch development environment
tmux new-window -n "investigation-name" -c "/Users/looneym/src/worktrees/ml-investigation-name-intercom" \; send-keys "muxup" Enter
```

## TMux Development Environment

### Muxup Integration

The `muxup` command creates standardized development layout:

```
┌─────────────┬─────────────┐
│             │             │
│    vim      │   claude    │
│ CLAUDE.md   │             │
│ +NERDTree   ├─────────────┤
│             │             │
│             │    shell    │
└─────────────┴─────────────┘
```

### Automated Setup
```bash
tmux new-window -n "[investigation-name]" -c "/path/to/worktree" \; send-keys "muxup" Enter
```

This automatically:
- **Left Pane**: Vim with CLAUDE.md open + NERDTree sidebar
- **Top Right**: Claude Code session with worktree context
- **Bottom Right**: Shell in worktree directory

## Worktree State Management

### Active Worktrees
- **Location**: `~/src/worktrees/[worktree-name]/`
- **Tech Plans**: Symlinked to `orc/tech-plans/in-progress/[worktree-name]/`
- **TMux Access**: Direct window creation via orchestrator

### Paused Worktrees  
- **Location**: `~/src/worktrees/paused/[worktree-name]/`
- **Tech Plans**: Remain in `orc/tech-plans/in-progress/[worktree-name]/`
- **Access**: Move back to active for resume

### Archived Worktrees
- **Worktree**: Deleted via `git worktree remove`
- **Tech Plans**: Moved to `orc/tech-plans/archive/`
- **History**: Preserved in archive for reference

## Status Reporting

### Worktree Status Query

When El Presidente asks: *"What's the status of [worktree-name]?"*

1. **Read CLAUDE.md Progress**: Check investigation context
2. **Git Status**: `git status` for uncommitted changes
3. **Recent Activity**: `git log --oneline -5` for latest work
4. **Diff Summary**: `git diff --stat` for change overview

**Report Format**:
```markdown
## Worktree Status: ml-investigation-name

**Progress from Investigation**:
- [Summary from CLAUDE.md Progress section]

**Git Status**:
- Clean working directory / 3 files modified, ready to commit
- 2 commits ahead of origin/master

**Recent Activity**:
- [Recent commits with timestamps]

**Tech Plans Status**:
- [Active plans and their states]
```

### Cross-Worktree Overview

For global status queries:

```bash
# List all active investigations
ls ~/src/worktrees/

# List paused work
ls ~/src/worktrees/paused/

# See all in-progress tech plans
ls /Users/looneym/src/orc/tech-plans/in-progress/
```

## Investigation Context Template

### Single-Repository CLAUDE.md

```markdown
# Investigation: [Descriptive Name Based on Problem]

## Environment Setup
You are working on focused investigation in the Intercom repository:

- **Repository**: Single intercom checkout on branch `ml/descriptive-name`
- **Tech Plans**: Available via `.tech-plans/` symlink 
- **Context**: Focused single-repository investigation

## Your Mission
[Full problem description from issue/context]

**Source**: [GitHub Issue URL / Investigation Request]

### Problem Summary
[Clear problem statement with full context]

## Available Resources
- **Issue/Context**: [All relevant links and resources]
- **Tech Plans**: Use `/tech-plan` to create focused planning documents
- **Bootstrap**: Use `/bootstrap` for context loading

## Status Update Protocol
**CRITICAL**: Update progress sections as work evolves.

### Current Status
🔄 **Investigation Started** - Initial setup complete

**Key Actions Completed**:
- [Timestamp] Worktree created and environment setup
- [Timestamp] [Next action]

**Next Steps**: [What's needed to progress]

## Progress Log
[Timestamped entries of findings and decisions]
```

## Command Integration

### Context-Aware Commands

Commands behave differently in orchestrator vs worktree context:

#### In ORC Context (Orchestrator)
- **`/tech-plan`**: Creates in `tech-plans/backlog/` for strategic planning
- **`/bootstrap`**: Provides ORC ecosystem overview
- **`/janitor`**: Manages cross-worktree cleanup and lifecycle

#### In Worktree Context (Investigation)  
- **`/tech-plan`**: Creates in `.tech-plans/` (symlinked to ORC)
- **`/bootstrap`**: Loads investigation context from tech plans + git
- **`/janitor`**: Local worktree maintenance and cleanup

## Safety Boundaries

### Orchestrator Scope Limits

**If El Presidente asks for direct technical work:**

❌ **Invalid Requests**:
- "Fix this bug in the worker code"
- "Update the configuration file" 
- "Debug this database query"

✅ **Correct Response**:
> "El Presidente, I'm the Orchestrator - I coordinate investigations but don't work directly on code. Switch to the `[investigation-name]` TMux window to work with the investigation-claude on that technical task."

### Investigation Handoff

When creating worktrees, orchestrator should:
1. **Setup Environment**: Create worktree + tech plans + TMux
2. **Provide Context**: Comprehensive CLAUDE.md with all relevant info  
3. **Clean Handoff**: "Environment ready - switch to `[investigation-name]` window to begin work"

## Troubleshooting

### Worktree Creation Issues
```bash
# Verify repo access
cd /Users/looneym/src/intercom && git status

# Check for stale worktrees  
git worktree list
git worktree prune

# Force remove problematic worktree
git worktree remove --force /path/to/worktree
```

### Tech Plans Symlink Issues
```bash
# Check symlink target
ls -la worktree/.tech-plans

# Recreate if broken
rm worktree/.tech-plans
ln -sf /Users/looneym/src/orc/tech-plans/in-progress/[worktree-name] worktree/.tech-plans
```

### TMux Environment Problems
```bash
# List active windows
tmux list-windows

# Kill problematic window
tmux kill-window -t [window-name]

# Recreate development environment
cd /path/to/worktree
tmux new-window -n "[name]" \; send-keys "muxup" Enter
```

## Best Practices

### Naming Conventions
- **Worktrees**: `ml-descriptive-problem-repo` (e.g., `ml-dlq-performance-intercom`)
- **Branches**: `ml/descriptive-problem` (e.g., `ml/dlq-performance`)
- **TMux Windows**: Short descriptive names (e.g., `dlq-perf`)

### Repository Selection
- **Primary Focus**: Choose the main repository where most work will happen
- **Single Repo**: Avoid multi-repo worktrees in new architecture
- **Common Patterns**:
  - `ml-feature-name-intercom`: Application development
  - `ml-config-name-infrastructure`: Infrastructure changes
  - `ml-analysis-name-intercom`: Performance or debugging investigations

### Lifecycle Management
- **Clean Creation**: Always create from fresh `origin/master`
- **Active Maintenance**: Regular status reports and progress updates
- **Proper Archiving**: Move completed work to archive with context preservation
- **Workspace Focus**: Use paused directory for valid but inactive work