class GlobalStatusTool < ApplicationTool
  description "Get status overview across all active worktrees (orchestrator context)"
  
  arguments do
    optional(:include_completed).filled(:bool).description("Include completed tasks in counts (default: false)")
  end
  
  def call(include_completed: false)
    worktrees = Worktree.active.includes(:repository, :tasks)
    
    if worktrees.empty?
      "🌍 **Global ORC Status**\n\n" +
      "No active worktrees found.\n\n" +
      "Available repositories:\n" +
      Repository.pluck(:name).map { |name| "• #{name}" }.join("\n")
    else
      format_global_status(worktrees, include_completed)
    end
  end
  
  private
  
  def format_global_status(worktrees, include_completed)
    header = "🌍 **Global ORC Status**\n\n"
    
    total_active = 0
    total_completed = 0
    urgent_tasks = []
    blocked_tasks = []
    
    worktree_summaries = worktrees.map do |wt|
      tasks = wt.tasks
      active_count = tasks.active.count
      completed_count = tasks.completed.count
      
      total_active += active_count
      total_completed += completed_count
      
      # Collect urgent and blocked tasks
      urgent_tasks.concat(tasks.urgent.active.to_a)
      blocked_tasks.concat(tasks.blocked.to_a)
      
      status_summary = if active_count == 0 && completed_count == 0
        "No tasks"
      elsif active_count == 0
        "✅ All complete (#{completed_count})"
      else
        parts = ["#{active_count} active"]
        parts << "#{completed_count} complete" if include_completed && completed_count > 0
        parts.join(", ")
      end
      
      "**#{wt.name}** (#{wt.repository.name})\n" +
      "   Branch: #{wt.current_branch || 'unknown'}\n" +
      "   Tasks: #{status_summary}\n"
    end
    
    # Summary stats
    summary = "**Summary**: #{total_active} active tasks across #{worktrees.count} worktrees"
    summary += ", #{total_completed} completed" if include_completed && total_completed > 0
    summary += "\n\n"
    
    # Urgent tasks alert
    if urgent_tasks.any?
      urgent_section = "🚨 **Urgent Tasks** (#{urgent_tasks.count}):\n"
      urgent_tasks.each do |task|
        urgent_section += "• ##{task.id}: #{task.title} (#{task.worktree.name})\n"
      end
      urgent_section += "\n"
    else
      urgent_section = ""
    end
    
    # Blocked tasks alert
    if blocked_tasks.any?
      blocked_section = "🚫 **Blocked Tasks** (#{blocked_tasks.count}):\n"
      blocked_tasks.each do |task|
        recent_history = task.task_histories.recent.first
        blocker_info = recent_history&.notes || "No details"
        blocked_section += "• ##{task.id}: #{task.title} - #{blocker_info}\n"
      end
      blocked_section += "\n"
    else
      blocked_section = ""
    end
    
    header + summary + urgent_section + blocked_section + worktree_summaries.join("\n")
  end
end