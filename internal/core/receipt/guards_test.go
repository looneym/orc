package receipt

import "testing"

func TestCanCreateREC(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreateRECContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create REC when task exists and has no receipt",
			ctx: CreateRECContext{
				TaskID:           "TASK-001",
				TaskExists:       true,
				TaskHasReceipt:   false,
				DeliveredOutcome: "Delivered complete feature set",
			},
			wantAllowed: true,
		},
		{
			name: "cannot create REC when task not found",
			ctx: CreateRECContext{
				TaskID:           "TASK-999",
				TaskExists:       false,
				TaskHasReceipt:   false,
				DeliveredOutcome: "Delivered complete feature set",
			},
			wantAllowed: false,
			wantReason:  "task TASK-999 not found",
		},
		{
			name: "cannot create REC when task already has receipt",
			ctx: CreateRECContext{
				TaskID:           "TASK-001",
				TaskExists:       true,
				TaskHasReceipt:   true,
				DeliveredOutcome: "Delivered complete feature set",
			},
			wantAllowed: false,
			wantReason:  "task TASK-001 already has a receipt",
		},
		{
			name: "cannot create REC with empty delivered outcome",
			ctx: CreateRECContext{
				TaskID:           "TASK-001",
				TaskExists:       true,
				TaskHasReceipt:   false,
				DeliveredOutcome: "",
			},
			wantAllowed: false,
			wantReason:  "delivered outcome cannot be empty",
		},
		{
			name: "cannot create REC with whitespace-only delivered outcome",
			ctx: CreateRECContext{
				TaskID:           "TASK-001",
				TaskExists:       true,
				TaskHasReceipt:   false,
				DeliveredOutcome: "   ",
			},
			wantAllowed: false,
			wantReason:  "delivered outcome cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateREC(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanSubmit(t *testing.T) {
	tests := []struct {
		name        string
		ctx         StatusTransitionContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can submit draft REC",
			ctx: StatusTransitionContext{
				RECID:         "REC-001",
				CurrentStatus: "draft",
			},
			wantAllowed: true,
		},
		{
			name: "cannot submit submitted REC",
			ctx: StatusTransitionContext{
				RECID:         "REC-001",
				CurrentStatus: "submitted",
			},
			wantAllowed: false,
			wantReason:  "can only submit draft RECs (current status: submitted)",
		},
		{
			name: "cannot submit verified REC",
			ctx: StatusTransitionContext{
				RECID:         "REC-001",
				CurrentStatus: "verified",
			},
			wantAllowed: false,
			wantReason:  "can only submit draft RECs (current status: verified)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanSubmit(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanVerify(t *testing.T) {
	tests := []struct {
		name        string
		ctx         StatusTransitionContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can verify submitted REC",
			ctx: StatusTransitionContext{
				RECID:         "REC-001",
				CurrentStatus: "submitted",
			},
			wantAllowed: true,
		},
		{
			name: "cannot verify draft REC",
			ctx: StatusTransitionContext{
				RECID:         "REC-001",
				CurrentStatus: "draft",
			},
			wantAllowed: false,
			wantReason:  "can only verify submitted RECs (current status: draft)",
		},
		{
			name: "cannot verify already verified REC",
			ctx: StatusTransitionContext{
				RECID:         "REC-001",
				CurrentStatus: "verified",
			},
			wantAllowed: false,
			wantReason:  "can only verify submitted RECs (current status: verified)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanVerify(tt.ctx)
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
