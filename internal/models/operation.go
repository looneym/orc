package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/looneym/orc/internal/db"
)

type Operation struct {
	ID          string
	MissionID   string
	Title       string
	Description sql.NullString
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt sql.NullTime
}

// CreateOperation creates a new operation
func CreateOperation(missionID, title, description string) (*Operation, error) {
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

	// Generate operation ID
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM operations").Scan(&count)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("OP-%03d", count+1)

	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	_, err = database.Exec(
		"INSERT INTO operations (id, mission_id, title, description, status) VALUES (?, ?, ?, ?, ?)",
		id, missionID, title, desc, "backlog",
	)
	if err != nil {
		return nil, err
	}

	return GetOperation(id)
}

// GetOperation retrieves an operation by ID
func GetOperation(id string) (*Operation, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	op := &Operation{}
	err = database.QueryRow(
		"SELECT id, mission_id, title, description, status, created_at, updated_at, completed_at FROM operations WHERE id = ?",
		id,
	).Scan(&op.ID, &op.MissionID, &op.Title, &op.Description, &op.Status, &op.CreatedAt, &op.UpdatedAt, &op.CompletedAt)

	if err != nil {
		return nil, err
	}

	return op, nil
}

// ListOperations retrieves operations, optionally filtered by mission and/or status
func ListOperations(missionID, status string) ([]*Operation, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, mission_id, title, description, status, created_at, updated_at, completed_at FROM operations WHERE 1=1"
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

	var operations []*Operation
	for rows.Next() {
		op := &Operation{}
		err := rows.Scan(&op.ID, &op.MissionID, &op.Title, &op.Description, &op.Status, &op.CreatedAt, &op.UpdatedAt, &op.CompletedAt)
		if err != nil {
			return nil, err
		}
		operations = append(operations, op)
	}

	return operations, nil
}

// UpdateOperationStatus updates the status of an operation
func UpdateOperationStatus(id, status string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var completedAt sql.NullTime
	if status == "complete" {
		completedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	_, err = database.Exec(
		"UPDATE operations SET status = ?, updated_at = CURRENT_TIMESTAMP, completed_at = ? WHERE id = ?",
		status, completedAt, id,
	)

	return err
}
