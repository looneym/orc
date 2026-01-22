package models

import (
	"database/sql"
	"time"
)

type Handoff struct {
	ID                    string
	CreatedAt             time.Time
	HandoffNote           string
	ActiveComcommissionID sql.NullString
	ActiveWorkbenchID     sql.NullString
	TodosSnapshot         sql.NullString
}
