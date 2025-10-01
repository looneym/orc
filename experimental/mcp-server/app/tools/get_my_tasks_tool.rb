class GetMyTasksTool < ApplicationTool
  description "Get tasks for current worktree context (implementer)"
  
  arguments do
    optional(:status).filled(:string, included_in?: %w[investigating in_progress blocked completed]).description("Filter by status")
  end
  
  def call(status: nil)
    worktree = current_context.current_worktree
    
    return "❌ No worktree context detected. Make sure you're in a worktree directory." unless worktree
    
    # Get tasks for this worktree
    tasks = worktree.tasks
    tasks = tasks.where(status: status) if status
    tasks = tasks.by_priority.includes(:task_histories)
    
    if tasks.empty?
      status_filter = status ? " with status '#{status}'" : ""
      "📭 No tasks found for **#{worktree.name}**#{status_filter}\n\n" +
      "Repository: #{worktree.repository.name}\n" +
      "Branch: #{worktree.current_branch || 'unknown'}"
    else
      format_task_list(tasks, worktree)
    end
  end
  
  private
  
  def format_task_list(tasks, worktree)
    header = "📋 **Tasks for #{worktree.name}**\n" +
             "Repository: #{worktree.repository.name} | Branch: #{worktree.current_branch || 'unknown'}\n\n"
    
    task_lines = tasks.map.with_index do |task, index|
      emoji = status_emoji(task.status)
      priority_flag = task.urgent? ? " 🚨" : (task.high? ? " ⚠️" : "")
      
      recent_history = task.task_histories.recent.limit(1).first
      last_update = recent_history ? " (#{recent_history.agent_id} #{time_ago(recent_history.created_at)})" : ""
      
      "#{emoji} **##{task.id}: #{task.title}**#{priority_flag}\n" +
      "   Status: #{task.status.humanize} | Priority: #{task.priority.humanize}#{last_update}\n" +
      "#{task.description ? "   #{task.description}\n" : ""}" +
      "\n"
    end
    
    header + task_lines.join("")
  end
  
  def status_emoji(status)
    {
      'investigating' => '🔍',
      'in_progress' => '⚡',
      'blocked' => '🚫', 
      'completed' => '✅'
    }[status] || '📋'
  end
  
  def time_ago(timestamp)
    seconds_ago = Time.current - timestamp
    
    case seconds_ago
    when 0..60
      "#{seconds_ago.to_i}s ago"
    when 61..3600
      "#{(seconds_ago / 60).to_i}m ago" 
    when 3601..86400
      "#{(seconds_ago / 3600).to_i}h ago"
    else
      "#{(seconds_ago / 86400).to_i}d ago"
    end
  end
end