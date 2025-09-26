# Command System

**Universal Slash Commands with Central Management and Global Access**

## Overview

The command system provides universal access to project management and development workflow commands through Claude Code's slash command system. Commands are centrally managed in the ORC repository but accessible globally via symlinks.

## Architecture

### Central Management
- **Universal Commands**: `/Users/looneym/src/orc/global-commands/`
- **ORC-Specific Commands**: `/Users/looneym/src/orc/.claude/commands/`
- **Version Control**: All commands tracked in ORC repository
- **Single Source of Truth**: One place to update, changes propagate everywhere

### Global Access
- **Symlink Directory**: `~/.claude/commands/`
- **Universal Availability**: Commands accessible from any Claude session
- **Automatic Discovery**: Claude Code finds commands via standard search paths

### Symlink Pattern
```bash
# Universal commands accessible everywhere
~/.claude/commands/command-name.md -> /Users/looneym/src/orc/global-commands/command-name.md

# ORC-specific commands available only in ORC context  
orc/.claude/commands/orc-command.md
```

## Available Commands

### Core Project Management
- **`/analyze-prompt`**: Advanced prompt quality assessment using latest Anthropic practices
- **`/bootstrap`**: Quick project orientation for new Claude sessions  
- **`/janitor`**: Complete project maintenance (CLAUDE.md validation, tech plan lifecycle, cleanup)
- **`/tech-plan`**: Structured technical planning with lightweight templates

### Development Workflows  
- **`/coda-nav`**: Navigate and extract data from team standup board
- **`/journal`**: Create and publish engineering session summaries as GitHub gists
- **`/pr-workflow`**: Complete git workflow management (branch, commit, PR, review)
- **`/rails-debug`**: Safe Rails console debugging without shell syntax errors

## Command Structure

### Standard Format
```markdown
# Command Title

**Brief description of command purpose**

**Just run `/command` for guided workflow** - summary of what it automates

## Role
You are a **[Specialist Type]** - expert in [domain]. Your expertise includes:
- **[Skill 1]** - Description
- **[Skill 2]** - Description  
- **[Skill 3]** - Description

Your mission is to [clear objective].

## Usage
```
/command [optional-args]
```

**Default Behavior** (no arguments): **[Primary action]**
- [What it does by default]

**Options**: [If command takes arguments]

## Protocol
**When called, execute ALL steps below for [comprehensive action].**

### Phase 1: [Action Category]
<step number="1" name="step_identifier">
**[Step description]:**
- [Action 1]
- [Action 2]
</step>

[Additional phases...]

## Completion Summary
After executing command:

```markdown
## 🎯 [Command Name] Complete

### [Summary sections with specific results]
```
```

### Context Awareness

Commands are designed to work in different contexts:

#### Orchestrator Context (ORC repository)
- **Global perspective**: Commands see all worktrees, strategic plans
- **Coordination focus**: Cross-project status, planning, maintenance
- **Creation scope**: Global tech plans, strategic documentation

#### Worktree Context (Individual repositories)  
- **Local focus**: Commands see worktree-specific context
- **Implementation focus**: Local tech plans, git history, specific work
- **Creation scope**: Worktree-specific tech plans and context

### Integration with New Architecture

#### `/tech-plan` Command Updates Needed
```bash
# Current: Creates in .claude/tech_plans/ directory
# Needed: Context-aware creation

# In ORC context:
# → Create in orc/tech-plans/backlog/

# In worktree context:  
# → Create in .tech-plans/ (symlinked to orc/tech-plans/in-progress/[worktree]/)
```

#### `/bootstrap` Command Updates Needed  
```bash
# Current: Reads from .claude/tech_plans/
# Needed: Symlink-aware reading

# In worktree context:
# → Read from .tech-plans/ symlink
# → Include recent git activity in single repo
# → Show worktree-specific context only
```

#### `/janitor` Command Updates Needed
```bash
# Current: Works with .claude/tech_plans/ structure  
# Needed: Multi-namespace management

# Cross-worktree capabilities:
# → Manage tech plans across orc/tech-plans/in-progress/*/
# → Handle worktree state transitions (active/paused/archived)
# → Archive completed plans from multiple worktrees
```

## Command Development

### Creating New Commands

1. **Define Role and Expertise**: Clear specialist identity
2. **Protocol Structure**: Step-by-step execution with numbered phases
3. **Context Awareness**: Handle both orchestrator and worktree contexts
4. **Completion Summary**: Structured output showing what was accomplished
5. **Integration**: Work seamlessly with symlink architecture

### Command File Management

```bash
# Create new universal command
vim /Users/looneym/src/orc/global-commands/new-command.md

# Create global symlink
cd ~/.claude/commands
ln -sf /Users/looneym/src/orc/global-commands/new-command.md .

# Create ORC-specific command (no symlink needed)
vim /Users/looneym/src/orc/.claude/commands/orc-specific.md

# Test accessibility
/new-command        # Works everywhere
/orc-specific      # Works only in ORC context
```

### Version Control
- **All changes tracked**: Commands are part of ORC repository
- **Change propagation**: Updates to master files immediately available globally
- **Rollback capability**: Git history provides command version control

## Integration Patterns

### With Worktree System
- **Context Detection**: Commands determine if running in worktree vs ORC
- **Symlink Awareness**: Commands work with `.tech-plans/` symlinked directories
- **State Management**: Commands handle worktree lifecycle (active/paused/archived)

### With Tech Plans System  
- **Lifecycle Integration**: Commands manage tech plan state transitions
- **Namespace Awareness**: Commands work across worktree namespaces
- **Archive Management**: Commands handle completed work archiving

### With Development Workflow
- **Git Integration**: Commands work with single-repo worktree model
- **TMux Integration**: Commands integrate with `muxup` automated setup
- **Session Handoff**: Commands facilitate context transfer between sessions

## Troubleshooting

### Command Not Found
```bash
# Check global symlinks exist
ls -la ~/.claude/commands/

# Recreate missing universal command symlink
cd ~/.claude/commands
ln -sf /Users/looneym/src/orc/global-commands/[command].md .
```

### Command Execution Issues
```bash
# Check universal command exists and is readable
cat /Users/looneym/src/orc/global-commands/[command].md

# Check ORC-specific command
cat /Users/looneym/src/orc/.claude/commands/[command].md

# Verify Claude can access the file
/analyze-prompt [command-file]
```

### Context Detection Problems
- **Worktree Detection**: Commands should check for `.tech-plans` symlink presence
- **ORC Detection**: Commands should check for `tech-plans/` directory structure  
- **Path Resolution**: Commands should use absolute paths for reliability

## Migration Tasks

### High Priority Updates
1. **`/tech-plan`**: Context-aware creation (worktree vs global)
2. **`/bootstrap`**: Symlink-aware reading and context loading
3. **`/janitor`**: Cross-worktree namespace management

### Medium Priority Updates
1. **`/coda-nav`**: Integration with worktree planning workflow
2. **`/journal`**: Include worktree context in session documentation
3. **`/pr-workflow`**: Single-repo worktree workflow optimization

### Testing Strategy
1. **Dual Context Testing**: Verify commands work in both ORC and worktree contexts
2. **Symlink Integration**: Ensure commands handle symlinked directories correctly
3. **Cross-Worktree Operations**: Test orchestrator commands across multiple worktrees
4. **Error Handling**: Verify graceful degradation when expected structures missing

## Future Enhancements

### Command Discovery
- **Auto-completion**: Shell integration for command discovery
- **Help System**: `/help [command]` for detailed usage guidance
- **Command Categories**: Organization by workflow area

### Advanced Integration
- **Workflow Orchestration**: Commands that coordinate multiple other commands
- **State Persistence**: Commands that maintain state across sessions
- **Cross-System Integration**: Commands that integrate with external tools (GitHub, Slack, etc.)