package memorystore

import (
	"context"
	"fmt"

	sentinel "github.com/entropylabsai/sentinel/server"
)

type MemoryStore struct {
	ReviewStore *ReviewStoreType
	// ProjectStore *ProjectStoreType
	// LLMStore     *LLMStoreType
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

func New() *MemoryStore {
	return &MemoryStore{
		ReviewStore: NewReviewStore(),
	}
}
