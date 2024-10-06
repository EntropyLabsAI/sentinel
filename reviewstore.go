package main

import "sync"

type ReviewStore struct {
	sync.RWMutex
	Reviews map[string]ReviewRequest
}

func NewReviewStore() *ReviewStore {
	return &ReviewStore{
		Reviews: make(map[string]ReviewRequest),
	}
}

func (rs *ReviewStore) Add(review ReviewRequest) {
	rs.Lock()
	defer rs.Unlock()
	rs.Reviews[review.RequestID] = review
}

func (rs *ReviewStore) Get(reviewID string) (ReviewRequest, bool) {
	rs.RLock()
	defer rs.RUnlock()
	review, exists := rs.Reviews[reviewID]
	return review, exists
}

func (rs *ReviewStore) Delete(reviewID string) {
	rs.Lock()
	defer rs.Unlock()
	delete(rs.Reviews, reviewID)
}

func (rs *ReviewStore) Count() int {
	rs.RLock()
	defer rs.RUnlock()
	return len(rs.Reviews)
}
