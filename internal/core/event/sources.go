// Package event defines core event types and source constants for the unified event system.
// This is part of the functional core — no non-core imports.
package event

import "time"

// Source constants identify where events originate.
const (
	SourceLedger     = "ledger"
	SourcePoll       = "poll"
	SourceTmuxApply  = "tmux-apply"
	SourceDeployGlue = "deploy-glue"
	SourceWorkbench  = "workbench"
)

// Level constants for operational events.
const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

// BaseEvent is the common foundation for all events.
type BaseEvent struct {
	ID        string
	Timestamp time.Time
	Actor     string
	Workshop  string
	Source    string
	Version   string
}

// AuditEvent records a change to a domain entity.
type AuditEvent struct {
	BaseEvent
	EntityType string
	EntityID   string
	Action     string
	FieldName  string
	OldValue   string
	NewValue   string
}

// OperationalEvent records system-level activity (logs, diagnostics, lifecycle).
type OperationalEvent struct {
	BaseEvent
	Level   string
	Message string
	Data    map[string]string
}
