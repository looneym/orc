// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// WatchdogRepository implements secondary.WatchdogRepository with SQLite.
type WatchdogRepository struct {
	db *sql.DB
}

// NewWatchdogRepository creates a new SQLite watchdog repository.
func NewWatchdogRepository(db *sql.DB) *WatchdogRepository {
	return &WatchdogRepository{db: db}
}

// Create persists a new watchdog.
func (r *WatchdogRepository) Create(ctx context.Context, watchdog *secondary.WatchdogRecord) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO watchdogs (id, workbench_id, status) VALUES (?, ?, ?)`,
		watchdog.ID,
		watchdog.WorkbenchID,
		watchdog.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create watchdog: %w", err)
	}

	return nil
}

// GetByID retrieves a watchdog by its ID.
func (r *WatchdogRepository) GetByID(ctx context.Context, id string) (*secondary.WatchdogRecord, error) {
	var (
		createdAt time.Time
		updatedAt time.Time
	)

	record := &secondary.WatchdogRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workbench_id, status, created_at, updated_at FROM watchdogs WHERE id = ?`,
		id,
	).Scan(&record.ID, &record.WorkbenchID, &record.Status, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("watchdog %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get watchdog: %w", err)
	}
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// GetByWorkbench retrieves a watchdog by workbench ID.
func (r *WatchdogRepository) GetByWorkbench(ctx context.Context, workbenchID string) (*secondary.WatchdogRecord, error) {
	var (
		createdAt time.Time
		updatedAt time.Time
	)

	record := &secondary.WatchdogRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workbench_id, status, created_at, updated_at FROM watchdogs WHERE workbench_id = ?`,
		workbenchID,
	).Scan(&record.ID, &record.WorkbenchID, &record.Status, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("watchdog for workbench %s not found", workbenchID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get watchdog: %w", err)
	}
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// List retrieves watchdogs matching the given filters.
func (r *WatchdogRepository) List(ctx context.Context, filters secondary.WatchdogFilters) ([]*secondary.WatchdogRecord, error) {
	query := `SELECT id, workbench_id, status, created_at, updated_at FROM watchdogs WHERE 1=1`
	args := []any{}

	if filters.WorkbenchID != "" {
		query += " AND workbench_id = ?"
		args = append(args, filters.WorkbenchID)
	}

	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list watchdogs: %w", err)
	}
	defer rows.Close()

	var watchdogs []*secondary.WatchdogRecord
	for rows.Next() {
		var (
			createdAt time.Time
			updatedAt time.Time
		)

		record := &secondary.WatchdogRecord{}
		err := rows.Scan(&record.ID, &record.WorkbenchID, &record.Status, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan watchdog: %w", err)
		}
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)

		watchdogs = append(watchdogs, record)
	}

	return watchdogs, nil
}

// Update updates an existing watchdog.
func (r *WatchdogRepository) Update(ctx context.Context, watchdog *secondary.WatchdogRecord) error {
	query := "UPDATE watchdogs SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if watchdog.Status != "" {
		query += ", status = ?"
		args = append(args, watchdog.Status)
	}

	query += " WHERE id = ?"
	args = append(args, watchdog.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update watchdog: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("watchdog %s not found", watchdog.ID)
	}

	return nil
}

// Delete removes a watchdog from persistence.
func (r *WatchdogRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM watchdogs WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete watchdog: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("watchdog %s not found", id)
	}

	return nil
}

// GetNextID returns the next available watchdog ID.
func (r *WatchdogRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	prefixLen := len("WATCH-") + 1
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COALESCE(MAX(CAST(SUBSTR(id, %d) AS INTEGER)), 0) FROM watchdogs", prefixLen),
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next watchdog ID: %w", err)
	}

	return fmt.Sprintf("WATCH-%03d", maxID+1), nil
}

// UpdateStatus updates the status of a watchdog.
func (r *WatchdogRepository) UpdateStatus(ctx context.Context, id, status string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE watchdogs SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update watchdog status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("watchdog %s not found", id)
	}

	return nil
}

// WorkbenchExists checks if a workbench exists.
func (r *WatchdogRepository) WorkbenchExists(ctx context.Context, workbenchID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM workbenches WHERE id = ?", workbenchID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check workbench existence: %w", err)
	}
	return count > 0, nil
}

// WorkbenchHasWatchdog checks if a workbench already has a watchdog (for 1:1 constraint).
func (r *WatchdogRepository) WorkbenchHasWatchdog(ctx context.Context, workbenchID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM watchdogs WHERE workbench_id = ?", workbenchID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check existing watchdog: %w", err)
	}
	return count > 0, nil
}

// Ensure WatchdogRepository implements the interface
var _ secondary.WatchdogRepository = (*WatchdogRepository)(nil)
