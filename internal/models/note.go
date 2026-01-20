package models

import (
	"database/sql"
	"time"
)

type Note struct {
	ID               string
	MissionID        string
	Title            string
	Content          sql.NullString
	Type             sql.NullString // learning, concern, finding, frq, bug, investigation_report
	ShipmentID       sql.NullString
	InvestigationID  sql.NullString
	ConclaveID       sql.NullString
	TomeID           sql.NullString
	Pinned           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	PromotedFromID   sql.NullString
	PromotedFromType sql.NullString
}

// Note types
const (
	NoteTypeLearning            = "learning"
	NoteTypeConcern             = "concern"
	NoteTypeFinding             = "finding"
	NoteTypeFRQ                 = "frq"
	NoteTypeBug                 = "bug"
	NoteTypeInvestigationReport = "investigation_report"
)
