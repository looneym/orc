// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// CycleRepository implements secondary.CycleRepository with SQLite.
type CycleRepository struct {
	db *sql.DB
}

// NewCycleRepository creates a new SQLite cycle repository.
func NewCycleRepository(db *sql.DB) *CycleRepository {
	return &CycleRepository{db: db}
}

// Create persists a new cycle.
func (r *CycleRepository) Create(ctx context.Context, cycle *secondary.CycleRecord) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)`,
		cycle.ID,
		cycle.ShipmentID,
		cycle.SequenceNumber,
		cycle.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create cycle: %w", err)
	}

	return nil
}

// GetByID retrieves a cycle by its ID.
func (r *CycleRepository) GetByID(ctx context.Context, id string) (*secondary.CycleRecord, error) {
	var (
		createdAt   time.Time
		updatedAt   time.Time
		startedAt   sql.NullTime
		completedAt sql.NullTime
	)

	record := &secondary.CycleRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, sequence_number, status, created_at, updated_at, started_at, completed_at FROM cycles WHERE id = ?`,
		id,
	).Scan(&record.ID,
		&record.ShipmentID,
		&record.SequenceNumber,
		&record.Status,
		&createdAt, &updatedAt, &startedAt, &completedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cycle %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cycle: %w", err)
	}
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if startedAt.Valid {
		record.StartedAt = startedAt.Time.Format(time.RFC3339)
	}
	if completedAt.Valid {
		record.CompletedAt = completedAt.Time.Format(time.RFC3339)
	}

	return record, nil
}

// List retrieves cycles matching the given filters.
func (r *CycleRepository) List(ctx context.Context, filters secondary.CycleFilters) ([]*secondary.CycleRecord, error) {
	query := `SELECT id, shipment_id, sequence_number, status, created_at, updated_at, started_at, completed_at FROM cycles WHERE 1=1`
	args := []any{}

	if filters.ShipmentID != "" {
		query += " AND shipment_id = ?"
		args = append(args, filters.ShipmentID)
	}

	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}

	query += " ORDER BY sequence_number ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list cycles: %w", err)
	}
	defer rows.Close()

	var cycles []*secondary.CycleRecord
	for rows.Next() {
		var (
			createdAt   time.Time
			updatedAt   time.Time
			startedAt   sql.NullTime
			completedAt sql.NullTime
		)

		record := &secondary.CycleRecord{}
		err := rows.Scan(&record.ID,
			&record.ShipmentID,
			&record.SequenceNumber,
			&record.Status,
			&createdAt, &updatedAt, &startedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cycle: %w", err)
		}
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if startedAt.Valid {
			record.StartedAt = startedAt.Time.Format(time.RFC3339)
		}
		if completedAt.Valid {
			record.CompletedAt = completedAt.Time.Format(time.RFC3339)
		}

		cycles = append(cycles, record)
	}

	return cycles, nil
}

// Update updates an existing cycle.
func (r *CycleRepository) Update(ctx context.Context, cycle *secondary.CycleRecord) error {
	query := "UPDATE cycles SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}
	if cycle.SequenceNumber != 0 {
		query += ", sequence_number = ?"
		args = append(args, cycle.SequenceNumber)
	}

	query += " WHERE id = ?"
	args = append(args, cycle.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update cycle: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("cycle %s not found", cycle.ID)
	}

	return nil
}

// Delete removes a cycle from persistence.
func (r *CycleRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM cycles WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete cycle: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("cycle %s not found", id)
	}

	return nil
}

// GetNextID returns the next available cycle ID.
func (r *CycleRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	prefixLen := len("CYC-") + 1
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COALESCE(MAX(CAST(SUBSTR(id, %d) AS INTEGER)), 0) FROM cycles", prefixLen),
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next cycle ID: %w", err)
	}

	return fmt.Sprintf("CYC-%03d", maxID+1), nil
}

// ShipmentExists checks if a shipment exists.
func (r *CycleRepository) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM shipments WHERE id = ?", shipmentID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check shipment existence: %w", err)
	}
	return count > 0, nil
}

// GetNextSequenceNumber returns the next sequence number for a shipment.
func (r *CycleRepository) GetNextSequenceNumber(ctx context.Context, shipmentID string) (int64, error) {
	var maxSeq sql.NullInt64
	err := r.db.QueryRowContext(ctx,
		"SELECT MAX(sequence_number) FROM cycles WHERE shipment_id = ?",
		shipmentID,
	).Scan(&maxSeq)
	if err != nil {
		return 0, fmt.Errorf("failed to get next sequence number: %w", err)
	}

	if maxSeq.Valid {
		return maxSeq.Int64 + 1, nil
	}
	return 1, nil
}

// GetActiveCycle returns the active cycle for a shipment (if any).
func (r *CycleRepository) GetActiveCycle(ctx context.Context, shipmentID string) (*secondary.CycleRecord, error) {
	var (
		createdAt   time.Time
		updatedAt   time.Time
		startedAt   sql.NullTime
		completedAt sql.NullTime
	)

	record := &secondary.CycleRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, sequence_number, status, created_at, updated_at, started_at, completed_at
		 FROM cycles WHERE shipment_id = ? AND status = 'active'`,
		shipmentID,
	).Scan(&record.ID,
		&record.ShipmentID,
		&record.SequenceNumber,
		&record.Status,
		&createdAt, &updatedAt, &startedAt, &completedAt)

	if err == sql.ErrNoRows {
		return nil, nil // No active cycle, not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active cycle: %w", err)
	}
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if startedAt.Valid {
		record.StartedAt = startedAt.Time.Format(time.RFC3339)
	}
	if completedAt.Valid {
		record.CompletedAt = completedAt.Time.Format(time.RFC3339)
	}

	return record, nil
}

// GetByShipmentAndSequence returns a specific cycle by shipment and sequence number.
func (r *CycleRepository) GetByShipmentAndSequence(ctx context.Context, shipmentID string, seq int64) (*secondary.CycleRecord, error) {
	var (
		createdAt   time.Time
		updatedAt   time.Time
		startedAt   sql.NullTime
		completedAt sql.NullTime
	)

	record := &secondary.CycleRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, sequence_number, status, created_at, updated_at, started_at, completed_at
		 FROM cycles WHERE shipment_id = ? AND sequence_number = ?`,
		shipmentID, seq,
	).Scan(&record.ID,
		&record.ShipmentID,
		&record.SequenceNumber,
		&record.Status,
		&createdAt, &updatedAt, &startedAt, &completedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cycle not found for shipment %s sequence %d", shipmentID, seq)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cycle: %w", err)
	}
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if startedAt.Valid {
		record.StartedAt = startedAt.Time.Format(time.RFC3339)
	}
	if completedAt.Valid {
		record.CompletedAt = completedAt.Time.Format(time.RFC3339)
	}

	return record, nil
}

// UpdateStatus updates cycle status and optional timestamps.
func (r *CycleRepository) UpdateStatus(ctx context.Context, id, status string, setStarted, setCompleted bool) error {
	query := "UPDATE cycles SET status = ?, updated_at = CURRENT_TIMESTAMP"
	args := []any{status}

	if setStarted {
		query += ", started_at = CURRENT_TIMESTAMP"
	}
	if setCompleted {
		query += ", completed_at = CURRENT_TIMESTAMP"
	}

	query += " WHERE id = ?"
	args = append(args, id)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update cycle status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("cycle %s not found", id)
	}

	return nil
}

// Ensure CycleRepository implements the interface
var _ secondary.CycleRepository = (*CycleRepository)(nil)
