package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// ManifestServiceImpl implements the ManifestService interface.
type ManifestServiceImpl struct {
	manifestRepo secondary.ManifestRepository
}

// NewManifestService creates a new ManifestService with injected dependencies.
func NewManifestService(manifestRepo secondary.ManifestRepository) *ManifestServiceImpl {
	return &ManifestServiceImpl{
		manifestRepo: manifestRepo,
	}
}

// GetManifest retrieves a manifest by ID.
func (s *ManifestServiceImpl) GetManifest(ctx context.Context, manifestID string) (*primary.Manifest, error) {
	record, err := s.manifestRepo.GetByID(ctx, manifestID)
	if err != nil {
		return nil, err
	}
	return s.recordToManifest(record), nil
}

// GetManifestByShipment retrieves a manifest by shipment ID.
func (s *ManifestServiceImpl) GetManifestByShipment(ctx context.Context, shipmentID string) (*primary.Manifest, error) {
	record, err := s.manifestRepo.GetByShipment(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	return s.recordToManifest(record), nil
}

// ListManifests lists manifests with optional filters.
func (s *ManifestServiceImpl) ListManifests(ctx context.Context, filters primary.ManifestFilters) ([]*primary.Manifest, error) {
	records, err := s.manifestRepo.List(ctx, secondary.ManifestFilters{
		ShipmentID: filters.ShipmentID,
		Status:     filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list manifests: %w", err)
	}

	manifests := make([]*primary.Manifest, len(records))
	for i, r := range records {
		manifests[i] = s.recordToManifest(r)
	}
	return manifests, nil
}

// Helper methods

func (s *ManifestServiceImpl) recordToManifest(r *secondary.ManifestRecord) *primary.Manifest {
	return &primary.Manifest{
		ID:            r.ID,
		ShipmentID:    r.ShipmentID,
		CreatedBy:     r.CreatedBy,
		Attestation:   r.Attestation,
		Tasks:         r.Tasks,
		OrderingNotes: r.OrderingNotes,
		Status:        r.Status,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

// Ensure ManifestServiceImpl implements the interface
var _ primary.ManifestService = (*ManifestServiceImpl)(nil)
