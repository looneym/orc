# ORC - Orchestrator Command Center

**Forest Factory Command Center for El Presidente's Development Ecosystem**

The ORC (Orchestrator) repository serves as the central command and coordination layer for development workflow management, providing universal commands, lightweight tech planning, and efficient worktree orchestration.

## Core Systems

### 🎯 Universal Commands
- **Global Access**: Commands accessible from any Claude Code session via symlinks
- **9 Core Commands**: `/tech-plan`, `/bootstrap`, `/janitor`, `/pr-workflow`, `/journal`, etc.
- **Central Management**: All commands maintained in `global-commands/` directory

### 📋 Tech Plans System  
- **Lightweight Planning**: 3-state system (backlog → in-progress → archive)
- **Worktree Integration**: Plans symlinked to individual investigations
- **No Ceremony**: Quick planning without complex project management overhead

### 🌳 Worktree Architecture
- **Single-Repo Focus**: Clean development environments per investigation
- **TMux Integration**: Automated workspace setup with `muxup`
- **Symlinked Plans**: Local `.tech-plans/` directories connect to central storage

## Quick Start

### Command Access
```bash
# Commands are globally accessible via symlinks
/tech-plan investigation-analysis    # Create new tech plan
/bootstrap                           # Load project context
/janitor                             # Maintenance and cleanup
```

### Tech Planning
```bash
# Plans flow through simple states
tech-plans/backlog/              # Future work
tech-plans/in-progress/          # Active investigations  
tech-plans/archive/              # Completed work
```

### Worktree Coordination
```bash
# Create new investigation
git worktree add ~/src/worktrees/ml-feature-name-repo -b ml/feature-name
cd ~/src/worktrees/ml-feature-name-repo
ln -sf /Users/looneym/src/orc/tech-plans/in-progress/ml-feature-name-repo .tech-plans
```

## Directory Structure

```
orc/
├── docs/                    # Complete ecosystem documentation
├── global-commands/         # Universal command definitions (symlinked globally)
├── tech-plans/              # Central planning system
│   ├── in-progress/         # Active worktree investigations
│   ├── backlog/            # Future work items
│   └── archive/            # Completed work
├── experimental/            # Experimental systems and prototypes
│   └── mcp-server/         # Rails-based MCP task management system
├── .claude/
│   └── commands/           # ORC-specific command definitions
├── work-trees -> ~/src/worktrees/  # Symlink to active worktrees
└── CLAUDE.md               # Central ecosystem context
```

## Documentation

Complete documentation available in `docs/`:
- **[Architecture Overview](docs/README.md)** - Complete ecosystem overview
- **[Command System](docs/command-system.md)** - Universal command management
- **[Tech Plans](docs/tech-plans-system.md)** - Lightweight planning system
- **[Worktree Architecture](docs/worktree-architecture.md)** - Single-repo worktree patterns
- **[Orchestrator Workflow](docs/orchestrator-workflow.md)** - Coordination procedures

## Experimental Systems

### MCP Task Management Server
Rails-based MCP server prototype for Claude Code integration:
- **Location**: `experimental/mcp-server/`
- **Purpose**: Native task management with worktree awareness
- **Status**: Experimental - foundations built, purpose evolving

## Key Principles

- **Lightweight**: Minimal ceremony, maximum workflow efficiency
- **Universal**: Commands accessible from any development context  
- **Coordinated**: Orchestrator manages, investigations implement
- **Preserved**: Core ORC workflow maintained and prioritized
- **Experimental**: New systems isolated until proven valuable

---

**Orchestrator Claude Coordinates. Investigation Claude Implements. El Presidente Commands.**