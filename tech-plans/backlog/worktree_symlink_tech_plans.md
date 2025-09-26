# Worktree Symlink Tech Plans

**Status**: investigating

## Problem & Solution
**Current Issue:** Tech plans are scattered - some in local repos (bot-test), some need central coordination (ORC), no consistent access pattern for worktrees
**Solution:** Symlink architecture that provides local access to centrally managed tech plans while maintaining clean separation

## Current System Analysis

### What Works Today
- **Bot-test pattern**: Tech plans committed directly to repository work perfectly for isolated projects
- **ORC central storage**: Strategic tech plans tracked in ORC repository with version control
- **Worktree isolation**: Each worktree is self-contained development environment

### The Core Challenge
- **Bot-test approach**: Can commit tech plans directly (isolated repo) ✅
- **Multi-repo projects**: Cannot commit tech plans to shared repos (infrastructure, intercom) ❌
- **Cross-project visibility**: Orchestrator Claude needs to see all active work across worktrees ❌
- **Local access**: Implementation Claude needs easy access to relevant plans ❌

## Implementation

### Approach
**Symlink strategy that mirrors your brilliant insight**: Store centrally in ORC, access locally via symlinks. Each worktree gets its own tech plan namespace while maintaining central coordination.

### Interface/API/Contract

#### Central Storage Architecture
```
/Users/looneym/src/orc/.claude/tech_plans/
├── global/                           # Cross-project strategic plans
│   ├── orc_ecosystem_refinement.md   # Current strategic work
│   └── claude_task_master_eval.md    # Tool evaluation
├── worktrees/                        # Worktree-specific plans
│   ├── ml-dlq-bot/                   # DLQ bot investigation plans
│   │   ├── queue_label_fix.md
│   │   └── alarm_automation.md
│   ├── ml-perfbot-improvements/      # PerfBot enhancement plans
│   │   ├── system_enhancements.md
│   │   └── mcp_integration.md
│   └── ml-claude-task-master-eval/   # Task master evaluation plans
│       └── ecosystem_integration.md
└── archive/                          # Completed plans
```

#### Worktree Symlink Pattern
```
# In each worktree
/Users/looneym/src/worktrees/ml-dlq-bot/
├── .claude/
│   ├── tech_plans -> /Users/looneym/src/orc/.claude/tech_plans/worktrees/ml-dlq-bot/
│   ├── global-tech-plans -> /Users/looneym/src/orc/.claude/tech_plans/global/
│   ├── commands -> ~/.claude/commands/     # Access to global commands
│   └── local-agents/                       # Worktree-specific agents
├── intercom/                              # Main repo
├── infrastructure/                        # Infrastructure repo
└── CLAUDE.md                             # Worktree context
```

#### Command Integration
- **`/tech-plan`**: Creates plans in appropriate location based on context
  - From worktree: Creates in `orc/.claude/tech_plans/worktrees/[worktree-name]/`
  - From ORC: Creates in `orc/.claude/tech_plans/global/`
- **`/janitor`**: Manages tech plans across both local and global scopes
- **`/bootstrap`**: Shows both worktree-specific and relevant global plans

## Testing Strategy

### Phase 1: Single Worktree Prototype
- Test with `ml-dlq-bot` worktree
- Create symlink structure
- Verify tech plan creation and access
- Test `/tech-plan` and `/janitor` integration

### Phase 2: Multi-Worktree Validation  
- Extend to `ml-perfbot-improvements` and `ml-claude-task-master-eval`
- Verify cross-worktree isolation
- Test orchestrator visibility of all worktree plans
- Validate no conflicts between worktree namespaces

### Phase 3: Integration Testing
- Test global commands work in all contexts
- Verify `/bootstrap` shows appropriate plans for context
- Test `/janitor` can manage both local and global plans
- Validate orchestrator can see comprehensive status

## Implementation Plan

### Phase 1: Architecture Setup
**Goal**: Create central storage structure and first worktree integration

1. **Create ORC Tech Plan Structure**
   - Create `worktrees/` subdirectory in ORC tech plans
   - Move strategic plans to `global/` subdirectory
   - Set up namespace directories for active worktrees

2. **Implement First Worktree Symlinks**  
   - Choose `ml-dlq-bot` as prototype
   - Create `.claude/tech_plans` symlink to ORC namespace
   - Create convenience symlinks for global access
   - Test tech plan creation and access

3. **Update Command Integration**
   - Modify `/tech-plan` to detect context and create in appropriate location
   - Update `/janitor` to handle both local and global plan management
   - Ensure `/bootstrap` shows relevant plans for current context

### Phase 2: Multi-Worktree Rollout
**Goal**: Extend architecture to all active worktrees

1. **Replicate Pattern**
   - Create ORC namespaces for other active worktrees
   - Set up symlink patterns in each worktree `.claude/` directory
   - Migrate existing worktree-specific plans to ORC storage

2. **Orchestrator Integration**
   - Verify orchestrator Claude can see all worktree plans via ORC storage
   - Test cross-worktree status reporting
   - Ensure clean separation between worktree contexts

3. **Documentation Update**
   - Update worktree CLAUDE.md template with symlink patterns
   - Document tech plan creation patterns for different contexts
   - Update ORC CLAUDE.md with worktree coordination guidance

### Phase 3: Workflow Integration
**Goal**: Seamless integration with existing workflows

1. **Command Refinement**
   - Fine-tune context detection in `/tech-plan` command
   - Enhance `/janitor` cross-scope plan management  
   - Update `/bootstrap` to prioritize relevant plans by context

2. **Template Evolution**
   - Simplify tech plan template with 4-state lifecycle
   - Remove complex work order ceremony
   - Focus on "investigating → in_progress → paused → done"

3. **Validation and Cleanup**
   - Test complete workflow from plan creation to completion
   - Verify archive patterns work across all contexts
   - Clean up any legacy tech plan locations

## Success Metrics

### Local Access Efficiency
- Worktree-claude can access relevant tech plans within 2 clicks in NerdTree
- Tech plan creation context-aware (worktree vs global)
- No confusion about where plans are stored

### Central Coordination
- Orchestrator Claude can see all active work across worktrees
- Tech plan lifecycle management works across all contexts  
- Cross-project dependencies visible when needed

### Clean Separation
- Worktree plans don't interfere with each other
- Global strategic plans remain accessible to all contexts
- Archive patterns maintain organization over time

### Workflow Integration
- Existing commands work seamlessly in new architecture
- Tech plan creation feels natural and lightweight
- No additional ceremony or complexity for day-to-day use

## Architecture Benefits

1. **Best of Both Worlds**
   - Central coordination like ORC strategic plans
   - Local access like bot-test direct commits
   - Clean separation without complexity

2. **Scalable Pattern**  
   - Easy to add new worktrees without changing architecture
   - Global vs local plans naturally organized
   - Archive patterns maintain long-term organization

3. **Tool Integration Ready**
   - Symlink pattern works with existing commands  
   - NerdTree navigation efficient
   - Future claude-task-master integration simplified

4. **Context-Aware Workflow**
   - Plans created in appropriate location automatically
   - Commands show relevant information for current context
   - Orchestrator maintains comprehensive view without confusion