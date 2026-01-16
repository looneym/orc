package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/db"
)

type Task struct {
	ID              string
	EpicID          sql.NullString
	RabbitHoleID    sql.NullString
	MissionID       string
	Title           string
	Description     sql.NullString
	Type            sql.NullString
	Status          string
	Priority        sql.NullString
	AssignedGroveID sql.NullString
	ContextRef      sql.NullString
	Pinned          bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ClaimedAt       sql.NullTime
	CompletedAt     sql.NullTime
}

// CreateTask creates a new task under an epic or rabbit hole
func CreateTask(epicID, rabbitHoleID, missionID, title, description, taskType string) (*Task, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	// Verify exactly one parent is specified (epic XOR rabbit hole)
	if (epicID == "" && rabbitHoleID == "") || (epicID != "" && rabbitHoleID != "") {
		return nil, fmt.Errorf("task must belong to either an epic OR a rabbit hole (not both, not neither)")
	}

	// Verify mission exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM missions WHERE id = ?", missionID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, fmt.Errorf("mission %s not found", missionID)
	}

	// If epic ID specified, verify epic exists and has no rabbit holes
	if epicID != "" {
		err = database.QueryRow("SELECT COUNT(*) FROM epics WHERE id = ?", epicID).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if exists == 0 {
			return nil, fmt.Errorf("epic %s not found", epicID)
		}

		// Check if epic has rabbit holes (no mixed children)
		var rhCount int
		err = database.QueryRow("SELECT COUNT(*) FROM rabbit_holes WHERE epic_id = ?", epicID).Scan(&rhCount)
		if err != nil {
			return nil, err
		}
		if rhCount > 0 {
			return nil, fmt.Errorf("epic %s has rabbit holes\nTasks must be created under rabbit holes, not directly under epic (no mixed children)", epicID)
		}
	}

	// If rabbit hole ID specified, verify it exists
	if rabbitHoleID != "" {
		err = database.QueryRow("SELECT COUNT(*) FROM rabbit_holes WHERE id = ?", rabbitHoleID).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if exists == 0 {
			return nil, fmt.Errorf("rabbit hole %s not found", rabbitHoleID)
		}
	}

	// Generate task ID by finding max existing ID
	var maxID int
	err = database.QueryRow("SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM tasks").Scan(&maxID)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("TASK-%03d", maxID+1)

	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	var typeVal sql.NullString
	if taskType != "" {
		typeVal = sql.NullString{String: taskType, Valid: true}
	}

	var epicIDVal sql.NullString
	if epicID != "" {
		epicIDVal = sql.NullString{String: epicID, Valid: true}
	}

	var rhIDVal sql.NullString
	if rabbitHoleID != "" {
		rhIDVal = sql.NullString{String: rabbitHoleID, Valid: true}
	}

	_, err = database.Exec(
		"INSERT INTO tasks (id, epic_id, rabbit_hole_id, mission_id, title, description, type, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		id, epicIDVal, rhIDVal, missionID, title, desc, typeVal, "ready",
	)
	if err != nil {
		return nil, err
	}

	return GetTask(id)
}

// GetTask retrieves a task by ID
func GetTask(id string) (*Task, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	task := &Task{}
	err = database.QueryRow(
		"SELECT id, epic_id, rabbit_hole_id, mission_id, title, description, type, status, priority, assigned_grove_id, context_ref, pinned, created_at, updated_at, claimed_at, completed_at FROM tasks WHERE id = ?",
		id,
	).Scan(&task.ID, &task.EpicID, &task.RabbitHoleID, &task.MissionID, &task.Title, &task.Description, &task.Type, &task.Status, &task.Priority, &task.AssignedGroveID, &task.ContextRef, &task.Pinned, &task.CreatedAt, &task.UpdatedAt, &task.ClaimedAt, &task.CompletedAt)

	if err != nil {
		return nil, err
	}

	return task, nil
}

// ListTasks retrieves tasks, optionally filtered by epic/rabbit hole/status
func ListTasks(epicID, rabbitHoleID, status string) ([]*Task, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, epic_id, rabbit_hole_id, mission_id, title, description, type, status, priority, assigned_grove_id, context_ref, pinned, created_at, updated_at, claimed_at, completed_at FROM tasks WHERE 1=1"
	args := []interface{}{}

	if epicID != "" {
		query += " AND epic_id = ?"
		args = append(args, epicID)
	}

	if rabbitHoleID != "" {
		query += " AND rabbit_hole_id = ?"
		args = append(args, rabbitHoleID)
	}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at ASC"

	rows, err := database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		err := rows.Scan(&task.ID, &task.EpicID, &task.RabbitHoleID, &task.MissionID, &task.Title, &task.Description, &task.Type, &task.Status, &task.Priority, &task.AssignedGroveID, &task.ContextRef, &task.Pinned, &task.CreatedAt, &task.UpdatedAt, &task.ClaimedAt, &task.CompletedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// ClaimTask claims a task (marks as "implement")
func ClaimTask(id, groveID string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var groveIDNullable sql.NullString
	if groveID != "" {
		groveIDNullable = sql.NullString{String: groveID, Valid: true}
	}

	_, err = database.Exec(
		"UPDATE tasks SET status = 'implement', assigned_grove_id = ?, claimed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		groveIDNullable, id,
	)

	return err
}

// CompleteTask marks a task as complete
func CompleteTask(id string) error {
	// First, get task to check if pinned
	task, err := GetTask(id)
	if err != nil {
		return err
	}

	// Prevent completing pinned task
	if task.Pinned {
		return fmt.Errorf("Cannot complete pinned task %s. Unpin first with: orc task unpin %s", id, id)
	}

	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(
		"UPDATE tasks SET status = 'complete', completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UpdateTask updates the title and/or description of a task
func UpdateTask(id, title, description string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify task exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM tasks WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("task %s not found", id)
	}

	// Build update query based on what's being updated
	if title != "" && description != "" {
		_, err = database.Exec(
			"UPDATE tasks SET title = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, description, id,
		)
	} else if title != "" {
		_, err = database.Exec(
			"UPDATE tasks SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, id,
		)
	} else if description != "" {
		_, err = database.Exec(
			"UPDATE tasks SET description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			description, id,
		)
	}

	return err
}

// PinTask pins a task to keep it visible at the top
func PinTask(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify task exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM tasks WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("task %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE tasks SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UnpinTask unpins a task
func UnpinTask(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify task exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM tasks WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("task %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE tasks SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// GetTasksByGrove retrieves all tasks assigned to a grove
func GetTasksByGrove(groveID string) ([]*Task, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, epic_id, rabbit_hole_id, mission_id, title, description, type, status, priority, assigned_grove_id, context_ref, pinned, created_at, updated_at, claimed_at, completed_at FROM tasks WHERE assigned_grove_id = ?"
	rows, err := database.Query(query, groveID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		err := rows.Scan(&task.ID, &task.EpicID, &task.RabbitHoleID, &task.MissionID, &task.Title, &task.Description, &task.Type, &task.Status, &task.Priority, &task.AssignedGroveID, &task.ContextRef, &task.Pinned, &task.CreatedAt, &task.UpdatedAt, &task.ClaimedAt, &task.CompletedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTaskTag retrieves the tag for a task (returns nil if no tag assigned)
func GetTaskTag(taskID string) (*Tag, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	var tagID string
	err = database.QueryRow(
		"SELECT tag_id FROM task_tags WHERE task_id = ?",
		taskID,
	).Scan(&tagID)

	if err == sql.ErrNoRows {
		return nil, nil // No tag assigned
	}
	if err != nil {
		return nil, err
	}

	return GetTag(tagID)
}

// AddTagToTask assigns a tag to a task (enforces one-tag constraint)
func AddTagToTask(taskID, tagName string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify task exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("task %s not found", taskID)
	}

	// Get tag by name
	tag, err := GetTagByName(tagName)
	if err != nil {
		return fmt.Errorf("tag '%s' not found", tagName)
	}

	// Check if task already has a tag
	existingTag, err := GetTaskTag(taskID)
	if err != nil {
		return err
	}
	if existingTag != nil {
		return fmt.Errorf("task %s already has tag '%s' (one tag per task limit)\nRemove existing tag first with: orc task untag %s", taskID, existingTag.Name, taskID)
	}

	// Generate task_tag ID by finding max existing ID
	var maxID int
	err = database.QueryRow("SELECT COALESCE(MAX(CAST(SUBSTR(id, 4) AS INTEGER)), 0) FROM task_tags").Scan(&maxID)
	if err != nil {
		return err
	}
	id := fmt.Sprintf("TT-%03d", maxID+1)

	// Create task-tag association
	_, err = database.Exec(
		"INSERT INTO task_tags (id, task_id, tag_id) VALUES (?, ?, ?)",
		id, taskID, tag.ID,
	)

	return err
}

// RemoveTagFromTask removes the tag from a task
func RemoveTagFromTask(taskID string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify task exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("task %s not found", taskID)
	}

	// Check if task has a tag
	tag, err := GetTaskTag(taskID)
	if err != nil {
		return err
	}
	if tag == nil {
		return fmt.Errorf("task %s has no tag assigned", taskID)
	}

	_, err = database.Exec("DELETE FROM task_tags WHERE task_id = ?", taskID)
	return err
}

// ListTasksByTag retrieves all tasks with a specific tag
func ListTasksByTag(tagName string) ([]*Task, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	// Get tag by name
	tag, err := GetTagByName(tagName)
	if err != nil {
		return nil, fmt.Errorf("tag '%s' not found", tagName)
	}

	query := `
		SELECT t.id, t.epic_id, t.rabbit_hole_id, t.mission_id, t.title, t.description,
		       t.type, t.status, t.priority, t.assigned_grove_id, t.context_ref,
		       t.pinned, t.created_at, t.updated_at, t.claimed_at, t.completed_at
		FROM tasks t
		INNER JOIN task_tags tt ON t.id = tt.task_id
		WHERE tt.tag_id = ?
		ORDER BY t.created_at ASC
	`

	rows, err := database.Query(query, tag.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		err := rows.Scan(&task.ID, &task.EpicID, &task.RabbitHoleID, &task.MissionID, &task.Title, &task.Description, &task.Type, &task.Status, &task.Priority, &task.AssignedGroveID, &task.ContextRef, &task.Pinned, &task.CreatedAt, &task.UpdatedAt, &task.ClaimedAt, &task.CompletedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
