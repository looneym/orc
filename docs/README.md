# ORC Ecosystem Documentation

**Forest Factory Command Center - Architecture and Operations**

This documentation describes the complete ORC (Orchestrator) development ecosystem that provides centralized command management, lightweight tech planning, and efficient worktree coordination for El Presidente's development workflow.

## Overview

The ORC ecosystem solves three core problems:
1. **Command Discoverability** - Universal access to project management commands
2. **Cross-Repository Confusion** - Clean single-repo worktree architecture  
3. **Lightweight Planning** - Quick tech plan creation without ceremony

## Architecture Components

### [Worktree System](worktree-architecture.md)
Single-repository worktrees with symlinked tech plans for clean development isolation.

### [Tech Plans System](tech-plans-system.md) 
Lightweight 3-state planning: `backlog → in-progress → archive`

### [Command System](command-system.md)
Universal slash commands managed centrally in ORC, accessible globally via symlinks.

### [Orchestrator Workflow](orchestrator-workflow.md)
Complete workflow for worktree creation, TMux setup, and investigation coordination.

### [Tools Evaluation](tools-evaluation.md)
Registry and framework for evaluating potential development tools and workflow enhancements.

### [Integration Patterns](integration-patterns.md)
How commands, worktrees, and tech plans work together seamlessly.

## Quick Reference

### Directory Structure
```
orc/
├── docs/                    # This documentation
├── global-commands/         # Universal command definitions
├── tech-plans/              # Central planning system
│   ├── in-progress/         # Active worktree investigations
│   ├── backlog/            # Future work items
│   └── archive/            # Completed work
├── .claude/
│   └── commands/           # ORC-specific command definitions
├── work-trees -> ~/src/worktrees/  # Symlink to active worktrees
└── CLAUDE.md               # Central ecosystem context
```

### Key Symlinks
```
~/.claude/commands/ → orc/global-commands/      # Global command access
worktree/.tech-plans → orc/tech-plans/in-progress/[worktree]/  # Local tech plans
```

### Worktree States
```
~/src/worktrees/[active]/   # Currently working on
~/src/worktrees/paused/     # Valid but not active focus
[deleted]                   # Completed and archived
```

## Getting Started

1. **Create New Investigation**: Use orchestrator to set up single-repo worktree
2. **Plan Work**: Use `/tech-plan` command for lightweight planning
3. **Manage Progress**: Tech plans flow from backlog → in-progress → archive
4. **Pause/Resume**: Move worktrees between active and paused states
5. **Maintain System**: Use `/janitor` for cleanup and organization

## Implementation Status

- ✅ **Command System**: 8 universal commands operational
- ✅ **Symlink Architecture**: Prototype validated with ml-symlink-test-intercom
- ✅ **Tech Plans Structure**: 3-state organization complete
- ✅ **Paused Worktrees**: Directory structure and workflow established
- 🔄 **Integration Work**: Commands need updates for new architecture
- 🔄 **Migration**: Existing worktrees need conversion to new pattern

See individual documentation files for detailed implementation guidance.