class Worktree < ApplicationRecord
  belongs_to :repository
  has_many :tasks, dependent: :destroy
  has_many :task_histories, through: :tasks
  
  validates :name, presence: true, uniqueness: true
  validates :path, presence: true
  validates :status, inclusion: { in: %w[active paused archived] }
  
  enum :status, { active: 0, paused: 1, archived: 2 }
  
  scope :active, -> { where(status: 'active') }
  
  def current_branch
    return branch if branch.present?
    return nil unless File.directory?(path)
    Dir.chdir(path) { `git branch --show-current`.strip }
  rescue
    nil
  end
  
  def active_tasks_count
    tasks.active.count
  end
  
  def completed_tasks_count  
    tasks.completed.count
  end
end
