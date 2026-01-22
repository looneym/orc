// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// CycleWorkOrderRepository implements secondary.CycleWorkOrderRepository with SQLite.
type CycleWorkOrderRepository struct {
	db *sql.DB
}

// NewCycleWorkOrderRepository creates a new SQLite cycle work order repository.
func NewCycleWorkOrderRepository(db *sql.DB) *CycleWorkOrderRepository {
	return &CycleWorkOrderRepository{db: db}
}

// Create persists a new cycle work order.
func (r *CycleWorkOrderRepository) Create(ctx context.Context, cwo *secondary.CycleWorkOrderRecord) error {
	var acceptanceCriteria sql.NullString
	if cwo.AcceptanceCriteria != "" {
		acceptanceCriteria = sql.NullString{String: cwo.AcceptanceCriteria, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO cycle_work_orders (id, cycle_id, shipment_id, outcome, acceptance_criteria, status) VALUES (?, ?, ?, ?, ?, ?)`,
		cwo.ID,
		cwo.CycleID,
		cwo.ShipmentID,
		cwo.Outcome,
		acceptanceCriteria,
		cwo.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create cycle work order: %w", err)
	}

	return nil
}

// GetByID retrieves a cycle work order by its ID.
func (r *CycleWorkOrderRepository) GetByID(ctx context.Context, id string) (*secondary.CycleWorkOrderRecord, error) {
	var (
		acceptanceCriteria sql.NullString
		createdAt          time.Time
		updatedAt          time.Time
	)

	record := &secondary.CycleWorkOrderRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, cycle_id, shipment_id, outcome, acceptance_criteria, status, created_at, updated_at FROM cycle_work_orders WHERE id = ?`,
		id,
	).Scan(&record.ID,
		&record.CycleID,
		&record.ShipmentID,
		&record.Outcome,
		&acceptanceCriteria,
		&record.Status,
		&createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cycle work order %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cycle work order: %w", err)
	}
	record.AcceptanceCriteria = acceptanceCriteria.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// GetByCycle retrieves a cycle work order by its cycle ID.
func (r *CycleWorkOrderRepository) GetByCycle(ctx context.Context, cycleID string) (*secondary.CycleWorkOrderRecord, error) {
	var (
		acceptanceCriteria sql.NullString
		createdAt          time.Time
		updatedAt          time.Time
	)

	record := &secondary.CycleWorkOrderRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, cycle_id, shipment_id, outcome, acceptance_criteria, status, created_at, updated_at FROM cycle_work_orders WHERE cycle_id = ?`,
		cycleID,
	).Scan(&record.ID,
		&record.CycleID,
		&record.ShipmentID,
		&record.Outcome,
		&acceptanceCriteria,
		&record.Status,
		&createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cycle work order for cycle %s not found", cycleID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cycle work order: %w", err)
	}
	record.AcceptanceCriteria = acceptanceCriteria.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// List retrieves cycle work orders matching the given filters.
func (r *CycleWorkOrderRepository) List(ctx context.Context, filters secondary.CycleWorkOrderFilters) ([]*secondary.CycleWorkOrderRecord, error) {
	query := `SELECT id, cycle_id, shipment_id, outcome, acceptance_criteria, status, created_at, updated_at FROM cycle_work_orders WHERE 1=1`
	args := []any{}

	if filters.CycleID != "" {
		query += " AND cycle_id = ?"
		args = append(args, filters.CycleID)
	}

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
		return nil, fmt.Errorf("failed to list cycle work orders: %w", err)
	}
	defer rows.Close()

	var cwos []*secondary.CycleWorkOrderRecord
	for rows.Next() {
		var (
			acceptanceCriteria sql.NullString
			createdAt          time.Time
			updatedAt          time.Time
		)

		record := &secondary.CycleWorkOrderRecord{}
		err := rows.Scan(&record.ID,
			&record.CycleID,
			&record.ShipmentID,
			&record.Outcome,
			&acceptanceCriteria,
			&record.Status,
			&createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cycle work order: %w", err)
		}
		record.AcceptanceCriteria = acceptanceCriteria.String
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)

		cwos = append(cwos, record)
	}

	return cwos, nil
}

// Update updates an existing cycle work order.
func (r *CycleWorkOrderRepository) Update(ctx context.Context, cwo *secondary.CycleWorkOrderRecord) error {
	query := "UPDATE cycle_work_orders SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}
	if cwo.Outcome != "" {
		query += ", outcome = ?"
		args = append(args, cwo.Outcome)
	}
	if cwo.AcceptanceCriteria != "" {
		query += ", acceptance_criteria = ?"
		args = append(args, sql.NullString{String: cwo.AcceptanceCriteria, Valid: true})
	}

	query += " WHERE id = ?"
	args = append(args, cwo.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update cycle work order: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("cycle work order %s not found", cwo.ID)
	}

	return nil
}

// Delete removes a cycle work order from persistence.
func (r *CycleWorkOrderRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM cycle_work_orders WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete cycle work order: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("cycle work order %s not found", id)
	}

	return nil
}

// GetNextID returns the next available cycle work order ID.
func (r *CycleWorkOrderRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	prefixLen := len("CWO-") + 1
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COALESCE(MAX(CAST(SUBSTR(id, %d) AS INTEGER)), 0) FROM cycle_work_orders", prefixLen),
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next cycle work order ID: %w", err)
	}

	return fmt.Sprintf("CWO-%03d", maxID+1), nil
}

// UpdateStatus updates the status of a cycle work order.
func (r *CycleWorkOrderRepository) UpdateStatus(ctx context.Context, id, status string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE cycle_work_orders SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update cycle work order status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("cycle work order %s not found", id)
	}

	return nil
}

// CycleExists checks if a cycle exists.
func (r *CycleWorkOrderRepository) CycleExists(ctx context.Context, cycleID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM cycles WHERE id = ?", cycleID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check cycle existence: %w", err)
	}
	return count > 0, nil
}

// ShipmentExists checks if a shipment exists.
func (r *CycleWorkOrderRepository) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM shipments WHERE id = ?", shipmentID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check shipment existence: %w", err)
	}
	return count > 0, nil
}

// CycleHasCWO checks if a cycle already has a CWO (for 1:1 constraint).
func (r *CycleWorkOrderRepository) CycleHasCWO(ctx context.Context, cycleID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM cycle_work_orders WHERE cycle_id = ?", cycleID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check existing cycle work order: %w", err)
	}
	return count > 0, nil
}

// GetCycleStatus retrieves the status of a cycle.
func (r *CycleWorkOrderRepository) GetCycleStatus(ctx context.Context, cycleID string) (string, error) {
	var status string
	err := r.db.QueryRowContext(ctx, "SELECT status FROM cycles WHERE id = ?", cycleID).Scan(&status)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("cycle %s not found", cycleID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get cycle status: %w", err)
	}
	return status, nil
}

// GetCycleShipmentID retrieves the shipment ID for a cycle.
func (r *CycleWorkOrderRepository) GetCycleShipmentID(ctx context.Context, cycleID string) (string, error) {
	var shipmentID string
	err := r.db.QueryRowContext(ctx, "SELECT shipment_id FROM cycles WHERE id = ?", cycleID).Scan(&shipmentID)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("cycle %s not found", cycleID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get cycle shipment ID: %w", err)
	}
	return shipmentID, nil
}

// Ensure CycleWorkOrderRepository implements the interface
var _ secondary.CycleWorkOrderRepository = (*CycleWorkOrderRepository)(nil)
