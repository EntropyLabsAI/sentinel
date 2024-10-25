package memorystore

import (
	"sync"

	sentinel "github.com/entropylabsai/sentinel/server"
)

type ReviewStoreType struct {
	sync.RWMutex
	Reviews map[string]sentinel.Review
}

func NewReviewStore() *ReviewStoreType {
	return &ReviewStoreType{
		Reviews: make(map[string]sentinel.Review),
	}
}

func (rs *ReviewStoreType) Add(review sentinel.Review) {
	rs.Lock()
	defer rs.Unlock()
	rs.Reviews[review.Id] = review
}

func (rs *ReviewStoreType) Get(reviewID string) (sentinel.Review, bool) {
	rs.RLock()
	defer rs.RUnlock()
	review, exists := rs.Reviews[reviewID]
	return review, exists
}

func (rs *ReviewStoreType) Delete(reviewID string) {
	rs.Lock()
	defer rs.Unlock()
	delete(rs.Reviews, reviewID)
}

func (rs *ReviewStoreType) Count() int {
	rs.RLock()
	defer rs.RUnlock()
	return len(rs.Reviews)
}
