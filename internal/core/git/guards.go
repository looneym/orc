// Package git contains domain logic for git operations.
package git

import "fmt"

// GuardResult represents the outcome of a guard check.
type GuardResult struct {
	Allowed bool
	Reason  string
}

// Error returns the guard result as an error (nil if allowed).
func (r GuardResult) Error() error {
	if r.Allowed {
		return nil
	}
	return fmt.Errorf("guard failed: %s", r.Reason)
}

// StashDanceContext contains context for stash dance operations.
type StashDanceContext struct {
	WorkbenchPath string
	TargetBranch  string
	IsDirty       bool
	BranchExists  bool
}

// CanPerformStashDance checks if a stash dance operation is allowed.
func CanPerformStashDance(ctx StashDanceContext) GuardResult {
	if ctx.WorkbenchPath == "" {
		return GuardResult{Allowed: false, Reason: "workbench path is required"}
	}
	if ctx.TargetBranch == "" {
		return GuardResult{Allowed: false, Reason: "target branch is required"}
	}
	return GuardResult{Allowed: true}
}

// CreateBranchContext contains context for branch creation.
type CreateBranchContext struct {
	RepoPath     string
	BranchName   string
	BaseBranch   string
	BranchExists bool
}

// CanCreateBranch checks if a branch can be created.
func CanCreateBranch(ctx CreateBranchContext) GuardResult {
	if ctx.RepoPath == "" {
		return GuardResult{Allowed: false, Reason: "repository path is required"}
	}
	if ctx.BranchName == "" {
		return GuardResult{Allowed: false, Reason: "branch name is required"}
	}
	if ctx.BranchExists {
		return GuardResult{Allowed: false, Reason: fmt.Sprintf("branch %s already exists", ctx.BranchName)}
	}
	return GuardResult{Allowed: true}
}
