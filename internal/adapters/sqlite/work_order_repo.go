// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// WorkOrderRepository implements secondary.WorkOrderRepository with SQLite.
type WorkOrderRepository struct {
	db *sql.DB
}

// NewWorkOrderRepository creates a new SQLite workOrder repository.
func NewWorkOrderRepository(db *sql.DB) *WorkOrderRepository {
	return &WorkOrderRepository{db: db}
}

// Create persists a new workOrder.
func (r *WorkOrderRepository) Create(ctx context.Context, workOrder *secondary.WorkOrderRecord) error {
	var acceptanceCriteria sql.NullString
	if workOrder.AcceptanceCriteria != "" {
		acceptanceCriteria = sql.NullString{String: workOrder.AcceptanceCriteria, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO work_orders (id, shipment_id, outcome, acceptance_criteria, status) VALUES (?, ?, ?, ?, ?)`,
		workOrder.ID,
		workOrder.ShipmentID,
		workOrder.Outcome,
		acceptanceCriteria,
		workOrder.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create workOrder: %w", err)
	}

	return nil
}

// GetByID retrieves a workOrder by its ID.
func (r *WorkOrderRepository) GetByID(ctx context.Context, id string) (*secondary.WorkOrderRecord, error) {
	var (
		acceptanceCriteria sql.NullString
		createdAt          time.Time
		updatedAt          time.Time
	)

	record := &secondary.WorkOrderRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, outcome, acceptance_criteria, status, created_at, updated_at FROM work_orders WHERE id = ?`,
		id,
	).Scan(&record.ID,
		&record.ShipmentID,
		&record.Outcome,
		&acceptanceCriteria,
		&record.Status,
		&createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("work order %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get work order: %w", err)
	}
	record.AcceptanceCriteria = acceptanceCriteria.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// GetByShipment retrieves a work order by its shipment ID.
func (r *WorkOrderRepository) GetByShipment(ctx context.Context, shipmentID string) (*secondary.WorkOrderRecord, error) {
	var (
		acceptanceCriteria sql.NullString
		createdAt          time.Time
		updatedAt          time.Time
	)

	record := &secondary.WorkOrderRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, outcome, acceptance_criteria, status, created_at, updated_at FROM work_orders WHERE shipment_id = ?`,
		shipmentID,
	).Scan(&record.ID,
		&record.ShipmentID,
		&record.Outcome,
		&acceptanceCriteria,
		&record.Status,
		&createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("work order for shipment %s not found", shipmentID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get work order: %w", err)
	}
	record.AcceptanceCriteria = acceptanceCriteria.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// List retrieves work_orders matching the given filters.
func (r *WorkOrderRepository) List(ctx context.Context, filters secondary.WorkOrderFilters) ([]*secondary.WorkOrderRecord, error) {
	query := `SELECT id, shipment_id, outcome, acceptance_criteria, status, created_at, updated_at FROM work_orders WHERE 1=1`
	args := []any{}

	if filters.ShipmentID != "" {
		query += " AND shipment_id = ?"
		args = append(args, filters.ShipmentID)
	}

	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list work_orders: %w", err)
	}
	defer rows.Close()

	var workOrders []*secondary.WorkOrderRecord
	for rows.Next() {
		var (
			acceptanceCriteria sql.NullString
			createdAt          time.Time
			updatedAt          time.Time
		)

		record := &secondary.WorkOrderRecord{}
		err := rows.Scan(&record.ID,
			&record.ShipmentID,
			&record.Outcome,
			&acceptanceCriteria,
			&record.Status,
			&createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workOrder: %w", err)
		}
		record.AcceptanceCriteria = acceptanceCriteria.String
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)

		workOrders = append(workOrders, record)
	}

	return workOrders, nil
}

// Update updates an existing workOrder.
func (r *WorkOrderRepository) Update(ctx context.Context, workOrder *secondary.WorkOrderRecord) error {
	query := "UPDATE work_orders SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}
	if workOrder.Outcome != "" {
		query += ", outcome = ?"
		args = append(args, workOrder.Outcome)
	}
	if workOrder.AcceptanceCriteria != "" {
		query += ", acceptance_criteria = ?"
		args = append(args, sql.NullString{String: workOrder.AcceptanceCriteria, Valid: true})
	}

	query += " WHERE id = ?"
	args = append(args, workOrder.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update workOrder: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("workOrder %s not found", workOrder.ID)
	}

	return nil
}

// Delete removes a workOrder from persistence.
func (r *WorkOrderRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM work_orders WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete workOrder: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("workOrder %s not found", id)
	}

	return nil
}

// GetNextID returns the next available workOrder ID.
func (r *WorkOrderRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	prefixLen := len("WO-") + 1
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COALESCE(MAX(CAST(SUBSTR(id, %d) AS INTEGER)), 0) FROM work_orders", prefixLen),
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next workOrder ID: %w", err)
	}

	return fmt.Sprintf("WO-%03d", maxID+1), nil
}

// ShipmentExists checks if a shipment exists.
func (r *WorkOrderRepository) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM shipments WHERE id = ?", shipmentID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check shipment existence: %w", err)
	}
	return count > 0, nil
}

// ShipmentHasWorkOrder checks if a shipment already has a workOrder (for 1:1 relationships).
func (r *WorkOrderRepository) ShipmentHasWorkOrder(ctx context.Context, shipmentID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM work_orders WHERE shipment_id = ?", shipmentID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check existing workOrder: %w", err)
	}
	return count > 0, nil
}

// UpdateStatus updates the status of a work order.
func (r *WorkOrderRepository) UpdateStatus(ctx context.Context, id, status string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE work_orders SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update work order status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("work order %s not found", id)
	}

	return nil
}

// Ensure WorkOrderRepository implements the interface
var _ secondary.WorkOrderRepository = (*WorkOrderRepository)(nil)
