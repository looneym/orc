class CreateWorktrees < ActiveRecord::Migration[8.0]
  def change
    create_table :worktrees do |t|
      t.string :name
      t.references :repository, null: false, foreign_key: true
      t.string :path
      t.string :branch
      t.string :status

      t.timestamps
    end
    add_index :worktrees, :name, unique: true
  end
end
