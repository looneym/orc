# Tech Plans System

**Lightweight 3-State Planning for Quick Development Planning**

## Overview

The tech plans system provides a lightweight approach to planning development work without ceremony or complex project management overhead. Plans flow through three simple states and are organized for both individual focus and cross-project coordination.

## Design Philosophy

### Lightweight Planning
- **Goal**: "Quickly cutting plans for things I'm working on"
- **Anti-Pattern**: Complex lifecycle tracking and project management ceremony  
- **Focus**: Individual developer workflow, not team coordination complexity

### Simple State Model
- **investigating**: Figuring out the problem/approach
- **in_progress**: Actively working  
- **paused**: Blocked or deprioritized
- **done**: Completed (moves to archive)

### Central Coordination
- **Individual Access**: Via worktree symlinks for local convenience
- **Orchestrator Visibility**: Central storage for cross-project awareness
- **Historical Context**: Archive preserves completed work for reference

## Directory Structure

```
orc/tech-plans/
├── in-progress/             # Active worktree investigations
│   ├── ml-feature-intercom/ # Tech plans for specific worktree
│   │   ├── investigation.md
│   │   └── implementation.md
│   └── ml-perfbot-analysis/ # Another worktree's plans
│       └── performance.md
├── backlog/                 # Future work items
│   ├── orc_ecosystem_refinement.md     # Strategic plans
│   ├── worktree_symlink_tech_plans.md  # Architecture work
│   └── WO-004_perfbot_system_enhancements.md  # Converted work orders
└── archive/                 # Completed work
    ├── WO-012_dlq_bot_foundations.md   # Finished work orders
    └── api_optimization_complete.md    # Completed investigations
```

## Tech Plan Template

### Basic Structure
```markdown
# [Plan Name]

**Status**: investigating | in_progress | paused | done

## Problem & Solution
**Current Issue:** [What's broken/missing/inefficient]
**Solution:** [What we're building in one sentence]

## Implementation
### Approach
[High-level solution strategy]

### Key Tasks
- [ ] Task 1
- [ ] Task 2
- [ ] Task 3

## Testing Strategy
[How we'll validate it works]

## Notes
[Implementation notes, discoveries, links]
```

### Status Lifecycle
1. **investigating**: Created with basic problem/solution, gathering information
2. **in_progress**: Implementation started, tasks being worked
3. **paused**: Work stopped (blocked, deprioritized, or waiting)  
4. **done**: Work completed → moves to archive

## Integration with Worktrees

### Local Access Pattern
```bash
# From any worktree
ls .tech-plans/                    # See all plans for this investigation
vim .tech-plans/feature-analysis.md  # Edit specific plan
```

### Symlink Architecture
```bash
# Worktree symlink points to ORC namespace
worktree/.tech-plans -> orc/tech-plans/in-progress/[worktree-name]/
```

### Creation Workflow
```bash
# From within worktree
/tech-plan feature-analysis        # Creates .tech-plans/feature-analysis.md
                                  # Actually stored in orc/tech-plans/in-progress/[worktree]/
```

## Command Integration

### `/tech-plan` Command
- **Context Detection**: Recognizes worktree vs ORC context
- **Automatic Placement**: Creates plans in correct location
- **Template Application**: Uses lightweight 4-state template

### `/bootstrap` Command  
- **Local Plans**: Reads from `.tech-plans/` symlink
- **Context Loading**: Combines tech plans with git history
- **Priority Focus**: Shows most relevant plans for current work

### `/janitor` Command
- **Lifecycle Management**: Helps transition plans between states
- **Cross-Worktree**: Manages plans across all namespaces  
- **Archive Management**: Moves completed work to archive

## Lifecycle Management

### State Transitions

#### investigating → in_progress
```bash
# Edit plan to update status and add implementation details
vim .tech-plans/plan-name.md
# Change: **Status**: investigating → in_progress
```

#### in_progress → paused
```bash
# Update status and note reason for pausing
vim .tech-plans/plan-name.md
# Change: **Status**: in_progress → paused
# Add notes about why paused and what's needed to resume
```

#### in_progress → done  
```bash
# Complete the work, then archive the plan
vim .tech-plans/plan-name.md
# Change: **Status**: in_progress → done

# Archive via janitor or manual move
mv orc/tech-plans/in-progress/[worktree]/plan-name.md \
   orc/tech-plans/archive/
```

### Worktree State Coordination

#### Active Worktree → Paused
```bash
# Move worktree to paused (tech plans stay in in-progress)
mv worktrees/ml-feature-intercom worktrees/paused/

# Tech plans remain accessible via:
# orc/tech-plans/in-progress/ml-feature-intercom/
```

#### Paused Worktree → Active
```bash
# Move worktree back to active
mv worktrees/paused/ml-feature-intercom worktrees/

# Symlinks automatically work again
cd worktrees/ml-feature-intercom
ls .tech-plans/  # Plans are back
```

## Migration from Work Orders

### Conversion Pattern
Work orders have been migrated to tech plans structure:

- **01-backlog + 02-next + 03-in-progress** → `backlog/`
- **04-complete** → `archive/`
- **Complex lifecycle fields** → Simplified 4-state approach

### Template Simplification
```markdown
# Before (Work Order)
**Created**: 2025-08-21  
**Category**: 🤖 Automation  
**Priority**: Medium  
**Effort**: L  
**IMP Assignment**: Unassigned

# After (Tech Plan)
**Status**: investigating
```

## Best Practices

### Plan Creation
- **Start Simple**: Just problem/solution initially
- **Iterate**: Add details as understanding develops
- **Stay Focused**: One investigation per plan
- **Link Context**: Reference issues, PRs, documentation

### State Management
- **Honest Status**: Update status when reality changes
- **Clear Pausing**: Note why work stopped and what's needed to resume
- **Archive Promptly**: Move completed work to keep in-progress clean

### Cross-Worktree Coordination
- **Orchestrator View**: Use ORC perspective for status across all work
- **Individual Focus**: Use worktree symlinks for local work
- **Dependencies**: Note cross-plan dependencies when they exist

## Troubleshooting

### Plans Not Appearing in Worktree
```bash
# Check symlink
ls -la .tech-plans
# Should point to: orc/tech-plans/in-progress/[worktree-name]

# Check ORC directory exists
ls orc/tech-plans/in-progress/[worktree-name]/
```

### `/tech-plan` Command Issues
- **Wrong Location**: Verify command detects worktree context correctly
- **Permission Issues**: Ensure ORC directory is writable
- **Template Problems**: Check command uses simplified 4-state template

### Cross-Worktree Visibility
```bash
# Orchestrator view of all active work
ls orc/tech-plans/in-progress/*/
# Should show all worktree namespaces and their plans
```