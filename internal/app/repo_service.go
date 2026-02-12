package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/core/repo"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// RepoServiceImpl implements the RepoService interface.
type RepoServiceImpl struct {
	repoRepo   secondary.RepoRepository
	gitService *GitService
}

// NewRepoService creates a new RepoService with injected dependencies.
func NewRepoService(repoRepo secondary.RepoRepository) *RepoServiceImpl {
	return &RepoServiceImpl{
		repoRepo:   repoRepo,
		gitService: NewGitService(),
	}
}

// CreateRepo creates a new repository.
func (s *RepoServiceImpl) CreateRepo(ctx context.Context, req primary.CreateRepoRequest) (*primary.CreateRepoResponse, error) {
	// Check if name already exists
	existing, err := s.repoRepo.GetByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check name uniqueness: %w", err)
	}

	// Evaluate guard
	result := repo.CanCreateRepo(repo.CreateRepoContext{
		Name:       req.Name,
		NameExists: existing != nil,
	})
	if err := result.Error(); err != nil {
		return nil, err
	}

	// Get next ID
	nextID, err := s.repoRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate repository ID: %w", err)
	}

	// Set default branch
	defaultBranch := req.DefaultBranch
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	// Build record
	record := &secondary.RepoRecord{
		ID:            nextID,
		Name:          req.Name,
		URL:           req.URL,
		LocalPath:     req.LocalPath,
		DefaultBranch: defaultBranch,
	}

	if err := s.repoRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// Fetch created repository
	created, err := s.repoRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created repository: %w", err)
	}

	return &primary.CreateRepoResponse{
		RepoID: created.ID,
		Repo:   s.recordToRepo(created),
	}, nil
}

// GetRepo retrieves a repository by ID.
func (s *RepoServiceImpl) GetRepo(ctx context.Context, repoID string) (*primary.Repo, error) {
	record, err := s.repoRepo.GetByID(ctx, repoID)
	if err != nil {
		return nil, err
	}
	return s.recordToRepo(record), nil
}

// GetRepoByName retrieves a repository by its unique name.
func (s *RepoServiceImpl) GetRepoByName(ctx context.Context, name string) (*primary.Repo, error) {
	record, err := s.repoRepo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, fmt.Errorf("repository with name %q not found", name)
	}
	return s.recordToRepo(record), nil
}

// ListRepos lists repositories with optional filters.
func (s *RepoServiceImpl) ListRepos(ctx context.Context, filters primary.RepoFilters) ([]*primary.Repo, error) {
	records, err := s.repoRepo.List(ctx, secondary.RepoFilters{
		Status: filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	repos := make([]*primary.Repo, len(records))
	for i, r := range records {
		repos[i] = s.recordToRepo(r)
	}
	return repos, nil
}

// UpdateRepo updates a repository's configuration.
func (s *RepoServiceImpl) UpdateRepo(ctx context.Context, req primary.UpdateRepoRequest) error {
	// Verify repository exists
	_, err := s.repoRepo.GetByID(ctx, req.RepoID)
	if err != nil {
		return err
	}

	record := &secondary.RepoRecord{
		ID:             req.RepoID,
		URL:            req.URL,
		LocalPath:      req.LocalPath,
		DefaultBranch:  req.DefaultBranch,
		UpstreamURL:    req.UpstreamURL,
		UpstreamBranch: req.UpstreamBranch,
	}
	return s.repoRepo.Update(ctx, record)
}

// ArchiveRepo archives a repository.
func (s *RepoServiceImpl) ArchiveRepo(ctx context.Context, repoID string) error {
	// Get current repository
	record, err := s.repoRepo.GetByID(ctx, repoID)
	if err != nil {
		return err
	}

	// Evaluate guard
	result := repo.CanArchiveRepo(repo.ArchiveRepoContext{
		RepoID: repoID,
		Status: record.Status,
	})
	if err := result.Error(); err != nil {
		return err
	}

	return s.repoRepo.UpdateStatus(ctx, repoID, "archived")
}

// RestoreRepo restores an archived repository.
func (s *RepoServiceImpl) RestoreRepo(ctx context.Context, repoID string) error {
	// Get current repository
	record, err := s.repoRepo.GetByID(ctx, repoID)
	if err != nil {
		return err
	}

	// Evaluate guard
	result := repo.CanRestoreRepo(repo.RestoreRepoContext{
		RepoID: repoID,
		Status: record.Status,
	})
	if err := result.Error(); err != nil {
		return err
	}

	return s.repoRepo.UpdateStatus(ctx, repoID, "active")
}

// DeleteRepo hard-deletes a repository.
func (s *RepoServiceImpl) DeleteRepo(ctx context.Context, repoID string) error {
	// Check for active PRs
	hasActivePRs, err := s.repoRepo.HasActivePRs(ctx, repoID)
	if err != nil {
		return fmt.Errorf("failed to check active PRs: %w", err)
	}

	// Evaluate guard
	result := repo.CanDeleteRepo(repo.DeleteRepoContext{
		RepoID:       repoID,
		HasActivePRs: hasActivePRs,
	})
	if err := result.Error(); err != nil {
		return err
	}

	return s.repoRepo.Delete(ctx, repoID)
}

// ForkRepo configures a repository as a fork, swapping origin and adding upstream.
func (s *RepoServiceImpl) ForkRepo(ctx context.Context, req primary.ForkRepoRequest) (*primary.ForkRepoResponse, error) {
	// Fetch current repo
	record, err := s.repoRepo.GetByID(ctx, req.RepoID)
	if err != nil {
		return nil, err
	}

	// Validate upstream URL
	urlResult := repo.ValidateUpstreamURL(repo.ValidateUpstreamURLContext{
		URL: record.URL,
	})
	if err := urlResult.Error(); err != nil {
		return nil, fmt.Errorf("current repo URL is invalid for upstream: %w", err)
	}

	// Validate fork URL
	forkURLResult := repo.ValidateUpstreamURL(repo.ValidateUpstreamURLContext{
		URL: req.ForkURL,
	})
	if err := forkURLResult.Error(); err != nil {
		return nil, fmt.Errorf("fork URL is invalid: %w", err)
	}

	// Evaluate fork guard
	result := repo.CanFork(repo.ForkContext{
		HasUpstream: record.UpstreamURL != "",
		ForkURL:     req.ForkURL,
		UpstreamURL: record.URL,
	})
	if err := result.Error(); err != nil {
		return nil, err
	}

	// Swap URLs: current url -> upstream_url, fork url -> new url
	oldURL := record.URL
	updateRecord := &secondary.RepoRecord{
		ID:          record.ID,
		URL:         req.ForkURL,
		UpstreamURL: oldURL,
	}
	if err := s.repoRepo.Update(ctx, updateRecord); err != nil {
		return nil, fmt.Errorf("failed to update repository for fork: %w", err)
	}

	// Execute git remote operations (if local path exists)
	if record.LocalPath != "" {
		// git remote set-url origin <fork-url>
		if err := s.gitService.runGitCommand(record.LocalPath, "remote", "set-url", "origin", req.ForkURL); err != nil {
			return nil, fmt.Errorf("failed to set origin URL: %w", err)
		}
		// git remote add upstream <old-url>
		if err := s.gitService.runGitCommand(record.LocalPath, "remote", "add", "upstream", oldURL); err != nil {
			// upstream remote might already exist; try set-url instead
			if err2 := s.gitService.runGitCommand(record.LocalPath, "remote", "set-url", "upstream", oldURL); err2 != nil {
				return nil, fmt.Errorf("failed to add/set upstream remote: %w", err2)
			}
		}
		// git fetch upstream (best-effort)
		_ = s.gitService.FetchUpstream(record.LocalPath)
	}

	return &primary.ForkRepoResponse{
		RepoID:      record.ID,
		UpstreamURL: oldURL,
		ForkURL:     req.ForkURL,
	}, nil
}

// Helper methods

func (s *RepoServiceImpl) recordToRepo(r *secondary.RepoRecord) *primary.Repo {
	upstreamBranch := r.UpstreamBranch
	if upstreamBranch == "" && r.UpstreamURL != "" {
		upstreamBranch = r.DefaultBranch
	}
	return &primary.Repo{
		ID:             r.ID,
		Name:           r.Name,
		URL:            r.URL,
		LocalPath:      r.LocalPath,
		DefaultBranch:  r.DefaultBranch,
		UpstreamURL:    r.UpstreamURL,
		UpstreamBranch: upstreamBranch,
		Status:         r.Status,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

// Ensure RepoServiceImpl implements the interface
var _ primary.RepoService = (*RepoServiceImpl)(nil)
