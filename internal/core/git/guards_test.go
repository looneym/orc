package git

import (
	"testing"
)

func TestCanPerformStashDance(t *testing.T) {
	tests := []struct {
		name        string
		ctx         StashDanceContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can perform stash dance with valid context",
			ctx: StashDanceContext{
				WorkbenchPath: "/path/to/workbench",
				TargetBranch:  "feature-branch",
				IsDirty:       true,
				BranchExists:  true,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "can perform stash dance on clean repo",
			ctx: StashDanceContext{
				WorkbenchPath: "/path/to/workbench",
				TargetBranch:  "main",
				IsDirty:       false,
				BranchExists:  true,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot perform stash dance without workbench path",
			ctx: StashDanceContext{
				WorkbenchPath: "",
				TargetBranch:  "feature-branch",
				IsDirty:       true,
				BranchExists:  true,
			},
			wantAllowed: false,
			wantReason:  "workbench path is required",
		},
		{
			name: "cannot perform stash dance without target branch",
			ctx: StashDanceContext{
				WorkbenchPath: "/path/to/workbench",
				TargetBranch:  "",
				IsDirty:       true,
				BranchExists:  true,
			},
			wantAllowed: false,
			wantReason:  "target branch is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanPerformStashDance(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanPerformStashDance() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanPerformStashDance() Reason = %q, want %q", result.Reason, tt.wantReason)
			}

			// Test Error() method
			err := result.Error()
			if tt.wantAllowed && err != nil {
				t.Errorf("CanPerformStashDance().Error() = %v, want nil", err)
			}
			if !tt.wantAllowed && err == nil {
				t.Error("CanPerformStashDance().Error() = nil, want error")
			}
		})
	}
}

func TestCanCreateBranch(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreateBranchContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create branch with valid context",
			ctx: CreateBranchContext{
				RepoPath:     "/path/to/repo",
				BranchName:   "feature-branch",
				BaseBranch:   "main",
				BranchExists: false,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot create branch without repo path",
			ctx: CreateBranchContext{
				RepoPath:     "",
				BranchName:   "feature-branch",
				BaseBranch:   "main",
				BranchExists: false,
			},
			wantAllowed: false,
			wantReason:  "repository path is required",
		},
		{
			name: "cannot create branch without branch name",
			ctx: CreateBranchContext{
				RepoPath:     "/path/to/repo",
				BranchName:   "",
				BaseBranch:   "main",
				BranchExists: false,
			},
			wantAllowed: false,
			wantReason:  "branch name is required",
		},
		{
			name: "cannot create branch that already exists",
			ctx: CreateBranchContext{
				RepoPath:     "/path/to/repo",
				BranchName:   "existing-branch",
				BaseBranch:   "main",
				BranchExists: true,
			},
			wantAllowed: false,
			wantReason:  "branch existing-branch already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateBranch(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanCreateBranch() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanCreateBranch() Reason = %q, want %q", result.Reason, tt.wantReason)
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
