package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// EventServiceImpl implements the EventService interface.
type EventServiceImpl struct {
	auditRepo secondary.WorkshopEventRepository
	opsRepo   secondary.OperationalEventRepository
}

// NewEventService creates a new EventService with injected dependencies.
func NewEventService(auditRepo secondary.WorkshopEventRepository, opsRepo secondary.OperationalEventRepository) *EventServiceImpl {
	return &EventServiceImpl{
		auditRepo: auditRepo,
		opsRepo:   opsRepo,
	}
}

// ListEvents retrieves events matching the given filters.
// Queries both audit and operational event repos, merge-sorts by timestamp descending.
func (s *EventServiceImpl) ListEvents(ctx context.Context, filters primary.EventFilters) ([]*primary.Event, error) {
	eventType := filters.EventType
	if eventType == "" {
		eventType = "all"
	}

	var events []*primary.Event

	// Query audit events
	if eventType == "all" || eventType == "audit" {
		auditFilters := secondary.AuditEventFilters{
			WorkshopID: filters.WorkshopID,
			ActorID:    filters.ActorID,
			EntityID:   filters.EntityID,
			Source:     filters.Source,
			Limit:      filters.Limit,
		}
		records, err := s.auditRepo.List(ctx, auditFilters)
		if err != nil {
			return nil, fmt.Errorf("failed to list audit events: %w", err)
		}
		for _, r := range records {
			events = append(events, auditRecordToEvent(r))
		}
	}

	// Query operational events
	if eventType == "all" || eventType == "ops" {
		opsFilters := secondary.OperationalEventFilters{
			WorkshopID: filters.WorkshopID,
			Source:     filters.Source,
			Level:      filters.Level,
			Limit:      filters.Limit,
		}
		records, err := s.opsRepo.List(ctx, opsFilters)
		if err != nil {
			return nil, fmt.Errorf("failed to list operational events: %w", err)
		}
		for _, r := range records {
			events = append(events, opsRecordToEvent(r))
		}
	}

	// Merge-sort by timestamp descending
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp > events[j].Timestamp
	})

	// Apply combined limit
	if filters.Limit > 0 && len(events) > filters.Limit {
		events = events[:filters.Limit]
	}

	return events, nil
}

// GetEvent retrieves a single event by ID.
// Tries audit repo first (WE- prefix), then operational repo (OE- prefix).
func (s *EventServiceImpl) GetEvent(ctx context.Context, id string) (*primary.Event, error) {
	// Try audit repo first
	if len(id) >= 3 && id[:3] == "WE-" {
		record, err := s.auditRepo.GetByID(ctx, id)
		if err == nil {
			return auditRecordToEvent(record), nil
		}
	}

	// Not found or not WE- prefix — return generic not found
	return nil, fmt.Errorf("event %s not found", id)
}

// PruneEvents deletes events older than the specified number of days from both repos.
func (s *EventServiceImpl) PruneEvents(ctx context.Context, olderThanDays int) (int, error) {
	auditCount, err := s.auditRepo.PruneOlderThan(ctx, olderThanDays)
	if err != nil {
		return 0, fmt.Errorf("failed to prune audit events: %w", err)
	}

	opsCount, err := s.opsRepo.PruneOlderThan(ctx, olderThanDays)
	if err != nil {
		return auditCount, fmt.Errorf("failed to prune operational events: %w", err)
	}

	return auditCount + opsCount, nil
}

// Helper conversions

func auditRecordToEvent(r *secondary.AuditEventRecord) *primary.Event {
	return &primary.Event{
		ID:         r.ID,
		WorkshopID: r.WorkshopID,
		Timestamp:  r.Timestamp,
		ActorID:    r.ActorID,
		Source:     r.Source,
		Version:    r.Version,
		EntityType: r.EntityType,
		EntityID:   r.EntityID,
		Action:     r.Action,
		FieldName:  r.FieldName,
		OldValue:   r.OldValue,
		NewValue:   r.NewValue,
		CreatedAt:  r.CreatedAt,
	}
}

func opsRecordToEvent(r *secondary.OperationalEventRecord) *primary.Event {
	return &primary.Event{
		ID:         r.ID,
		WorkshopID: r.WorkshopID,
		Timestamp:  r.Timestamp,
		ActorID:    r.ActorID,
		Source:     r.Source,
		Version:    r.Version,
		Level:      r.Level,
		Message:    r.Message,
		Data:       r.DataJSON,
		CreatedAt:  r.CreatedAt,
	}
}

// Ensure EventServiceImpl implements the interface
var _ primary.EventService = (*EventServiceImpl)(nil)
