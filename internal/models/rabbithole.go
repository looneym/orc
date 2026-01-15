package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/db"
)

type RabbitHole struct {
	ID          string
	EpicID      string
	Title       string
	Description sql.NullString
	Status      string
	Priority    sql.NullString
	Pinned      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt sql.NullTime
}

// CreateRabbitHole creates a new rabbit hole under an epic
func CreateRabbitHole(epicID, title, description string) (*RabbitHole, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	// Verify epic exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM epics WHERE id = ?", epicID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, fmt.Errorf("epic %s not found", epicID)
	}

	// Verify epic doesn't already have direct tasks (no mixed children)
	var taskCount int
	err = database.QueryRow("SELECT COUNT(*) FROM tasks WHERE epic_id = ? AND rabbit_hole_id IS NULL", epicID).Scan(&taskCount)
	if err != nil {
		return nil, err
	}
	if taskCount > 0 {
		return nil, fmt.Errorf("epic %s already has direct tasks\nCannot add rabbit holes to epic with direct tasks (no mixed children)", epicID)
	}

	// Generate rabbit hole ID by finding max existing ID
	var maxID int
	err = database.QueryRow("SELECT COALESCE(MAX(CAST(SUBSTR(id, 4) AS INTEGER)), 0) FROM rabbit_holes").Scan(&maxID)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("RH-%03d", maxID+1)

	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	_, err = database.Exec(
		"INSERT INTO rabbit_holes (id, epic_id, title, description, status) VALUES (?, ?, ?, ?, ?)",
		id, epicID, title, desc, "ready",
	)
	if err != nil {
		return nil, err
	}

	return GetRabbitHole(id)
}

// GetRabbitHole retrieves a rabbit hole by ID
func GetRabbitHole(id string) (*RabbitHole, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	rh := &RabbitHole{}
	err = database.QueryRow(
		"SELECT id, epic_id, title, description, status, priority, pinned, created_at, updated_at, completed_at FROM rabbit_holes WHERE id = ?",
		id,
	).Scan(&rh.ID, &rh.EpicID, &rh.Title, &rh.Description, &rh.Status, &rh.Priority, &rh.Pinned, &rh.CreatedAt, &rh.UpdatedAt, &rh.CompletedAt)

	if err != nil {
		return nil, err
	}

	return rh, nil
}

// ListRabbitHoles retrieves rabbit holes, optionally filtered by epic and/or status
func ListRabbitHoles(epicID, status string) ([]*RabbitHole, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, epic_id, title, description, status, priority, pinned, created_at, updated_at, completed_at FROM rabbit_holes WHERE 1=1"
	args := []interface{}{}

	if epicID != "" {
		query += " AND epic_id = ?"
		args = append(args, epicID)
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

	var rabbitHoles []*RabbitHole
	for rows.Next() {
		rh := &RabbitHole{}
		err := rows.Scan(&rh.ID, &rh.EpicID, &rh.Title, &rh.Description, &rh.Status, &rh.Priority, &rh.Pinned, &rh.CreatedAt, &rh.UpdatedAt, &rh.CompletedAt)
		if err != nil {
			return nil, err
		}
		rabbitHoles = append(rabbitHoles, rh)
	}

	return rabbitHoles, nil
}

// CompleteRabbitHole marks a rabbit hole as complete
func CompleteRabbitHole(id string) error {
	// First, get rabbit hole to check if pinned
	rh, err := GetRabbitHole(id)
	if err != nil {
		return err
	}

	// Prevent completing pinned rabbit hole
	if rh.Pinned {
		return fmt.Errorf("Cannot complete pinned rabbit hole %s. Unpin first with: orc rabbit-hole unpin %s", id, id)
	}

	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(
		"UPDATE rabbit_holes SET status = 'complete', completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UpdateRabbitHole updates the title and/or description of a rabbit hole
func UpdateRabbitHole(id, title, description string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify rabbit hole exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM rabbit_holes WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("rabbit hole %s not found", id)
	}

	// Build update query based on what's being updated
	if title != "" && description != "" {
		_, err = database.Exec(
			"UPDATE rabbit_holes SET title = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, description, id,
		)
	} else if title != "" {
		_, err = database.Exec(
			"UPDATE rabbit_holes SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, id,
		)
	} else if description != "" {
		_, err = database.Exec(
			"UPDATE rabbit_holes SET description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			description, id,
		)
	}

	return err
}

// PinRabbitHole pins a rabbit hole to keep it visible
func PinRabbitHole(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify rabbit hole exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM rabbit_holes WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("rabbit hole %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE rabbit_holes SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UnpinRabbitHole unpins a rabbit hole
func UnpinRabbitHole(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify rabbit hole exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM rabbit_holes WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("rabbit hole %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE rabbit_holes SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// GetRabbitHoleTasks retrieves all tasks under a rabbit hole
func GetRabbitHoleTasks(rabbitHoleID string) ([]*Task, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, epic_id, rabbit_hole_id, mission_id, title, description, type, status, priority, assigned_grove_id, context_ref, pinned, created_at, updated_at, claimed_at, completed_at FROM tasks WHERE rabbit_hole_id = ? ORDER BY created_at ASC"
	rows, err := database.Query(query, rabbitHoleID)
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
