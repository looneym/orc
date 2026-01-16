# 🏭 ORC 2.0 - Forest Factory Orchestration System

**Forest Factory Command Center for El Presidente's Development Ecosystem**

![Forest Factory](assets/orc.png)

## 🎯 What It Does

ORC 2.0 coordinates multiple Claude agents working across repositories simultaneously. 

## 🗂️ Memory Architecture

- **SQLite Ledger** - Single source of truth for all operational data: missions, epics, tasks, groves, and handoffs
- **Handoff System** - Narrative-based context transfer between sessions for seamless continuity

## 🌲 Key Concepts

- **🎭 ORCs (Orchestrators)** - Coordinate missions, manage work orders, and facilitate cross-grove communication
- **⚒️ IMPs (Implementers)** - Work in groves on actual code, completing work orders and reporting discoveries back
- **🌳 Groves** - Git worktrees where IMPs do their work; one mission can have multiple groves from different repositories
- **📋 Handoffs** - Context preservation across sessions via narrative summaries

## 🤖 Systematic + Intelligent

Claude agents work autonomously on separate cross-repo tasks while sharing context through the handoff system. The system combines systematic execution (structured work tracking, TMux coordination) with context preservation (handoff narratives). Currently powering all development work with full context preservation across sessions.

---

## Quick Start

```bash
# Initialize ORC context
orc prime

# View current mission status
orc status

# List work orders
orc work-order list

# Create a new mission
orc mission create "Mission Title" -d "Description"

# Create work orders
orc work-order create "Task title" --mission MISSION-001

# View mission summary
orc summary
```

## Command System Architecture

**Central Management + Global Access**
- Commands stored in `global-commands/` (universal) and `.claude/commands/` (ORC-specific)
- Symlinked to `~/.claude/commands/` for global availability
- Single source of truth - update once, available everywhere

**Key Commands**
- `/handoff` - Create handoff for session continuity
- `orc prime` - Lightweight context injection (agents auto-run this on startup via direct prompt)
- `orc status --handoff` - View latest handoff

---

**Orchestrator Claude Coordinates. Investigation Claude Implements. El Presidente Commands.**
