# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# This file is the source Rails uses to define your schema when running `bin/rails
# db:schema:load`. When creating a new database, `bin/rails db:schema:load` tends to
# be faster and is potentially less error prone than running all of your
# migrations from scratch. Old migrations may fail to apply correctly if those
# migrations use external dependencies or application code.
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema[8.0].define(version: 2025_09_26_053538) do
  create_table "repositories", force: :cascade do |t|
    t.string "name"
    t.string "path"
    t.string "primary_branch", default: "master"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["name"], name: "index_repositories_on_name", unique: true
  end

  create_table "task_histories", force: :cascade do |t|
    t.integer "task_id", null: false
    t.string "action"
    t.string "old_value"
    t.string "new_value"
    t.text "notes"
    t.string "agent_id"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["agent_id"], name: "index_task_histories_on_agent_id"
    t.index ["task_id", "created_at"], name: "index_task_histories_on_task_id_and_created_at"
    t.index ["task_id"], name: "index_task_histories_on_task_id"
  end

  create_table "tasks", force: :cascade do |t|
    t.string "title"
    t.text "description"
    t.string "status", default: "investigating"
    t.string "priority", default: "medium"
    t.integer "worktree_id", null: false
    t.string "created_by"
    t.string "assigned_agent"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["priority"], name: "index_tasks_on_priority"
    t.index ["status"], name: "index_tasks_on_status"
    t.index ["worktree_id", "status"], name: "index_tasks_on_worktree_id_and_status"
    t.index ["worktree_id"], name: "index_tasks_on_worktree_id"
  end

  create_table "worktrees", force: :cascade do |t|
    t.string "name"
    t.integer "repository_id", null: false
    t.string "path"
    t.string "branch"
    t.string "status", default: "active"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["name"], name: "index_worktrees_on_name", unique: true
    t.index ["repository_id"], name: "index_worktrees_on_repository_id"
    t.index ["status"], name: "index_worktrees_on_status"
  end

  add_foreign_key "task_histories", "tasks"
  add_foreign_key "tasks", "worktrees"
  add_foreign_key "worktrees", "repositories"
end
