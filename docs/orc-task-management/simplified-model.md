# Simplified ORC Task Management Model

**Single-Repo, Database-Only, MCP-First Approach**

## Design Decisions

✅ **Single-Repo Worktrees Only** - No multi-repo complexity  
✅ **No TMux Integration** - Pure MCP tool coordination  
✅ **Full Database** - No file-based tech plans, everything in DB  
✅ **Spike Approach** - Build new system alongside existing  

## Simplified Domain Model

### Core Entities (Minimal Set)

```ruby
# Physical repositories
Repository
- name (string): 'intercom', 'ems', 'infrastructure'
- path (string): '/Users/looneym/src/intercom'
- primary_branch (string): 'master'

# Working investigations  
Worktree  
- name (string): 'ml-dlq-investigation-ems'
- repository_id (FK)
- path (string): '/Users/looneym/src/worktrees/ml-dlq-investigation-ems'
- branch (string): 'ml/dlq-investigation'
- status (enum): active, paused, archived

# Work items
Task
- title (string): 'Fix DLQ bot label length issue'
- description (text): 'Remove queue name labels when...'
- status (enum): investigating, in_progress, blocked, completed
- priority (enum): low, medium, high, urgent
- worktree_id (FK)
- created_by (string): 'orchestrator'
- assigned_agent (string): 'implementer'

# Audit trail
TaskHistory  
- task_id (FK)
- action (string): 'created', 'status_changed', 'completed'
- old_value (string)
- new_value (string)
- notes (text)
- agent_id (string): 'orchestrator', 'implementer_dlq'
```

### Simplified Relationships
```
Repository (1) → (many) Worktree
Worktree (1) → (many) Task  
Task (1) → (many) TaskHistory
```

## Context Detection (Simplified)

```ruby
class ContextDetector
  def current_worktree
    pwd = ENV['PWD']
    return nil unless pwd.include?('/worktrees/')
    
    # Extract worktree name from path
    worktree_name = pwd.split('/worktrees/')[1]&.split('/')&.first
    Worktree.find_by(name: worktree_name)
  end
  
  def agent_type
    if ENV['PWD'].include?('/orc')
      'orchestrator'
    elsif current_worktree
      'implementer'  
    else
      'maintenance'
    end
  end
end
```

## Core MCP Tools (Minimal Set)

### 1. CreateTaskTool (Orchestrator)
```ruby
class CreateTaskTool < ApplicationTool
  description "Create new task for investigation"
  
  arguments do
    required(:title).filled(:string)
    required(:worktree_name).filled(:string) 
    optional(:description).filled(:string)
    optional(:priority).filled(:string, included_in?: %w[low medium high urgent])
  end
  
  def call(title:, worktree_name:, description: nil, priority: 'medium')
    worktree = Worktree.find_by!(name: worktree_name)
    
    task = Task.create!(
      title: title,
      description: description,
      worktree: worktree,
      priority: priority,
      status: 'investigating',
      created_by: 'orchestrator'
    )
    
    TaskHistory.create!(
      task: task,
      action: 'created',
      notes: "Task created by orchestrator",
      agent_id: 'orchestrator'
    )
    
    "✅ Created task '#{title}' for #{worktree_name}"
  end
end
```

### 2. GetMyTasksTool (Implementer)
```ruby
class GetMyTasksTool < ApplicationTool
  description "Get tasks for current worktree (implementer context)"
  
  def call
    worktree = current_context.current_worktree
    return "❌ No worktree context detected" unless worktree
    
    tasks = worktree.tasks.active.order(:priority, :created_at)
    
    if tasks.empty?
      "No active tasks for #{worktree.name}"
    else
      format_task_list(tasks, worktree.name)
    end
  end
  
  private
  
  def format_task_list(tasks, worktree_name)
    header = "📋 **Active Tasks for #{worktree_name}**\n\n"
    
    task_lines = tasks.map do |task|
      emoji = status_emoji(task.status)
      "#{emoji} **#{task.title}** (#{task.priority} priority)\n" +
      "   Status: #{task.status}\n" + 
      "   #{task.description || 'No description'}\n"
    end
    
    header + task_lines.join("\n")
  end
  
  def status_emoji(status)
    {
      'investigating' => '🔍',
      'in_progress' => '⚡',
      'blocked' => '🚫', 
      'completed' => '✅'
    }[status] || '📋'
  end
end
```

### 3. UpdateTaskTool (Any Agent)
```ruby
class UpdateTaskTool < ApplicationTool
  description "Update task status and add progress notes"
  
  arguments do
    required(:task_id).filled(:integer)
    required(:status).filled(:string, included_in?: %w[investigating in_progress blocked completed])
    optional(:notes).filled(:string)
  end
  
  def call(task_id:, status:, notes: nil)
    task = Task.find(task_id)
    old_status = task.status
    
    task.update!(status: status)
    
    TaskHistory.create!(
      task: task,
      action: 'status_changed',
      old_value: old_status,
      new_value: status,
      notes: notes,
      agent_id: current_context.agent_id
    )
    
    "✅ Updated '#{task.title}' from #{old_status} to #{status}"
  rescue ActiveRecord::RecordNotFound
    "❌ Task not found"
  end
end
```

### 4. GlobalStatusTool (Orchestrator)
```ruby
class GlobalStatusTool < ApplicationTool
  description "Get status across all active worktrees (orchestrator context)"
  
  def call
    worktrees = Worktree.active.includes(:tasks, :repository)
    
    if worktrees.empty?
      "No active worktrees found"
    else
      format_global_status(worktrees)
    end
  end
  
  private
  
  def format_global_status(worktrees)
    header = "🌍 **Global ORC Status**\n\n"
    
    worktree_summaries = worktrees.map do |wt|
      active_tasks = wt.tasks.active.count
      completed_tasks = wt.tasks.completed.count
      
      "**#{wt.name}** (#{wt.repository.name})\n" +
      "   Branch: #{wt.branch}\n" +
      "   Tasks: #{active_tasks} active, #{completed_tasks} completed\n"
    end
    
    header + worktree_summaries.join("\n")
  end
end
```

## Database Schema (Simplified)

```ruby
# db/migrate/001_create_repositories.rb
class CreateRepositories < ActiveRecord::Migration[8.0]
  def change
    create_table :repositories do |t|
      t.string :name, null: false, index: { unique: true }
      t.string :path, null: false
      t.string :primary_branch, default: 'master'
      t.timestamps
    end
  end
end

# db/migrate/002_create_worktrees.rb  
class CreateWorktrees < ActiveRecord::Migration[8.0]
  def change
    create_table :worktrees do |t|
      t.string :name, null: false, index: { unique: true }
      t.references :repository, null: false, foreign_key: true
      t.string :path, null: false
      t.string :branch
      t.string :status, null: false, default: 'active'
      t.timestamps
    end
    
    add_index :worktrees, :status
  end
end

# db/migrate/003_create_tasks.rb
class CreateTasks < ActiveRecord::Migration[8.0]
  def change
    create_table :tasks do |t|
      t.string :title, null: false
      t.text :description
      t.string :status, null: false, default: 'investigating'
      t.string :priority, null: false, default: 'medium'
      t.references :worktree, null: false, foreign_key: true
      t.string :created_by
      t.string :assigned_agent
      t.timestamps
    end
    
    add_index :tasks, :status
    add_index :tasks, [:worktree_id, :status]
  end
end

# db/migrate/004_create_task_histories.rb
class CreateTaskHistories < ActiveRecord::Migration[8.0]
  def change
    create_table :task_histories do |t|
      t.references :task, null: false, foreign_key: true
      t.string :action, null: false
      t.string :old_value
      t.string :new_value
      t.text :notes
      t.string :agent_id, null: false
      t.timestamps
    end
    
    add_index :task_histories, [:task_id, :created_at]
  end
end
```

## Workflow Examples (Simplified)

### Example 1: Create Investigation
```bash
# In orc directory (orchestrator context)
$ claude

> "I need to investigate DLQ bot label length issues in EMS"

Claude automatically:
1. Calls create_task("Fix DLQ bot label length", "ml-dlq-labels-ems", ...)
2. Task stored in database
3. Ready for implementation

# Manually create worktree (no auto-creation for now)
$ git worktree add ~/src/worktrees/ml-dlq-labels-ems -b ml/dlq-labels origin/master
```

### Example 2: Work on Tasks  
```bash
# In worktree directory (implementer context)
$ cd ~/src/worktrees/ml-dlq-labels-ems
$ claude

> "What tasks do I have?"

Claude automatically:
1. Detects worktree context from PWD
2. Calls get_my_tasks() 
3. Shows tasks for this specific worktree

> "Mark the first task as in progress"

Claude automatically:
1. Calls update_task(task_id, "in_progress", "Starting implementation")
2. Updates database + audit trail
```

### Example 3: Global Coordination
```bash
# Back in orc directory (orchestrator context)
$ cd ~/src/orc
$ claude

> "What's the status across all investigations?"

Claude automatically:
1. Calls global_status()
2. Shows summary of all active worktrees and their tasks
```

## Implementation Status Update

### ✅ COMPLETED (Phase 1)
- [x] **Rails Foundation**: Converted ORC repo to Rails 8.0 API app
- [x] **FastMCP Integration**: Added FastMCP gem with tool auto-discovery
- [x] **Domain Models**: Created Repository, Worktree, Task, TaskHistory models
- [x] **Database Schema**: Full migrations with proper indexes and relationships
- [x] **Core MCP Tools**: Built 5 essential tools:
  - `CreateTaskTool` - Orchestrator creates tasks
  - `GetMyTasksTool` - Implementer gets context-aware tasks
  - `UpdateTaskTool` - Any agent updates task status/notes
  - `GlobalStatusTool` - Orchestrator sees cross-worktree status
  - `TestTool` - Basic connectivity testing
- [x] **Context Detection**: Auto-detects orchestrator vs implementer mode via PWD
- [x] **Seed Data**: Sample repositories, worktrees, and tasks for testing
- [x] **Task History**: Full audit trail with agent attribution

### 🔄 CURRENT STATUS (January 26, 2025)
**Rails server running on port 6970** with all MCP tools registered and operational.

**Database contains**:
- 3 repositories (intercom, event-management-system, infrastructure)
- 3 worktrees (2 active: dlq-investigation, perfbot-enhancements)  
- 4 tasks (1 urgent blocked, 1 high in-progress, 2 investigating)
- 8 history entries with full audit trail

**Tools ready for testing** but need MCP transport configuration for Claude Code integration.

### 🚧 NEXT STEPS (Immediate)
- [ ] **MCP Transport Setup**: Configure FastMCP HTTP endpoint for Claude Code
- [ ] **End-to-End Testing**: Test complete workflow (create → get → update tasks)
- [ ] **Claude Code Config**: Set up ~/.claude.json MCP client configuration
- [ ] **Context Validation**: Test orchestrator vs implementer context detection

### 🎯 SUCCESS CRITERIA (Updated)

#### MVP (Minimum Viable Product) 
- [x] Create tasks from orchestrator context
- [x] Auto-detect worktree context for implementers  
- [x] Update task status from any context
- [x] Global visibility for orchestrator
- [x] Basic audit trail
- [ ] **MCP Client Integration** ← CURRENT BLOCKER
- [ ] **Real-world Workflow Test** ← PENDING MCP CONNECTION

#### Phase 2 (After MVP Integration)
- [ ] Create worktree tool (optional automation)
- [ ] Archive completed investigations  
- [ ] Task dependencies and subtasks
- [ ] Time tracking and estimates
- [ ] Integration with existing /bootstrap, /janitor commands

## Technical Architecture (As Built)

### Database Schema
```sql
repositories: id, name, path, primary_branch
worktrees: id, name, repository_id, path, branch, status
tasks: id, title, description, status, priority, worktree_id, created_by, assigned_agent
task_histories: id, task_id, action, old_value, new_value, notes, agent_id, created_at
```

### MCP Tools Available
```ruby
CreateTaskTool    # orchestrator → create_task(title, worktree_name, ...)
GetMyTasksTool    # implementer → get_my_tasks(status?)
UpdateTaskTool    # any → update_task(task_id, status, notes?)
GlobalStatusTool  # orchestrator → global_status(include_completed?)
TestTool          # any → test_tool(message?)
```

### Context Detection Logic
```ruby
ENV['PWD'].include?('/orc') → 'orchestrator'
ENV['PWD'].include?('/worktrees/') → 'implementer' 
else → 'maintenance'
```

## Lessons Learned

1. **Rails Generators**: Excellent for rapid model/migration creation
2. **FastMCP**: Tool auto-discovery works well, but transport setup needs more investigation
3. **Context Detection**: Simple PWD-based detection is effective for role determination
4. **Database Design**: Single-repo, DB-only approach eliminates complexity
5. **Seed Data**: Essential for realistic testing and development

## Next Session Goals

1. **Complete MCP Integration**: Get Claude Code talking to Rails server
2. **Validate Workflow**: Full orchestrator → implementer → status cycle
3. **Performance Test**: Ensure response times are acceptable
4. **Error Handling**: Robust error messages and recovery

**Current State**: Fully functional Rails MCP server waiting for client integration!