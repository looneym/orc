package sqlite

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/example/orc/internal/core/event"
	"github.com/example/orc/internal/ctxutil"
	"github.com/example/orc/internal/ports/secondary"
)

// EventWriterAdapter implements secondary.EventWriter using WorkshopEventRepository
// and OperationalEventRepository.
type EventWriterAdapter struct {
	workshopEventRepo    secondary.WorkshopEventRepository
	operationalEventRepo secondary.OperationalEventRepository
	workbenchRepo        secondary.WorkbenchRepository
	versionString        string
}

// NewEventWriterAdapter creates a new EventWriterAdapter.
// workbenchRepo is used to resolve workshop from workbench actors.
// versionString is stamped on every event for forward compatibility.
func NewEventWriterAdapter(
	workshopEventRepo secondary.WorkshopEventRepository,
	operationalEventRepo secondary.OperationalEventRepository,
	workbenchRepo secondary.WorkbenchRepository,
	versionString string,
) *EventWriterAdapter {
	return &EventWriterAdapter{
		workshopEventRepo:    workshopEventRepo,
		operationalEventRepo: operationalEventRepo,
		workbenchRepo:        workbenchRepo,
		versionString:        versionString,
	}
}

// EmitAuditCreate emits an audit event for a create operation.
func (w *EventWriterAdapter) EmitAuditCreate(ctx context.Context, entityType, entityID string) error {
	return w.writeAudit(ctx, entityType, entityID, "create", "", "", "")
}

// EmitAuditUpdate emits an audit event for an update operation.
func (w *EventWriterAdapter) EmitAuditUpdate(ctx context.Context, entityType, entityID, fieldName, oldValue, newValue string) error {
	return w.writeAudit(ctx, entityType, entityID, "update", fieldName, oldValue, newValue)
}

// EmitAuditDelete emits an audit event for a delete operation.
func (w *EventWriterAdapter) EmitAuditDelete(ctx context.Context, entityType, entityID string) error {
	return w.writeAudit(ctx, entityType, entityID, "delete", "", "", "")
}

// EmitOperational emits an operational event.
func (w *EventWriterAdapter) EmitOperational(ctx context.Context, source, level, message string, data map[string]string) error {
	actorID := ctxutil.ActorFromContext(ctx)
	workshopID := w.resolveWorkshop(ctx, actorID)

	id, err := w.operationalEventRepo.GetNextID(ctx)
	if err != nil {
		return err
	}

	var dataJSON string
	if len(data) > 0 {
		b, err := json.Marshal(data)
		if err != nil {
			return err
		}
		dataJSON = string(b)
	}

	record := &secondary.OperationalEventRecord{
		ID:         id,
		WorkshopID: workshopID,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		ActorID:    actorID,
		Source:     source,
		Version:    w.versionString,
		Level:      level,
		Message:    message,
		DataJSON:   dataJSON,
	}

	return w.operationalEventRepo.Create(ctx, record)
}

// writeAudit writes an audit event with common logic.
func (w *EventWriterAdapter) writeAudit(ctx context.Context, entityType, entityID, action, fieldName, oldValue, newValue string) error {
	actorID := ctxutil.ActorFromContext(ctx)
	workshopID := w.resolveWorkshop(ctx, actorID)

	if workshopID == "" {
		// No workshop context — skip audit logging.
		// This happens for operations outside workshop scope.
		return nil
	}

	id, err := w.workshopEventRepo.GetNextID(ctx)
	if err != nil {
		return err
	}

	record := &secondary.AuditEventRecord{
		ID:         id,
		WorkshopID: workshopID,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		ActorID:    actorID,
		Source:     event.SourceLedger,
		Version:    w.versionString,
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		FieldName:  fieldName,
		OldValue:   oldValue,
		NewValue:   newValue,
	}

	return w.workshopEventRepo.Create(ctx, record)
}

// resolveWorkshop resolves the workshop ID from the actor.
// For BENCH-xxx actors, looks up the workbench's workshop.
// Returns empty string if workshop cannot be resolved.
func (w *EventWriterAdapter) resolveWorkshop(ctx context.Context, actorID string) string {
	if actorID == "" {
		return ""
	}

	// Parse actor to find workbench.
	// Actor IDs are like "IMP-BENCH-014" or just "GOBLIN".
	if strings.Contains(actorID, "BENCH-") {
		parts := strings.Split(actorID, "-")
		for i, p := range parts {
			if p == "BENCH" && i+1 < len(parts) {
				workbenchID := "BENCH-" + parts[i+1]
				if w.workbenchRepo != nil {
					bench, err := w.workbenchRepo.GetByID(ctx, workbenchID)
					if err == nil && bench != nil {
						return bench.WorkshopID
					}
				}
				return ""
			}
		}
	}

	return ""
}

// Ensure EventWriterAdapter implements the interface.
var _ secondary.EventWriter = (*EventWriterAdapter)(nil)
