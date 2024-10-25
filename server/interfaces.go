package sentinel

import (
	"context"
)

// Store defines the interface for all storage operations
type Store interface {
	ProjectStore
	ReviewStore
	// LLMStore
	ProjectToolStore
}

type ProjectStore interface {
	CreateProject(ctx context.Context, project Project) error
	GetProject(ctx context.Context, id string) (*Project, error)
	ListProjects(ctx context.Context) ([]Project, error)
}

type ReviewStore interface {
	CreateReview(ctx context.Context, review Review) error
	GetReview(ctx context.Context, id string) (*Review, error)
	UpdateReview(ctx context.Context, review Review) error
	ListHumanReviews(ctx context.Context) ([]Review, error)
	ListLLMReviews(ctx context.Context) ([]Review, error)
	DeleteReview(ctx context.Context, id string) error
	CountReviews(ctx context.Context) (int, error)
}

type LLMStore interface {
	SetPrompt(ctx context.Context, prompt string) error
	GetPrompt(ctx context.Context) (string, error)
}

type ProjectToolStore interface {
	GetProjectTools(ctx context.Context, id string) ([]Tool, error)
	CreateProjectTool(ctx context.Context, id string, tool Tool) error
}
