package db

const schemaSQL = `
-- Missions (Strategic work streams)
CREATE TABLE IF NOT EXISTS missions (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL CHECK(status IN ('active', 'paused', 'complete', 'archived')) DEFAULT 'active',
	pinned INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	completed_at DATETIME
);

-- Epics (Top-level work containers)
CREATE TABLE IF NOT EXISTS epics (
	id TEXT PRIMARY KEY,
	mission_id TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL CHECK(status IN ('ready', 'design', 'implement', 'deploy', 'blocked', 'paused', 'complete')) DEFAULT 'ready',
	priority TEXT CHECK(priority IN ('low', 'medium', 'high')),
	assigned_grove_id TEXT,
	context_ref TEXT,
	pinned INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	completed_at DATETIME,
	FOREIGN KEY (mission_id) REFERENCES missions(id),
	FOREIGN KEY (assigned_grove_id) REFERENCES groves(id)
);

-- Rabbit Holes (Optional grouping layer within epics)
CREATE TABLE IF NOT EXISTS rabbit_holes (
	id TEXT PRIMARY KEY,
	epic_id TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL CHECK(status IN ('ready', 'design', 'implement', 'deploy', 'blocked', 'paused', 'complete')) DEFAULT 'ready',
	priority TEXT CHECK(priority IN ('low', 'medium', 'high')),
	pinned INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	completed_at DATETIME,
	FOREIGN KEY (epic_id) REFERENCES epics(id) ON DELETE CASCADE
);

-- Tasks (Atomic units of work)
CREATE TABLE IF NOT EXISTS tasks (
	id TEXT PRIMARY KEY,
	epic_id TEXT,
	rabbit_hole_id TEXT,
	mission_id TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT,
	type TEXT CHECK(type IN ('research', 'implementation', 'fix', 'documentation', 'maintenance')),
	status TEXT NOT NULL CHECK(status IN ('ready', 'design', 'implement', 'deploy', 'blocked', 'paused', 'complete')) DEFAULT 'ready',
	priority TEXT CHECK(priority IN ('low', 'medium', 'high')),
	assigned_grove_id TEXT,
	context_ref TEXT,
	pinned INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	claimed_at DATETIME,
	completed_at DATETIME,
	FOREIGN KEY (epic_id) REFERENCES epics(id) ON DELETE CASCADE,
	FOREIGN KEY (rabbit_hole_id) REFERENCES rabbit_holes(id) ON DELETE CASCADE,
	FOREIGN KEY (mission_id) REFERENCES missions(id),
	FOREIGN KEY (assigned_grove_id) REFERENCES groves(id),
	CHECK ((epic_id IS NOT NULL AND rabbit_hole_id IS NULL) OR (epic_id IS NULL AND rabbit_hole_id IS NOT NULL))
);

-- Groves (Physical workspaces) - Mission-level worktrees
CREATE TABLE IF NOT EXISTS groves (
	id TEXT PRIMARY KEY,
	mission_id TEXT NOT NULL,
	name TEXT NOT NULL,
	path TEXT NOT NULL UNIQUE,
	repos TEXT,
	status TEXT NOT NULL CHECK(status IN ('active', 'archived')) DEFAULT 'active',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (mission_id) REFERENCES missions(id)
);

-- Handoffs (Claude-to-Claude context transfer)
CREATE TABLE IF NOT EXISTS handoffs (
	id TEXT PRIMARY KEY,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	handoff_note TEXT NOT NULL,
	active_mission_id TEXT,
	active_work_orders TEXT,
	active_grove_id TEXT,
	todos_snapshot TEXT,
	FOREIGN KEY (active_mission_id) REFERENCES missions(id),
	FOREIGN KEY (active_grove_id) REFERENCES groves(id)
);

-- Messages (Agent mail system)
CREATE TABLE IF NOT EXISTS messages (
	id TEXT PRIMARY KEY,
	sender TEXT NOT NULL,
	recipient TEXT NOT NULL,
	subject TEXT NOT NULL,
	body TEXT NOT NULL,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	read INTEGER DEFAULT 0,
	mission_id TEXT NOT NULL,
	FOREIGN KEY (mission_id) REFERENCES missions(id)
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_missions_status ON missions(status);
CREATE INDEX IF NOT EXISTS idx_epics_mission ON epics(mission_id);
CREATE INDEX IF NOT EXISTS idx_epics_status ON epics(status);
CREATE INDEX IF NOT EXISTS idx_epics_grove ON epics(assigned_grove_id);
CREATE INDEX IF NOT EXISTS idx_rabbit_holes_epic ON rabbit_holes(epic_id);
CREATE INDEX IF NOT EXISTS idx_rabbit_holes_status ON rabbit_holes(status);
CREATE INDEX IF NOT EXISTS idx_tasks_epic ON tasks(epic_id);
CREATE INDEX IF NOT EXISTS idx_tasks_rabbit_hole ON tasks(rabbit_hole_id);
CREATE INDEX IF NOT EXISTS idx_tasks_mission ON tasks(mission_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_grove ON tasks(assigned_grove_id);
CREATE INDEX IF NOT EXISTS idx_groves_mission ON groves(mission_id);
CREATE INDEX IF NOT EXISTS idx_groves_status ON groves(status);
CREATE INDEX IF NOT EXISTS idx_handoffs_created ON handoffs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_recipient ON messages(recipient, read);
CREATE INDEX IF NOT EXISTS idx_messages_mission ON messages(mission_id);
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp DESC);
`

// InitSchema creates the database schema
func InitSchema() error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	// Check if schema_version table exists to determine if this is a fresh install
	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_version'").Scan(&tableCount)
	if err != nil {
		return err
	}

	if tableCount == 0 {
		// Fresh install - check if we have old schema tables
		var oldTableCount int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('operations', 'expeditions')").Scan(&oldTableCount)
		if err != nil {
			return err
		}

		if oldTableCount > 0 {
			// Old schema exists - run migrations
			return RunMigrations()
		} else {
			// Completely fresh install - create new schema directly
			_, err = db.Exec(schemaSQL)
			return err
		}
	}

	// schema_version table exists - run any pending migrations
	return RunMigrations()
}
