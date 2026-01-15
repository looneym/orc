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

// EpicAssignment represents an epic (with children) assigned to a grove
type EpicAssignment struct {
	EpicID          string              `json:"epic_id"`
	EpicTitle       string              `json:"epic_title"`
	EpicDescription string              `json:"epic_description"`
	MissionID       string              `json:"mission_id"`
	AssignedBy      string              `json:"assigned_by"`
	AssignedAt      string              `json:"assigned_at"`
	Status          string              `json:"status"` // assigned, in_progress, complete
	ChildWorkOrders []ChildWorkOrderInfo `json:"child_work_orders"`
	Progress        ProgressInfo        `json:"progress"`
}

// ChildWorkOrderInfo represents a child work order within an epic
type ChildWorkOrderInfo struct {
	WorkOrderID string `json:"work_order_id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	ClaimedAt   string `json:"claimed_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
}

// ProgressInfo tracks progress across epic's child work orders
type ProgressInfo struct {
	TotalChildren int `json:"total_children"`
	Completed     int `json:"completed"`
	InProgress    int `json:"in_progress"`
	Ready         int `json:"ready"`
}

// WriteEpicAssignment writes an epic assignment to a grove's .orc directory
func WriteEpicAssignment(groveDir string, epic *WorkOrder, children []*WorkOrder, assignedBy string) error {
	// Build child work order info
	var childInfos []ChildWorkOrderInfo
	progress := ProgressInfo{
		TotalChildren: len(children),
		Completed:     0,
		InProgress:    0,
		Ready:         0,
	}

	for _, child := range children {
		childInfo := ChildWorkOrderInfo{
			WorkOrderID: child.ID,
			Title:       child.Title,
			Status:      child.Status,
		}

		if child.ClaimedAt.Valid {
			childInfo.ClaimedAt = child.ClaimedAt.Time.Format(time.RFC3339)
		}
		if child.CompletedAt.Valid {
			childInfo.CompletedAt = child.CompletedAt.Time.Format(time.RFC3339)
		}

		childInfos = append(childInfos, childInfo)

		// Update progress counts
		switch child.Status {
		case "complete":
			progress.Completed++
		case "implement", "in_progress":
			progress.InProgress++
		case "ready":
			progress.Ready++
		}
	}

	assignment := &EpicAssignment{
		EpicID:          epic.ID,
		EpicTitle:       epic.Title,
		EpicDescription: epic.Description.String,
		MissionID:       epic.MissionID,
		AssignedBy:      assignedBy,
		AssignedAt:      time.Now().Format(time.RFC3339),
		Status:          "assigned",
		ChildWorkOrders: childInfos,
		Progress:        progress,
	}

	orcDir := filepath.Join(groveDir, ".orc")
	if err := os.MkdirAll(orcDir, 0755); err != nil {
		return fmt.Errorf("failed to create .orc directory: %w", err)
	}

	path := filepath.Join(orcDir, "assigned-work.json")
	data, err := json.MarshalIndent(assignment, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal epic assignment: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write assignment file: %w", err)
	}

	return nil
}

// ReadEpicAssignment reads an epic assignment from a grove's .orc directory
func ReadEpicAssignment(groveDir string) (*EpicAssignment, error) {
	path := filepath.Join(groveDir, ".orc", "assigned-work.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read assignment file: %w", err)
	}

	var assignment EpicAssignment
	if err := json.Unmarshal(data, &assignment); err != nil {
		return nil, fmt.Errorf("failed to parse epic assignment: %w", err)
	}

	return &assignment, nil
}

// UpdateEpicAssignmentStatus updates the status of an epic assignment file
func UpdateEpicAssignmentStatus(groveDir, status string) error {
	assignment, err := ReadEpicAssignment(groveDir)
	if err != nil {
		return err
	}

	assignment.Status = status

	path := filepath.Join(groveDir, ".orc", "assigned-work.json")
	data, err := json.MarshalIndent(assignment, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal epic assignment: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write assignment file: %w", err)
	}

	return nil
}
