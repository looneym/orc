package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/db"
)

type Expedition struct {
	ID          string
	Name        string
	WorkOrderID sql.NullString
	AssignedIMP sql.NullString
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateExpedition creates a new expedition
func CreateExpedition(name string) (*Expedition, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	// Generate expedition ID by finding max existing ID
	var maxID int
	err = database.QueryRow("SELECT COALESCE(MAX(CAST(SUBSTR(id, 5) AS INTEGER)), 0) FROM expeditions").Scan(&maxID)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("EXP-%03d", maxID+1)

	_, err = database.Exec(
		"INSERT INTO expeditions (id, name, status) VALUES (?, ?, ?)",
		id, name, "planning",
	)
	if err != nil {
		return nil, err
	}

	return GetExpedition(id)
}

// GetExpedition retrieves an expedition by ID
func GetExpedition(id string) (*Expedition, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	exp := &Expedition{}
	err = database.QueryRow(
		"SELECT id, name, work_order_id, assigned_imp, status, created_at, updated_at FROM expeditions WHERE id = ?",
		id,
	).Scan(&exp.ID, &exp.Name, &exp.WorkOrderID, &exp.AssignedIMP, &exp.Status, &exp.CreatedAt, &exp.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return exp, nil
}

// ListExpeditions retrieves all expeditions
func ListExpeditions() ([]*Expedition, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := database.Query(
		"SELECT id, name, work_order_id, assigned_imp, status, created_at, updated_at FROM expeditions ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expeditions []*Expedition
	for rows.Next() {
		exp := &Expedition{}
		err := rows.Scan(&exp.ID, &exp.Name, &exp.WorkOrderID, &exp.AssignedIMP, &exp.Status, &exp.CreatedAt, &exp.UpdatedAt)
		if err != nil {
			return nil, err
		}
		expeditions = append(expeditions, exp)
	}

	return expeditions, nil
}

// UpdateExpeditionStatus updates the status of an expedition
func UpdateExpeditionStatus(id, status string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(
		"UPDATE expeditions SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, id,
	)
	return err
}

// AssignIMP assigns an IMP to an expedition
func AssignIMP(expeditionID, impName string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(
		"UPDATE expeditions SET assigned_imp = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		impName, expeditionID,
	)
	return err
}
