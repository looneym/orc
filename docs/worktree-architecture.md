# Worktree Architecture

**Single-Repository Worktrees with Symlinked Tech Plans**

## Overview

The worktree system provides clean, focused development environments by using single-repository worktrees instead of multi-repository containers. Each worktree maintains its own context while connecting to centrally managed tech plans via symlinks.

## Core Principles

### Single-Repository Focus
- **Before**: Multi-repo worktrees (ml-dlq-bot/intercom/, ml-dlq-bot/infrastructure/, etc.)
- **After**: Single-repo worktrees (ml-dlq-investigation-intercom/, ml-perfbot-intercom/, etc.)
- **Benefit**: Eliminates cross-repository navigation confusion

### Symlinked Tech Plans
- **Local Access**: `.tech-plans/` directory in repo root
- **Central Storage**: `orc/tech-plans/in-progress/[worktree-name]/`
- **Coordination**: Orchestrator sees all worktree plans via ORC

### State Management
- **Active**: `~/src/worktrees/[worktree-name]/`
- **Paused**: `~/src/worktrees/paused/[worktree-name]/`
- **Archived**: Worktree deleted, tech plans moved to archive

## Directory Structure

### Active Worktree Layout
```
~/src/worktrees/ml-feature-intercom/
├── .tech-plans -> /Users/looneym/src/orc/tech-plans/in-progress/ml-feature-intercom/
├── .claude/                    # Standard Claude configuration
├── CLAUDE.md                   # Worktree-specific context
├── .git                       # Git worktree pointing to feature branch
└── [intercom repo files]      # Full intercom repository checkout
```

### Central Tech Plans Storage
```
orc/tech-plans/in-progress/
├── ml-feature-intercom/        # Tech plans for specific worktree
│   ├── investigation_plan.md
│   └── implementation_notes.md
├── ml-perfbot-improvements/    # Another worktree's plans
│   └── enhancement_strategy.md
└── ml-api-optimization/        # Third worktree's plans
    └── performance_analysis.md
```

## Symlink Architecture

### Creation Pattern
```bash
# In worktree root
ln -sf /Users/looneym/src/orc/tech-plans/in-progress/[worktree-name] .tech-plans
```

### Git Integration
- **Global gitignore**: `.tech-plans` automatically ignored
- **Clean status**: No tracking issues in any repository
- **No pollution**: Symlinks don't affect repository state

### Benefits
- **Local Convenience**: `ls .tech-plans/` from anywhere in worktree
- **Central Coordination**: All plans visible to orchestrator via ORC
- **Clean Separation**: No cross-contamination between worktrees
- **Version Control**: Tech plans tracked in ORC, code tracked in respective repos

## Worktree Lifecycle

### 1. Creation
```bash
# From appropriate repository
git worktree add ~/src/worktrees/ml-feature-intercom -b ml/feature-name origin/master

# Setup symlink
cd ~/src/worktrees/ml-feature-intercom
mkdir -p /Users/looneym/src/orc/tech-plans/in-progress/ml-feature-intercom
ln -sf /Users/looneym/src/orc/tech-plans/in-progress/ml-feature-intercom .tech-plans
```

### 2. Development
```bash
# Work in the worktree
cd ~/src/worktrees/ml-feature-intercom

# Create/edit tech plans
/tech-plan feature-analysis    # Creates plan in .tech-plans/
vim .tech-plans/feature-analysis.md

# Normal development workflow
git add .
git commit -m "Implement feature"
```

### 3. State Transitions

#### Active → Paused
```bash
cd ~/src/worktrees
mv ml-feature-intercom paused/
# Tech plans remain accessible via ORC
```

#### Paused → Active
```bash
cd ~/src/worktrees
mv paused/ml-feature-intercom .
```

#### Active → Archived
```bash
# Move tech plans to archive
mv /Users/looneym/src/orc/tech-plans/in-progress/ml-feature-intercom \
   /Users/looneym/src/orc/tech-plans/archive/

# Remove worktree
git worktree remove ~/src/worktrees/ml-feature-intercom
```

## TMux Integration

### Automated Setup
```bash
# Create worktree with TMux window
tmux new-window -n "feature-name" -c "~/src/worktrees/ml-feature-intercom" \; send-keys "muxup" Enter
```

### Muxup Layout
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

## Context Management

### Worktree-Specific Context
- **CLAUDE.md**: Investigation-specific context in repo root
- **Git History**: Feature branch with focused commits
- **Tech Plans**: Accessible via `.tech-plans/` symlink

### Global Context Access
- **Commands**: All universal commands via `~/.claude/commands/`
- **Bootstrap**: `/bootstrap` loads context from tech plans + git history
- **Coordination**: Orchestrator accesses via `orc/tech-plans/in-progress/`

## Migration Strategy

### Converting Existing Worktrees
1. **Backup Current State**: Ensure no uncommitted work
2. **Create Tech Plans Directory**: `mkdir -p orc/tech-plans/in-progress/[worktree-name]`
3. **Migrate Existing Plans**: Move any existing tech plans to ORC location
4. **Create Symlink**: Add `.tech-plans` symlink in worktree root
5. **Test Access**: Verify tech plans accessible via symlink
6. **Update Context**: Modify CLAUDE.md for single-repo focus

### Command Updates Required
- **`/tech-plan`**: Context-aware creation in correct ORC location
- **`/bootstrap`**: Read from symlinked tech plans directory
- **`/janitor`**: Manage tech plans across worktree namespaces
- **Orchestrator**: Navigate via `orc/tech-plans/in-progress/` structure

## Troubleshooting

### Symlink Issues
```bash
# Check symlink target
ls -la .tech-plans
# Should show: .tech-plans -> /Users/looneym/src/orc/tech-plans/in-progress/[worktree-name]

# Recreate if broken
rm .tech-plans
ln -sf /Users/looneym/src/orc/tech-plans/in-progress/[worktree-name] .tech-plans
```

### Git Status Issues
```bash
# Should be clean, if not check global gitignore
git status
# If .tech-plans appears, verify ~/.gitignore.global contains:
# .tech-plans
```

### Tech Plans Not Appearing
```bash
# Check ORC directory exists
ls /Users/looneym/src/orc/tech-plans/in-progress/[worktree-name]/
# Create if missing:
mkdir -p /Users/looneym/src/orc/tech-plans/in-progress/[worktree-name]
```