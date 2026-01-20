// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/example/orc/internal/ports/secondary"
)

// TagRepository provides tag lookup operations for services.
type TagRepository struct {
	db *sql.DB
}

// NewTagRepository creates a new SQLite tag repository.
func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{db: db}
}

// GetTagByName retrieves a tag by its name.
func (r *TagRepository) GetTagByName(ctx context.Context, name string) (*secondary.TagRecord, error) {
	var tagID, tagName string
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name FROM tags WHERE name = ?",
		name,
	).Scan(&tagID, &tagName)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tag '%s' not found", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return &secondary.TagRecord{ID: tagID, Name: tagName}, nil
}
