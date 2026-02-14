package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockFactoryRepoForService implements secondary.FactoryRepository for testing.
type mockFactoryRepoForService struct {
	factories       map[string]*secondary.FactoryRecord
	factoriesByName map[string]*secondary.FactoryRecord
	nextID          int
	workshopCounts  map[string]int
	commissionCount map[string]int
	createErr       error
	getErr          error
	updateErr       error
	deleteErr       error
}

func newMockFactoryRepoForService() *mockFactoryRepoForService {
	return &mockFactoryRepoForService{
		factories:       make(map[string]*secondary.FactoryRecord),
		factoriesByName: make(map[string]*secondary.FactoryRecord),
		workshopCounts:  make(map[string]int),
		commissionCount: make(map[string]int),
		nextID:          1,
	}
}

func (m *mockFactoryRepoForService) Create(ctx context.Context, factory *secondary.FactoryRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.factories[factory.ID] = factory
	m.factoriesByName[factory.Name] = factory
	return nil
}

func (m *mockFactoryRepoForService) GetByID(ctx context.Context, id string) (*secondary.FactoryRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if f, ok := m.factories[id]; ok {
		return f, nil
	}
	return nil, errors.New("not found")
}

func (m *mockFactoryRepoForService) GetByName(ctx context.Context, name string) (*secondary.FactoryRecord, error) {
	if f, ok := m.factoriesByName[name]; ok {
		return f, nil
	}
	return nil, errors.New("not found")
}

func (m *mockFactoryRepoForService) List(ctx context.Context, filters secondary.FactoryFilters) ([]*secondary.FactoryRecord, error) {
	var result []*secondary.FactoryRecord
	for _, f := range m.factories {
		if filters.Status != "" && f.Status != filters.Status {
			continue
		}
		result = append(result, f)
		if filters.Limit > 0 && len(result) >= filters.Limit {
			break
		}
	}
	return result, nil
}

func (m *mockFactoryRepoForService) Update(ctx context.Context, factory *secondary.FactoryRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.factories[factory.ID]; !ok {
		return errors.New("not found")
	}
	existing := m.factories[factory.ID]
	if factory.Name != "" {
		delete(m.factoriesByName, existing.Name)
		existing.Name = factory.Name
		m.factoriesByName[factory.Name] = existing
	}
	if factory.Status != "" {
		existing.Status = factory.Status
	}
	return nil
}

func (m *mockFactoryRepoForService) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.factories[id]; !ok {
		return errors.New("not found")
	}
	f := m.factories[id]
	delete(m.factoriesByName, f.Name)
	delete(m.factories, id)
	return nil
}

func (m *mockFactoryRepoForService) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("FACT-%03d", id), nil
}

func (m *mockFactoryRepoForService) CountWorkshops(ctx context.Context, factoryID string) (int, error) {
	return m.workshopCounts[factoryID], nil
}

func (m *mockFactoryRepoForService) CountCommissions(ctx context.Context, factoryID string) (int, error) {
	return m.commissionCount[factoryID], nil
}

func newTestFactoryService() (*FactoryServiceImpl, *mockFactoryRepoForService) {
	repo := newMockFactoryRepoForService()
	service := NewFactoryService(repo, &mockTransactor{})
	return service, repo
}

func TestFactoryService_CreateFactory(t *testing.T) {
	service, _ := newTestFactoryService()
	ctx := context.Background()

	resp, err := service.CreateFactory(ctx, primary.CreateFactoryRequest{
		Name: "test-factory",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.FactoryID == "" {
		t.Error("expected factory ID to be set")
	}
	if resp.Factory.Name != "test-factory" {
		t.Errorf("expected name 'test-factory', got %q", resp.Factory.Name)
	}
	if resp.Factory.Status != "active" {
		t.Errorf("expected status 'active', got %q", resp.Factory.Status)
	}
}

func TestFactoryService_CreateFactory_DuplicateName(t *testing.T) {
	service, repo := newTestFactoryService()
	ctx := context.Background()

	repo.factories["FACT-001"] = &secondary.FactoryRecord{ID: "FACT-001", Name: "existing"}
	repo.factoriesByName["existing"] = repo.factories["FACT-001"]

	_, err := service.CreateFactory(ctx, primary.CreateFactoryRequest{
		Name: "existing",
	})

	if err == nil {
		t.Error("expected error for duplicate name")
	}
}

func TestFactoryService_GetFactory(t *testing.T) {
	service, repo := newTestFactoryService()
	ctx := context.Background()

	repo.factories["FACT-001"] = &secondary.FactoryRecord{
		ID:     "FACT-001",
		Name:   "test-factory",
		Status: "active",
	}

	factory, err := service.GetFactory(ctx, "FACT-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if factory.Name != "test-factory" {
		t.Errorf("expected name 'test-factory', got %q", factory.Name)
	}
}

func TestFactoryService_GetFactory_NotFound(t *testing.T) {
	service, _ := newTestFactoryService()
	ctx := context.Background()

	_, err := service.GetFactory(ctx, "FACT-999")
	if err == nil {
		t.Error("expected error for non-existent factory")
	}
}

func TestFactoryService_GetFactoryByName(t *testing.T) {
	service, repo := newTestFactoryService()
	ctx := context.Background()

	repo.factories["FACT-001"] = &secondary.FactoryRecord{
		ID:     "FACT-001",
		Name:   "unique-name",
		Status: "active",
	}
	repo.factoriesByName["unique-name"] = repo.factories["FACT-001"]

	factory, err := service.GetFactoryByName(ctx, "unique-name")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if factory.ID != "FACT-001" {
		t.Errorf("expected ID 'FACT-001', got %q", factory.ID)
	}
}

func TestFactoryService_GetFactoryByName_NotFound(t *testing.T) {
	service, _ := newTestFactoryService()
	ctx := context.Background()

	_, err := service.GetFactoryByName(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent name")
	}
}

func TestFactoryService_ListFactories(t *testing.T) {
	service, repo := newTestFactoryService()
	ctx := context.Background()

	repo.factories["FACT-001"] = &secondary.FactoryRecord{ID: "FACT-001", Name: "factory-1", Status: "active"}
	repo.factories["FACT-002"] = &secondary.FactoryRecord{ID: "FACT-002", Name: "factory-2", Status: "active"}

	factories, err := service.ListFactories(ctx, primary.FactoryFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(factories) != 2 {
		t.Errorf("expected 2 factories, got %d", len(factories))
	}
}

func TestFactoryService_ListFactories_FilterByStatus(t *testing.T) {
	service, repo := newTestFactoryService()
	ctx := context.Background()

	repo.factories["FACT-001"] = &secondary.FactoryRecord{ID: "FACT-001", Name: "factory-1", Status: "active"}
	repo.factories["FACT-002"] = &secondary.FactoryRecord{ID: "FACT-002", Name: "factory-2", Status: "archived"}

	factories, err := service.ListFactories(ctx, primary.FactoryFilters{Status: "active"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(factories) != 1 {
		t.Errorf("expected 1 active factory, got %d", len(factories))
	}
}

func TestFactoryService_UpdateFactory(t *testing.T) {
	service, repo := newTestFactoryService()
	ctx := context.Background()

	repo.factories["FACT-001"] = &secondary.FactoryRecord{ID: "FACT-001", Name: "original", Status: "active"}
	repo.factoriesByName["original"] = repo.factories["FACT-001"]

	err := service.UpdateFactory(ctx, primary.UpdateFactoryRequest{
		FactoryID: "FACT-001",
		Name:      "updated",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	factory, _ := service.GetFactory(ctx, "FACT-001")
	if factory.Name != "updated" {
		t.Errorf("expected name 'updated', got %q", factory.Name)
	}
}

func TestFactoryService_UpdateFactory_NotFound(t *testing.T) {
	service, _ := newTestFactoryService()
	ctx := context.Background()

	err := service.UpdateFactory(ctx, primary.UpdateFactoryRequest{
		FactoryID: "FACT-999",
		Name:      "updated",
	})
	if err == nil {
		t.Error("expected error for non-existent factory")
	}
}

func TestFactoryService_DeleteFactory(t *testing.T) {
	service, repo := newTestFactoryService()
	ctx := context.Background()

	repo.factories["FACT-001"] = &secondary.FactoryRecord{ID: "FACT-001", Name: "to-delete", Status: "active"}
	repo.factoriesByName["to-delete"] = repo.factories["FACT-001"]

	err := service.DeleteFactory(ctx, primary.DeleteFactoryRequest{
		FactoryID: "FACT-001",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = service.GetFactory(ctx, "FACT-001")
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestFactoryService_DeleteFactory_HasChildren(t *testing.T) {
	service, repo := newTestFactoryService()
	ctx := context.Background()

	repo.factories["FACT-001"] = &secondary.FactoryRecord{ID: "FACT-001", Name: "with-children", Status: "active"}
	repo.workshopCounts["FACT-001"] = 2

	err := service.DeleteFactory(ctx, primary.DeleteFactoryRequest{
		FactoryID: "FACT-001",
		Force:     false,
	})
	if err == nil {
		t.Error("expected error for factory with children")
	}
}

func TestFactoryService_DeleteFactory_HasChildren_Force(t *testing.T) {
	service, repo := newTestFactoryService()
	ctx := context.Background()

	repo.factories["FACT-001"] = &secondary.FactoryRecord{ID: "FACT-001", Name: "with-children", Status: "active"}
	repo.factoriesByName["with-children"] = repo.factories["FACT-001"]
	repo.workshopCounts["FACT-001"] = 2

	err := service.DeleteFactory(ctx, primary.DeleteFactoryRequest{
		FactoryID: "FACT-001",
		Force:     true,
	})
	if err != nil {
		t.Fatalf("expected no error with force=true, got %v", err)
	}
}

func TestFactoryService_DeleteFactory_NotFound(t *testing.T) {
	service, _ := newTestFactoryService()
	ctx := context.Background()

	err := service.DeleteFactory(ctx, primary.DeleteFactoryRequest{
		FactoryID: "FACT-999",
	})
	if err == nil {
		t.Error("expected error for non-existent factory")
	}
}
