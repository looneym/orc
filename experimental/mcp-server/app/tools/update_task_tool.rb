class UpdateTaskTool < ApplicationTool
  description "Update task status and add progress notes"
  
  arguments do
    required(:task_id).filled(:integer).description("Task ID to update")
    required(:status).filled(:string, included_in?: %w[investigating in_progress blocked completed]).description("New status")
    optional(:notes).filled(:string).description("Progress notes or comments")
    optional(:priority).filled(:string, included_in?: %w[low medium high urgent]).description("Update priority (optional)")
  end
  
  def call(task_id:, status:, notes: nil, priority: nil)
    task = Task.find(task_id)
    
    old_status = task.status
    old_priority = task.priority
    
    updates = { status: status }
    updates[:priority] = priority if priority
    
    task.update!(updates)
    
    # Add status change history
    if old_status != status
      task.add_history!(
        action: 'status_changed',
        old_value: old_status,
        new_value: status,
        notes: notes,
        agent_id: current_context.agent_id
      )
    end
    
    # Add priority change history if changed
    if priority && old_priority != priority
      task.add_history!(
        action: 'priority_changed', 
        old_value: old_priority,
        new_value: priority,
        notes: "Priority updated#{notes ? " - #{notes}" : ""}",
        agent_id: current_context.agent_id
      )
    end
    
    # Add notes if provided (even without status change)
    if notes && old_status == status
      task.add_history!(
        action: 'notes_added',
        notes: notes,
        agent_id: current_context.agent_id
      )
    end
    
    format_update_response(task, old_status, old_priority, status, priority, notes)
    
  rescue ActiveRecord::RecordNotFound
    "❌ Task ##{task_id} not found"
  rescue ActiveRecord::RecordInvalid => e
    "❌ Update failed: #{e.message}"
  rescue => e
    "❌ Error: #{e.message}"
  end
  
  private
  
  def format_update_response(task, old_status, old_priority, new_status, new_priority, notes)
    emoji = status_emoji(new_status)
    
    response = "#{emoji} **Updated Task ##{task.id}: #{task.title}**\n\n"
    
    if old_status != new_status
      response += "**Status**: #{old_status.humanize} → #{new_status.humanize}\n"
    end
    
    if new_priority && old_priority != new_priority
      response += "**Priority**: #{old_priority.humanize} → #{new_priority.humanize}\n"
    end
    
    response += "**Worktree**: #{task.worktree.name}\n"
    response += "**Updated by**: #{current_context.agent_id}\n"
    
    if notes
      response += "\n**Notes**: #{notes}"
    end
    
    response
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