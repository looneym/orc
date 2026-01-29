package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/ports/primary"
)

// SummaryServiceImpl implements the SummaryService interface.
type SummaryServiceImpl struct {
	commissionService primary.CommissionService
	conclaveService   primary.ConclaveService
	tomeService       primary.TomeService
	shipmentService   primary.ShipmentService
	taskService       primary.TaskService
	noteService       primary.NoteService
	workbenchService  primary.WorkbenchService
}

// NewSummaryService creates a new SummaryService with injected dependencies.
func NewSummaryService(
	commissionService primary.CommissionService,
	conclaveService primary.ConclaveService,
	tomeService primary.TomeService,
	shipmentService primary.ShipmentService,
	taskService primary.TaskService,
	noteService primary.NoteService,
	workbenchService primary.WorkbenchService,
) *SummaryServiceImpl {
	return &SummaryServiceImpl{
		commissionService: commissionService,
		conclaveService:   conclaveService,
		tomeService:       tomeService,
		shipmentService:   shipmentService,
		taskService:       taskService,
		noteService:       noteService,
		workbenchService:  workbenchService,
	}
}

// GetCommissionSummary returns a hierarchical summary with role-based filtering.
func (s *SummaryServiceImpl) GetCommissionSummary(ctx context.Context, req primary.SummaryRequest) (*primary.CommissionSummary, error) {
	// Validate commission exists
	commission, err := s.commissionService.GetCommission(ctx, req.CommissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get commission: %w", err)
	}

	// 1. Fetch all conclaves for this commission
	conclaves, err := s.conclaveService.ListConclaves(ctx, primary.ConclaveFilters{CommissionID: req.CommissionID})
	if err != nil {
		return nil, fmt.Errorf("failed to list conclaves: %w", err)
	}

	// 2. Fetch all tomes for this commission
	allTomes, err := s.tomeService.ListTomes(ctx, primary.TomeFilters{CommissionID: req.CommissionID})
	if err != nil {
		return nil, fmt.Errorf("failed to list tomes: %w", err)
	}

	// 3. Fetch all shipments for this commission
	allShipments, err := s.shipmentService.ListShipments(ctx, primary.ShipmentFilters{CommissionID: req.CommissionID})
	if err != nil {
		return nil, fmt.Errorf("failed to list shipments: %w", err)
	}

	// Build lookup maps for grouping
	tomesByContainer := s.groupTomesByContainer(allTomes)
	shipsByContainer := s.groupShipmentsByContainer(allShipments)

	// Track hidden shipments for IMP view
	hiddenShipmentCount := 0
	isIMP := req.Role == config.RoleIMP

	// Build conclave summaries
	var conclaveSummaries []primary.ConclaveSummary
	for _, con := range conclaves {
		// Skip closed conclaves
		if con.Status == "closed" {
			continue
		}

		conSummary := primary.ConclaveSummary{
			ID:        con.ID,
			Title:     con.Title,
			Status:    con.Status,
			Pinned:    con.Pinned,
			IsFocused: con.ID == req.FocusID,
		}

		// Get tomes for this conclave
		if tomes, ok := tomesByContainer[con.ID]; ok {
			for _, tome := range tomes {
				if tome.Status == "closed" {
					continue
				}
				tomeSummary, err := s.buildTomeSummary(ctx, tome, req.FocusID)
				if err != nil {
					continue // Skip on error
				}
				conSummary.Tomes = append(conSummary.Tomes, *tomeSummary)
			}
		}

		// Get shipments for this conclave
		if ships, ok := shipsByContainer[con.ID]; ok {
			for _, ship := range ships {
				if ship.Status == "complete" {
					continue
				}

				// IMP filtering: hide shipments not assigned to this workbench
				if isIMP && !req.ShowAllShipments {
					if ship.AssignedWorkbenchID != "" && ship.AssignedWorkbenchID != req.WorkbenchID {
						hiddenShipmentCount++
						continue
					}
				}

				shipSummary, err := s.buildShipmentSummary(ctx, ship, req.FocusID)
				if err != nil {
					continue // Skip on error
				}
				conSummary.Shipments = append(conSummary.Shipments, *shipSummary)
			}
		}

		conclaveSummaries = append(conclaveSummaries, conSummary)
	}

	// Build library summary (tomes with container_type="library")
	librarySummary := s.buildLibrarySummary(allTomes)

	// Build shipyard summary (shipments with container_type="shipyard")
	shipyardSummary := s.buildShipyardSummary(allShipments, isIMP, req.ShowAllShipments, req.WorkbenchID, &hiddenShipmentCount)

	return &primary.CommissionSummary{
		ID:                  commission.ID,
		Title:               commission.Title,
		Conclaves:           conclaveSummaries,
		Library:             librarySummary,
		Shipyard:            shipyardSummary,
		HiddenShipmentCount: hiddenShipmentCount,
	}, nil
}

// groupTomesByContainer groups tomes by their container ID (conclave or library).
func (s *SummaryServiceImpl) groupTomesByContainer(tomes []*primary.Tome) map[string][]*primary.Tome {
	result := make(map[string][]*primary.Tome)
	for _, t := range tomes {
		containerID := t.ContainerID
		// Fall back to ConclaveID for backwards compatibility
		if containerID == "" && t.ConclaveID != "" {
			containerID = t.ConclaveID
		}
		if containerID != "" {
			result[containerID] = append(result[containerID], t)
		}
	}
	return result
}

// groupShipmentsByContainer groups shipments by their container ID (conclave or shipyard).
func (s *SummaryServiceImpl) groupShipmentsByContainer(shipments []*primary.Shipment) map[string][]*primary.Shipment {
	result := make(map[string][]*primary.Shipment)
	for _, ship := range shipments {
		if ship.ContainerID != "" {
			result[ship.ContainerID] = append(result[ship.ContainerID], ship)
		}
	}
	return result
}

// buildTomeSummary creates a TomeSummary with note count.
func (s *SummaryServiceImpl) buildTomeSummary(ctx context.Context, tome *primary.Tome, focusID string) (*primary.TomeSummary, error) {
	// Count notes for this tome
	notes, err := s.tomeService.GetTomeNotes(ctx, tome.ID)
	noteCount := 0
	if err == nil {
		for _, n := range notes {
			if n.Status != "closed" {
				noteCount++
			}
		}
	}

	return &primary.TomeSummary{
		ID:        tome.ID,
		Title:     tome.Title,
		Status:    tome.Status,
		NoteCount: noteCount,
		IsFocused: tome.ID == focusID,
		Pinned:    tome.Pinned,
	}, nil
}

// buildShipmentSummary creates a ShipmentSummary with task progress.
func (s *SummaryServiceImpl) buildShipmentSummary(ctx context.Context, ship *primary.Shipment, focusID string) (*primary.ShipmentSummary, error) {
	// Count tasks for this shipment
	tasks, err := s.shipmentService.GetShipmentTasks(ctx, ship.ID)
	tasksDone := 0
	tasksTotal := 0
	if err == nil {
		for _, t := range tasks {
			tasksTotal++
			if t.Status == "complete" {
				tasksDone++
			}
		}
	}

	// Get workbench name if assigned
	benchName := ""
	if ship.AssignedWorkbenchID != "" {
		bench, err := s.workbenchService.GetWorkbench(ctx, ship.AssignedWorkbenchID)
		if err == nil && bench != nil {
			benchName = bench.Name
		}
	}

	return &primary.ShipmentSummary{
		ID:         ship.ID,
		Title:      ship.Title,
		Status:     ship.Status,
		IsFocused:  ship.ID == focusID,
		Pinned:     ship.Pinned,
		BenchID:    ship.AssignedWorkbenchID,
		BenchName:  benchName,
		TasksDone:  tasksDone,
		TasksTotal: tasksTotal,
	}, nil
}

// buildLibrarySummary counts tomes in the Library.
func (s *SummaryServiceImpl) buildLibrarySummary(tomes []*primary.Tome) primary.LibrarySummary {
	count := 0
	for _, t := range tomes {
		if t.ContainerType == "library" && t.Status != "closed" {
			count++
		}
	}
	return primary.LibrarySummary{TomeCount: count}
}

// buildShipyardSummary counts shipments in the Shipyard.
func (s *SummaryServiceImpl) buildShipyardSummary(shipments []*primary.Shipment, isIMP, showAll bool, workbenchID string, hiddenCount *int) primary.ShipyardSummary {
	count := 0
	for _, ship := range shipments {
		if ship.ContainerType == "shipyard" && ship.Status != "complete" {
			// For IMP, check visibility
			if isIMP && !showAll {
				if ship.AssignedWorkbenchID != "" && ship.AssignedWorkbenchID != workbenchID {
					*hiddenCount++
					continue
				}
			}
			count++
		}
	}
	return primary.ShipyardSummary{ShipmentCount: count}
}

// Ensure SummaryServiceImpl implements the interface
var _ primary.SummaryService = (*SummaryServiceImpl)(nil)
