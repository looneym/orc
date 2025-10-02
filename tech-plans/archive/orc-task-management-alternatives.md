# ORC Task Management Alternatives Research

**Status**: in_progress

## Problem & Solution
**Current Issue:** TaskMaster rejected due to lack of git worktree support - need task management/PM system that works with Claude Code CLI and supports our parallel development workflow
**Solution:** Research and evaluate alternative task management solutions that integrate with Claude Code CLI and support git worktree-based parallel development

## Research Context
Post-TaskMaster evaluation (Sept 2025) - comprehensive investigation revealed TaskMaster's fundamental incompatibility with worktree-based parallel development. Need to explore alternative solutions that can provide:

- Task management with AI/Claude Code CLI integration
- Git worktree support or compatibility
- Fluid workflows without ceremony
- Excellent context management across parallel development streams
- Individual developer focus (not team collaboration overhead)

## Research Phase 1: Solution Discovery

### Search Categories

#### 1. Anthropic Ecosystem
- Official Anthropic tools and extensions
- Claude Code CLI plugins/extensions
- Anthropic-recommended development workflows
- Community tools built specifically for Claude integration

#### 2. GitHub Ecosystem  
- Popular task management CLIs with git integration
- Tools with explicit worktree support
- AI-enhanced development workflow tools
- Context-aware project management solutions

#### 3. Developer Blogs & Communities
- AI-driven development workflow articles
- Git worktree + task management combinations
- Individual developer productivity systems
- Context switching minimization tools

### Key Requirements Matrix
| Feature | Priority | Notes |
|---------|----------|--------|
| Claude Code CLI Integration | CRITICAL | Must work seamlessly with Claude sessions |
| Git Worktree Support | CRITICAL | Essential for parallel development isolation |
| Context Management | HIGH | Track context across multiple worktrees |
| Fluid Workflows | HIGH | Minimal ceremony, focus on doing work |
| Individual Focus | MEDIUM | Optimize for single developer, not teams |
| Local Storage | MEDIUM | Avoid cloud dependencies for speed |

## Implementation Plan

### ✅ Phase 1: Wide Discovery Research (COMPLETE)
- ✅ Survey Anthropic ecosystem for Claude Code integrations
- ✅ Search GitHub for worktree + task management combinations  
- ✅ Review developer productivity blogs from last 12 months
- ✅ Identify candidate tools and patterns

### Phase 2: ORC Enhancement Design (NEXT)
- Evaluate @johnlindquist/worktree CLI integration potential
- Design Linear CLI + ORC integration patterns  
- Plan enhanced TMux automation based on community workflows
- Design automated worktree lifecycle commands for ORC ecosystem

### Phase 3: Proof of Concept Implementation
- Build prototype integrations for top enhancement opportunities
- Test with actual ORC worktree workflows
- Compare enhanced ORC vs current system benefits
- Document integration patterns and implementation approach

## Research Sources

### Anthropic Official Channels
- [ ] Claude Code CLI documentation and extensions
- [ ] Anthropic developer community discussions
- [ ] Official Anthropic blog posts about development workflows
- [ ] GitHub repos under Anthropic organization

### GitHub Search Targets
- [ ] "claude code" + "task management" repositories
- [ ] "git worktree" + "task" + "cli" search combinations
- [ ] Popular CLI task managers with recent activity
- [ ] AI-enhanced development workflow tools

### Blog & Community Research
- [ ] Hacker News discussions about AI development workflows
- [ ] Developer Twitter/X discussions about Claude Code workflows  
- [ ] Medium/Dev.to articles about git worktree productivity
- [ ] Reddit r/programming AI tool discussions

## Testing Strategy
Each candidate tool will be evaluated against:
1. **Worktree Integration**: Can it work across multiple isolated worktrees?
2. **Claude Code Compatibility**: Does it enhance or conflict with Claude sessions?
3. **Context Preservation**: How well does it maintain context across parallel work?
4. **Workflow Fluidity**: Does it reduce friction or add ceremony?
5. **ORC Ecosystem Fit**: How would it integrate with existing commands and patterns?

## Research Findings (Phase 1 - Discovery)

### 1. Anthropic Ecosystem Analysis
**Claude Code CLI Official Direction**: 
- Feature request #4963 for native parallel task orchestration with `/fork`, `/tasks`, `/tasks merge` commands
- Shows Anthropic recognizes the need for worktree-based parallel development
- Currently in proposal stage, no implementation timeline

**Existing Claude Code Extensions**:
- **Claude Code PM** (Ran Aroussi): Comprehensive project management workflow with specialized agents
- **Project Workflow System** (harperreed): Commands for task management, code review, deployment
- **Scopecraft**: Comprehensive SDLC commands
- **Steadystart**: Structured project bootstrapping and meta-command creation
- **Simone** (Helmi): Broader project management workflow encompassing documents and processes

**Assessment**: Strong Claude Code integration ecosystem exists, but no worktree-native solutions found.

### 2. GitHub Ecosystem - Worktree-Aware Tools
**@johnlindquist/worktree CLI**:
- ✅ Modern worktree management with GitHub integration
- ✅ PR checkout directly into worktrees (`wt new pr/123`)  
- ✅ Automated dependency installation per worktree
- ✅ Multi-editor support (cursor, code, etc.)
- ✅ Designed for AI agent + parallel development workflows

**Linear CLI Tools** (Community):
- **schpet/linear-cli**: Git-aware, manages branches, writes PR details
- **czottmann/linearis**: JSON output, designed for LLM agents
- **evangodon/linear-cli**: Basic issue management
- **linear-4-terminal**: Rust-based comprehensive Linear interface

**Git Worktree Enhancement Tools**:
- **git-worktree-manager**: Bash script for bare clone + worktree workflow
- **wt CLI**: Fast worktree switching (mentioned in multiple articles)

### 3. AI Development Task Management
**AI Dev Tasks** (snarktank):
- ✅ Structured feature development with AI assistants
- ✅ Breaks complex features into digestible tasks  
- ✅ Works with Claude Code, Cursor, other AI tools
- ❌ No explicit worktree support mentioned

**GitHub Spec Kit** (Microsoft):
- ✅ Spec-driven development with `/specify`, `/plan`, `/tasks` commands
- ✅ Works with Claude Code, GitHub Copilot, Gemini CLI
- ❌ No worktree integration mentioned

**GitHub Copilot Coding Agent**:
- ✅ Background task execution via GitHub Actions
- ✅ Autonomous issue assignment and PR creation
- ❌ Cloud-based, not local worktree compatible

### 4. Emerging Patterns (2024-2025)
**AI + Worktree Workflow Pattern**:
- Create worktree for issue → Launch AI agent → Continue main work
- Recognized by Nx, various developer blogs
- Supported by @johnlindquist/worktree and similar tools

**Context Switching Solutions**:
- Git worktree adoption increasing specifically for AI agent workflows
- Tools focusing on "jumping between contexts" without stash/switch overhead
- Integration with PR review workflows becoming standard

## Initial Assessment Against Requirements

| Tool/Approach | Claude Code | Worktree Support | Context Management | Individual Focus | Assessment |
|---------------|-------------|------------------|-------------------|------------------|------------|
| Claude Code Feature #4963 | ✅ Native | ✅ Proposed | ✅ Planned | ✅ Yes | 🟡 Future Solution |
| @johnlindquist/worktree + Linear CLI | ✅ Compatible | ✅ Excellent | ✅ Strong | ✅ Yes | 🟢 Promising Combo |
| AI Dev Tasks | ✅ Yes | ❌ No mention | ✅ Structured | ✅ Yes | 🟡 Partial Match |
| GitHub Spec Kit | ✅ Yes | ❌ No mention | ✅ Planning | ⚠️ Team Focus | 🟡 Partial Match |
| Current ORC System | ✅ Native | ✅ Core Feature | ✅ Excellent | ✅ Yes | 🟢 Reference Point |

## Notes
Research reveals two main paths:
1. **Wait for Claude Code native support** (Feature #4963) - timeline unknown
2. **Enhance ORC with discovered patterns** - @johnlindquist/worktree + Linear CLI integration

### 5. Developer Community Patterns (2024-2025 Blogs)

**Common Workflow Pattern**:
1. Create worktrees for different branches/tasks
2. Open multiple terminal panes (iTerm2, tmux)
3. Start separate Claude Code sessions in each worktree
4. Work on multiple tasks simultaneously without context switching

**Specific Techniques Documented**:
- **iTerm2 Integration**: Split panes (Cmd+D, Cmd+Shift+D) with worktree per pane
- **Shell Aliases**: Automate worktree creation and navigation
- **Cleanup Routines**: Scripts for worktree lifecycle management
- **AI Agent Coordination**: While one AI works, continue other tasks in parallel

**Developer-Reported Benefits**:
- "10x developer productivity" through parallel task handling
- Preserved context per investigation
- Independent AI conversations per worktree
- Elimination of "idle waiting time" during AI generation

**Tools Mentioned**:
- Standard git worktree commands
- iTerm2 for terminal pane management
- Shell scripts for automation
- Various git worktree management CLIs

## Phase 1 Conclusion

**Key Finding**: No comprehensive TaskMaster replacement exists, but strong pattern of **ORC Enhancement** emerges.

**Two Primary Paths Identified**:

### Path 1: Wait for Claude Code Native Support
- **Pros**: Official Anthropic solution (Feature #4963), integrated experience
- **Cons**: No timeline, may not match ORC's specific patterns
- **Risk**: Could be 6+ months or longer

### Path 2: Enhance ORC with Community Patterns  
- **Pros**: Immediate implementation, leverages proven patterns, maintains ORC benefits
- **Cons**: Integration work required, multiple tools vs single solution
- **Risk**: Lower - builds on existing successful architecture

**Recommended Direction**: **Path 2 - ORC Enhancement**

**Specific Integration Opportunities**:
1. **@johnlindquist/worktree** for improved worktree management
2. **Linear CLI** integration for issue tracking  
3. **Enhanced TMux integration** based on community patterns
4. **Automated worktree lifecycle** commands in ORC

Key insight: The development community is actively solving worktree + AI agent integration, but most solutions are complementary tools rather than replacement task management systems.

Key lesson from TaskMaster: Surface-level Claude integration isn't enough - the underlying architecture must align with parallel development patterns.