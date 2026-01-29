// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	coreshipyard "github.com/example/orc/internal/core/shipyard"
	"github.com/example/orc/internal/ports/secondary"
)

// ShipyardRepository implements secondary.ShipyardRepository with SQLite.
type ShipyardRepository struct {
	db *sql.DB
}

// NewShipyardRepository creates a new SQLite shipyard repository.
func NewShipyardRepository(db *sql.DB) *ShipyardRepository {
	return &ShipyardRepository{db: db}
}

// Create persists a new shipyard.
func (r *ShipyardRepository) Create(ctx context.Context, shipyard *secondary.ShipyardRecord) error {
	// Verify commission exists
	exists, err := r.CommissionExists(ctx, shipyard.CommissionID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("commission %s not found", shipyard.CommissionID)
	}

	// Check if shipyard already exists for this commission
	var count int
	err = r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM shipyards WHERE commission_id = ?", shipyard.CommissionID,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing shipyard: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("shipyard already exists for commission %s", shipyard.CommissionID)
	}

	// Generate shipyard ID if not provided
	id := shipyard.ID
	if id == "" {
		var maxID int
		err = r.db.QueryRowContext(ctx,
			"SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM shipyards",
		).Scan(&maxID)
		if err != nil {
			return fmt.Errorf("failed to generate shipyard ID: %w", err)
		}
		id = coreshipyard.GenerateShipyardID(maxID)
	}

	_, err = r.db.ExecContext(ctx,
		"INSERT INTO shipyards (id, commission_id) VALUES (?, ?)",
		id, shipyard.CommissionID,
	)
	if err != nil {
		return fmt.Errorf("failed to create shipyard: %w", err)
	}

	shipyard.ID = id
	return nil
}

// GetByID retrieves a shipyard by its ID.
func (r *ShipyardRepository) GetByID(ctx context.Context, id string) (*secondary.ShipyardRecord, error) {
	var (
		createdAt time.Time
		updatedAt time.Time
	)

	record := &secondary.ShipyardRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, commission_id, created_at, updated_at FROM shipyards WHERE id = ?",
		id,
	).Scan(&record.ID, &record.CommissionID, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("shipyard %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get shipyard: %w", err)
	}

	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	return record, nil
}

// GetByCommissionID retrieves the shipyard for a commission.
func (r *ShipyardRepository) GetByCommissionID(ctx context.Context, commissionID string) (*secondary.ShipyardRecord, error) {
	var (
		createdAt time.Time
		updatedAt time.Time
	)

	record := &secondary.ShipyardRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, commission_id, created_at, updated_at FROM shipyards WHERE commission_id = ?",
		commissionID,
	).Scan(&record.ID, &record.CommissionID, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("shipyard for commission %s not found", commissionID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get shipyard by commission: %w", err)
	}

	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	return record, nil
}

// GetNextID returns the next available shipyard ID.
func (r *ShipyardRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM shipyards",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next shipyard ID: %w", err)
	}

	return coreshipyard.GenerateShipyardID(maxID), nil
}

// CommissionExists checks if a commission exists.
func (r *ShipyardRepository) CommissionExists(ctx context.Context, commissionID string) (bool, error) {
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

// Ensure ShipyardRepository implements the interface
var _ secondary.ShipyardRepository = (*ShipyardRepository)(nil)
