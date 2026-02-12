package db

import (
	_ "embed"
)

// SchemaSQL is the complete modern schema for fresh ORC installs.
// This schema reflects the current state after all migrations.
//
// # Schema Drift Protection
//
// This is the SINGLE SOURCE OF TRUTH for the database schema. All tests use
// this schema via GetSchemaSQL(), which provides two layers of protection:
//
//  1. No hardcoded schemas: `make schema-check` fails if any test file contains
//     CREATE TABLE statements. Tests must use db.GetSchemaSQL() instead.
//
//  2. Immediate failure on drift: If repository code references a column that
//     doesn't exist in this schema, tests fail immediately with "no such column".
//     This catches drift at development time, not production.
//
// # Keeping Schema in Sync
//
// Schema changes use the Atlas workflow:
//
//  1. Edit internal/db/schema.sql
//  2. Run: make schema-diff   (preview changes)
//  3. Run: make schema-apply  (apply to local DB)
//  4. Run: make test          (verify alignment)
//
//go:embed schema.sql
var SchemaSQL string

// InitSchema creates the database schema.
// The schema.sql uses IF NOT EXISTS so this is idempotent.
func InitSchema() error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	_, err = db.Exec(SchemaSQL)
	return err
}

// GetSchemaSQL returns the authoritative schema SQL for use by tests.
// Tests should use this instead of hardcoding their own schema to prevent drift.
func GetSchemaSQL() string {
	return SchemaSQL
}
