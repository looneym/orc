package app

import (
	"context"
	"fmt"

	corefactory "github.com/example/orc/internal/core/factory"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// FactoryServiceImpl implements the FactoryService interface.
type FactoryServiceImpl struct {
	factoryRepo secondary.FactoryRepository
	transactor  secondary.Transactor
}

// NewFactoryService creates a new FactoryService with injected dependencies.
func NewFactoryService(factoryRepo secondary.FactoryRepository, transactor secondary.Transactor) *FactoryServiceImpl {
	return &FactoryServiceImpl{
		factoryRepo: factoryRepo,
		transactor:  transactor,
	}
}

// CreateFactory creates a new factory.
func (s *FactoryServiceImpl) CreateFactory(ctx context.Context, req primary.CreateFactoryRequest) (*primary.CreateFactoryResponse, error) {
	// 1. Check if name already exists
	_, err := s.factoryRepo.GetByName(ctx, req.Name)
	nameExists := err == nil

	// 2. Guard check
	guardCtx := corefactory.CreateFactoryContext{
		Name:       req.Name,
		NameExists: nameExists,
	}
	if result := corefactory.CanCreateFactory(guardCtx); !result.Allowed {
		return nil, result.Error()
	}

	var record *secondary.FactoryRecord
	err = s.transactor.WithImmediateTx(ctx, func(txCtx context.Context) error {
		// 3. Generate ID
		id, err := s.factoryRepo.GetNextID(txCtx)
		if err != nil {
			return fmt.Errorf("failed to generate factory ID: %w", err)
		}

		// 4. Create factory record
		record = &secondary.FactoryRecord{
			ID:     id,
			Name:   req.Name,
			Status: "active",
		}
		if err := s.factoryRepo.Create(txCtx, record); err != nil {
			return fmt.Errorf("failed to create factory: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &primary.CreateFactoryResponse{
		FactoryID: record.ID,
		Factory:   s.recordToFactory(record),
	}, nil
}

// GetFactory retrieves a factory by ID.
func (s *FactoryServiceImpl) GetFactory(ctx context.Context, factoryID string) (*primary.Factory, error) {
	record, err := s.factoryRepo.GetByID(ctx, factoryID)
	if err != nil {
		return nil, fmt.Errorf("factory not found: %w", err)
	}
	return s.recordToFactory(record), nil
}

// GetFactoryByName retrieves a factory by its unique name.
func (s *FactoryServiceImpl) GetFactoryByName(ctx context.Context, name string) (*primary.Factory, error) {
	record, err := s.factoryRepo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("factory not found: %w", err)
	}
	return s.recordToFactory(record), nil
}

// ListFactories lists factories with optional filters.
func (s *FactoryServiceImpl) ListFactories(ctx context.Context, filters primary.FactoryFilters) ([]*primary.Factory, error) {
	records, err := s.factoryRepo.List(ctx, secondary.FactoryFilters{
		Status: filters.Status,
		Limit:  filters.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list factories: %w", err)
	}

	factories := make([]*primary.Factory, len(records))
	for i, r := range records {
		factories[i] = s.recordToFactory(r)
	}
	return factories, nil
}

// UpdateFactory updates factory name.
func (s *FactoryServiceImpl) UpdateFactory(ctx context.Context, req primary.UpdateFactoryRequest) error {
	// 1. Check factory exists
	_, err := s.factoryRepo.GetByID(ctx, req.FactoryID)
	if err != nil {
		return fmt.Errorf("factory not found: %w", err)
	}

	// 2. If name is being changed, check it's unique
	if req.Name != "" {
		existing, err := s.factoryRepo.GetByName(ctx, req.Name)
		if err == nil && existing.ID != req.FactoryID {
			return fmt.Errorf("factory with name %q already exists", req.Name)
		}
	}

	// 3. Update
	record := &secondary.FactoryRecord{
		ID:   req.FactoryID,
		Name: req.Name,
	}
	if err := s.factoryRepo.Update(ctx, record); err != nil {
		return fmt.Errorf("failed to update factory: %w", err)
	}

	return nil
}

// DeleteFactory deletes a factory.
func (s *FactoryServiceImpl) DeleteFactory(ctx context.Context, req primary.DeleteFactoryRequest) error {
	// 1. Check factory exists
	_, err := s.factoryRepo.GetByID(ctx, req.FactoryID)
	factoryExists := err == nil

	// 2. Count workshops and commissions
	workshopCount, _ := s.factoryRepo.CountWorkshops(ctx, req.FactoryID)
	commissionCount, _ := s.factoryRepo.CountCommissions(ctx, req.FactoryID)

	// 3. Guard check
	guardCtx := corefactory.DeleteFactoryContext{
		FactoryID:       req.FactoryID,
		FactoryExists:   factoryExists,
		WorkshopCount:   workshopCount,
		CommissionCount: commissionCount,
		ForceDelete:     req.Force,
	}
	if result := corefactory.CanDeleteFactory(guardCtx); !result.Allowed {
		return result.Error()
	}

	// 4. Delete
	return s.factoryRepo.Delete(ctx, req.FactoryID)
}

// Helper methods

func (s *FactoryServiceImpl) recordToFactory(r *secondary.FactoryRecord) *primary.Factory {
	return &primary.Factory{
		ID:        r.ID,
		Name:      r.Name,
		Status:    r.Status,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// Ensure FactoryServiceImpl implements the interface
var _ primary.FactoryService = (*FactoryServiceImpl)(nil)
