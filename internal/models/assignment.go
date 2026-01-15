package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// EpicAssignment represents an epic (with children) assigned to a grove
type EpicAssignment struct {
	EpicID          string            `json:"epic_id"`
	EpicTitle       string            `json:"epic_title"`
	EpicDescription string            `json:"epic_description"`
	MissionID       string            `json:"mission_id"`
	AssignedBy      string            `json:"assigned_by"`
	AssignedAt      string            `json:"assigned_at"`
	Status          string            `json:"status"` // assigned, in_progress, complete
	Structure       string            `json:"structure"` // "tasks" or "rabbit_holes"
	RabbitHoles     []RabbitHoleInfo  `json:"rabbit_holes,omitempty"`
	Tasks           []TaskInfo        `json:"tasks,omitempty"`
	ChildWorkOrders []ChildWorkOrderInfo `json:"child_work_orders,omitempty"` // Legacy format
	Progress        ProgressInfo      `json:"progress"`
}

// RabbitHoleInfo represents a rabbit hole within an epic assignment
type RabbitHoleInfo struct {
	RabbitHoleID string     `json:"rabbit_hole_id"`
	Title        string     `json:"title"`
	Status       string     `json:"status"`
	Tasks        []TaskInfo `json:"tasks"`
}

// TaskInfo represents a task within an epic or rabbit hole
type TaskInfo struct {
	TaskID      string `json:"task_id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Type        string `json:"type,omitempty"`
	ClaimedAt   string `json:"claimed_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
}

// ChildWorkOrderInfo represents a child work order within an epic (LEGACY)
type ChildWorkOrderInfo struct {
	WorkOrderID string `json:"work_order_id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	ClaimedAt   string `json:"claimed_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
}

// ProgressInfo tracks progress across epic's children
type ProgressInfo struct {
	TotalRabbitHoles int `json:"total_rabbit_holes,omitempty"`
	TotalTasks       int `json:"total_tasks"`
	CompletedTasks   int `json:"completed_tasks"`
	InProgressTasks  int `json:"in_progress_tasks"`
	ReadyTasks       int `json:"ready_tasks"`
	// Legacy fields
	TotalChildren int `json:"total_children,omitempty"`
	Completed     int `json:"completed,omitempty"`
	InProgress    int `json:"in_progress,omitempty"`
	Ready         int `json:"ready,omitempty"`
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

// WriteEpicAssignmentWithRabbitHoles writes an epic with rabbit holes to assignment file
func WriteEpicAssignmentWithRabbitHoles(groveDir string, epic *Epic, rabbitHoles []*RabbitHole, assignedBy string) error {
	// Build rabbit hole info with tasks
	var rhInfos []RabbitHoleInfo
	totalTasks := 0
	completedTasks := 0
	inProgressTasks := 0
	readyTasks := 0

	for _, rh := range rabbitHoles {
		// Get tasks under this rabbit hole
		tasks, err := GetRabbitHoleTasks(rh.ID)
		if err != nil {
			return fmt.Errorf("failed to get tasks for rabbit hole %s: %w", rh.ID, err)
		}

		var taskInfos []TaskInfo
		for _, task := range tasks {
			taskInfo := TaskInfo{
				TaskID: task.ID,
				Title:  task.Title,
				Status: task.Status,
			}
			if task.Type.Valid {
				taskInfo.Type = task.Type.String
			}
			if task.ClaimedAt.Valid {
				taskInfo.ClaimedAt = task.ClaimedAt.Time.Format(time.RFC3339)
			}
			if task.CompletedAt.Valid {
				taskInfo.CompletedAt = task.CompletedAt.Time.Format(time.RFC3339)
			}

			taskInfos = append(taskInfos, taskInfo)
			totalTasks++

			// Update progress counts
			switch task.Status {
			case "complete":
				completedTasks++
			case "implement", "in_progress":
				inProgressTasks++
			case "ready":
				readyTasks++
			}
		}

		rhInfo := RabbitHoleInfo{
			RabbitHoleID: rh.ID,
			Title:        rh.Title,
			Status:       rh.Status,
			Tasks:        taskInfos,
		}
		rhInfos = append(rhInfos, rhInfo)
	}

	assignment := &EpicAssignment{
		EpicID:          epic.ID,
		EpicTitle:       epic.Title,
		EpicDescription: epic.Description.String,
		MissionID:       epic.MissionID,
		AssignedBy:      assignedBy,
		AssignedAt:      time.Now().Format(time.RFC3339),
		Status:          "assigned",
		Structure:       "rabbit_holes",
		RabbitHoles:     rhInfos,
		Progress: ProgressInfo{
			TotalRabbitHoles: len(rabbitHoles),
			TotalTasks:       totalTasks,
			CompletedTasks:   completedTasks,
			InProgressTasks:  inProgressTasks,
			ReadyTasks:       readyTasks,
		},
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

// WriteEpicAssignmentWithTasks writes an epic with direct tasks to assignment file
func WriteEpicAssignmentWithTasks(groveDir string, epic *Epic, tasks []*Task, assignedBy string) error {
	// Build task info
	var taskInfos []TaskInfo
	completedTasks := 0
	inProgressTasks := 0
	readyTasks := 0

	for _, task := range tasks {
		taskInfo := TaskInfo{
			TaskID: task.ID,
			Title:  task.Title,
			Status: task.Status,
		}
		if task.Type.Valid {
			taskInfo.Type = task.Type.String
		}
		if task.ClaimedAt.Valid {
			taskInfo.ClaimedAt = task.ClaimedAt.Time.Format(time.RFC3339)
		}
		if task.CompletedAt.Valid {
			taskInfo.CompletedAt = task.CompletedAt.Time.Format(time.RFC3339)
		}

		taskInfos = append(taskInfos, taskInfo)

		// Update progress counts
		switch task.Status {
		case "complete":
			completedTasks++
		case "implement", "in_progress":
			inProgressTasks++
		case "ready":
			readyTasks++
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
		Structure:       "tasks",
		Tasks:           taskInfos,
		Progress: ProgressInfo{
			TotalTasks:      len(tasks),
			CompletedTasks:  completedTasks,
			InProgressTasks: inProgressTasks,
			ReadyTasks:      readyTasks,
		},
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
