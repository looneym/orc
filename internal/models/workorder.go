package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/looneym/orc/internal/db"
)

type WorkOrder struct {
	ID          string
	OperationID string
	Title       string
	Description sql.NullString
	Status      string
	AssignedImp sql.NullString
	ContextRef  sql.NullString
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ClaimedAt   sql.NullTime
	CompletedAt sql.NullTime
}

// CreateWorkOrder creates a new work order
func CreateWorkOrder(operationID, title, description, contextRef string) (*WorkOrder, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	// Verify operation exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM operations WHERE id = ?", operationID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, fmt.Errorf("operation %s not found", operationID)
	}

	// Generate work order ID
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM work_orders").Scan(&count)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("WO-%03d", count+1)

	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	var ctxRef sql.NullString
	if contextRef != "" {
		ctxRef = sql.NullString{String: contextRef, Valid: true}
	}

	_, err = database.Exec(
		"INSERT INTO work_orders (id, operation_id, title, description, context_ref, status) VALUES (?, ?, ?, ?, ?, ?)",
		id, operationID, title, desc, ctxRef, "backlog",
	)
	if err != nil {
		return nil, err
	}

	return GetWorkOrder(id)
}

// GetWorkOrder retrieves a work order by ID
func GetWorkOrder(id string) (*WorkOrder, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	wo := &WorkOrder{}
	err = database.QueryRow(
		"SELECT id, operation_id, title, description, status, assigned_imp, context_ref, created_at, updated_at, claimed_at, completed_at FROM work_orders WHERE id = ?",
		id,
	).Scan(&wo.ID, &wo.OperationID, &wo.Title, &wo.Description, &wo.Status, &wo.AssignedImp, &wo.ContextRef, &wo.CreatedAt, &wo.UpdatedAt, &wo.ClaimedAt, &wo.CompletedAt)

	if err != nil {
		return nil, err
	}

	return wo, nil
}

// ListWorkOrders retrieves work orders, optionally filtered by operation and/or status
func ListWorkOrders(operationID, status string) ([]*WorkOrder, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, operation_id, title, description, status, assigned_imp, context_ref, created_at, updated_at, claimed_at, completed_at FROM work_orders WHERE 1=1"
	args := []interface{}{}

	if operationID != "" {
		query += " AND operation_id = ?"
		args = append(args, operationID)
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

	var orders []*WorkOrder
	for rows.Next() {
		wo := &WorkOrder{}
		err := rows.Scan(&wo.ID, &wo.OperationID, &wo.Title, &wo.Description, &wo.Status, &wo.AssignedImp, &wo.ContextRef, &wo.CreatedAt, &wo.UpdatedAt, &wo.ClaimedAt, &wo.CompletedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, wo)
	}

	return orders, nil
}

// ClaimWorkOrder assigns a work order to an IMP and marks it as in_progress
func ClaimWorkOrder(id, impName string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(
		"UPDATE work_orders SET status = 'in_progress', assigned_imp = ?, claimed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		impName, id,
	)

	return err
}

// CompleteWorkOrder marks a work order as complete
func CompleteWorkOrder(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(
		"UPDATE work_orders SET status = 'complete', completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}
