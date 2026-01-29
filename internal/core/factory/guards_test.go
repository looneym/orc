package factory

import (
	"testing"
)

func TestCanCreateFactory(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreateFactoryContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create factory with unique name",
			ctx: CreateFactoryContext{
				Name:       "my-factory",
				NameExists: false,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot create factory with existing name",
			ctx: CreateFactoryContext{
				Name:       "existing-factory",
				NameExists: true,
			},
			wantAllowed: false,
			wantReason:  `factory with name "existing-factory" already exists`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateFactory(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanCreateFactory() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanCreateFactory() Reason = %q, want %q", result.Reason, tt.wantReason)
			}

			// Test Error() method
			err := result.Error()
			if tt.wantAllowed && err != nil {
				t.Errorf("CanCreateFactory().Error() = %v, want nil", err)
			}
			if !tt.wantAllowed && err == nil {
				t.Error("CanCreateFactory().Error() = nil, want error")
			}
		})
	}
}

func TestCanDeleteFactory(t *testing.T) {
	tests := []struct {
		name        string
		ctx         DeleteFactoryContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can delete factory with no dependents",
			ctx: DeleteFactoryContext{
				FactoryID:       "FACT-001",
				FactoryExists:   true,
				WorkshopCount:   0,
				CommissionCount: 0,
				ForceDelete:     false,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot delete non-existent factory",
			ctx: DeleteFactoryContext{
				FactoryID:       "FACT-999",
				FactoryExists:   false,
				WorkshopCount:   0,
				CommissionCount: 0,
				ForceDelete:     false,
			},
			wantAllowed: false,
			wantReason:  "factory FACT-999 not found",
		},
		{
			name: "cannot delete factory with workshops without force",
			ctx: DeleteFactoryContext{
				FactoryID:       "FACT-002",
				FactoryExists:   true,
				WorkshopCount:   3,
				CommissionCount: 0,
				ForceDelete:     false,
			},
			wantAllowed: false,
			wantReason:  "factory FACT-002 has 3 workshops and 0 commissions. Use --force to delete anyway",
		},
		{
			name: "cannot delete factory with commissions without force",
			ctx: DeleteFactoryContext{
				FactoryID:       "FACT-003",
				FactoryExists:   true,
				WorkshopCount:   0,
				CommissionCount: 2,
				ForceDelete:     false,
			},
			wantAllowed: false,
			wantReason:  "factory FACT-003 has 0 workshops and 2 commissions. Use --force to delete anyway",
		},
		{
			name: "cannot delete factory with both workshops and commissions without force",
			ctx: DeleteFactoryContext{
				FactoryID:       "FACT-004",
				FactoryExists:   true,
				WorkshopCount:   5,
				CommissionCount: 3,
				ForceDelete:     false,
			},
			wantAllowed: false,
			wantReason:  "factory FACT-004 has 5 workshops and 3 commissions. Use --force to delete anyway",
		},
		{
			name: "can force delete factory with dependents",
			ctx: DeleteFactoryContext{
				FactoryID:       "FACT-005",
				FactoryExists:   true,
				WorkshopCount:   5,
				CommissionCount: 3,
				ForceDelete:     true,
			},
			wantAllowed: true,
			wantReason:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanDeleteFactory(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanDeleteFactory() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanDeleteFactory() Reason = %q, want %q", result.Reason, tt.wantReason)
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
