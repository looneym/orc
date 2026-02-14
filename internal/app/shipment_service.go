package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	coreshipment "github.com/example/orc/internal/core/shipment"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// ShipmentServiceImpl implements the ShipmentService interface.
type ShipmentServiceImpl struct {
	shipmentRepo secondary.ShipmentRepository
	taskRepo     secondary.TaskRepository
	noteService  primary.NoteService
	transactor   secondary.Transactor
}

// NewShipmentService creates a new ShipmentService with injected dependencies.
func NewShipmentService(
	shipmentRepo secondary.ShipmentRepository,
	taskRepo secondary.TaskRepository,
	noteService primary.NoteService,
	transactor secondary.Transactor,
) *ShipmentServiceImpl {
	return &ShipmentServiceImpl{
		shipmentRepo: shipmentRepo,
		taskRepo:     taskRepo,
		noteService:  noteService,
		transactor:   transactor,
	}
}

// CreateShipment creates a new shipment for a commission.
func (s *ShipmentServiceImpl) CreateShipment(ctx context.Context, req primary.CreateShipmentRequest) (*primary.CreateShipmentResponse, error) {
	// Validate commission exists
	exists, err := s.shipmentRepo.CommissionExists(ctx, req.CommissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate commission: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("commission %s not found", req.CommissionID)
	}

	var nextID string
	err = s.transactor.WithImmediateTx(ctx, func(txCtx context.Context) error {
		// Get next ID
		var err error
		nextID, err = s.shipmentRepo.GetNextID(txCtx)
		if err != nil {
			return fmt.Errorf("failed to generate shipment ID: %w", err)
		}

		// Generate branch name if repo is specified
		var branch string
		if req.RepoID != "" {
			if req.Branch != "" {
				branch = req.Branch // Use provided branch name
			} else {
				// Auto-generate branch name: {initials}/SHIP-{id}-{slug}
				branch = GenerateShipmentBranchName(UserInitials, nextID, req.Title)
			}
		}

		// Create record - shipments go directly under commissions
		record := &secondary.ShipmentRecord{
			ID:           nextID,
			CommissionID: req.CommissionID,
			Title:        req.Title,
			Description:  req.Description,
			RepoID:       req.RepoID,
			Branch:       branch,
		}

		if err := s.shipmentRepo.Create(txCtx, record); err != nil {
			return fmt.Errorf("failed to create shipment: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Fetch created shipment
	created, err := s.shipmentRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created shipment: %w", err)
	}

	return &primary.CreateShipmentResponse{
		ShipmentID: created.ID,
		Shipment:   s.recordToShipment(created),
	}, nil
}

// GetShipment retrieves a shipment by ID.
func (s *ShipmentServiceImpl) GetShipment(ctx context.Context, shipmentID string) (*primary.Shipment, error) {
	record, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	return s.recordToShipment(record), nil
}

// ListShipments lists shipments with optional filters.
func (s *ShipmentServiceImpl) ListShipments(ctx context.Context, filters primary.ShipmentFilters) ([]*primary.Shipment, error) {
	records, err := s.shipmentRepo.List(ctx, secondary.ShipmentFilters{
		CommissionID: filters.CommissionID,
		Status:       filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list shipments: %w", err)
	}

	shipments := make([]*primary.Shipment, len(records))
	for i, r := range records {
		shipments[i] = s.recordToShipment(r)
	}
	return shipments, nil
}

// CloseShipment marks a shipment as closed.
// If force is true, closes even if tasks are not closed.
// Closes any type=spec notes attached to this shipment with reason "resolved".
func (s *ShipmentServiceImpl) CloseShipment(ctx context.Context, shipmentID string, force bool) error {
	record, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return err
	}

	// Get tasks for this shipment
	taskRecords, err := s.taskRepo.List(ctx, secondary.TaskFilters{ShipmentID: shipmentID})
	if err != nil {
		return fmt.Errorf("failed to get tasks for shipment: %w", err)
	}

	// Build task summaries for guard
	tasks := make([]coreshipment.TaskSummary, len(taskRecords))
	for i, t := range taskRecords {
		tasks[i] = coreshipment.TaskSummary{
			ID:     t.ID,
			Status: t.Status,
		}
	}

	// Guard: check all close preconditions
	guardCtx := coreshipment.CloseShipmentContext{
		ShipmentID:      shipmentID,
		IsPinned:        record.Pinned,
		Tasks:           tasks,
		ForceCompletion: force,
	}
	if result := coreshipment.CanCloseShipment(guardCtx); !result.Allowed {
		return result.Error()
	}

	// Update shipment status to closed
	if err := s.shipmentRepo.UpdateStatus(ctx, shipmentID, "closed", true); err != nil {
		return err
	}

	// Close any spec notes attached to this shipment
	if s.noteService != nil {
		notes, err := s.noteService.GetNotesByContainer(ctx, "shipment", shipmentID)
		if err != nil {
			fmt.Printf("Warning: failed to query notes for shipment %s: %v\n", shipmentID, err)
			return nil
		}
		for _, note := range notes {
			if note.Type == "spec" && note.Status != "closed" {
				closeReq := primary.CloseNoteRequest{
					NoteID: note.ID,
					Reason: "resolved",
				}
				if err := s.noteService.CloseNote(ctx, closeReq); err != nil {
					fmt.Printf("Warning: failed to close spec note %s: %v\n", note.ID, err)
				}
			}
		}
	}

	return nil
}

// CompleteShipment is an alias for CloseShipment for backwards compatibility.
func (s *ShipmentServiceImpl) CompleteShipment(ctx context.Context, shipmentID string, force bool) error {
	return s.CloseShipment(ctx, shipmentID, force)
}

// UpdateShipment updates a shipment's title, description, and/or branch.
func (s *ShipmentServiceImpl) UpdateShipment(ctx context.Context, req primary.UpdateShipmentRequest) error {
	record := &secondary.ShipmentRecord{
		ID:          req.ShipmentID,
		Title:       req.Title,
		Description: req.Description,
		Branch:      req.Branch,
	}
	return s.shipmentRepo.Update(ctx, record)
}

// UpdateStatus sets a shipment's status directly.
func (s *ShipmentServiceImpl) UpdateStatus(ctx context.Context, shipmentID, status string) error {
	return s.shipmentRepo.UpdateStatus(ctx, shipmentID, status, false)
}

// SetStatus sets a shipment's status with escape hatch protection.
// If force is true, allows backwards transitions.
func (s *ShipmentServiceImpl) SetStatus(ctx context.Context, shipmentID, status string, force bool) error {
	record, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return err
	}

	// Guard: check for backwards transitions
	guardCtx := coreshipment.OverrideStatusContext{
		ShipmentID:    shipmentID,
		CurrentStatus: record.Status,
		NewStatus:     status,
		Force:         force,
	}
	if result := coreshipment.CanOverrideStatus(guardCtx); !result.Allowed {
		return result.Error()
	}

	// Set completed flag if transitioning to closed
	setCompleted := status == "closed"

	return s.shipmentRepo.UpdateStatus(ctx, shipmentID, status, setCompleted)
}

// PinShipment pins a shipment.
func (s *ShipmentServiceImpl) PinShipment(ctx context.Context, shipmentID string) error {
	return s.shipmentRepo.Pin(ctx, shipmentID)
}

// UnpinShipment unpins a shipment.
func (s *ShipmentServiceImpl) UnpinShipment(ctx context.Context, shipmentID string) error {
	return s.shipmentRepo.Unpin(ctx, shipmentID)
}

// AssignShipmentToWorkbench assigns a shipment to a workbench.
func (s *ShipmentServiceImpl) AssignShipmentToWorkbench(ctx context.Context, shipmentID, workbenchID string) error {
	// Verify shipment exists
	_, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return err
	}

	// Check if workbench is already assigned to another shipment
	otherShipmentID, err := s.shipmentRepo.WorkbenchAssignedToOther(ctx, workbenchID, shipmentID)
	if err != nil {
		return fmt.Errorf("failed to check workbench assignment: %w", err)
	}
	if otherShipmentID != "" {
		return fmt.Errorf("workbench already assigned to shipment %s", otherShipmentID)
	}

	// Assign workbench to shipment
	if err := s.shipmentRepo.AssignWorkbench(ctx, shipmentID, workbenchID); err != nil {
		return err
	}

	// Cascade to tasks
	return s.taskRepo.AssignWorkbenchByShipment(ctx, shipmentID, workbenchID)
}

// GetShipmentsByWorkbench retrieves shipments assigned to a workbench.
func (s *ShipmentServiceImpl) GetShipmentsByWorkbench(ctx context.Context, workbenchID string) ([]*primary.Shipment, error) {
	records, err := s.shipmentRepo.GetByWorkbench(ctx, workbenchID)
	if err != nil {
		return nil, err
	}

	shipments := make([]*primary.Shipment, len(records))
	for i, r := range records {
		shipments[i] = s.recordToShipment(r)
	}
	return shipments, nil
}

// GetShipmentTasks retrieves all tasks for a shipment.
func (s *ShipmentServiceImpl) GetShipmentTasks(ctx context.Context, shipmentID string) ([]*primary.Task, error) {
	records, err := s.taskRepo.GetByShipment(ctx, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shipment tasks: %w", err)
	}

	tasks := make([]*primary.Task, len(records))
	for i, r := range records {
		tasks[i] = recordToTask(r)
	}
	return tasks, nil
}

// DeleteShipment deletes a shipment.
func (s *ShipmentServiceImpl) DeleteShipment(ctx context.Context, shipmentID string) error {
	return s.shipmentRepo.Delete(ctx, shipmentID)
}

// Helper methods

func (s *ShipmentServiceImpl) recordToShipment(r *secondary.ShipmentRecord) *primary.Shipment {
	return &primary.Shipment{
		ID:                  r.ID,
		CommissionID:        r.CommissionID,
		Title:               r.Title,
		Description:         r.Description,
		Status:              r.Status,
		AssignedWorkbenchID: r.AssignedWorkbenchID,
		RepoID:              r.RepoID,
		Branch:              r.Branch,
		Pinned:              r.Pinned,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
		CompletedAt:         r.CompletedAt,
	}
}

// recordToTask converts a TaskRecord to a Task (shared helper).
func recordToTask(r *secondary.TaskRecord) *primary.Task {
	var dependsOn []string
	if r.DependsOn != "" {
		_ = json.Unmarshal([]byte(r.DependsOn), &dependsOn)
	}

	return &primary.Task{
		ID:                  r.ID,
		ShipmentID:          r.ShipmentID,
		TomeID:              r.TomeID,
		CommissionID:        r.CommissionID,
		Title:               r.Title,
		Description:         r.Description,
		Type:                r.Type,
		Status:              r.Status,
		Priority:            r.Priority,
		AssignedWorkbenchID: r.AssignedWorkbenchID,
		Pinned:              r.Pinned,
		DependsOn:           dependsOn,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
		ClaimedAt:           r.ClaimedAt,
		CompletedAt:         r.CompletedAt,
	}
}

// MoveShipmentToCommission moves a shipment and its children to a different commission.
func (s *ShipmentServiceImpl) MoveShipmentToCommission(ctx context.Context, shipmentID, targetCommissionID string) (*primary.MoveShipmentResult, error) {
	// Validate shipment exists
	_, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	// Validate target commission exists
	exists, err := s.shipmentRepo.CommissionExists(ctx, targetCommissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate commission: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("commission %s not found", targetCommissionID)
	}

	// Move shipment and cascade to children
	tasksUpdated, notesUpdated, prsUpdated, err := s.shipmentRepo.MoveToCommission(ctx, shipmentID, targetCommissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to move shipment: %w", err)
	}

	return &primary.MoveShipmentResult{
		TasksUpdated: tasksUpdated,
		NotesUpdated: notesUpdated,
		PRsUpdated:   prsUpdated,
	}, nil
}

// Ensure ShipmentServiceImpl implements the interface
var _ primary.ShipmentService = (*ShipmentServiceImpl)(nil)

// Sentinel error for pinned shipment
var ErrShipmentPinned = errors.New("cannot close pinned shipment")
