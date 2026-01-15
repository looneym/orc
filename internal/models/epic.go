package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/db"
)

type Epic struct {
	ID              string
	MissionID       string
	Title           string
	Description     sql.NullString
	Status          string
	Priority        sql.NullString
	AssignedGroveID sql.NullString
	ContextRef      sql.NullString
	Pinned          bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CompletedAt     sql.NullTime
}

// CreateEpic creates a new epic
func CreateEpic(missionID, title, description, contextRef string) (*Epic, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
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

	// Generate epic ID by finding max existing ID
	var maxID int
	err = database.QueryRow("SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM epics").Scan(&maxID)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("EPIC-%03d", maxID+1)

	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	var ctxRef sql.NullString
	if contextRef != "" {
		ctxRef = sql.NullString{String: contextRef, Valid: true}
	}

	_, err = database.Exec(
		"INSERT INTO epics (id, mission_id, title, description, context_ref, status) VALUES (?, ?, ?, ?, ?, ?)",
		id, missionID, title, desc, ctxRef, "ready",
	)
	if err != nil {
		return nil, err
	}

	return GetEpic(id)
}

// GetEpic retrieves an epic by ID
func GetEpic(id string) (*Epic, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	epic := &Epic{}
	err = database.QueryRow(
		"SELECT id, mission_id, title, description, status, priority, assigned_grove_id, context_ref, pinned, created_at, updated_at, completed_at FROM epics WHERE id = ?",
		id,
	).Scan(&epic.ID, &epic.MissionID, &epic.Title, &epic.Description, &epic.Status, &epic.Priority, &epic.AssignedGroveID, &epic.ContextRef, &epic.Pinned, &epic.CreatedAt, &epic.UpdatedAt, &epic.CompletedAt)

	if err != nil {
		return nil, err
	}

	return epic, nil
}

// ListEpics retrieves epics, optionally filtered by mission and/or status
func ListEpics(missionID, status string) ([]*Epic, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, mission_id, title, description, status, priority, assigned_grove_id, context_ref, pinned, created_at, updated_at, completed_at FROM epics WHERE 1=1"
	args := []interface{}{}

	if missionID != "" {
		query += " AND mission_id = ?"
		args = append(args, missionID)
	}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var epics []*Epic
	for rows.Next() {
		epic := &Epic{}
		err := rows.Scan(&epic.ID, &epic.MissionID, &epic.Title, &epic.Description, &epic.Status, &epic.Priority, &epic.AssignedGroveID, &epic.ContextRef, &epic.Pinned, &epic.CreatedAt, &epic.UpdatedAt, &epic.CompletedAt)
		if err != nil {
			return nil, err
		}
		epics = append(epics, epic)
	}

	return epics, nil
}

// CompleteEpic marks an epic as complete
func CompleteEpic(id string) error {
	// First, get epic to check if pinned
	epic, err := GetEpic(id)
	if err != nil {
		return err
	}

	// Prevent completing pinned epic
	if epic.Pinned {
		return fmt.Errorf("Cannot complete pinned epic %s. Unpin first with: orc epic unpin %s", id, id)
	}

	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(
		"UPDATE epics SET status = 'complete', completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UpdateEpic updates the title and/or description of an epic
func UpdateEpic(id, title, description string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify epic exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM epics WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("epic %s not found", id)
	}

	// Build update query based on what's being updated
	if title != "" && description != "" {
		_, err = database.Exec(
			"UPDATE epics SET title = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, description, id,
		)
	} else if title != "" {
		_, err = database.Exec(
			"UPDATE epics SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, id,
		)
	} else if description != "" {
		_, err = database.Exec(
			"UPDATE epics SET description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			description, id,
		)
	}

	return err
}

// PinEpic pins an epic to keep it visible at the top
func PinEpic(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify epic exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM epics WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("epic %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE epics SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UnpinEpic unpins an epic
func UnpinEpic(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify epic exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM epics WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("epic %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE epics SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// AssignEpicToGrove assigns an epic to a grove (will be expanded to include children)
func AssignEpicToGrove(epicID, groveID string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify epic exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM epics WHERE id = ?", epicID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("epic %s not found", epicID)
	}

	// Get epic to check status
	epic, err := GetEpic(epicID)
	if err != nil {
		return err
	}

	// Verify epic is in ready status
	if epic.Status != "ready" {
		return fmt.Errorf("epic must be in 'ready' status (current: %s)", epic.Status)
	}

	// Verify grove is available (no other epic assigned)
	err = ValidateGroveAvailability(groveID)
	if err != nil {
		return err
	}

	// Assign epic
	_, err = database.Exec(
		"UPDATE epics SET status = 'implement', assigned_grove_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		groveID, epicID,
	)
	if err != nil {
		return err
	}

	// Also assign all child tasks to the same grove
	_, err = database.Exec(
		"UPDATE tasks SET assigned_grove_id = ?, updated_at = CURRENT_TIMESTAMP WHERE epic_id = ?",
		groveID, epicID,
	)
	if err != nil {
		return err
	}

	// Also assign all tasks under rabbit holes of this epic
	_, err = database.Exec(`
		UPDATE tasks SET assigned_grove_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE rabbit_hole_id IN (SELECT id FROM rabbit_holes WHERE epic_id = ?)
	`, groveID, epicID)

	return err
}

// ValidateGroveAvailability checks if a grove is available for epic assignment
func ValidateGroveAvailability(groveID string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Check if any epics are already assigned to this grove
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM epics WHERE assigned_grove_id = ?", groveID).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		// Get the assigned epic ID
		var assignedEpicID string
		err = database.QueryRow("SELECT id FROM epics WHERE assigned_grove_id = ? LIMIT 1", groveID).Scan(&assignedEpicID)
		if err != nil {
			return err
		}
		return fmt.Errorf("grove already assigned to epic %s\n1:1:1 relationship: one grove can only work on one epic", assignedEpicID)
	}

	return nil
}

// GetEpicsByGrove retrieves all epics assigned to a grove (should be 0 or 1)
func GetEpicsByGrove(groveID string) ([]*Epic, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, mission_id, title, description, status, priority, assigned_grove_id, context_ref, pinned, created_at, updated_at, completed_at FROM epics WHERE assigned_grove_id = ?"
	rows, err := database.Query(query, groveID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var epics []*Epic
	for rows.Next() {
		epic := &Epic{}
		err := rows.Scan(&epic.ID, &epic.MissionID, &epic.Title, &epic.Description, &epic.Status, &epic.Priority, &epic.AssignedGroveID, &epic.ContextRef, &epic.Pinned, &epic.CreatedAt, &epic.UpdatedAt, &epic.CompletedAt)
		if err != nil {
			return nil, err
		}
		epics = append(epics, epic)
	}

	return epics, nil
}

// HasRabbitHoles checks if an epic has rabbit holes (vs direct tasks)
func HasRabbitHoles(epicID string) (bool, error) {
	database, err := db.GetDB()
	if err != nil {
		return false, err
	}

	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM rabbit_holes WHERE epic_id = ?", epicID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetDirectTasks gets tasks that are direct children of an epic (not in rabbit holes)
func GetDirectTasks(epicID string) ([]*Task, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, epic_id, rabbit_hole_id, mission_id, title, description, type, status, priority, assigned_grove_id, context_ref, pinned, created_at, updated_at, claimed_at, completed_at FROM tasks WHERE epic_id = ? AND rabbit_hole_id IS NULL ORDER BY created_at ASC"
	rows, err := database.Query(query, epicID)
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
