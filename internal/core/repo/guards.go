// Package repo contains the pure business logic for repository operations.
// Guards are pure functions that evaluate preconditions without side effects.
package repo

import (
	"fmt"
	"regexp"
	"strings"
)

// GuardResult represents the outcome of a guard evaluation.
type GuardResult struct {
	Allowed bool
	Reason  string
}

// Error converts the guard result to an error if not allowed.
func (r GuardResult) Error() error {
	if r.Allowed {
		return nil
	}
	return fmt.Errorf("%s", r.Reason)
}

// CreateRepoContext provides context for repository creation guards.
type CreateRepoContext struct {
	Name       string
	NameExists bool // true if a repo with this name already exists
}

// ArchiveRepoContext provides context for repository archive guards.
type ArchiveRepoContext struct {
	RepoID string
	Status string
}

// RestoreRepoContext provides context for repository restore guards.
type RestoreRepoContext struct {
	RepoID string
	Status string
}

// DeleteRepoContext provides context for repository deletion guards.
type DeleteRepoContext struct {
	RepoID       string
	HasActivePRs bool
}

// CanCreateRepo evaluates whether a repository can be created.
// Rules:
// - Name must not be empty
// - Name must be unique
func CanCreateRepo(ctx CreateRepoContext) GuardResult {
	// Rule 1: Name must not be empty
	if strings.TrimSpace(ctx.Name) == "" {
		return GuardResult{
			Allowed: false,
			Reason:  "repository name cannot be empty",
		}
	}

	// Rule 2: Name must be unique
	if ctx.NameExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("repository with name %q already exists", ctx.Name),
		}
	}

	return GuardResult{Allowed: true}
}

// CanArchiveRepo evaluates whether a repository can be archived.
// Rules:
// - Status must be "active"
func CanArchiveRepo(ctx ArchiveRepoContext) GuardResult {
	if ctx.Status != "active" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only archive active repositories (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}

// CanRestoreRepo evaluates whether a repository can be restored.
// Rules:
// - Status must be "archived"
func CanRestoreRepo(ctx RestoreRepoContext) GuardResult {
	if ctx.Status != "archived" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only restore archived repositories (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}

// CanDeleteRepo evaluates whether a repository can be deleted.
// Rules:
// - No active PRs can reference this repository
func CanDeleteRepo(ctx DeleteRepoContext) GuardResult {
	if ctx.HasActivePRs {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cannot delete repository %s with active pull requests", ctx.RepoID),
		}
	}

	return GuardResult{Allowed: true}
}

// ValidateUpstreamURLContext provides context for upstream URL validation.
type ValidateUpstreamURLContext struct {
	URL string
}

// Patterns for valid git remote URLs.
var (
	sshURLPattern   = regexp.MustCompile(`^git@[\w.\-]+:[\w.\-]+/[\w.\-]+\.git$`)
	httpsURLPattern = regexp.MustCompile(`^https://[\w.\-]+/[\w.\-]+/[\w.\-]+\.git$`)
)

// ValidateUpstreamURL evaluates whether an upstream URL is valid.
// Rules:
// - URL must not be empty
// - URL must match SSH format (git@host:owner/repo.git) or HTTPS format (https://host/owner/repo.git)
func ValidateUpstreamURL(ctx ValidateUpstreamURLContext) GuardResult {
	if strings.TrimSpace(ctx.URL) == "" {
		return GuardResult{
			Allowed: false,
			Reason:  "upstream URL cannot be empty",
		}
	}

	if sshURLPattern.MatchString(ctx.URL) || httpsURLPattern.MatchString(ctx.URL) {
		return GuardResult{Allowed: true}
	}

	return GuardResult{
		Allowed: false,
		Reason:  fmt.Sprintf("invalid upstream URL %q: must be SSH (git@host:owner/repo.git) or HTTPS (https://host/owner/repo.git)", ctx.URL),
	}
}

// ForkContext provides context for fork eligibility guards.
type ForkContext struct {
	HasUpstream bool   // true if upstream_url is already set
	ForkURL     string // the origin/fork URL
	UpstreamURL string // the upstream URL being configured
}

// CanFork evaluates whether a repository can be configured as a fork.
// Rules:
// - Repository must not already have an upstream (no double-forking)
// - Fork URL must differ from upstream URL
func CanFork(ctx ForkContext) GuardResult {
	if ctx.HasUpstream {
		return GuardResult{
			Allowed: false,
			Reason:  "repository already has an upstream configured (cannot double-fork)",
		}
	}

	if ctx.ForkURL == ctx.UpstreamURL {
		return GuardResult{
			Allowed: false,
			Reason:  "fork URL must differ from upstream URL",
		}
	}

	return GuardResult{Allowed: true}
}
