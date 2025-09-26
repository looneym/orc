class TaskHistory < ApplicationRecord
  belongs_to :task
  
  validates :action, presence: true
  validates :agent_id, presence: true
  
  scope :recent, -> { order(created_at: :desc) }
  scope :by_action, ->(action) { where(action: action) }
end
