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

// WorkshopEventRepository implements secondary.WorkshopEventRepository with SQLite.
type WorkshopEventRepository struct {
	db *sql.DB
}

// NewWorkshopEventRepository creates a new SQLite workshop event repository.
func NewWorkshopEventRepository(db *sql.DB) *WorkshopEventRepository {
	return &WorkshopEventRepository{db: db}
}

// conn returns the context-carried transaction if present, otherwise r.db.
func (r *WorkshopEventRepository) conn(ctx context.Context) db.DBTX {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	return r.db
}

// Create persists a new audit event.
func (r *WorkshopEventRepository) Create(ctx context.Context, event *secondary.AuditEventRecord) error {
	var workshopID, actorID, source, version, fieldName, oldValue, newValue sql.NullString
	if event.WorkshopID != "" {
		workshopID = sql.NullString{String: event.WorkshopID, Valid: true}
	}
	if event.ActorID != "" {
		actorID = sql.NullString{String: event.ActorID, Valid: true}
	}
	if event.Source != "" {
		source = sql.NullString{String: event.Source, Valid: true}
	}
	if event.Version != "" {
		version = sql.NullString{String: event.Version, Valid: true}
	}
	if event.FieldName != "" {
		fieldName = sql.NullString{String: event.FieldName, Valid: true}
	}
	if event.OldValue != "" {
		oldValue = sql.NullString{String: event.OldValue, Valid: true}
	}
	if event.NewValue != "" {
		newValue = sql.NullString{String: event.NewValue, Valid: true}
	}

	_, err := r.conn(ctx).ExecContext(ctx,
		`INSERT INTO workshop_events (id, workshop_id, actor_id, source, version, entity_type, entity_id, action, field_name, old_value, new_value) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.ID,
		workshopID,
		actorID,
		source,
		version,
		event.EntityType,
		event.EntityID,
		event.Action,
		fieldName,
		oldValue,
		newValue,
	)
	if err != nil {
		return fmt.Errorf("failed to create workshop event: %w", err)
	}

	return nil
}

// GetByID retrieves an audit event by its ID.
func (r *WorkshopEventRepository) GetByID(ctx context.Context, id string) (*secondary.AuditEventRecord, error) {
	var (
		workshopID sql.NullString
		actorID    sql.NullString
		source     sql.NullString
		version    sql.NullString
		fieldName  sql.NullString
		oldValue   sql.NullString
		newValue   sql.NullString
		timestamp  time.Time
		createdAt  time.Time
	)

	record := &secondary.AuditEventRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workshop_id, timestamp, actor_id, source, version, entity_type, entity_id, action, field_name, old_value, new_value, created_at FROM workshop_events WHERE id = ?`,
		id,
	).Scan(&record.ID,
		&workshopID,
		&timestamp,
		&actorID,
		&source,
		&version,
		&record.EntityType,
		&record.EntityID,
		&record.Action,
		&fieldName,
		&oldValue,
		&newValue,
		&createdAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("workshop event %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workshop event: %w", err)
	}
	record.WorkshopID = workshopID.String
	record.Timestamp = timestamp.Format(time.RFC3339)
	record.ActorID = actorID.String
	record.Source = source.String
	record.Version = version.String
	record.FieldName = fieldName.String
	record.OldValue = oldValue.String
	record.NewValue = newValue.String
	record.CreatedAt = createdAt.Format(time.RFC3339)

	return record, nil
}

// List retrieves audit events matching the given filters.
func (r *WorkshopEventRepository) List(ctx context.Context, filters secondary.AuditEventFilters) ([]*secondary.AuditEventRecord, error) {
	query := `SELECT id, workshop_id, timestamp, actor_id, source, version, entity_type, entity_id, action, field_name, old_value, new_value, created_at FROM workshop_events WHERE 1=1`
	args := []any{}

	if filters.WorkshopID != "" {
		query += " AND workshop_id = ?"
		args = append(args, filters.WorkshopID)
	}

	if filters.EntityType != "" {
		query += " AND entity_type = ?"
		args = append(args, filters.EntityType)
	}

	if filters.EntityID != "" {
		query += " AND entity_id = ?"
		args = append(args, filters.EntityID)
	}

	if filters.ActorID != "" {
		query += " AND actor_id = ?"
		args = append(args, filters.ActorID)
	}

	if filters.Action != "" {
		query += " AND action = ?"
		args = append(args, filters.Action)
	}

	if filters.Source != "" {
		query += " AND source = ?"
		args = append(args, filters.Source)
	}

	query += " ORDER BY timestamp DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list workshop events: %w", err)
	}
	defer rows.Close()

	var events []*secondary.AuditEventRecord
	for rows.Next() {
		var (
			workshopID sql.NullString
			actorID    sql.NullString
			source     sql.NullString
			version    sql.NullString
			fieldName  sql.NullString
			oldValue   sql.NullString
			newValue   sql.NullString
			timestamp  time.Time
			createdAt  time.Time
		)

		record := &secondary.AuditEventRecord{}
		err := rows.Scan(&record.ID,
			&workshopID,
			&timestamp,
			&actorID,
			&source,
			&version,
			&record.EntityType,
			&record.EntityID,
			&record.Action,
			&fieldName,
			&oldValue,
			&newValue,
			&createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workshop event: %w", err)
		}
		record.WorkshopID = workshopID.String
		record.Timestamp = timestamp.Format(time.RFC3339)
		record.ActorID = actorID.String
		record.Source = source.String
		record.Version = version.String
		record.FieldName = fieldName.String
		record.OldValue = oldValue.String
		record.NewValue = newValue.String
		record.CreatedAt = createdAt.Format(time.RFC3339)

		events = append(events, record)
	}

	return events, nil
}

// GetNextID returns the next available audit event ID.
func (r *WorkshopEventRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	prefixLen := len("WE-") + 1
	err := r.conn(ctx).QueryRowContext(ctx,
		fmt.Sprintf("SELECT COALESCE(MAX(CAST(SUBSTR(id, %d) AS INTEGER)), 0) FROM workshop_events", prefixLen),
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next workshop event ID: %w", err)
	}

	return fmt.Sprintf("WE-%04d", maxID+1), nil
}

// WorkshopExists checks if a workshop exists (for validation).
func (r *WorkshopEventRepository) WorkshopExists(ctx context.Context, workshopID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM workshops WHERE id = ?", workshopID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check workshop existence: %w", err)
	}
	return count > 0, nil
}

// PruneOlderThan deletes audit events older than the given number of days.
func (r *WorkshopEventRepository) PruneOlderThan(ctx context.Context, days int) (int, error) {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM workshop_events WHERE timestamp < datetime('now', ?)",
		fmt.Sprintf("-%d days", days),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to prune workshop events: %w", err)
	}

	count, _ := result.RowsAffected()
	return int(count), nil
}

// Ensure WorkshopEventRepository implements the interface
var _ secondary.WorkshopEventRepository = (*WorkshopEventRepository)(nil)
