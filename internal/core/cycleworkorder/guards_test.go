package cycleworkorder

import "testing"

func TestCanCreateCWO(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreateCWOContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create CWO when cycle exists and has no CWO",
			ctx: CreateCWOContext{
				CycleID:     "CYC-001",
				CycleExists: true,
				CycleHasCWO: false,
				Outcome:     "Implement feature X",
				ShipmentID:  "SHIP-001",
			},
			wantAllowed: true,
		},
		{
			name: "cannot create CWO when cycle not found",
			ctx: CreateCWOContext{
				CycleID:     "CYC-999",
				CycleExists: false,
				CycleHasCWO: false,
				Outcome:     "Implement feature X",
				ShipmentID:  "SHIP-001",
			},
			wantAllowed: false,
			wantReason:  "cycle CYC-999 not found",
		},
		{
			name: "cannot create CWO when cycle already has CWO",
			ctx: CreateCWOContext{
				CycleID:     "CYC-001",
				CycleExists: true,
				CycleHasCWO: true,
				Outcome:     "Implement feature X",
				ShipmentID:  "SHIP-001",
			},
			wantAllowed: false,
			wantReason:  "cycle CYC-001 already has a CWO",
		},
		{
			name: "cannot create CWO with empty outcome",
			ctx: CreateCWOContext{
				CycleID:     "CYC-001",
				CycleExists: true,
				CycleHasCWO: false,
				Outcome:     "",
				ShipmentID:  "SHIP-001",
			},
			wantAllowed: false,
			wantReason:  "outcome cannot be empty",
		},
		{
			name: "cannot create CWO with whitespace-only outcome",
			ctx: CreateCWOContext{
				CycleID:     "CYC-001",
				CycleExists: true,
				CycleHasCWO: false,
				Outcome:     "   ",
				ShipmentID:  "SHIP-001",
			},
			wantAllowed: false,
			wantReason:  "outcome cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateCWO(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanApprove(t *testing.T) {
	tests := []struct {
		name        string
		ctx         StatusTransitionContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can approve draft CWO",
			ctx: StatusTransitionContext{
				CWOID:         "CWO-001",
				CurrentStatus: "draft",
				Outcome:       "Implement feature X",
				CycleExists:   true,
				CycleStatus:   "draft",
			},
			wantAllowed: true,
		},
		{
			name: "cannot approve active CWO",
			ctx: StatusTransitionContext{
				CWOID:         "CWO-001",
				CurrentStatus: "active",
				Outcome:       "Implement feature X",
				CycleExists:   true,
				CycleStatus:   "approved",
			},
			wantAllowed: false,
			wantReason:  "can only approve draft CWOs (current status: active)",
		},
		{
			name: "cannot approve complete CWO",
			ctx: StatusTransitionContext{
				CWOID:         "CWO-001",
				CurrentStatus: "complete",
				Outcome:       "Implement feature X",
				CycleExists:   true,
				CycleStatus:   "complete",
			},
			wantAllowed: false,
			wantReason:  "can only approve draft CWOs (current status: complete)",
		},
		{
			name: "cannot approve CWO with empty outcome",
			ctx: StatusTransitionContext{
				CWOID:         "CWO-001",
				CurrentStatus: "draft",
				Outcome:       "",
				CycleExists:   true,
				CycleStatus:   "draft",
			},
			wantAllowed: false,
			wantReason:  "cannot approve CWO: outcome is empty",
		},
		{
			name: "cannot approve CWO when cycle does not exist",
			ctx: StatusTransitionContext{
				CWOID:         "CWO-001",
				CurrentStatus: "draft",
				Outcome:       "Implement feature X",
				CycleExists:   false,
				CycleStatus:   "",
			},
			wantAllowed: false,
			wantReason:  "cannot approve CWO: parent cycle no longer exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanApprove(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanComplete(t *testing.T) {
	tests := []struct {
		name        string
		ctx         StatusTransitionContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can complete active CWO when cycle exists",
			ctx: StatusTransitionContext{
				CWOID:         "CWO-001",
				CurrentStatus: "active",
				Outcome:       "Implement feature X",
				CycleExists:   true,
				CycleStatus:   "implementing",
			},
			wantAllowed: true,
		},
		{
			name: "can complete active CWO regardless of cycle status",
			ctx: StatusTransitionContext{
				CWOID:         "CWO-001",
				CurrentStatus: "active",
				Outcome:       "Implement feature X",
				CycleExists:   true,
				CycleStatus:   "approved",
			},
			wantAllowed: true,
		},
		{
			name: "cannot complete draft CWO",
			ctx: StatusTransitionContext{
				CWOID:         "CWO-001",
				CurrentStatus: "draft",
				Outcome:       "Implement feature X",
				CycleExists:   true,
				CycleStatus:   "implementing",
			},
			wantAllowed: false,
			wantReason:  "can only complete active CWOs (current status: draft)",
		},
		{
			name: "cannot complete already complete CWO",
			ctx: StatusTransitionContext{
				CWOID:         "CWO-001",
				CurrentStatus: "complete",
				Outcome:       "Implement feature X",
				CycleExists:   true,
				CycleStatus:   "complete",
			},
			wantAllowed: false,
			wantReason:  "can only complete active CWOs (current status: complete)",
		},
		{
			name: "cannot complete CWO when cycle does not exist",
			ctx: StatusTransitionContext{
				CWOID:         "CWO-001",
				CurrentStatus: "active",
				Outcome:       "Implement feature X",
				CycleExists:   false,
				CycleStatus:   "",
			},
			wantAllowed: false,
			wantReason:  "cannot complete CWO: parent cycle no longer exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanComplete(tt.ctx)
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
