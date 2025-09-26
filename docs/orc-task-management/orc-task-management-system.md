# ORC Task Management System

**Status**: investigating

## Problem & Solution

**Current Issue:** No existing task management tool provides native Claude Code CLI integration with worktree awareness, TMux session coordination, and MCP-based cross-agent communication.

**Solution:** Build a custom MCP-based task coordination system that natively understands ORC's worktree architecture and TMux workflow patterns.

## Core Requirements

### 1. Native ORC Integration
- **Worktree Awareness**: Tasks automatically linked to specific investigations
- **TMux Mapping**: Agent sessions mapped to TMux windows
- **Tech Plans Integration**: Tasks flow from tech plans, sync back to progress
- **Git Context**: Tasks understand branch state, commit history, PR status

### 2. MCP-First Architecture  
- **Claude Code Native**: Primary interface is slash commands
- **Cross-Session Communication**: Agents coordinate via centralized database
- **Real-time Updates**: Task state changes propagate to all connected agents
- **Natural Language**: No complex UI, just conversation with tasks

### 3. Role-Based Coordination
- **Orchestrator Agent**: Creates investigations, assigns tasks, coordinates cross-worktree
- **Implementation Agents**: Work on specific tasks within their worktree context
- **Automatic Role Detection**: Context-aware based on working directory
- **Handoff Protocols**: Clean context transfer between orchestrator and implementers

## Architecture Design

### Core Components

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   ORC Task DB   │    │   MCP Server     │    │  Claude Agents  │
│                 │◄──►│                  │◄──►│                 │
│ - Tasks         │    │ - Task CRUD      │    │ - Orchestrator  │
│ - Worktrees     │    │ - Role Detection │    │ - Implementers  │  
│ - Assignments   │    │ - Notifications  │    │ - Context Aware │
│ - State History │    │ - Sync Engine    │    │ - Command Driven│
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Data Model

#### Tasks
```json
{
  "id": "task_001",
  "title": "Fix DLQ bot label length issue",
  "description": "Remove queue name labels when they exceed API limits",
  "status": "investigating|in_progress|blocked|completed",
  "priority": "high|medium|low",
  "worktree": "ml-dlqbot-label-fix-ems",
  "repository": "event-management-system",
  "branch": "ml/dlqbot-label-fix",
  "assigned_agent": "implementer",
  "created_by": "orchestrator",
  "created_at": "2025-01-25T10:00:00Z",
  "updated_at": "2025-01-25T10:30:00Z",
  "tech_plan_link": ".tech-plans/dlq-label-fix.md",
  "context": {
    "github_issue": null,
    "related_files": ["cloudbot/issue_creator.py"],
    "testing_notes": "Test with long queue names"
  }
}
```

#### Worktrees
```json
{
  "name": "ml-dlqbot-label-fix-ems",
  "repository": "event-management-system", 
  "branch": "ml/dlqbot-label-fix",
  "path": "/Users/looneym/src/worktrees/ml-dlqbot-label-fix-ems",
  "tmux_window": "dlqbot-fix",
  "agent_session": "implementer_001",
  "status": "active|paused|archived",
  "created_at": "2025-01-25T09:45:00Z"
}
```

## MCP Interface Design

### Core Tools

#### For Orchestrator Context
```javascript
// Task creation and assignment
create_task({
  title: "Fix bug description",
  worktree: "ml-investigation-repo", 
  priority: "high",
  context: {...}
})

// Cross-worktree visibility  
list_all_tasks({status?: "in_progress"})
get_worktree_status("ml-investigation-repo")

// Investigation setup
create_investigation({
  name: "dlq-label-fix",
  repository: "event-management-system",
  branch: "ml/dlq-label-fix"  
})
```

#### For Implementation Context
```javascript
// Context-aware task access
get_my_tasks() // Auto-detects worktree context
update_task_status(task_id, "in_progress")
add_task_note(task_id, "Found issue in cloudbot/issue_creator.py:42")

// Progress reporting
report_progress(task_id, {
  status: "blocked",
  notes: "Need clarification on API limits",
  next_action: "Research Intercom API documentation"
})
```

### Command Integration

#### Enhanced `/hpmboot` → `/taskboot`
```markdown
**Just run `/taskboot` for automatic task discovery and context loading**

1. Detect current worktree context
2. Register as implementation agent for this investigation
3. Load active tasks for current worktree
4. Show task status and next actions
5. Update CLAUDE.md with task context
```

#### New `/task` Command
```markdown
**Universal task management command**

/task list                    # Show my tasks (context-aware)
/task create "Fix bug X"      # Create task (orchestrator context)
/task update <id> blocked     # Update task status
/task assign <id> <worktree>  # Assign to specific investigation
/task complete <id>           # Mark complete with notes
```

## Implementation Strategy

### Phase 1: Core MCP Server (2 weeks)
- [ ] SQLite database with task/worktree models
- [ ] Basic MCP server with CRUD operations
- [ ] Context detection (orchestrator vs worktree)
- [ ] Authentication and multi-session support

### Phase 2: Command Integration (1 week)
- [ ] Update `/hpmboot` → `/taskboot` 
- [ ] Create `/task` universal command
- [ ] Integrate with existing ORC commands
- [ ] Test multi-session coordination

### Phase 3: Workflow Integration (1 week)  
- [ ] Auto-task creation from tech plans
- [ ] Git integration (branch/commit awareness)
- [ ] TMux window mapping
- [ ] Progress sync back to tech plans

### Phase 4: Advanced Features (2 weeks)
- [ ] Task dependencies and blocking
- [ ] Time tracking and estimates
- [ ] Cross-worktree task relationships
- [ ] Automated status reporting

## Technical Specifications

### MCP Server Stack
- **FastAPI**: HTTP MCP server on port 6970
- **SQLite**: Local database with migrations
- **Pydantic**: Type-safe models and validation  
- **Authentication**: API key based (like HPM pattern)
- **Logging**: Structured logging for debugging

### Database Schema
```sql
CREATE TABLE tasks (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  description TEXT,
  status TEXT CHECK(status IN ('investigating','in_progress','blocked','completed')),
  priority TEXT CHECK(priority IN ('low','medium','high')),
  worktree TEXT,
  repository TEXT,
  branch TEXT,
  assigned_agent TEXT,
  created_by TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  tech_plan_link TEXT,
  context JSON
);

CREATE TABLE worktrees (
  name TEXT PRIMARY KEY,
  repository TEXT NOT NULL,
  branch TEXT,
  path TEXT,
  tmux_window TEXT,
  agent_session TEXT,
  status TEXT CHECK(status IN ('active','paused','archived')),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE task_history (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  task_id TEXT,
  agent_id TEXT,
  action TEXT,
  notes TEXT,
  timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(task_id) REFERENCES tasks(id)
);
```

### Directory Structure
```
orc/
├── task-management/
│   ├── server/              # MCP server implementation  
│   │   ├── main.py         # FastAPI app
│   │   ├── models.py       # Pydantic models
│   │   ├── database.py     # SQLite operations  
│   │   ├── mcp_tools.py    # MCP tool definitions
│   │   └── migrations/     # Database migrations
│   ├── client/             # Python client library (if needed)
│   ├── tests/              # Test suite
│   └── README.md           # Setup and usage docs
└── global-commands/
    ├── taskboot.md         # Enhanced bootstrap command
    └── task.md             # Universal task command
```

## Testing Strategy

### Unit Tests
- Database operations and migrations
- MCP tool registration and execution
- Context detection logic
- Task state transitions

### Integration Tests  
- Multi-agent coordination scenarios
- Worktree lifecycle integration
- TMux window mapping
- Cross-session communication

### Manual Workflow Tests
1. **Orchestrator Creates Investigation**: New task + worktree setup
2. **Implementation Agent Connects**: Auto-discovery and context loading
3. **Cross-Session Updates**: Status changes propagate immediately
4. **Task Completion**: Progress syncs back to tech plans

## Success Metrics

### Immediate Goals
- [ ] Replace HPM with native ORC solution
- [ ] Zero-config agent registration per worktree
- [ ] Real-time task coordination between sessions
- [ ] Natural language task management via `/task` commands

### Long-term Vision
- [ ] Full tech plan ↔ task synchronization
- [ ] Automated progress reporting
- [ ] Git-aware task lifecycle management
- [ ] Cross-investigation dependency tracking

## Next Steps

1. **Architecture Review**: Validate design with El Presidente
2. **Prototype Setup**: Create basic MCP server structure  
3. **Database Design**: Implement core models and migrations
4. **MCP Integration**: Build and test core tools
5. **Command Integration**: Update existing ORC commands

## Notes

This system will be purpose-built for our exact workflow:
- Single developer with multiple concurrent investigations
- TMux-based session management  
- Git worktree architecture
- Claude Code CLI as primary interface
- Natural language task coordination

No compromises, no external dependencies that don't fit our patterns.