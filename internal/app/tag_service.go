package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// TagServiceImpl implements the TagService interface.
type TagServiceImpl struct {
	tagRepo    secondary.TagRepository
	transactor secondary.Transactor
}

// NewTagService creates a new TagService with injected dependencies.
func NewTagService(tagRepo secondary.TagRepository, transactor secondary.Transactor) *TagServiceImpl {
	return &TagServiceImpl{
		tagRepo:    tagRepo,
		transactor: transactor,
	}
}

// CreateTag creates a new tag.
func (s *TagServiceImpl) CreateTag(ctx context.Context, req primary.CreateTagRequest) (*primary.CreateTagResponse, error) {
	var nextID string
	err := s.transactor.WithImmediateTx(ctx, func(txCtx context.Context) error {
		// Get next ID
		var err error
		nextID, err = s.tagRepo.GetNextID(txCtx)
		if err != nil {
			return fmt.Errorf("failed to generate tag ID: %w", err)
		}

		// Create record
		record := &secondary.TagRecord{
			ID:          nextID,
			Name:        req.Name,
			Description: req.Description,
		}

		if err := s.tagRepo.Create(txCtx, record); err != nil {
			return fmt.Errorf("failed to create tag: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Fetch created tag
	created, err := s.tagRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created tag: %w", err)
	}

	return &primary.CreateTagResponse{
		TagID: created.ID,
		Tag:   s.recordToTag(created),
	}, nil
}

// GetTag retrieves a tag by ID.
func (s *TagServiceImpl) GetTag(ctx context.Context, tagID string) (*primary.Tag, error) {
	record, err := s.tagRepo.GetByID(ctx, tagID)
	if err != nil {
		return nil, err
	}
	return s.recordToTag(record), nil
}

// GetTagByName retrieves a tag by name.
func (s *TagServiceImpl) GetTagByName(ctx context.Context, name string) (*primary.Tag, error) {
	record, err := s.tagRepo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return s.recordToTag(record), nil
}

// ListTags retrieves all tags.
func (s *TagServiceImpl) ListTags(ctx context.Context) ([]*primary.Tag, error) {
	records, err := s.tagRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	tags := make([]*primary.Tag, len(records))
	for i, r := range records {
		tags[i] = s.recordToTag(r)
	}
	return tags, nil
}

// DeleteTag deletes a tag.
func (s *TagServiceImpl) DeleteTag(ctx context.Context, tagID string) error {
	return s.tagRepo.Delete(ctx, tagID)
}

// GetEntityTag retrieves the tag for an entity.
func (s *TagServiceImpl) GetEntityTag(ctx context.Context, entityID, entityType string) (*primary.Tag, error) {
	record, err := s.tagRepo.GetEntityTag(ctx, entityID, entityType)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	return s.recordToTag(record), nil
}

// Helper methods

func (s *TagServiceImpl) recordToTag(r *secondary.TagRecord) *primary.Tag {
	return &primary.Tag{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// Ensure TagServiceImpl implements the interface.
var _ primary.TagService = (*TagServiceImpl)(nil)
