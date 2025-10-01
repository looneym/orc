# ORC Ecosystem Integration Model

**How All The Pieces Fit Together**

## Current State Analysis

### Physical Directory Structure
```
~/src/
├── orc/                           # Rails app + MCP server (THIS REPO)
├── intercom/                      # Core repository
├── infrastructure/                # Infrastructure repository  
├── event-management-system/       # EMS repository
├── other-repos...
└── worktrees/
    ├── ml-investigation-name-repo/     # Single-repo worktrees
    ├── ml-multi-container/             # Multi-repo worktrees  
    │   ├── intercom/
    │   ├── infrastructure/
    │   └── ems/
    └── paused/
        └── ml-old-investigation-repo/
```

### TMux Session Architecture
```
El Presidente's Main Session:
├── Window 0: "orc"          (orchestrator claude)
├── Window 1: "investigation-1"  (implementer claude)  
├── Window 2: "investigation-2"  (implementer claude)
├── Window 3: "maintenance"      (janitor/bootstrap/etc.)
└── ...

Each Investigation Window (muxup layout):
┌─────────────┬─────────────┐
│     vim     │   claude    │  <- implementer claude session
│  CLAUDE.md  │             │
│ +NERDTree   ├─────────────┤
│             │    shell    │
│             │             │
└─────────────┴─────────────┘
```

## Integration Flow Model

### 1. Task Creation Flow
```
El Presidente (orc window) 
  ↓ "Create a task to fix DLQ bot labels"
ORC MCP Server (Rails app)
  ↓ create_task tool
ORC Database (SQLite)
  ↓ stores task + worktree association
Tech Plans System
  ↓ optional .tech-plans/ integration
Worktree Creation
  ↓ if new investigation needed
TMux Window Setup
  ↓ new window with implementer claude
```

### 2. Cross-Session Communication Flow  
```
Orchestrator (Window 0: orc):
- Creates tasks for investigations
- Assigns to specific worktrees
- Gets global status across all work

Implementer (Window N: investigation):
- Auto-detects current worktree via PWD
- Gets tasks assigned to current worktree  
- Updates task status/progress
- Creates subtasks for detailed work

Maintenance (Window M: maintenance):
- Runs system cleanup
- Updates git status for all worktrees
- Archives completed investigations
```

### 3. Data Flow Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    ORC Rails App                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ MCP Server  │  │   Models    │  │  Database   │        │ 
│  │ (Port 6970) │◄─┤ Repository  │◄─┤  SQLite     │        │
│  │             │  │ Worktree    │  │             │        │
│  │             │  │ Task        │  │             │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
        ▲                    ▲                    ▲
        │                    │                    │
┌───────▼─────┐    ┌─────────▼──────┐    ┌───────▼─────┐
│ Claude Code │    │ File System    │    │ Git Repos   │
│ MCP Client  │    │ Integration    │    │ Integration │
│             │    │                │    │             │
│ - Tools     │    │ ~/src/repos/   │    │ Branch      │
│ - Context   │    │ ~/src/worktrees│    │ Status      │
│ - Commands  │    │ .tech-plans/   │    │ Commits     │
└─────────────┘    └────────────────┘    └─────────────┘
```

## Entity Relationship Integration

### Core Relationships
```ruby
# Physical Infrastructure
Repository (~/src/repo-name)
  ↓ has_many worktree_repositories
Worktree (~/src/worktrees/name) 
  ↓ belongs_to tmux_window
  ↓ has_many tasks
TMuxWindow
  ↓ has_many agent_sessions

# Work Organization  
Project (strategic level)
  ↓ has_many epics
Epic (feature grouping)
  ↓ has_many tasks
Task (individual work items)
  ↓ belongs_to worktree
  ↓ has_many task_histories

# Agent Coordination
Agent (claude instances)
  ↓ has_many agent_sessions
AgentSession (active MCP connections)
  ↓ belongs_to tmux_window
  ↓ belongs_to worktree
```

## Claude Command Integration

### Current Commands → MCP Tools Evolution
```
OLD: Slash commands in individual sessions
/bootstrap   → bootstrap_tool (context loading)
/janitor     → janitor_tool (maintenance)
/tech-plan   → tech_plan_tool (planning)
/hpmboot     → taskboot_tool (agent registration)

NEW: MCP tools with cross-session awareness
create_task → Orchestrator creates tasks
get_my_tasks → Implementer gets current context
update_task → Any agent updates progress
global_status → Orchestrator sees all work
```

### Tool Context Detection
```ruby
class ContextDetector
  def agent_type
    case ENV['PWD']
    when /\/orc$/
      'orchestrator'  # In ORC root directory
    when /\/worktrees\/[^\/]+$/  
      'implementer'   # In specific worktree
    when /\/worktrees$/
      'coordinator'   # In worktrees management
    else
      'maintenance'   # General system work
    end
  end
  
  def available_tools
    case agent_type
    when 'orchestrator'
      [:create_task, :create_worktree, :global_status, :assign_tasks]
    when 'implementer'
      [:get_my_tasks, :update_task, :create_subtask, :report_progress]
    when 'maintenance' 
      [:cleanup_tasks, :archive_worktrees, :sync_git_status]
    end
  end
end
```

## Workflow Integration Scenarios

### Scenario 1: New Investigation
```
1. El Presidente in orc window:
   "I need to investigate DLQ bot performance issues"

2. ORC MCP handles:
   - create_task("Investigate DLQ performance") 
   - create_worktree("ml-dlq-perf-investigation-ems")
   - create_tmux_window("dlq-perf")

3. Physical Setup:
   - Git worktree created at ~/src/worktrees/ml-dlq-perf-investigation-ems
   - Tech plan symlink: .tech-plans/ → orc/tech-plans/in-progress/ml-dlq-perf-investigation-ems/
   - TMux window with muxup layout
   - Implementer claude session auto-connects

4. Implementer Context:
   - get_my_tasks() returns DLQ performance task
   - CLAUDE.md has full investigation context
   - Can update_task() with progress
```

### Scenario 2: Cross-Investigation Coordination
```
1. Multiple investigations running:
   - ml-dlq-perf-investigation-ems (Window 1)
   - ml-perfbot-enhancements-intercom (Window 2) 
   - ml-infrastructure-migration (Window 3)

2. Orchestrator coordination:
   - global_status() sees all active work
   - Can reassign tasks between investigations
   - Identifies blockers and dependencies

3. Task Dependencies:
   - DLQ investigation blocks PerfBot work
   - Infrastructure migration affects both
   - Cross-investigation communication via MCP
```

### Scenario 3: Maintenance and Lifecycle
```
1. Weekly cleanup (janitor in maintenance window):
   - archive_completed_tasks()
   - cleanup_stale_worktrees() 
   - sync_git_status_all()

2. Investigation completion:
   - Mark tasks as completed
   - Move worktree to paused/ or delete
   - Archive tech plans to archive/
   - Update project/epic progress

3. Context handoff:
   - Export investigation summary
   - Create follow-up tasks
   - Transfer knowledge to documentation
```

## Tech Plans Integration

### Current → Future Evolution
```
CURRENT: .claude/tech_plans/ in each worktree
- Local to specific investigation
- No cross-worktree visibility
- Manual lifecycle management

FUTURE: Symlinked + Database hybrid
- .tech-plans/ → orc/tech-plans/in-progress/[worktree]/
- Database tracks task ↔ tech plan relationships  
- Automatic lifecycle (backlog → in-progress → archive)
- Cross-investigation planning coordination
```

### Tech Plan ↔ Task Sync
```ruby
class TechPlanSyncService
  def sync_tech_plan_to_tasks(tech_plan_file)
    # Parse .md file for task items
    # Create/update database tasks
    # Maintain bidirectional sync
  end
  
  def sync_task_to_tech_plans(task)
    # Update relevant .md files
    # Add progress notes
    # Update status sections
  end
end
```

## Migration Strategy

### Phase 1: Basic MCP Integration
- ✅ Rails app with FastMCP
- ✅ Basic domain models
- ✅ Context detection
- 🔄 Core tools (create/get/update tasks)

### Phase 2: Worktree Integration
- 📋 Auto-discover existing worktrees
- 📋 TMux window mapping
- 📋 Git status integration
- 📋 Agent session tracking

### Phase 3: Command Migration
- 📋 Replace /hpmboot with /taskboot
- 📋 Update /bootstrap for MCP context
- 📋 Enhance /janitor with task management
- 📋 Integrate /tech-plan with database

### Phase 4: Advanced Coordination
- 📋 Cross-investigation dependencies
- 📋 Project/Epic strategic planning
- 📋 Automated progress reporting
- 📋 Full tech plan ↔ task synchronization

## Success Metrics

### Immediate (Phase 1-2)
- [ ] Create tasks from orchestrator context
- [ ] Auto-detect worktree context for implementers
- [ ] Real-time task updates between sessions
- [ ] Basic cross-session coordination

### Medium Term (Phase 3-4)  
- [ ] Replace existing command workflow
- [ ] Eliminate manual task tracking
- [ ] Automated investigation lifecycle
- [ ] Strategic project management

### Long Term Vision
- [ ] AI-driven task prioritization
- [ ] Predictive investigation planning  
- [ ] Automated progress reporting
- [ ] Cross-team coordination capabilities

## Questions for El Presidente

1. **Worktree Architecture**: Keep both single-repo and multi-repo support, or migrate everything to single-repo?

2. **TMux Integration**: Auto-create windows, or manual window creation with auto-detection?

3. **Command Migration**: Gradual replacement of existing commands, or big-bang switch?

4. **Tech Plans**: Full database integration, or keep file-based with database augmentation?

5. **Agent Coordination**: How detailed should cross-session communication be?

This gives us the complete picture of how ORC Task Management integrates with your existing workflow while enhancing cross-session coordination!