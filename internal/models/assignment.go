package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// WorkAssignment represents a work order assigned to a grove
type WorkAssignment struct {
	WorkOrderID string `json:"work_order_id"`
	AssignedBy  string `json:"assigned_by"`
	AssignedAt  string `json:"assigned_at"`
	Title       string `json:"title"`
	Description string `json:"description"`
	MissionID   string `json:"mission_id"`
	Status      string `json:"status"` // assigned, in_progress, complete
}

// WriteAssignment writes a work assignment to a grove's .orc directory
func WriteAssignment(groveDir string, wo *WorkOrder, assignedBy string) error {
	assignment := &WorkAssignment{
		WorkOrderID: wo.ID,
		AssignedBy:  assignedBy,
		AssignedAt:  time.Now().Format(time.RFC3339),
		Title:       wo.Title,
		Description: wo.Description.String,
		MissionID:   wo.MissionID,
		Status:      "assigned",
	}

	orcDir := filepath.Join(groveDir, ".orc")
	if err := os.MkdirAll(orcDir, 0755); err != nil {
		return fmt.Errorf("failed to create .orc directory: %w", err)
	}

	path := filepath.Join(orcDir, "assigned-work.json")
	data, err := json.MarshalIndent(assignment, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal assignment: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write assignment file: %w", err)
	}

	return nil
}

// ReadAssignment reads a work assignment from a grove's .orc directory
func ReadAssignment(groveDir string) (*WorkAssignment, error) {
	path := filepath.Join(groveDir, ".orc", "assigned-work.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read assignment file: %w", err)
	}

	var assignment WorkAssignment
	if err := json.Unmarshal(data, &assignment); err != nil {
		return nil, fmt.Errorf("failed to parse assignment: %w", err)
	}

	return &assignment, nil
}

// UpdateAssignmentStatus updates the status of an assignment file
func UpdateAssignmentStatus(groveDir, status string) error {
	assignment, err := ReadAssignment(groveDir)
	if err != nil {
		return err
	}

	assignment.Status = status

	path := filepath.Join(groveDir, ".orc", "assigned-work.json")
	data, err := json.MarshalIndent(assignment, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal assignment: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write assignment file: %w", err)
	}

	return nil
}
