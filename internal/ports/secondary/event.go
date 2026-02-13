package secondary

import "context"

// EventWriter defines the interface for emitting audit and operational events.
// Implementations extract actor from context and handle workshop resolution.
type EventWriter interface {
	// EmitAuditCreate emits an audit event for a create operation.
	EmitAuditCreate(ctx context.Context, entityType, entityID string) error

	// EmitAuditUpdate emits an audit event for an update operation.
	// fieldName, oldValue, newValue describe what changed.
	EmitAuditUpdate(ctx context.Context, entityType, entityID, fieldName, oldValue, newValue string) error

	// EmitAuditDelete emits an audit event for a delete operation.
	EmitAuditDelete(ctx context.Context, entityType, entityID string) error

	// EmitOperational emits an operational event (logs, diagnostics, lifecycle).
	EmitOperational(ctx context.Context, source, level, message string, data map[string]string) error
}
