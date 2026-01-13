package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/looneym/orc/internal/db"
)

type Grove struct {
	ID           string
	Path         string
	Repos        sql.NullString
	ExpeditionID sql.NullString
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CreateGrove creates a new grove
func CreateGrove(id, path string, expeditionID *string) (*Grove, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	var expID sql.NullString
	if expeditionID != nil {
		expID = sql.NullString{String: *expeditionID, Valid: true}
	}

	_, err = database.Exec(
		"INSERT INTO groves (id, path, expedition_id, status) VALUES (?, ?, ?, ?)",
		id, path, expID, "active",
	)
	if err != nil {
		return nil, err
	}

	return GetGrove(id)
}

// GetGrove retrieves a grove by ID
func GetGrove(id string) (*Grove, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	grove := &Grove{}
	err = database.QueryRow(
		"SELECT id, path, repos, expedition_id, status, created_at, updated_at FROM groves WHERE id = ?",
		id,
	).Scan(&grove.ID, &grove.Path, &grove.Repos, &grove.ExpeditionID, &grove.Status, &grove.CreatedAt, &grove.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return grove, nil
}

// ListGroves retrieves all groves
func ListGroves() ([]*Grove, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := database.Query(
		"SELECT id, path, repos, expedition_id, status, created_at, updated_at FROM groves ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groves []*Grove
	for rows.Next() {
		grove := &Grove{}
		err := rows.Scan(&grove.ID, &grove.Path, &grove.Repos, &grove.ExpeditionID, &grove.Status, &grove.CreatedAt, &grove.UpdatedAt)
		if err != nil {
			return nil, err
		}
		groves = append(groves, grove)
	}

	return groves, nil
}

// GetGrovesByExpedition retrieves all groves for an expedition
func GetGrovesByExpedition(expeditionID string) ([]*Grove, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := database.Query(
		"SELECT id, path, repos, expedition_id, status, created_at, updated_at FROM groves WHERE expedition_id = ? ORDER BY created_at DESC",
		expeditionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groves []*Grove
	for rows.Next() {
		grove := &Grove{}
		err := rows.Scan(&grove.ID, &grove.Path, &grove.Repos, &grove.ExpeditionID, &grove.Status, &grove.CreatedAt, &grove.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning grove: %w", err)
		}
		groves = append(groves, grove)
	}

	return groves, nil
}
