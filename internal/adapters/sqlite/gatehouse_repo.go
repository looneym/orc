// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// GatehouseRepository implements secondary.GatehouseRepository with SQLite.
type GatehouseRepository struct {
	db *sql.DB
}

// NewGatehouseRepository creates a new SQLite gatehouse repository.
func NewGatehouseRepository(db *sql.DB) *GatehouseRepository {
	return &GatehouseRepository{db: db}
}

// Create persists a new gatehouse.
func (r *GatehouseRepository) Create(ctx context.Context, gatehouse *secondary.GatehouseRecord) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO gatehouses (id, workshop_id, status) VALUES (?, ?, ?)`,
		gatehouse.ID,
		gatehouse.WorkshopID,
		gatehouse.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create gatehouse: %w", err)
	}

	return nil
}

// GetByID retrieves a gatehouse by its ID.
func (r *GatehouseRepository) GetByID(ctx context.Context, id string) (*secondary.GatehouseRecord, error) {
	var (
		createdAt time.Time
		updatedAt time.Time
		focusedID sql.NullString
	)

	record := &secondary.GatehouseRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workshop_id, status, focused_id, created_at, updated_at FROM gatehouses WHERE id = ?`,
		id,
	).Scan(&record.ID, &record.WorkshopID, &record.Status, &focusedID, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("gatehouse %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get gatehouse: %w", err)
	}
	record.FocusedID = focusedID.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// GetByWorkshop retrieves a gatehouse by workshop ID.
func (r *GatehouseRepository) GetByWorkshop(ctx context.Context, workshopID string) (*secondary.GatehouseRecord, error) {
	var (
		createdAt time.Time
		updatedAt time.Time
		focusedID sql.NullString
	)

	record := &secondary.GatehouseRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workshop_id, status, focused_id, created_at, updated_at FROM gatehouses WHERE workshop_id = ?`,
		workshopID,
	).Scan(&record.ID, &record.WorkshopID, &record.Status, &focusedID, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("gatehouse for workshop %s not found", workshopID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get gatehouse: %w", err)
	}
	record.FocusedID = focusedID.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// List retrieves gatehouses matching the given filters.
func (r *GatehouseRepository) List(ctx context.Context, filters secondary.GatehouseFilters) ([]*secondary.GatehouseRecord, error) {
	query := `SELECT id, workshop_id, status, focused_id, created_at, updated_at FROM gatehouses WHERE 1=1`
	args := []any{}

	if filters.WorkshopID != "" {
		query += " AND workshop_id = ?"
		args = append(args, filters.WorkshopID)
	}

	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list gatehouses: %w", err)
	}
	defer rows.Close()

	var gatehouses []*secondary.GatehouseRecord
	for rows.Next() {
		var (
			createdAt time.Time
			updatedAt time.Time
			focusedID sql.NullString
		)

		record := &secondary.GatehouseRecord{}
		err := rows.Scan(&record.ID, &record.WorkshopID, &record.Status, &focusedID, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan gatehouse: %w", err)
		}
		record.FocusedID = focusedID.String
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)

		gatehouses = append(gatehouses, record)
	}

	return gatehouses, nil
}

// Update updates an existing gatehouse.
func (r *GatehouseRepository) Update(ctx context.Context, gatehouse *secondary.GatehouseRecord) error {
	query := "UPDATE gatehouses SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if gatehouse.Status != "" {
		query += ", status = ?"
		args = append(args, gatehouse.Status)
	}

	query += " WHERE id = ?"
	args = append(args, gatehouse.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update gatehouse: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("gatehouse %s not found", gatehouse.ID)
	}

	return nil
}

// Delete removes a gatehouse from persistence.
func (r *GatehouseRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM gatehouses WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete gatehouse: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("gatehouse %s not found", id)
	}

	return nil
}

// GetNextID returns the next available gatehouse ID.
func (r *GatehouseRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	prefixLen := len("GATE-") + 1
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COALESCE(MAX(CAST(SUBSTR(id, %d) AS INTEGER)), 0) FROM gatehouses", prefixLen),
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next gatehouse ID: %w", err)
	}

	return fmt.Sprintf("GATE-%03d", maxID+1), nil
}

// UpdateStatus updates the status of a gatehouse.
func (r *GatehouseRepository) UpdateStatus(ctx context.Context, id, status string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE gatehouses SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update gatehouse status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("gatehouse %s not found", id)
	}

	return nil
}

// WorkshopExists checks if a workshop exists.
func (r *GatehouseRepository) WorkshopExists(ctx context.Context, workshopID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM workshops WHERE id = ?", workshopID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check workshop existence: %w", err)
	}
	return count > 0, nil
}

// WorkshopHasGatehouse checks if a workshop already has a gatehouse (for 1:1 constraint).
func (r *GatehouseRepository) WorkshopHasGatehouse(ctx context.Context, workshopID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM gatehouses WHERE workshop_id = ?", workshopID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check existing gatehouse: %w", err)
	}
	return count > 0, nil
}

// UpdateFocusedID updates the focused container ID for a gatehouse.
// Pass empty string to clear focus.
func (r *GatehouseRepository) UpdateFocusedID(ctx context.Context, id, focusedID string) error {
	var focusedValue any
	if focusedID == "" {
		focusedValue = nil
	} else {
		focusedValue = focusedID
	}

	result, err := r.db.ExecContext(ctx,
		"UPDATE gatehouses SET focused_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		focusedValue, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update gatehouse focus: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("gatehouse %s not found", id)
	}

	return nil
}

// Ensure GatehouseRepository implements the interface
var _ secondary.GatehouseRepository = (*GatehouseRepository)(nil)
