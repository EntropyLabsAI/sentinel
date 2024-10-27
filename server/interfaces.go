package sentinel

import (
	"context"

	"github.com/google/uuid"
)

// Store defines the interface for all storage operations
type Store interface {
	ProjectStore
	ReviewRequestStore
	ProjectToolStore
	ToolStore
	SupervisorStore
	RunStore
	ReviewResultsStore
	ReviewStatusStore
}

type ReviewStatusStore interface {
	CreateReviewStatus(ctx context.Context, requestID uuid.UUID, status ReviewStatus) error
	CountReviewRequests(ctx context.Context, status ReviewStatusStatus) (int, error)
}

type ProjectStore interface {
	CreateProject(ctx context.Context, project Project) error
	GetProject(ctx context.Context, id uuid.UUID) (*Project, error)
	GetProjects(ctx context.Context) ([]Project, error)
	GetProjectRuns(ctx context.Context, id uuid.UUID) ([]Run, error)
}

type ReviewRequestStore interface {
	CreateReviewRequest(ctx context.Context, request ReviewRequest) (uuid.UUID, error)
	GetReviewRequest(ctx context.Context, id uuid.UUID) (*ReviewRequest, error)
	GetReviewRequests(ctx context.Context) ([]ReviewRequest, error)
	UpdateReviewRequest(ctx context.Context, reviewRequest ReviewRequest) error
	GetPendingReviewRequests(ctx context.Context) ([]ReviewRequest, error)
}

type ReviewResultsStore interface {
	GetReviewResults(ctx context.Context, id uuid.UUID) ([]*ReviewResult, error)
	CreateReviewResult(ctx context.Context, result ReviewResult) error
}

type ProjectToolStore interface {
	GetProjectTools(ctx context.Context, id uuid.UUID) ([]Tool, error)
	CreateProjectTool(ctx context.Context, id uuid.UUID, tool Tool) error
}

type ToolStore interface {
	CreateTool(ctx context.Context, tool Tool) (uuid.UUID, error)
	GetTool(ctx context.Context, id uuid.UUID) (*Tool, error)
	GetTools(ctx context.Context) ([]Tool, error)
	GetRunTools(ctx context.Context, id uuid.UUID) ([]Tool, error)
	GetReviewToolRequests(ctx context.Context, id uuid.UUID) ([]ToolRequest, error)
}

type SupervisorStore interface {
	GetSupervisorFromToolID(ctx context.Context, id uuid.UUID) (*Supervisor, error)
	GetSupervisor(ctx context.Context, id uuid.UUID) (*Supervisor, error)
	GetSupervisors(ctx context.Context) ([]Supervisor, error)
	CreateSupervisor(ctx context.Context, supervisor Supervisor) (uuid.UUID, error)
	GetRunToolSupervisors(ctx context.Context, runId uuid.UUID, toolId uuid.UUID) ([]Supervisor, error)
	AssignSupervisorsToTool(ctx context.Context, runId uuid.UUID, toolID uuid.UUID, supervisorIds []uuid.UUID) error
}

type RunStore interface {
	CreateRun(ctx context.Context, run Run) (uuid.UUID, error)
	GetRun(ctx context.Context, id uuid.UUID) (*Run, error)
	GetRuns(ctx context.Context, projectId uuid.UUID) ([]Run, error)
	GetProjectRuns(ctx context.Context, id uuid.UUID) ([]Run, error)
}
