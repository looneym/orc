# ORC - Orchestrator Command Center

**Forest Factory Command Center for El Presidente's Development Ecosystem**

ORC coordinates development workflow through universal commands, lightweight planning, and efficient worktree orchestration. One command system accessible everywhere, one planning approach that scales, one coordination layer that just works.

## Command System Architecture

**Central Management + Global Access**
- Commands stored in `global-commands/` (universal) and `.claude/commands/` (ORC-specific)
- Symlinked to `~/.claude/commands/` for global availability via git post-commit hook
- Single source of truth - update once, available everywhere

**Automatic Symlink Management**
```bash
# Git post-commit hook automatically maintains symlinks:
~/.claude/commands/bootstrap.md -> /Users/looneym/src/orc/global-commands/bootstrap.md
~/.claude/commands/tech-plan.md -> /Users/looneym/src/orc/global-commands/tech-plan.md
# ORC-specific commands stay local to .claude/commands/
```

## Available Commands

**Planning & Organization**
- `/tech-plan` - Create structured technical plans with lightweight templates
- `/bootstrap` - Load project context and recent work for new Claude sessions  
- `/janitor` - Local worktree maintenance and tech plan lifecycle management

**Development Workflow**
- `/worktree` - Create development environments from existing tech plans
- `/cleanup` - Intelligent worktree and TMux cleanup with safety recommendations
- `/commit` - Automatic conventional commits with intelligent staging

**Specialized Tools**  
- `/create-prompt` - Advanced prompt engineering and quality assessment
- `/rails-debug` - Generate Rails console debugging code with error handling

## Quick Examples

```bash
# Universal commands work everywhere
/tech-plan feature-name          # Create focused tech plan
/worktree                        # Select plan and create worktree
/bootstrap                       # Load project context in investigation

# Simple planning workflow
tech-plans/backlog/     → in-progress/     → archive/
   (future work)        (active projects)    (completed)
```

## How It Works

**Commands** (`global-commands/`) are symlinked globally for universal access  
**Plans** (`tech-plans/`) flow through backlog → in-progress → archive states  
**Worktrees** link to plans via symlinks for integrated development

## Documentation & Architecture

Complete technical documentation, architecture details, and implementation guides available in the `docs/` directory.

## Experimental Work

The `experimental/` directory contains prototypes and experimental systems, including an MCP task management server for potential future integration.

---

**Orchestrator Claude Coordinates. Investigation Claude Implements. El Presidente Commands.**