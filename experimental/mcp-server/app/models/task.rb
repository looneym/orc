class Task < ApplicationRecord
  belongs_to :worktree
  has_many :task_histories, dependent: :destroy
  
  validates :title, presence: true
  validates :status, inclusion: { in: %w[investigating in_progress blocked completed] }
  validates :priority, inclusion: { in: %w[low medium high urgent] }
  
  enum :status, { investigating: 0, in_progress: 1, blocked: 2, completed: 3 }
  enum :priority, { low: 0, medium: 1, high: 2, urgent: 3 }
  
  scope :active, -> { where.not(status: 'completed') }
  scope :by_priority, -> { order(:priority) }
  
  def repository
    worktree.repository
  end
  
  def add_history!(action:, old_value: nil, new_value: nil, notes: nil, agent_id:)
    task_histories.create!(
      action: action,
      old_value: old_value,
      new_value: new_value, 
      notes: notes,
      agent_id: agent_id
    )
  end
end
