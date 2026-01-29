package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockManifestRepository implements secondary.ManifestRepository for testing.
type mockManifestRepository struct {
	manifests           map[string]*secondary.ManifestRecord
	manifestsByShipment map[string]*secondary.ManifestRecord
	shipmentExists      map[string]bool
	shipmentHasManifest map[string]bool
	nextID              int
}

func newMockManifestRepository() *mockManifestRepository {
	return &mockManifestRepository{
		manifests:           make(map[string]*secondary.ManifestRecord),
		manifestsByShipment: make(map[string]*secondary.ManifestRecord),
		shipmentExists:      make(map[string]bool),
		shipmentHasManifest: make(map[string]bool),
		nextID:              1,
	}
}

func (m *mockManifestRepository) Create(ctx context.Context, manifest *secondary.ManifestRecord) error {
	m.manifests[manifest.ID] = manifest
	m.manifestsByShipment[manifest.ShipmentID] = manifest
	m.shipmentHasManifest[manifest.ShipmentID] = true
	return nil
}

func (m *mockManifestRepository) GetByID(ctx context.Context, id string) (*secondary.ManifestRecord, error) {
	if man, ok := m.manifests[id]; ok {
		return man, nil
	}
	return nil, errors.New("not found")
}

func (m *mockManifestRepository) GetByShipment(ctx context.Context, shipmentID string) (*secondary.ManifestRecord, error) {
	if man, ok := m.manifestsByShipment[shipmentID]; ok {
		return man, nil
	}
	return nil, errors.New("not found")
}

func (m *mockManifestRepository) List(ctx context.Context, filters secondary.ManifestFilters) ([]*secondary.ManifestRecord, error) {
	var result []*secondary.ManifestRecord
	for _, man := range m.manifests {
		if filters.ShipmentID != "" && man.ShipmentID != filters.ShipmentID {
			continue
		}
		if filters.Status != "" && man.Status != filters.Status {
			continue
		}
		result = append(result, man)
	}
	return result, nil
}

func (m *mockManifestRepository) Update(ctx context.Context, manifest *secondary.ManifestRecord) error {
	if _, ok := m.manifests[manifest.ID]; !ok {
		return errors.New("not found")
	}
	return nil
}

func (m *mockManifestRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.manifests[id]; !ok {
		return errors.New("not found")
	}
	man := m.manifests[id]
	delete(m.manifestsByShipment, man.ShipmentID)
	delete(m.manifests, id)
	m.shipmentHasManifest[man.ShipmentID] = false
	return nil
}

func (m *mockManifestRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("MAN-%03d", id), nil
}

func (m *mockManifestRepository) UpdateStatus(ctx context.Context, id, status string) error {
	if man, ok := m.manifests[id]; ok {
		man.Status = status
		return nil
	}
	return errors.New("not found")
}

func (m *mockManifestRepository) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	return m.shipmentExists[shipmentID], nil
}

func (m *mockManifestRepository) ShipmentHasManifest(ctx context.Context, shipmentID string) (bool, error) {
	return m.shipmentHasManifest[shipmentID], nil
}

func newTestManifestService() (*ManifestServiceImpl, *mockManifestRepository) {
	repo := newMockManifestRepository()
	service := NewManifestService(repo)
	return service, repo
}

func TestManifestService_GetManifest(t *testing.T) {
	service, repo := newTestManifestService()
	ctx := context.Background()

	repo.manifests["MAN-001"] = &secondary.ManifestRecord{
		ID:         "MAN-001",
		ShipmentID: "SHIP-001",
		CreatedBy:  "ORC",
		Status:     "draft",
	}

	manifest, err := service.GetManifest(ctx, "MAN-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if manifest.ShipmentID != "SHIP-001" {
		t.Errorf("expected shipmentID 'SHIP-001', got %q", manifest.ShipmentID)
	}
}

func TestManifestService_GetManifest_NotFound(t *testing.T) {
	service, _ := newTestManifestService()
	ctx := context.Background()

	_, err := service.GetManifest(ctx, "MAN-999")
	if err == nil {
		t.Error("expected error for non-existent manifest")
	}
}

func TestManifestService_GetManifestByShipment(t *testing.T) {
	service, repo := newTestManifestService()
	ctx := context.Background()

	repo.manifests["MAN-001"] = &secondary.ManifestRecord{
		ID:         "MAN-001",
		ShipmentID: "SHIP-001",
		CreatedBy:  "ORC",
		Status:     "draft",
	}
	repo.manifestsByShipment["SHIP-001"] = repo.manifests["MAN-001"]

	manifest, err := service.GetManifestByShipment(ctx, "SHIP-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if manifest.ID != "MAN-001" {
		t.Errorf("expected ID 'MAN-001', got %q", manifest.ID)
	}
}

func TestManifestService_ListManifests(t *testing.T) {
	service, repo := newTestManifestService()
	ctx := context.Background()

	repo.manifests["MAN-001"] = &secondary.ManifestRecord{ID: "MAN-001", ShipmentID: "SHIP-001", CreatedBy: "ORC", Status: "draft"}
	repo.manifests["MAN-002"] = &secondary.ManifestRecord{ID: "MAN-002", ShipmentID: "SHIP-002", CreatedBy: "ORC", Status: "launched"}

	manifests, err := service.ListManifests(ctx, primary.ManifestFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(manifests) != 2 {
		t.Errorf("expected 2 manifests, got %d", len(manifests))
	}
}

func TestManifestService_ListManifests_FilterByStatus(t *testing.T) {
	service, repo := newTestManifestService()
	ctx := context.Background()

	repo.manifests["MAN-001"] = &secondary.ManifestRecord{ID: "MAN-001", ShipmentID: "SHIP-001", CreatedBy: "ORC", Status: "draft"}
	repo.manifests["MAN-002"] = &secondary.ManifestRecord{ID: "MAN-002", ShipmentID: "SHIP-002", CreatedBy: "ORC", Status: "launched"}

	manifests, err := service.ListManifests(ctx, primary.ManifestFilters{Status: "launched"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(manifests) != 1 {
		t.Errorf("expected 1 manifest, got %d", len(manifests))
	}
}
