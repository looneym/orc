// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// ConclaveRepository implements secondary.ConclaveRepository with SQLite.
type ConclaveRepository struct {
	db *sql.DB
}

// NewConclaveRepository creates a new SQLite conclave repository.
func NewConclaveRepository(db *sql.DB) *ConclaveRepository {
	return &ConclaveRepository{db: db}
}

// Create persists a new conclave.
func (r *ConclaveRepository) Create(ctx context.Context, conclave *secondary.ConclaveRecord) error {
	var desc, shipmentID sql.NullString
	if conclave.Description != "" {
		desc = sql.NullString{String: conclave.Description, Valid: true}
	}
	if conclave.ShipmentID != "" {
		shipmentID = sql.NullString{String: conclave.ShipmentID, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO conclaves (id, commission_id, shipment_id, title, description, status) VALUES (?, ?, ?, ?, ?, ?)",
		conclave.ID, conclave.CommissionID, shipmentID, conclave.Title, desc, "open",
	)
	if err != nil {
		return fmt.Errorf("failed to create conclave: %w", err)
	}

	return nil
}

// GetByID retrieves a conclave by its ID.
func (r *ConclaveRepository) GetByID(ctx context.Context, id string) (*secondary.ConclaveRecord, error) {
	var (
		desc       sql.NullString
		shipmentID sql.NullString
		decision   sql.NullString
		pinned     bool
		createdAt  time.Time
		updatedAt  time.Time
		decidedAt  sql.NullTime
	)

	record := &secondary.ConclaveRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, commission_id, shipment_id, title, description, status, decision, pinned, created_at, updated_at, decided_at FROM conclaves WHERE id = ?",
		id,
	).Scan(&record.ID, &record.CommissionID, &shipmentID, &record.Title, &desc, &record.Status, &decision, &pinned, &createdAt, &updatedAt, &decidedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("conclave %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conclave: %w", err)
	}

	record.Description = desc.String
	record.ShipmentID = shipmentID.String
	record.Decision = decision.String
	record.Pinned = pinned
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if decidedAt.Valid {
		record.DecidedAt = decidedAt.Time.Format(time.RFC3339)
	}

	return record, nil
}

// List retrieves conclaves matching the given filters.
func (r *ConclaveRepository) List(ctx context.Context, filters secondary.ConclaveFilters) ([]*secondary.ConclaveRecord, error) {
	query := "SELECT id, commission_id, shipment_id, title, description, status, decision, pinned, created_at, updated_at, decided_at FROM conclaves WHERE 1=1"
	args := []any{}

	if filters.CommissionID != "" {
		query += " AND commission_id = ?"
		args = append(args, filters.CommissionID)
	}

	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list conclaves: %w", err)
	}
	defer rows.Close()

	var conclaves []*secondary.ConclaveRecord
	for rows.Next() {
		var (
			desc       sql.NullString
			shipmentID sql.NullString
			decision   sql.NullString
			pinned     bool
			createdAt  time.Time
			updatedAt  time.Time
			decidedAt  sql.NullTime
		)

		record := &secondary.ConclaveRecord{}
		err := rows.Scan(&record.ID, &record.CommissionID, &shipmentID, &record.Title, &desc, &record.Status, &decision, &pinned, &createdAt, &updatedAt, &decidedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conclave: %w", err)
		}

		record.Description = desc.String
		record.ShipmentID = shipmentID.String
		record.Decision = decision.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if decidedAt.Valid {
			record.DecidedAt = decidedAt.Time.Format(time.RFC3339)
		}

		conclaves = append(conclaves, record)
	}

	return conclaves, nil
}

// Update updates an existing conclave.
func (r *ConclaveRepository) Update(ctx context.Context, conclave *secondary.ConclaveRecord) error {
	query := "UPDATE conclaves SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if conclave.Title != "" {
		query += ", title = ?"
		args = append(args, conclave.Title)
	}

	if conclave.Description != "" {
		query += ", description = ?"
		args = append(args, sql.NullString{String: conclave.Description, Valid: true})
	}

	if conclave.Decision != "" {
		query += ", decision = ?"
		args = append(args, sql.NullString{String: conclave.Decision, Valid: true})
	}

	query += " WHERE id = ?"
	args = append(args, conclave.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update conclave: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("conclave %s not found", conclave.ID)
	}

	return nil
}

// Delete removes a conclave from persistence.
func (r *ConclaveRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM conclaves WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete conclave: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("conclave %s not found", id)
	}

	return nil
}

// Pin pins a conclave.
func (r *ConclaveRepository) Pin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE conclaves SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to pin conclave: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("conclave %s not found", id)
	}

	return nil
}

// Unpin unpins a conclave.
func (r *ConclaveRepository) Unpin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE conclaves SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to unpin conclave: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("conclave %s not found", id)
	}

	return nil
}

// GetNextID returns the next available conclave ID.
func (r *ConclaveRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 5) AS INTEGER)), 0) FROM conclaves",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next conclave ID: %w", err)
	}

	return fmt.Sprintf("CON-%03d", maxID+1), nil
}

// UpdateStatus updates the status and optionally decided_at timestamp.
func (r *ConclaveRepository) UpdateStatus(ctx context.Context, id, status string, setDecided bool) error {
	var query string
	var args []any

	if setDecided {
		query = "UPDATE conclaves SET status = ?, decided_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
		args = []any{status, id}
	} else {
		query = "UPDATE conclaves SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
		args = []any{status, id}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update conclave status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("conclave %s not found", id)
	}

	return nil
}

// CommissionExists checks if a commission exists.
func (r *ConclaveRepository) CommissionExists(ctx context.Context, commissionID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM commissions WHERE id = ?", commissionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check commission existence: %w", err)
	}
	return count > 0, nil
}

// GetTasksByConclave retrieves tasks belonging to a conclave.
func (r *ConclaveRepository) GetTasksByConclave(ctx context.Context, conclaveID string) ([]*secondary.ConclaveTaskRecord, error) {
	query := `SELECT id, shipment_id, commission_id, title, description, type, status, priority,
		assigned_workbench_id, pinned, created_at, updated_at, claimed_at, completed_at
		FROM tasks WHERE conclave_id = ? ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, conclaveID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by conclave: %w", err)
	}
	defer rows.Close()

	var tasks []*secondary.ConclaveTaskRecord
	for rows.Next() {
		var (
			shipmentID          sql.NullString
			desc                sql.NullString
			taskType            sql.NullString
			priority            sql.NullString
			assignedWorkbenchID sql.NullString
			pinned              bool
			createdAt           time.Time
			updatedAt           time.Time
			claimedAt           sql.NullTime
			completedAt         sql.NullTime
		)

		record := &secondary.ConclaveTaskRecord{}
		err := rows.Scan(&record.ID, &shipmentID, &record.CommissionID, &record.Title, &desc, &taskType, &record.Status, &priority,
			&assignedWorkbenchID, &pinned, &createdAt, &updatedAt, &claimedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		record.ShipmentID = shipmentID.String
		record.Description = desc.String
		record.Type = taskType.String
		record.Priority = priority.String
		record.AssignedWorkbenchID = assignedWorkbenchID.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if claimedAt.Valid {
			record.ClaimedAt = claimedAt.Time.Format(time.RFC3339)
		}
		if completedAt.Valid {
			record.CompletedAt = completedAt.Time.Format(time.RFC3339)
		}

		tasks = append(tasks, record)
	}

	return tasks, nil
}

// GetPlansByConclave retrieves plans belonging to a conclave.
func (r *ConclaveRepository) GetPlansByConclave(ctx context.Context, conclaveID string) ([]*secondary.ConclavePlanRecord, error) {
	query := `SELECT p.id, p.commission_id, p.task_id, p.title, p.description, p.content, p.status, p.pinned,
		p.created_at, p.updated_at, p.approved_at
		FROM plans p
		INNER JOIN tasks t ON p.task_id = t.id
		WHERE t.conclave_id = ?
		ORDER BY p.created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, conclaveID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plans by conclave: %w", err)
	}
	defer rows.Close()

	var plans []*secondary.ConclavePlanRecord
	for rows.Next() {
		var (
			taskID     sql.NullString
			desc       sql.NullString
			content    sql.NullString
			pinned     bool
			createdAt  time.Time
			updatedAt  time.Time
			approvedAt sql.NullTime
		)

		record := &secondary.ConclavePlanRecord{}
		err := rows.Scan(&record.ID, &record.CommissionID, &taskID, &record.Title, &desc, &content, &record.Status, &pinned,
			&createdAt, &updatedAt, &approvedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plan: %w", err)
		}

		record.TaskID = taskID.String
		record.Description = desc.String
		record.Content = content.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if approvedAt.Valid {
			record.ApprovedAt = approvedAt.Time.Format(time.RFC3339)
		}

		plans = append(plans, record)
	}

	return plans, nil
}

// Ensure ConclaveRepository implements the interface
var _ secondary.ConclaveRepository = (*ConclaveRepository)(nil)
