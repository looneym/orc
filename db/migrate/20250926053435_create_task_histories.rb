class CreateTaskHistories < ActiveRecord::Migration[8.0]
  def change
    create_table :task_histories do |t|
      t.references :task, null: false, foreign_key: true
      t.string :action
      t.string :old_value
      t.string :new_value
      t.text :notes
      t.string :agent_id

      t.timestamps
    end
  end
end
