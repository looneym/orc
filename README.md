# 🌲 ORC - Forest Orchestration Command Center

## Overview

**ORC (Orchestrator Claude)** is the command center for managing a distributed forest of implementation agents (IMPs) working in isolated development environments (worktrees). This system enables sophisticated multi-repository development with clear coordination and tracking.

## Forest Architecture

```
🏛️ EL PRESIDENTE (Supreme Commander)
    │
    └── 🧙‍♂️ ORC (Orchestrator)
            │
            ├── 👹 IMP-ZSH (ZSH Prompt Grove)
            ├── 👹 IMP-PERFBOT (PerfBot Grove) 
            ├── 👹 IMP-ZEROCODE (ZeroCode Grove)
            └── 👹 [Future IMPs...]
```

### Components

- **🏛️ El Presidente**: Strategic decision maker and supreme forest commander
- **🧙‍♂️ ORC**: Forest keeper who coordinates groves and manages IMP workforce  
- **👹 IMPs**: Specialized woodland workers in isolated worktree groves
- **🌲 Groves**: Individual worktree environments with complete context isolation

## Repository Structure

```
orc/
├── README.md                    # This file - forest overview
├── CLAUDE.md                    # ORC coordination guidelines and worktree templates
├── STATUS.md                    # Current forest status and active groves
├── BACKLOG.md                   # Future work items extracted from IMP suggestions
├── METHODOLOGY_ANALYSIS.md      # Project management patterns and approaches
├── work-orders/                 # Individual work order specifications (future)
└── archive/                     # Completed work order archive (future)
```

## Key Concepts

### Work Orders
Structured task definitions that move through defined states:
1. **📝 BACKLOG**: Ideas awaiting evaluation  
2. **⚡ READY**: Evaluated and prioritized
3. **👹 ASSIGNED**: Assigned to specific IMP
4. **🔨 IN PROGRESS**: IMP actively working
5. **🔍 REVIEW**: Complete, awaiting validation
6. **✅ COMPLETE**: Delivered and accepted

### Forest Work Order Categories
- **🧪 Investigation**: Open-ended research and exploration
- **⚙️ Feature**: Structured development with deliverables
- **🔧 Enhancement**: Improvements to existing systems  
- **🚨 Fix**: Problem resolution and debugging
- **🛠️ Tooling**: Development utilities and automation

### IMP Specialization
Each IMP operates in an isolated grove (worktree) with:
- Complete repository context for their investigation
- Dedicated tmux development environment  
- CLAUDE.md file with work order specifications
- Status update protocol for progress tracking

## Current Forest Status

The ORC manages multiple concurrent investigations:
- **Active Groves**: Development work in progress
- **Awaiting Review**: Completed work pending validation
- **Backlog Items**: Future enhancements identified by IMPs

## Project Management Pattern

This system implements a **Manufacturing + Kanban Hybrid**:
- **Manufacturing**: Work orders, quality gates, foreman-worker hierarchy
- **Kanban**: Visual workflow, pull system, continuous flow
- **Distributed Cognition**: Specialized processing with centralized coordination

## Integration Points

### Git Worktrees
- Each grove uses git worktrees for repository isolation
- Branches follow `ml/grove-name` pattern
- Multiple repositories can be checked out per investigation

### TMux Integration  
- Automated window setup with `muxup` command
- Forest theme support (pine green for ORC environments)
- Each grove gets dedicated tmux window for development

### Status Tracking
IMPs follow mandatory status update triggers:
1. Investigation complete → Create GitHub issue
2. Implementation started → Update status  
3. PRs created → Include PR links
4. Work complete → Mark as complete
5. Blockers encountered → Document and escalate

## Future Development

The ORC system is designed to evolve:
- **Work Order Management**: Formal work order creation and tracking
- **IMP Automation**: Shell utilities for worktree management
- **Forest Analytics**: Completion metrics and productivity insights
- **Integration Expansion**: Additional tool integrations and workflows

---

**🌲 Welcome to the Forest - Where IMPs toil and the ORC coordinates the grand symphony of distributed development! 🌲**