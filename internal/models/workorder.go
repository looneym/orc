package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/db"
)

type WorkOrder struct {
	ID              string
	MissionID       string
	Title           string
	Description     sql.NullString
	Type            sql.NullString
	Status          string
	Priority        sql.NullString
	ParentID        sql.NullString
	AssignedGroveID sql.NullString
	ContextRef      sql.NullString
	Pinned          bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ClaimedAt       sql.NullTime
	CompletedAt     sql.NullTime
}

// CreateWorkOrder creates a new work order
func CreateWorkOrder(missionID, title, description, contextRef, parentID string) (*WorkOrder, error) {
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

	// Verify parent exists if specified
	if parentID != "" {
		err = database.QueryRow("SELECT COUNT(*) FROM work_orders WHERE id = ?", parentID).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if exists == 0 {
			return nil, fmt.Errorf("parent work order %s not found", parentID)
		}

		// Enforce flat hierarchy: reject if parent itself has a parent (would create 3 levels)
		var parentOfParent sql.NullString
		err = database.QueryRow("SELECT parent_id FROM work_orders WHERE id = ?", parentID).Scan(&parentOfParent)
		if err != nil {
			return nil, err
		}
		if parentOfParent.Valid {
			return nil, fmt.Errorf("cannot create nested epics\nChildren cannot have children (2 level max: epic → children)")
		}
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

	var parent sql.NullString
	if parentID != "" {
		parent = sql.NullString{String: parentID, Valid: true}
	}

	_, err = database.Exec(
		"INSERT INTO work_orders (id, mission_id, title, description, context_ref, parent_id, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id, missionID, title, desc, ctxRef, parent, "ready",
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
		"SELECT id, mission_id, title, description, type, status, priority, parent_id, assigned_grove_id, context_ref, pinned, created_at, updated_at, claimed_at, completed_at FROM work_orders WHERE id = ?",
		id,
	).Scan(&wo.ID, &wo.MissionID, &wo.Title, &wo.Description, &wo.Type, &wo.Status, &wo.Priority, &wo.ParentID, &wo.AssignedGroveID, &wo.ContextRef, &wo.Pinned, &wo.CreatedAt, &wo.UpdatedAt, &wo.ClaimedAt, &wo.CompletedAt)

	if err != nil {
		return nil, err
	}

	return wo, nil
}

// ListWorkOrders retrieves work orders, optionally filtered by mission and/or status
func ListWorkOrders(missionID, status string) ([]*WorkOrder, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, mission_id, title, description, type, status, priority, parent_id, assigned_grove_id, context_ref, pinned, created_at, updated_at, claimed_at, completed_at FROM work_orders WHERE 1=1"
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

	var orders []*WorkOrder
	for rows.Next() {
		wo := &WorkOrder{}
		err := rows.Scan(&wo.ID, &wo.MissionID, &wo.Title, &wo.Description, &wo.Type, &wo.Status, &wo.Priority, &wo.ParentID, &wo.AssignedGroveID, &wo.ContextRef, &wo.Pinned, &wo.CreatedAt, &wo.UpdatedAt, &wo.ClaimedAt, &wo.CompletedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, wo)
	}

	return orders, nil
}

// ClaimWorkOrder assigns a work order to a grove and marks it as implement
func ClaimWorkOrder(id, groveID string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var groveIDNullable sql.NullString
	if groveID != "" {
		groveIDNullable = sql.NullString{String: groveID, Valid: true}
	}

	_, err = database.Exec(
		"UPDATE work_orders SET status = 'implement', assigned_grove_id = ?, claimed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		groveIDNullable, id,
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

// UpdateWorkOrder updates the title and/or description of a work order
func UpdateWorkOrder(id, title, description string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify work order exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM work_orders WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("work order %s not found", id)
	}

	// Build update query based on what's being updated
	if title != "" && description != "" {
		_, err = database.Exec(
			"UPDATE work_orders SET title = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, description, id,
		)
	} else if title != "" {
		_, err = database.Exec(
			"UPDATE work_orders SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, id,
		)
	} else if description != "" {
		_, err = database.Exec(
			"UPDATE work_orders SET description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			description, id,
		)
	}

	return err
}

// SetParentWorkOrder sets or updates the parent of a work order
func SetParentWorkOrder(id, parentID string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify work order exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM work_orders WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("work order %s not found", id)
	}

	// Verify parent exists if specified
	if parentID != "" {
		err = database.QueryRow("SELECT COUNT(*) FROM work_orders WHERE id = ?", parentID).Scan(&exists)
		if err != nil {
			return err
		}
		if exists == 0 {
			return fmt.Errorf("parent work order %s not found", parentID)
		}

		// Prevent circular reference (work order can't be its own parent)
		if id == parentID {
			return fmt.Errorf("work order cannot be its own parent")
		}

		// Check if parent would create a cycle (parent's parent is this work order)
		var parentOfParent sql.NullString
		err = database.QueryRow("SELECT parent_id FROM work_orders WHERE id = ?", parentID).Scan(&parentOfParent)
		if err != nil {
			return err
		}
		if parentOfParent.Valid && parentOfParent.String == id {
			return fmt.Errorf("cannot create circular parent relationship")
		}
	}

	var parent sql.NullString
	if parentID != "" {
		parent = sql.NullString{String: parentID, Valid: true}
	}

	_, err = database.Exec(
		"UPDATE work_orders SET parent_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		parent, id,
	)

	return err
}

// GetChildWorkOrders retrieves all child work orders for a given parent
func GetChildWorkOrders(parentID string) ([]*WorkOrder, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, mission_id, title, description, type, status, priority, parent_id, assigned_grove_id, context_ref, pinned, created_at, updated_at, claimed_at, completed_at FROM work_orders WHERE parent_id = ? ORDER BY created_at ASC"

	rows, err := database.Query(query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*WorkOrder
	for rows.Next() {
		wo := &WorkOrder{}
		err := rows.Scan(&wo.ID, &wo.MissionID, &wo.Title, &wo.Description, &wo.Type, &wo.Status, &wo.Priority, &wo.ParentID, &wo.AssignedGroveID, &wo.ContextRef, &wo.Pinned, &wo.CreatedAt, &wo.UpdatedAt, &wo.ClaimedAt, &wo.CompletedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, wo)
	}

	return orders, nil
}

// PinWorkOrder pins a work order to keep it visible at the top
func PinWorkOrder(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify work order exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM work_orders WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("work order %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE work_orders SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UnpinWorkOrder unpins a work order
func UnpinWorkOrder(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify work order exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM work_orders WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("work order %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE work_orders SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// AssignWorkOrderToGrove assigns a work order to a grove
func AssignWorkOrderToGrove(workOrderID, groveID string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify work order exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM work_orders WHERE id = ?", workOrderID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("work order %s not found", workOrderID)
	}

	// Update work order: status = implement, assigned_grove_id = grove_id
	_, err = database.Exec(
		"UPDATE work_orders SET status = 'implement', assigned_grove_id = ?, claimed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		groveID, workOrderID,
	)

	return err
}

// IsEpic checks if a work order is an epic (has children)
func IsEpic(workOrderID string) (bool, error) {
	children, err := GetChildWorkOrders(workOrderID)
	if err != nil {
		return false, err
	}
	return len(children) > 0, nil
}

// GetWorkOrdersByGrove retrieves all work orders assigned to a grove
func GetWorkOrdersByGrove(groveID string) ([]*WorkOrder, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, mission_id, title, description, type, status, priority, parent_id, assigned_grove_id, context_ref, pinned, created_at, updated_at, claimed_at, completed_at FROM work_orders WHERE assigned_grove_id = ?"
	rows, err := database.Query(query, groveID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*WorkOrder
	for rows.Next() {
		wo := &WorkOrder{}
		err := rows.Scan(&wo.ID, &wo.MissionID, &wo.Title, &wo.Description, &wo.Type, &wo.Status, &wo.Priority, &wo.ParentID, &wo.AssignedGroveID, &wo.ContextRef, &wo.Pinned, &wo.CreatedAt, &wo.UpdatedAt, &wo.ClaimedAt, &wo.CompletedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, wo)
	}

	return orders, nil
}

// ValidateGroveAvailability checks if a grove is available for epic assignment
func ValidateGroveAvailability(groveID string) error {
	// Get all work orders assigned to this grove
	wos, err := GetWorkOrdersByGrove(groveID)
	if err != nil {
		return err
	}

	if len(wos) == 0 {
		return nil // Grove available
	}

	// Check if all existing assignments share same parent (epic)
	var firstParent sql.NullString
	firstParent = wos[0].ParentID

	for _, wo := range wos {
		// If work order has no parent, it's an epic itself
		if !wo.ParentID.Valid {
			return fmt.Errorf("grove already assigned to epic %s", wo.ID)
		}

		// Check if all children belong to same parent
		if wo.ParentID != firstParent {
			return fmt.Errorf("grove already assigned to different epic")
		}
	}

	// If we get here, all work orders belong to same epic
	// Check if that epic is the one we're trying to assign
	return nil
}

// AssignEpicToGrove assigns entire epic (parent + all children) to grove
func AssignEpicToGrove(epicID, groveID string) error {
	// 1. Get the epic work order
	epic, err := GetWorkOrder(epicID)
	if err != nil {
		return fmt.Errorf("epic not found: %w", err)
	}

	// 2. Verify it's a top-level work order (no parent)
	if epic.ParentID.Valid {
		parent, _ := GetWorkOrder(epic.ParentID.String)
		return fmt.Errorf("%s is a child work order\nAssign the parent epic instead: %s", epicID, parent.ID)
	}

	// 3. Verify epic is in ready status
	if epic.Status != "ready" {
		return fmt.Errorf("epic must be in 'ready' status (current: %s)", epic.Status)
	}

	// 4. Verify grove is available (no other epics assigned)
	err = ValidateGroveAvailability(groveID)
	if err != nil {
		return err
	}

	// 5. Get all children
	children, err := GetChildWorkOrders(epicID)
	if err != nil {
		return fmt.Errorf("failed to get child work orders: %w", err)
	}

	// 6. Validate flat hierarchy (no nested epics)
	for _, child := range children {
		grandchildren, err := GetChildWorkOrders(child.ID)
		if err != nil {
			return fmt.Errorf("failed to validate hierarchy: %w", err)
		}
		if len(grandchildren) > 0 {
			return fmt.Errorf("epic has nested children - not supported\nFlatten hierarchy to 2 levels before assigning")
		}
	}

	// 7. Assign parent (epic)
	err = AssignWorkOrderToGrove(epicID, groveID)
	if err != nil {
		return fmt.Errorf("failed to assign epic: %w", err)
	}

	// 8. Assign all children
	for _, child := range children {
		err = AssignWorkOrderToGrove(child.ID, groveID)
		if err != nil {
			return fmt.Errorf("failed to assign child %s: %w", child.ID, err)
		}
	}

	return nil
}
