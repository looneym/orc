class AddIndexesAndDefaults < ActiveRecord::Migration[8.0]
  def change
    # Add default values
    change_column_default :repositories, :primary_branch, 'master'
    change_column_default :worktrees, :status, 'active'
    change_column_default :tasks, :status, 'investigating'
    change_column_default :tasks, :priority, 'medium'
    
    # Add indexes for performance
    add_index :worktrees, :status
    add_index :tasks, :status  
    add_index :tasks, [:worktree_id, :status]
    add_index :tasks, :priority
    add_index :task_histories, [:task_id, :created_at]
    add_index :task_histories, :agent_id
  end
end
