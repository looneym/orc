# ORC Task Management Seed Data

puts "🌱 Seeding ORC Task Management database..."

# Create repositories
repositories = [
  {
    name: 'intercom',
    path: '/Users/looneym/src/intercom',
    primary_branch: 'master'
  },
  {
    name: 'event-management-system',
    path: '/Users/looneym/src/event-management-system', 
    primary_branch: 'master'
  },
  {
    name: 'infrastructure',
    path: '/Users/looneym/src/infrastructure',
    primary_branch: 'master'
  }
]

repositories.each do |repo_data|
  repo = Repository.find_or_create_by(name: repo_data[:name]) do |r|
    r.path = repo_data[:path]
    r.primary_branch = repo_data[:primary_branch]
  end
  puts "📁 Repository: #{repo.name}"
end

# Create sample worktrees
worktrees_data = [
  {
    name: 'ml-dlq-investigation-ems',
    repository: 'event-management-system',
    path: '/Users/looneym/src/worktrees/ml-dlq-investigation-ems',
    branch: 'ml/dlq-investigation',
    status: 'active'
  },
  {
    name: 'ml-perfbot-enhancements-intercom',
    repository: 'intercom',
    path: '/Users/looneym/src/worktrees/ml-perfbot-enhancements-intercom',
    branch: 'ml/perfbot-enhancements',
    status: 'active'
  },
  {
    name: 'ml-infrastructure-cleanup',
    repository: 'infrastructure', 
    path: '/Users/looneym/src/worktrees/ml-infrastructure-cleanup',
    branch: 'ml/cleanup-unused-resources',
    status: 'paused'
  }
]

worktrees_data.each do |wt_data|
  repository = Repository.find_by!(name: wt_data[:repository])
  
  worktree = Worktree.find_or_create_by(name: wt_data[:name]) do |wt|
    wt.repository = repository
    wt.path = wt_data[:path]
    wt.branch = wt_data[:branch]
    wt.status = wt_data[:status]
  end
  puts "🌳 Worktree: #{worktree.name} (#{worktree.repository.name})"
end

# Create sample tasks
tasks_data = [
  {
    title: 'Fix DLQ bot label length issue',
    description: 'Remove queue name labels when they exceed API limits to prevent creation failures',
    worktree: 'ml-dlq-investigation-ems',
    status: 'in_progress',
    priority: 'high',
    created_by: 'orchestrator',
    assigned_agent: 'implementer'
  },
  {
    title: 'Add DLQ monitoring dashboard',
    description: 'Create Honeycomb dashboard for DLQ processing metrics and alerts',
    worktree: 'ml-dlq-investigation-ems', 
    status: 'investigating',
    priority: 'medium',
    created_by: 'orchestrator',
    assigned_agent: 'implementer'
  },
  {
    title: 'Optimize PerfBot memory usage',
    description: 'Reduce memory footprint by implementing lazy loading for large datasets',
    worktree: 'ml-perfbot-enhancements-intercom',
    status: 'blocked',
    priority: 'urgent',
    created_by: 'orchestrator', 
    assigned_agent: 'implementer'
  },
  {
    title: 'Update PerfBot notification system',
    description: 'Migrate from Slack webhooks to proper Slack app integration',
    worktree: 'ml-perfbot-enhancements-intercom',
    status: 'investigating',
    priority: 'low',
    created_by: 'orchestrator',
    assigned_agent: 'implementer'
  }
]

tasks_data.each do |task_data|
  worktree = Worktree.find_by!(name: task_data[:worktree])
  
  task = Task.find_or_initialize_by(
    title: task_data[:title],
    worktree: worktree
  )
  
  if task.new_record?
    task.assign_attributes(
      description: task_data[:description],
      status: task_data[:status],
      priority: task_data[:priority],
      created_by: task_data[:created_by],
      assigned_agent: task_data[:assigned_agent]
    )
    task.save!
    
    # Add creation history
    task.add_history!(
      action: 'created',
      new_value: task.status,
      notes: 'Initial task creation (seed data)',
      agent_id: 'system_seed'
    )
    
    puts "📋 Task: #{task.title} (#{task.status})"
    
    # Add some sample history for demonstration
    if task.status != 'investigating'
      task.add_history!(
        action: 'status_changed',
        old_value: 'investigating', 
        new_value: task.status,
        notes: "Status updated during initial setup",
        agent_id: 'system_seed'
      )
    end
  end
end

puts "\n✅ Seeding complete!"
puts "\n📊 Summary:"
puts "• #{Repository.count} repositories"
puts "• #{Worktree.count} worktrees (#{Worktree.active.count} active)" 
puts "• #{Task.count} tasks (#{Task.active.count} active)"
puts "• #{TaskHistory.count} history entries"
