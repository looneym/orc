package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// WatchdogServiceImpl implements the WatchdogService interface.
type WatchdogServiceImpl struct {
	watchdogRepo secondary.WatchdogRepository
}

// NewWatchdogService creates a new WatchdogService with injected dependencies.
func NewWatchdogService(watchdogRepo secondary.WatchdogRepository) *WatchdogServiceImpl {
	return &WatchdogServiceImpl{
		watchdogRepo: watchdogRepo,
	}
}

// GetWatchdog retrieves a watchdog by ID.
func (s *WatchdogServiceImpl) GetWatchdog(ctx context.Context, watchdogID string) (*primary.Watchdog, error) {
	record, err := s.watchdogRepo.GetByID(ctx, watchdogID)
	if err != nil {
		return nil, err
	}
	return s.recordToWatchdog(record), nil
}

// GetWatchdogByWorkbench retrieves a watchdog by workbench ID.
func (s *WatchdogServiceImpl) GetWatchdogByWorkbench(ctx context.Context, workbenchID string) (*primary.Watchdog, error) {
	record, err := s.watchdogRepo.GetByWorkbench(ctx, workbenchID)
	if err != nil {
		return nil, err
	}
	return s.recordToWatchdog(record), nil
}

// ListWatchdogs lists watchdogs with optional filters.
func (s *WatchdogServiceImpl) ListWatchdogs(ctx context.Context, filters primary.WatchdogFilters) ([]*primary.Watchdog, error) {
	records, err := s.watchdogRepo.List(ctx, secondary.WatchdogFilters{
		WorkbenchID: filters.WorkbenchID,
		Status:      filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list watchdogs: %w", err)
	}

	watchdogs := make([]*primary.Watchdog, len(records))
	for i, r := range records {
		watchdogs[i] = s.recordToWatchdog(r)
	}
	return watchdogs, nil
}

// Helper methods

func (s *WatchdogServiceImpl) recordToWatchdog(r *secondary.WatchdogRecord) *primary.Watchdog {
	return &primary.Watchdog{
		ID:          r.ID,
		WorkbenchID: r.WorkbenchID,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// Ensure WatchdogServiceImpl implements the interface
var _ primary.WatchdogService = (*WatchdogServiceImpl)(nil)
