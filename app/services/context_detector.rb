class ContextDetector
  def current_worktree
    pwd = ENV['PWD'] || Dir.pwd
    return nil unless pwd.include?('/worktrees/')
    
    # Extract worktree name from path
    worktree_name = pwd.split('/worktrees/')[1]&.split('/')&.first
    return nil unless worktree_name
    
    Worktree.find_by(name: worktree_name)
  end
  
  def agent_type
    if ENV['PWD']&.include?('/orc')
      'orchestrator'
    elsif current_worktree
      'implementer'  
    else
      'maintenance'
    end
  end
  
  def agent_id
    case agent_type
    when 'orchestrator'
      'orchestrator'
    when 'implementer'
      current_worktree ? "implementer_#{current_worktree.name}" : 'implementer_unknown'
    else
      'maintenance'
    end
  end
  
  def in_worktree?
    !current_worktree.nil?
  end
  
  def in_orchestrator_context?
    agent_type == 'orchestrator'
  end
end