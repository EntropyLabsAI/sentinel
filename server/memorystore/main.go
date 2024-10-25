package memorystore

import (
	"context"
	"fmt"

	sentinel "github.com/entropylabsai/sentinel/server"
)

type MemoryStore struct {
	ReviewStore  *ReviewStoreType
	ProjectStore *ProjectStoreType
	// LLMStore     *LLMStoreType
	ProjectToolStore *ProjectToolStoreType
}

// Ensure MemoryStore implements sentinel.Store.
var _ sentinel.Store = &MemoryStore{}

func New() *MemoryStore {
	return &MemoryStore{
		ReviewStore:  NewReviewStore(),
		ProjectStore: NewProjectStore(),
	}
}

// CountReviews implements sentinel.Store.
func (m *MemoryStore) CountReviews(ctx context.Context) (int, error) {
	return m.ReviewStore.Count(), nil
}

// CreateReview implements sentinel.Store.
func (m *MemoryStore) CreateReview(ctx context.Context, review sentinel.Review) error {
	m.ReviewStore.Add(review)
	return nil
}

// DeleteReview implements sentinel.Store.
func (m *MemoryStore) DeleteReview(ctx context.Context, id string) error {
	m.ReviewStore.Delete(id)
	return nil
}

// GetReview implements sentinel.Store.
func (m *MemoryStore) GetReview(ctx context.Context, id string) (*sentinel.Review, error) {
	review, exists := m.ReviewStore.Get(id)
	if !exists {
		return nil, fmt.Errorf("review not found")
	}
	return &review, nil
}

// ListHumanReviews implements sentinel.Store.
func (m *MemoryStore) ListHumanReviews(ctx context.Context) ([]sentinel.Review, error) {
	return nil, nil
}

// ListLLMReviews implements sentinel.Store.
func (m *MemoryStore) ListLLMReviews(ctx context.Context) ([]sentinel.Review, error) {
	return nil, nil
}

// UpdateReview implements sentinel.Store.
func (m *MemoryStore) UpdateReview(ctx context.Context, review sentinel.Review) error {
	m.ReviewStore.Add(review)
	return nil
}

// CreateProject implements sentinel.Store.
func (m *MemoryStore) CreateProject(ctx context.Context, project sentinel.Project) error {
	if err := m.ProjectStore.Add(ctx, project); err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

// GetProject implements sentinel.Store.
func (m *MemoryStore) GetProject(ctx context.Context, id string) (*sentinel.Project, error) {
	project, err := m.ProjectStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// ListProjects implements sentinel.Store.
func (m *MemoryStore) ListProjects(ctx context.Context) ([]sentinel.Project, error) {
	return m.ProjectStore.List(ctx)
}

// GetProjectTools implements sentinel.Store.
func (m *MemoryStore) GetProjectTools(ctx context.Context, id string) ([]sentinel.Tool, error) {
	tools, exists := m.ProjectToolStore.Get(id)
	if !exists {
		return nil, fmt.Errorf("tools not found")
	}
	return tools, nil
}

// CreateProjectTool implements sentinel.Store.
func (m *MemoryStore) CreateProjectTool(ctx context.Context, id string, tool sentinel.Tool) error {
	m.ProjectToolStore.Add(id, tool)
	return nil
}
