# ORC Task Management Domain Model

**Complete Entity Relationship Design for ORC Ecosystem**

## Core Entities & Relationships

### Repository Layer
```ruby
class Repository < ApplicationRecord
  # Core source repositories at ~/src/repo-name
  # Examples: intercom, infrastructure, event-management-system
  
  has_many :worktree_repositories
  has_many :worktrees, through: :worktree_repositories
  has_many :tasks, through: :worktrees
  
  validates :name, presence: true, uniqueness: true
  validates :path, presence: true # ~/src/intercom
  
  # Attributes
  # name: 'intercom'
  # path: '/Users/looneym/src/intercom' 
  # primary_branch: 'master'
  # language: 'ruby'
  # description: 'Main Intercom application'
  
  def current_branch
    Dir.chdir(path) { `git branch --show-current`.strip }
  end
  
  def latest_commit
    Dir.chdir(path) { `git log -1 --oneline`.strip }
  end
end
```

### Worktree Layer (Flexible Architecture)
```ruby
class Worktree < ApplicationRecord
  # Working copies at ~/src/worktrees/worktree-name
  # Can be single-repo OR multi-repo containers
  
  has_many :worktree_repositories, dependent: :destroy
  has_many :repositories, through: :worktree_repositories
  has_many :tasks, dependent: :destroy
  has_one :tmux_window, dependent: :destroy
  belongs_to :project, optional: true
  
  validates :name, presence: true, uniqueness: true
  validates :path, presence: true
  validates :architecture, inclusion: { in: %w[single_repo multi_repo] }
  
  enum status: { active: 0, paused: 1, archived: 2 }
  
  # Attributes
  # name: 'ml-dlq-investigation-ems'
  # path: '/Users/looneym/src/worktrees/ml-dlq-investigation-ems'
  # architecture: 'single_repo' | 'multi_repo'
  # primary_branch: 'ml/dlq-investigation'
  # status: 'active' | 'paused' | 'archived'
  # description: 'DLQ bot label length fix investigation'
  
  def primary_repository
    # For single_repo: the one repo
    # For multi_repo: the main repo (largest, most activity)
    repositories.order(:created_at).first
  end
  
  def tech_plans_path
    "/Users/looneym/src/orc/tech-plans/in-progress/#{name}"
  end
end

# Join table for worktree ↔ repository many-to-many
class WorktreeRepository < ApplicationRecord
  belongs_to :worktree
  belongs_to :repository
  
  # For multi-repo worktrees, track individual repo paths
  # path: 'intercom/' or 'infrastructure/' within worktree
  # branch: 'ml/feature-branch' - branch for this repo in this worktree
  # is_primary: true/false - main repo for this worktree
end
```

### Task/Work Order Layer
```ruby
class Task < ApplicationRecord
  # Individual work items (evolved from work orders)
  
  belongs_to :worktree
  belongs_to :project, optional: true
  belongs_to :epic, optional: true
  belongs_to :parent_task, class_name: 'Task', optional: true
  has_many :subtasks, class_name: 'Task', foreign_key: 'parent_task_id'
  has_many :task_histories, dependent: :destroy
  
  validates :title, presence: true
  validates :status, inclusion: { in: %w[investigating in_progress blocked completed archived] }
  validates :priority, inclusion: { in: %w[low medium high urgent] }
  validates :task_type, inclusion: { in: %w[feature bug investigation maintenance] }
  
  # Attributes
  # title: 'Fix DLQ bot label length issue'
  # description: 'Remove queue name labels when they exceed API limits'
  # status: 'investigating' | 'in_progress' | 'blocked' | 'completed' | 'archived'
  # priority: 'low' | 'medium' | 'high' | 'urgent'
  # task_type: 'feature' | 'bug' | 'investigation' | 'maintenance'
  # estimated_hours: 4.0
  # actual_hours: 6.5
  # assigned_agent: 'implementer' | 'orchestrator' | 'maintenance'
  # context: JSON field with GitHub issues, files, notes, etc.
  
  def primary_repository
    worktree.primary_repository
  end
  
  def tech_plan_files
    # Link to associated .tech-plans/*.md files
    Dir.glob("#{worktree.tech_plans_path}/*#{title.parameterize}*.md")
  end
end
```

### Project/Epic Layer (Strategic Organization)
```ruby
class Project < ApplicationRecord
  # Large initiatives that span multiple worktrees/repos
  # Examples: "DLQ Bot Overhaul", "Performance Bot v2", "Infrastructure Migration"
  
  has_many :worktrees
  has_many :epics, dependent: :destroy
  has_many :tasks
  belongs_to :owner_agent, class_name: 'Agent', optional: true
  
  validates :name, presence: true
  validates :status, inclusion: { in: %w[planning active on_hold completed archived] }
  
  # Attributes
  # name: 'DLQ Bot System Overhaul'
  # description: 'Complete redesign of DLQ bot architecture'
  # status: 'active'
  # estimated_weeks: 8
  # actual_weeks: 12
  # start_date: Date
  # target_date: Date
  # completion_date: Date
end

class Epic < ApplicationRecord
  # Mid-level groupings within projects
  # Examples: "Label Management", "Queue Processing", "Monitoring Integration"
  
  belongs_to :project
  has_many :tasks
  has_many :worktrees, through: :tasks
  
  validates :name, presence: true
  validates :status, inclusion: { in: %w[planning active blocked completed] }
  
  # Attributes similar to Project but scoped within project
end
```

### TMux Integration Layer
```ruby
class TmuxWindow < ApplicationRecord
  # TMux window sessions mapped to worktrees
  
  belongs_to :worktree
  has_many :agent_sessions, dependent: :destroy
  
  validates :window_name, presence: true, uniqueness: true
  validates :window_index, presence: true, numericality: { integer_only: true }
  
  # Attributes
  # window_name: 'dlq-investigation'
  # window_index: 3
  # session_name: 'main' (usually 'main' for El Presidente)
  # current_directory: '/Users/looneym/src/worktrees/ml-dlq-investigation-ems'
  # is_active: true/false
  # created_at: timestamp
  
  def muxup_layout_active?
    # Check if standard 3-pane muxup layout is running
    pane_count = `tmux list-panes -t #{session_name}:#{window_name} | wc -l`.strip.to_i
    pane_count == 3
  end
  
  def claude_pane_active?
    # Check if Claude Code is running in expected pane
    # This would require parsing tmux capture-pane output
  end
end
```

### Agent/Command Layer
```ruby
class Agent < ApplicationRecord
  # Claude agents and maintenance commands
  # Examples: orchestrator, implementer_dlq, janitor, bootstrap
  
  has_many :agent_sessions, dependent: :destroy
  has_many :task_histories, dependent: :destroy
  has_many :owned_projects, class_name: 'Project', foreign_key: 'owner_agent_id'
  
  validates :name, presence: true, uniqueness: true
  validates :agent_type, inclusion: { in: %w[orchestrator implementer maintenance command] }
  validates :role, inclusion: { in: %w[task_management development maintenance coordination] }
  
  # Attributes
  # name: 'orchestrator' | 'implementer_dlq' | 'janitor' | 'bootstrap'
  # agent_type: 'orchestrator' | 'implementer' | 'maintenance' | 'command'
  # role: 'task_management' | 'development' | 'maintenance' | 'coordination'
  # capabilities: JSON array of what this agent can do
  # active: true/false
  # last_seen_at: timestamp
end

class AgentSession < ApplicationRecord
  # Active Claude sessions connected to MCP server
  
  belongs_to :agent
  belongs_to :tmux_window, optional: true
  belongs_to :worktree, optional: true
  
  # Attributes
  # session_id: UUID from MCP connection
  # connected_at: timestamp
  # last_activity_at: timestamp
  # environment: JSON of PWD, git branch, etc.
  # is_active: true/false
end

class TaskHistory < ApplicationRecord
  # Audit trail of all task changes and agent actions
  
  belongs_to :task
  belongs_to :agent
  belongs_to :agent_session, optional: true
  
  # Attributes
  # action: 'created' | 'status_changed' | 'assigned' | 'commented' | 'archived'
  # old_value: previous state (for changes)
  # new_value: new state
  # notes: human-readable description
  # metadata: JSON with additional context
end
```

## Key Relationships

### Worktree Architecture Flexibility
```ruby
# Single-repo worktree (new approach)
worktree = Worktree.create!(
  name: 'ml-dlq-fix-ems',
  architecture: 'single_repo'
)
worktree.repositories << Repository.find_by(name: 'event-management-system')

# Multi-repo worktree (current approach) 
worktree = Worktree.create!(
  name: 'ml-dlq-infrastructure',
  architecture: 'multi_repo'
)
worktree.worktree_repositories.create!(
  repository: Repository.find_by(name: 'intercom'),
  path: 'intercom/',
  branch: 'ml/dlq-support',
  is_primary: true
)
worktree.worktree_repositories.create!(
  repository: Repository.find_by(name: 'infrastructure'),
  path: 'infrastructure/', 
  branch: 'ml/dlq-monitoring',
  is_primary: false
)
```

### Task Hierarchy Examples
```ruby
# Project → Epic → Tasks
project = Project.create!(name: 'DLQ Bot System Overhaul')

epic = project.epics.create!(name: 'Label Management')

parent_task = epic.tasks.create!(
  title: 'Implement label validation system',
  worktree: Worktree.find_by(name: 'ml-dlq-validation-ems'),
  task_type: 'feature'
)

# Subtasks
parent_task.subtasks.create!(
  title: 'Add label length validation',
  worktree: parent_task.worktree,
  task_type: 'feature'
)
```

### Agent ↔ Worktree ↔ TMux Flow
```ruby
# Orchestrator creates investigation
orchestrator = Agent.find_by(name: 'orchestrator')
session = orchestrator.agent_sessions.create!(session_id: SecureRandom.uuid)

worktree = Worktree.create!(name: 'ml-new-investigation-ems')
tmux_window = worktree.create_tmux_window!(
  window_name: 'new-investigation',
  session_name: 'main'
)

# Implementation agent connects in TMux window
implementer = Agent.find_by(name: 'implementer_investigation')
impl_session = implementer.agent_sessions.create!(
  tmux_window: tmux_window,
  worktree: worktree
)
```

## MCP Tools by Context

### Context Detection via Relationships
```ruby
class ContextDetector
  def current_worktree
    pwd = ENV['PWD']
    Worktree.find_by("path = ? OR path LIKE ?", pwd, "#{pwd}%")
  end
  
  def current_agent_session
    # Identify which agent is making the MCP call
    AgentSession.where(is_active: true, worktree: current_worktree).first
  end
  
  def available_tools
    # Tools available depend on agent type and worktree context
    agent = current_agent_session&.agent
    case agent&.agent_type
    when 'orchestrator'
      [:create_project, :create_worktree, :assign_tasks, :global_status]
    when 'implementer' 
      [:get_my_tasks, :update_task_status, :create_subtasks]
    when 'maintenance'
      [:cleanup_archived_tasks, :sync_git_status, :health_check]
    else
      [:basic_query_tools]
    end
  end
end
```

## Migration from Current Work Orders

### Data Migration Strategy
```ruby
# Convert existing work orders to new structure
class WorkOrderMigration
  def migrate!
    # 1. Create repositories for known repos
    Repository.find_or_create_by(name: 'intercom', path: '/Users/looneym/src/intercom')
    Repository.find_or_create_by(name: 'infrastructure', path: '/Users/looneym/src/infrastructure')
    
    # 2. Convert existing worktrees
    Dir.glob('/Users/looneym/src/worktrees/*').each do |worktree_path|
      next if File.basename(worktree_path) == 'paused'
      
      worktree_name = File.basename(worktree_path)
      
      # Detect if single or multi-repo
      architecture = detect_architecture(worktree_path)
      
      worktree = Worktree.create!(
        name: worktree_name,
        path: worktree_path,
        architecture: architecture,
        status: 'active'
      )
      
      # Link repositories
      link_repositories(worktree, worktree_path)
    end
    
    # 3. Convert work orders from tech-plans
    migrate_work_orders_from_tech_plans()
  end
end
```

This model gives you:

✅ **Flexible worktree architecture** (single or multi-repo)  
✅ **Strategic organization** (Projects → Epics → Tasks)  
✅ **TMux integration** with session tracking  
✅ **Agent coordination** with role-based capabilities  
✅ **Work order evolution** with full audit trail  
✅ **Context-aware MCP tools** based on agent type and location

Ready to build this comprehensive domain model, El Presidente?