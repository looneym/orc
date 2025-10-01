class CreateTaskTool < ApplicationTool
  description "Create new task for investigation (orchestrator context)"
  
  arguments do
    required(:title).filled(:string).description("Task title")
    required(:worktree_name).filled(:string).description("Target worktree name")
    optional(:description).filled(:string).description("Detailed task description") 
    optional(:priority).filled(:string, included_in?: %w[low medium high urgent]).description("Task priority (default: medium)")
  end
  
  def call(title:, worktree_name:, description: nil, priority: 'medium')
    # Find the worktree
    worktree = Worktree.find_by(name: worktree_name)
    
    return "❌ Worktree '#{worktree_name}' not found. Available: #{available_worktrees}" unless worktree
    
    # Create the task
    task = Task.create!(
      title: title,
      description: description,
      worktree: worktree,
      priority: priority,
      status: 'investigating',
      created_by: 'orchestrator',
      assigned_agent: 'implementer'
    )
    
    # Add creation history
    task.add_history!(
      action: 'created',
      new_value: 'investigating',
      notes: "Task created by orchestrator",
      agent_id: current_context.agent_id
    )
    
    format_task_created(task, worktree)
    
  rescue ActiveRecord::RecordInvalid => e
    "❌ Failed to create task: #{e.message}"
  rescue => e
    "❌ Error: #{e.message}"
  end
  
  private
  
  def available_worktrees
    Worktree.active.pluck(:name).join(', ')
  end
  
  def format_task_created(task, worktree)
    "✅ **Created Task ##{task.id}**\n\n" +
    "**Title**: #{task.title}\n" +
    "**Worktree**: #{worktree.name} (#{worktree.repository.name})\n" +
    "**Priority**: #{task.priority}\n" +
    "**Status**: #{task.status}\n" +
    "#{task.description ? "\n**Description**: #{task.description}" : ""}"
  end
end