// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	corefactory "github.com/example/orc/internal/core/factory"
	"github.com/example/orc/internal/ports/secondary"
)

// FactoryRepository implements secondary.FactoryRepository with SQLite.
type FactoryRepository struct {
	db *sql.DB
}

// NewFactoryRepository creates a new SQLite factory repository.
func NewFactoryRepository(db *sql.DB) *FactoryRepository {
	return &FactoryRepository{db: db}
}

// Create persists a new factory.
func (r *FactoryRepository) Create(ctx context.Context, factory *secondary.FactoryRecord) error {
	if factory.ID == "" {
		return fmt.Errorf("factory ID must be pre-populated by service layer")
	}
	if factory.Status == "" {
		factory.Status = "active"
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO factories (id, name, status) VALUES (?, ?, ?)",
		factory.ID, factory.Name, factory.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create factory: %w", err)
	}

	return nil
}

// GetByID retrieves a factory by its ID.
func (r *FactoryRepository) GetByID(ctx context.Context, id string) (*secondary.FactoryRecord, error) {
	var (
		createdAt time.Time
		updatedAt time.Time
	)

	record := &secondary.FactoryRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, status, created_at, updated_at FROM factories WHERE id = ?",
		id,
	).Scan(&record.ID, &record.Name, &record.Status, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("factory %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get factory: %w", err)
	}

	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	return record, nil
}

// GetByName retrieves a factory by its unique name.
func (r *FactoryRepository) GetByName(ctx context.Context, name string) (*secondary.FactoryRecord, error) {
	var (
		createdAt time.Time
		updatedAt time.Time
	)

	record := &secondary.FactoryRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, status, created_at, updated_at FROM factories WHERE name = ?",
		name,
	).Scan(&record.ID, &record.Name, &record.Status, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("factory with name %q not found", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get factory by name: %w", err)
	}

	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	return record, nil
}

// List retrieves factories matching the given filters.
func (r *FactoryRepository) List(ctx context.Context, filters secondary.FactoryFilters) ([]*secondary.FactoryRecord, error) {
	query := "SELECT id, name, status, created_at, updated_at FROM factories"
	args := []any{}

	if filters.Status != "" {
		query += " WHERE status = ?"
		args = append(args, filters.Status)
	}

	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list factories: %w", err)
	}
	defer rows.Close()

	var factories []*secondary.FactoryRecord
	for rows.Next() {
		var (
			createdAt time.Time
			updatedAt time.Time
		)

		record := &secondary.FactoryRecord{}
		err := rows.Scan(&record.ID, &record.Name, &record.Status, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan factory: %w", err)
		}

		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		factories = append(factories, record)
	}

	return factories, nil
}

// Update updates an existing factory.
func (r *FactoryRepository) Update(ctx context.Context, factory *secondary.FactoryRecord) error {
	query := "UPDATE factories SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if factory.Name != "" {
		query += ", name = ?"
		args = append(args, factory.Name)
	}

	if factory.Status != "" {
		query += ", status = ?"
		args = append(args, factory.Status)
	}

	query += " WHERE id = ?"
	args = append(args, factory.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update factory: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("factory %s not found", factory.ID)
	}

	return nil
}

// Delete removes a factory from persistence.
func (r *FactoryRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM factories WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete factory: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("factory %s not found", id)
	}

	return nil
}

// GetNextID returns the next available factory ID.
func (r *FactoryRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM factories",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next factory ID: %w", err)
	}

	return corefactory.GenerateFactoryID(maxID), nil
}

// CountWorkshops returns the number of workshops for a factory.
func (r *FactoryRepository) CountWorkshops(ctx context.Context, factoryID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM workshops WHERE factory_id = ?",
		factoryID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count workshops: %w", err)
	}

	return count, nil
}

// CountCommissions returns the number of commissions for a factory.
func (r *FactoryRepository) CountCommissions(ctx context.Context, factoryID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM commissions WHERE factory_id = ?",
		factoryID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count commissions: %w", err)
	}

	return count, nil
}

// Ensure FactoryRepository implements the interface
var _ secondary.FactoryRepository = (*FactoryRepository)(nil)
