// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	corelibrary "github.com/example/orc/internal/core/library"
	"github.com/example/orc/internal/ports/secondary"
)

// LibraryRepository implements secondary.LibraryRepository with SQLite.
type LibraryRepository struct {
	db *sql.DB
}

// NewLibraryRepository creates a new SQLite library repository.
func NewLibraryRepository(db *sql.DB) *LibraryRepository {
	return &LibraryRepository{db: db}
}

// Create persists a new library.
func (r *LibraryRepository) Create(ctx context.Context, library *secondary.LibraryRecord) error {
	// Verify commission exists
	exists, err := r.CommissionExists(ctx, library.CommissionID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("commission %s not found", library.CommissionID)
	}

	// Check if library already exists for this commission
	var count int
	err = r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM libraries WHERE commission_id = ?", library.CommissionID,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing library: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("library already exists for commission %s", library.CommissionID)
	}

	// Generate library ID if not provided
	id := library.ID
	if id == "" {
		var maxID int
		err = r.db.QueryRowContext(ctx,
			"SELECT COALESCE(MAX(CAST(SUBSTR(id, 5) AS INTEGER)), 0) FROM libraries",
		).Scan(&maxID)
		if err != nil {
			return fmt.Errorf("failed to generate library ID: %w", err)
		}
		id = corelibrary.GenerateLibraryID(maxID)
	}

	_, err = r.db.ExecContext(ctx,
		"INSERT INTO libraries (id, commission_id) VALUES (?, ?)",
		id, library.CommissionID,
	)
	if err != nil {
		return fmt.Errorf("failed to create library: %w", err)
	}

	library.ID = id
	return nil
}

// GetByID retrieves a library by its ID.
func (r *LibraryRepository) GetByID(ctx context.Context, id string) (*secondary.LibraryRecord, error) {
	var (
		createdAt time.Time
		updatedAt time.Time
	)

	record := &secondary.LibraryRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, commission_id, created_at, updated_at FROM libraries WHERE id = ?",
		id,
	).Scan(&record.ID, &record.CommissionID, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("library %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get library: %w", err)
	}

	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	return record, nil
}

// GetByCommissionID retrieves the library for a commission.
func (r *LibraryRepository) GetByCommissionID(ctx context.Context, commissionID string) (*secondary.LibraryRecord, error) {
	var (
		createdAt time.Time
		updatedAt time.Time
	)

	record := &secondary.LibraryRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, commission_id, created_at, updated_at FROM libraries WHERE commission_id = ?",
		commissionID,
	).Scan(&record.ID, &record.CommissionID, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("library for commission %s not found", commissionID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get library by commission: %w", err)
	}

	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	return record, nil
}

// GetNextID returns the next available library ID.
func (r *LibraryRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 5) AS INTEGER)), 0) FROM libraries",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next library ID: %w", err)
	}

	return corelibrary.GenerateLibraryID(maxID), nil
}

// CommissionExists checks if a commission exists.
func (r *LibraryRepository) CommissionExists(ctx context.Context, commissionID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM commissions WHERE id = ?",
		commissionID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check commission existence: %w", err)
	}

	return count > 0, nil
}

// Ensure LibraryRepository implements the interface
var _ secondary.LibraryRepository = (*LibraryRepository)(nil)
