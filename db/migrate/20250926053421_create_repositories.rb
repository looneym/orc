class CreateRepositories < ActiveRecord::Migration[8.0]
  def change
    create_table :repositories do |t|
      t.string :name
      t.string :path
      t.string :primary_branch

      t.timestamps
    end
    add_index :repositories, :name, unique: true
  end
end
