package workshop

import (
	"testing"
)

func TestCanCreateWorkshop(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreateWorkshopContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create workshop when factory exists",
			ctx: CreateWorkshopContext{
				FactoryID:     "FACT-001",
				FactoryExists: true,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot create workshop when factory does not exist",
			ctx: CreateWorkshopContext{
				FactoryID:     "FACT-999",
				FactoryExists: false,
			},
			wantAllowed: false,
			wantReason:  "factory FACT-999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateWorkshop(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanCreateWorkshop() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanCreateWorkshop() Reason = %q, want %q", result.Reason, tt.wantReason)
			}

			// Test Error() method
			err := result.Error()
			if tt.wantAllowed && err != nil {
				t.Errorf("CanCreateWorkshop().Error() = %v, want nil", err)
			}
			if !tt.wantAllowed && err == nil {
				t.Error("CanCreateWorkshop().Error() = nil, want error")
			}
		})
	}
}

func TestCanDeleteWorkshop(t *testing.T) {
	tests := []struct {
		name        string
		ctx         DeleteWorkshopContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can delete workshop with no workbenches",
			ctx: DeleteWorkshopContext{
				WorkshopID:     "WORK-001",
				WorkshopExists: true,
				WorkbenchCount: 0,
				ForceDelete:    false,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot delete non-existent workshop",
			ctx: DeleteWorkshopContext{
				WorkshopID:     "WORK-999",
				WorkshopExists: false,
				WorkbenchCount: 0,
				ForceDelete:    false,
			},
			wantAllowed: false,
			wantReason:  "workshop WORK-999 not found",
		},
		{
			name: "cannot delete workshop with workbenches without force",
			ctx: DeleteWorkshopContext{
				WorkshopID:     "WORK-002",
				WorkshopExists: true,
				WorkbenchCount: 3,
				ForceDelete:    false,
			},
			wantAllowed: false,
			wantReason:  "workshop WORK-002 has 3 workbenches. Use --force to delete anyway",
		},
		{
			name: "can force delete workshop with workbenches",
			ctx: DeleteWorkshopContext{
				WorkshopID:     "WORK-003",
				WorkshopExists: true,
				WorkbenchCount: 5,
				ForceDelete:    true,
			},
			wantAllowed: true,
			wantReason:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanDeleteWorkshop(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanDeleteWorkshop() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanDeleteWorkshop() Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestGuardResult_Error(t *testing.T) {
	tests := []struct {
		name      string
		result    GuardResult
		wantError bool
	}{
		{
			name:      "allowed result returns nil error",
			result:    GuardResult{Allowed: true, Reason: ""},
			wantError: false,
		},
		{
			name:      "disallowed result returns error",
			result:    GuardResult{Allowed: false, Reason: "not allowed"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.result.Error()
			if (err != nil) != tt.wantError {
				t.Errorf("GuardResult.Error() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
