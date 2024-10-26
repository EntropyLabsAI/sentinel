package sentinel

import (
	"context"

	"github.com/google/uuid"
)

// Store defines the interface for all storage operations
type Store interface {
	ProjectStore
	ReviewStore
	// LLMStore
	ProjectToolStore
	ToolStore
	SupervisorStore
	RunStore
	ReviewResultsStore
}

type ProjectStore interface {
	CreateProject(ctx context.Context, project Project) error
	GetProject(ctx context.Context, id uuid.UUID) (*Project, error)
	ListProjects(ctx context.Context) ([]Project, error)
}

type ReviewStore interface {
	CreateReview(ctx context.Context, review Review) error
	GetReview(ctx context.Context, id uuid.UUID) (*Review, error)
	UpdateReview(ctx context.Context, review Review) error
	ListHumanReviews(ctx context.Context) ([]Review, error)
	ListLLMReviews(ctx context.Context) ([]Review, error)
	DeleteReview(ctx context.Context, id uuid.UUID) error
	CountReviews(ctx context.Context) (int, error)
}

type ReviewResultsStore interface {
	GetReviewResults(ctx context.Context, id uuid.UUID) ([]*ReviewResult, error)
}

type LLMStore interface {
	SetPrompt(ctx context.Context, prompt string) error
	GetPrompt(ctx context.Context) (string, error)
}

type ProjectToolStore interface {
	GetProjectTools(ctx context.Context, id uuid.UUID) ([]Tool, error)
	CreateProjectTool(ctx context.Context, id uuid.UUID, tool Tool) error
}

type ToolStore interface {
	GetTool(ctx context.Context, id uuid.UUID) (*Tool, error)
	CreateTool(ctx context.Context, tool Tool) error
}

type SupervisorStore interface {
	GetSupervisorFromToolID(ctx context.Context, id uuid.UUID) (*Supervisor, error)
}

type RunStore interface {
	CreateRun(ctx context.Context, run Run) error
}
