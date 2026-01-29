// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// ManifestRepository implements secondary.ManifestRepository with SQLite.
type ManifestRepository struct {
	db *sql.DB
}

// NewManifestRepository creates a new SQLite manifest repository.
func NewManifestRepository(db *sql.DB) *ManifestRepository {
	return &ManifestRepository{db: db}
}

// Create persists a new manifest.
func (r *ManifestRepository) Create(ctx context.Context, manifest *secondary.ManifestRecord) error {
	var attestation, tasks, orderingNotes sql.NullString
	if manifest.Attestation != "" {
		attestation = sql.NullString{String: manifest.Attestation, Valid: true}
	}
	if manifest.Tasks != "" {
		tasks = sql.NullString{String: manifest.Tasks, Valid: true}
	}
	if manifest.OrderingNotes != "" {
		orderingNotes = sql.NullString{String: manifest.OrderingNotes, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO manifests (id, shipment_id, created_by, attestation, tasks, ordering_notes, status) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		manifest.ID,
		manifest.ShipmentID,
		manifest.CreatedBy,
		attestation,
		tasks,
		orderingNotes,
		manifest.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create manifest: %w", err)
	}

	return nil
}

// GetByID retrieves a manifest by its ID.
func (r *ManifestRepository) GetByID(ctx context.Context, id string) (*secondary.ManifestRecord, error) {
	var (
		attestation   sql.NullString
		tasks         sql.NullString
		orderingNotes sql.NullString
		createdAt     time.Time
		updatedAt     time.Time
	)

	record := &secondary.ManifestRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, created_by, attestation, tasks, ordering_notes, status, created_at, updated_at FROM manifests WHERE id = ?`,
		id,
	).Scan(&record.ID, &record.ShipmentID, &record.CreatedBy, &attestation, &tasks, &orderingNotes, &record.Status, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("manifest %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %w", err)
	}
	record.Attestation = attestation.String
	record.Tasks = tasks.String
	record.OrderingNotes = orderingNotes.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// GetByShipment retrieves a manifest by shipment ID.
func (r *ManifestRepository) GetByShipment(ctx context.Context, shipmentID string) (*secondary.ManifestRecord, error) {
	var (
		attestation   sql.NullString
		tasks         sql.NullString
		orderingNotes sql.NullString
		createdAt     time.Time
		updatedAt     time.Time
	)

	record := &secondary.ManifestRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, created_by, attestation, tasks, ordering_notes, status, created_at, updated_at FROM manifests WHERE shipment_id = ?`,
		shipmentID,
	).Scan(&record.ID, &record.ShipmentID, &record.CreatedBy, &attestation, &tasks, &orderingNotes, &record.Status, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("manifest for shipment %s not found", shipmentID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %w", err)
	}
	record.Attestation = attestation.String
	record.Tasks = tasks.String
	record.OrderingNotes = orderingNotes.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// List retrieves manifests matching the given filters.
func (r *ManifestRepository) List(ctx context.Context, filters secondary.ManifestFilters) ([]*secondary.ManifestRecord, error) {
	query := `SELECT id, shipment_id, created_by, attestation, tasks, ordering_notes, status, created_at, updated_at FROM manifests WHERE 1=1`
	args := []any{}

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
		return nil, fmt.Errorf("failed to list manifests: %w", err)
	}
	defer rows.Close()

	var manifests []*secondary.ManifestRecord
	for rows.Next() {
		var (
			attestation   sql.NullString
			tasks         sql.NullString
			orderingNotes sql.NullString
			createdAt     time.Time
			updatedAt     time.Time
		)

		record := &secondary.ManifestRecord{}
		err := rows.Scan(&record.ID, &record.ShipmentID, &record.CreatedBy, &attestation, &tasks, &orderingNotes, &record.Status, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan manifest: %w", err)
		}
		record.Attestation = attestation.String
		record.Tasks = tasks.String
		record.OrderingNotes = orderingNotes.String
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)

		manifests = append(manifests, record)
	}

	return manifests, nil
}

// Update updates an existing manifest.
func (r *ManifestRepository) Update(ctx context.Context, manifest *secondary.ManifestRecord) error {
	query := "UPDATE manifests SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if manifest.Attestation != "" {
		query += ", attestation = ?"
		args = append(args, sql.NullString{String: manifest.Attestation, Valid: true})
	}
	if manifest.Tasks != "" {
		query += ", tasks = ?"
		args = append(args, sql.NullString{String: manifest.Tasks, Valid: true})
	}
	if manifest.OrderingNotes != "" {
		query += ", ordering_notes = ?"
		args = append(args, sql.NullString{String: manifest.OrderingNotes, Valid: true})
	}

	query += " WHERE id = ?"
	args = append(args, manifest.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("manifest %s not found", manifest.ID)
	}

	return nil
}

// Delete removes a manifest from persistence.
func (r *ManifestRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM manifests WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete manifest: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("manifest %s not found", id)
	}

	return nil
}

// GetNextID returns the next available manifest ID.
func (r *ManifestRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	prefixLen := len("MAN-") + 1
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COALESCE(MAX(CAST(SUBSTR(id, %d) AS INTEGER)), 0) FROM manifests", prefixLen),
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next manifest ID: %w", err)
	}

	return fmt.Sprintf("MAN-%03d", maxID+1), nil
}

// UpdateStatus updates the status of a manifest.
func (r *ManifestRepository) UpdateStatus(ctx context.Context, id, status string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE manifests SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update manifest status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("manifest %s not found", id)
	}

	return nil
}

// ShipmentExists checks if a shipment exists.
func (r *ManifestRepository) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM shipments WHERE id = ?", shipmentID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check shipment existence: %w", err)
	}
	return count > 0, nil
}

// ShipmentHasManifest checks if a shipment already has a manifest (for 1:1 constraint).
func (r *ManifestRepository) ShipmentHasManifest(ctx context.Context, shipmentID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM manifests WHERE shipment_id = ?", shipmentID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check existing manifest: %w", err)
	}
	return count > 0, nil
}

// Ensure ManifestRepository implements the interface
var _ secondary.ManifestRepository = (*ManifestRepository)(nil)
