package memorystore

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	sentinel "github.com/entropylabsai/sentinel/server"
)

type ProjectStoreType struct {
	sync.RWMutex
	Projects map[string]sentinel.Project
}

func NewProjectStore() *ProjectStoreType {
	return &ProjectStoreType{
		Projects: make(map[string]sentinel.Project),
	}
}

func (ps *ProjectStoreType) Add(ctx context.Context, project sentinel.Project) error {
	ps.Lock()
	defer ps.Unlock()
	ps.Projects[project.Id] = project
	return nil
}

func (ps *ProjectStoreType) Get(ctx context.Context, id string) (sentinel.Project, error) {
	ps.RLock()
	defer ps.RUnlock()
	project, exists := ps.Projects[id]
	if !exists {
		return sentinel.Project{}, fmt.Errorf("project not found")
	}
	return project, nil
}

func (ps *ProjectStoreType) List(ctx context.Context) ([]sentinel.Project, error) {
	ps.RLock()
	defer ps.RUnlock()
	projects := make([]sentinel.Project, 0, len(ps.Projects))
	for _, project := range ps.Projects {
		projects = append(projects, project)
	}
	return projects, nil
}

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

type ProjectToolStoreType struct {
	sync.RWMutex
	Tools map[string][]sentinel.Tool
}

func NewProjectToolStore() *ProjectToolStoreType {
	return &ProjectToolStoreType{
		Tools: make(map[string][]sentinel.Tool),
	}
}

func (ts *ProjectToolStoreType) Add(projectID string, tool sentinel.Tool) {
	ts.Lock()
	defer ts.Unlock()

	// Iterate over the tools and check if any tool exists with the same desc, name and attributes
	for _, existingTool := range ts.Tools[projectID] {
		if existingTool.Description == tool.Description && existingTool.Name == tool.Name && reflect.DeepEqual(existingTool.Attributes, tool.Attributes) {
			return
		}
	}

	ts.Tools[projectID] = append(ts.Tools[projectID], tool)
}

func (ts *ProjectToolStoreType) Get(projectID string) ([]sentinel.Tool, bool) {
	ts.RLock()
	defer ts.RUnlock()
	tools, exists := ts.Tools[projectID]
	return tools, exists
}
