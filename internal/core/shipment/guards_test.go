package shipment

import "testing"

func TestCanCreateShipment(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreateShipmentContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create shipment when commission exists",
			ctx: CreateShipmentContext{
				CommissionID:     "COMM-001",
				CommissionExists: true,
			},
			wantAllowed: true,
		},
		{
			name: "cannot create shipment when commission not found",
			ctx: CreateShipmentContext{
				CommissionID:     "COMM-999",
				CommissionExists: false,
			},
			wantAllowed: false,
			wantReason:  "commission COMM-999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateShipment(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanCloseShipment(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CloseShipmentContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can close unpinned shipment with all tasks closed",
			ctx: CloseShipmentContext{
				ShipmentID: "SHIP-001",
				IsPinned:   false,
				Tasks: []TaskSummary{
					{ID: "TASK-001", Status: "closed"},
					{ID: "TASK-002", Status: "closed"},
				},
			},
			wantAllowed: true,
		},
		{
			name: "can close shipment with no tasks",
			ctx: CloseShipmentContext{
				ShipmentID: "SHIP-001",
				IsPinned:   false,
				Tasks:      []TaskSummary{},
			},
			wantAllowed: true,
		},
		{
			name: "cannot close pinned shipment",
			ctx: CloseShipmentContext{
				ShipmentID: "SHIP-001",
				IsPinned:   true,
				Tasks: []TaskSummary{
					{ID: "TASK-001", Status: "closed"},
				},
			},
			wantAllowed: false,
			wantReason:  "cannot close pinned shipment SHIP-001. Unpin first with: orc shipment unpin SHIP-001",
		},
		{
			name: "cannot close shipment with non-closed tasks",
			ctx: CloseShipmentContext{
				ShipmentID: "SHIP-001",
				IsPinned:   false,
				Tasks: []TaskSummary{
					{ID: "TASK-001", Status: "closed"},
					{ID: "TASK-002", Status: "open"},
					{ID: "TASK-003", Status: "in-progress"},
				},
			},
			wantAllowed: false,
			wantReason:  "cannot close shipment: 2 task(s) not closed (TASK-002, TASK-003). Use --force to close anyway",
		},
		{
			name: "can force close shipment with non-closed tasks",
			ctx: CloseShipmentContext{
				ShipmentID:      "SHIP-001",
				IsPinned:        false,
				ForceCompletion: true,
				Tasks: []TaskSummary{
					{ID: "TASK-001", Status: "closed"},
					{ID: "TASK-002", Status: "open"},
				},
			},
			wantAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCloseShipment(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanOverrideStatus(t *testing.T) {
	tests := []struct {
		name        string
		ctx         OverrideStatusContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can override to valid forward status",
			ctx: OverrideStatusContext{
				ShipmentID:    "SHIP-001",
				CurrentStatus: "draft",
				NewStatus:     "ready",
			},
			wantAllowed: true,
		},
		{
			name: "cannot override to invalid status",
			ctx: OverrideStatusContext{
				ShipmentID:    "SHIP-001",
				CurrentStatus: "draft",
				NewStatus:     "exploring",
			},
			wantAllowed: false,
			wantReason:  "invalid status 'exploring'. Valid statuses: draft, ready, in-progress, closed",
		},
		{
			name: "cannot go backwards without force",
			ctx: OverrideStatusContext{
				ShipmentID:    "SHIP-001",
				CurrentStatus: "in-progress",
				NewStatus:     "draft",
			},
			wantAllowed: false,
			wantReason:  "backwards transition from 'in-progress' to 'draft' requires --force flag",
		},
		{
			name: "can go backwards with force",
			ctx: OverrideStatusContext{
				ShipmentID:    "SHIP-001",
				CurrentStatus: "in-progress",
				NewStatus:     "draft",
				Force:         true,
			},
			wantAllowed: true,
		},
		{
			name: "allows transition from unknown status",
			ctx: OverrideStatusContext{
				ShipmentID:    "SHIP-001",
				CurrentStatus: "legacy_status",
				NewStatus:     "ready",
			},
			wantAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanOverrideStatus(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanAssignWorkbench(t *testing.T) {
	tests := []struct {
		name        string
		ctx         AssignWorkbenchContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can assign unassigned workbench",
			ctx: AssignWorkbenchContext{
				ShipmentID:            "SHIP-001",
				WorkbenchID:           "BENCH-001",
				ShipmentExists:        true,
				WorkbenchAssignedToID: "",
			},
			wantAllowed: true,
		},
		{
			name: "can assign workbench already assigned to same shipment (idempotent)",
			ctx: AssignWorkbenchContext{
				ShipmentID:            "SHIP-001",
				WorkbenchID:           "BENCH-001",
				ShipmentExists:        true,
				WorkbenchAssignedToID: "SHIP-001",
			},
			wantAllowed: true,
		},
		{
			name: "cannot assign workbench assigned to another shipment",
			ctx: AssignWorkbenchContext{
				ShipmentID:            "SHIP-001",
				WorkbenchID:           "BENCH-001",
				ShipmentExists:        true,
				WorkbenchAssignedToID: "SHIP-002",
			},
			wantAllowed: false,
			wantReason:  "workbench already assigned to shipment SHIP-002",
		},
		{
			name: "cannot assign workbench to non-existent shipment",
			ctx: AssignWorkbenchContext{
				ShipmentID:            "SHIP-999",
				WorkbenchID:           "BENCH-001",
				ShipmentExists:        false,
				WorkbenchAssignedToID: "",
			},
			wantAllowed: false,
			wantReason:  "shipment SHIP-999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanAssignWorkbench(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestGuardResult_Error(t *testing.T) {
	t.Run("allowed result returns nil error", func(t *testing.T) {
		result := GuardResult{Allowed: true}
		if err := result.Error(); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("not allowed result returns error with reason", func(t *testing.T) {
		result := GuardResult{Allowed: false, Reason: "test reason"}
		err := result.Error()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "test reason" {
			t.Errorf("error = %q, want %q", err.Error(), "test reason")
		}
	})
}
