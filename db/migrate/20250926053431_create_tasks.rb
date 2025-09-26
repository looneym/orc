class CreateTasks < ActiveRecord::Migration[8.0]
  def change
    create_table :tasks do |t|
      t.string :title
      t.text :description
      t.string :status
      t.string :priority
      t.references :worktree, null: false, foreign_key: true
      t.string :created_by
      t.string :assigned_agent

      t.timestamps
    end
  end
end
