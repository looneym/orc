class Repository < ApplicationRecord
  has_many :worktrees, dependent: :destroy
  has_many :tasks, through: :worktrees
  
  validates :name, presence: true, uniqueness: true
  validates :path, presence: true
  
  def current_branch
    return nil unless File.directory?(path)
    Dir.chdir(path) { `git branch --show-current`.strip }
  rescue
    primary_branch
  end
end
