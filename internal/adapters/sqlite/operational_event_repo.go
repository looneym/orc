// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/db"
	"github.com/example/orc/internal/ports/secondary"
)

// OperationalEventRepository implements secondary.OperationalEventRepository with SQLite.
type OperationalEventRepository struct {
	db *sql.DB
}

// NewOperationalEventRepository creates a new SQLite operational event repository.
func NewOperationalEventRepository(db *sql.DB) *OperationalEventRepository {
	return &OperationalEventRepository{db: db}
}

// conn returns the context-carried transaction if present, otherwise r.db.
func (r *OperationalEventRepository) conn(ctx context.Context) db.DBTX {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	return r.db
}

// Create persists a new operational event.
func (r *OperationalEventRepository) Create(ctx context.Context, event *secondary.OperationalEventRecord) error {
	var workshopID, actorID, version, dataJSON sql.NullString
	if event.WorkshopID != "" {
		workshopID = sql.NullString{String: event.WorkshopID, Valid: true}
	}
	if event.ActorID != "" {
		actorID = sql.NullString{String: event.ActorID, Valid: true}
	}
	if event.Version != "" {
		version = sql.NullString{String: event.Version, Valid: true}
	}
	if event.DataJSON != "" {
		dataJSON = sql.NullString{String: event.DataJSON, Valid: true}
	}

	_, err := r.conn(ctx).ExecContext(ctx,
		`INSERT INTO operational_events (id, workshop_id, actor_id, source, version, level, message, data_json) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		event.ID,
		workshopID,
		actorID,
		event.Source,
		version,
		event.Level,
		event.Message,
		dataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create operational event: %w", err)
	}

	return nil
}

// List retrieves operational events matching the given filters.
func (r *OperationalEventRepository) List(ctx context.Context, filters secondary.OperationalEventFilters) ([]*secondary.OperationalEventRecord, error) {
	query := `SELECT id, workshop_id, timestamp, actor_id, source, version, level, message, data_json, created_at FROM operational_events WHERE 1=1`
	args := []any{}

	if filters.WorkshopID != "" {
		query += " AND workshop_id = ?"
		args = append(args, filters.WorkshopID)
	}

	if filters.Source != "" {
		query += " AND source = ?"
		args = append(args, filters.Source)
	}

	if filters.Level != "" {
		query += " AND level = ?"
		args = append(args, filters.Level)
	}

	query += " ORDER BY timestamp DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list operational events: %w", err)
	}
	defer rows.Close()

	var events []*secondary.OperationalEventRecord
	for rows.Next() {
		var (
			workshopID sql.NullString
			actorID    sql.NullString
			version    sql.NullString
			dataJSON   sql.NullString
			timestamp  time.Time
			createdAt  time.Time
		)

		record := &secondary.OperationalEventRecord{}
		err := rows.Scan(&record.ID,
			&workshopID,
			&timestamp,
			&actorID,
			&record.Source,
			&version,
			&record.Level,
			&record.Message,
			&dataJSON,
			&createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan operational event: %w", err)
		}
		record.WorkshopID = workshopID.String
		record.Timestamp = timestamp.Format(time.RFC3339)
		record.ActorID = actorID.String
		record.Version = version.String
		record.DataJSON = dataJSON.String
		record.CreatedAt = createdAt.Format(time.RFC3339)

		events = append(events, record)
	}

	return events, nil
}

// GetNextID returns the next available operational event ID.
func (r *OperationalEventRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	prefixLen := len("OE-") + 1
	err := r.conn(ctx).QueryRowContext(ctx,
		fmt.Sprintf("SELECT COALESCE(MAX(CAST(SUBSTR(id, %d) AS INTEGER)), 0) FROM operational_events", prefixLen),
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next operational event ID: %w", err)
	}

	return fmt.Sprintf("OE-%04d", maxID+1), nil
}

// PruneOlderThan deletes operational events older than the given number of days.
func (r *OperationalEventRepository) PruneOlderThan(ctx context.Context, days int) (int, error) {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM operational_events WHERE timestamp < datetime('now', ?)",
		fmt.Sprintf("-%d days", days),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to prune operational events: %w", err)
	}

	count, _ := result.RowsAffected()
	return int(count), nil
}

// Ensure OperationalEventRepository implements the interface
var _ secondary.OperationalEventRepository = (*OperationalEventRepository)(nil)
