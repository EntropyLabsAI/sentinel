package sentinel

import (
	"context"

	"github.com/google/uuid"
)

// Store defines the interface for all storage operations
type Store interface {
	ProjectStore
	ReviewStore
	ProjectToolStore
	ToolStore
	SupervisorStore
	RunStore
	ReviewResultsStore
}

type ProjectStore interface {
	CreateProject(ctx context.Context, project Project) error
	GetProject(ctx context.Context, id uuid.UUID) (*Project, error)
	GetProjects(ctx context.Context) ([]Project, error)
	GetProjectRuns(ctx context.Context, id uuid.UUID) ([]Run, error)
}

type ReviewStore interface {
	CreateReviewRequest(ctx context.Context, request ReviewRequest) (uuid.UUID, error)
	GetReview(ctx context.Context, id uuid.UUID) (*Review, error)
	UpdateReview(ctx context.Context, review Review) error
	DeleteReview(ctx context.Context, id uuid.UUID) error
	GetReviews(ctx context.Context) ([]Review, error)
	CountReviews(ctx context.Context) (int, error)
	GetReviewToolRequests(ctx context.Context, id uuid.UUID) ([]ToolRequest, error)
	AssignSupervisorToTool(ctx context.Context, supervisorID uuid.UUID, toolID uuid.UUID) error
}

type ReviewResultsStore interface {
	GetReviewResults(ctx context.Context, id uuid.UUID) ([]*ReviewResult, error)
}

type ProjectToolStore interface {
	GetProjectTools(ctx context.Context, id uuid.UUID) ([]Tool, error)
	CreateProjectTool(ctx context.Context, id uuid.UUID, tool Tool) error
}

type ToolStore interface {
	GetTool(ctx context.Context, id uuid.UUID) (*Tool, error)
	GetTools(ctx context.Context) ([]Tool, error)
	GetRunTools(ctx context.Context, id uuid.UUID) ([]Tool, error)
	CreateRunTool(ctx context.Context, runId uuid.UUID, tool Tool) (uuid.UUID, error)
}

type SupervisorStore interface {
	GetSupervisorFromToolID(ctx context.Context, id uuid.UUID) (*Supervisor, error)
	GetSupervisor(ctx context.Context, id uuid.UUID) (*Supervisor, error)
	GetSupervisors(ctx context.Context) ([]Supervisor, error)
	CreateSupervisor(ctx context.Context, supervisor Supervisor) (uuid.UUID, error)
}

type RunStore interface {
	CreateRun(ctx context.Context, run Run) (uuid.UUID, error)
	GetRun(ctx context.Context, projectId uuid.UUID, id uuid.UUID) (*Run, error)
	GetRuns(ctx context.Context, projectId uuid.UUID) ([]Run, error)
	GetProjectRuns(ctx context.Context, id uuid.UUID) ([]Run, error)
}
