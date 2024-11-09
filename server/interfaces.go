package sentinel

import (
	"context"

	"github.com/google/uuid"
)

// Store defines the interface for all storage operations
type Store interface {
	ProjectStore
	RunStore
	ToolStore
	ToolRequestStore
	SupervisorStore
	SupervisorRequestStore
	SupervisorResultsStore
	SupervisorStatusStore
}

type SupervisorStatusStore interface {
	CreateSupervisionStatus(ctx context.Context, requestID uuid.UUID, status SupervisionStatus) error
	CountSupervisionRequests(ctx context.Context, status Status) (int, error)
	// GetSupervisionStatusesForRequest(ctx context.Context, requestId uuid.UUID) ([]SupervisionStatus, error)
}

type ProjectStore interface {
	CreateProject(ctx context.Context, project Project) error
	GetProject(ctx context.Context, id uuid.UUID) (*Project, error)
	GetProjectFromName(ctx context.Context, name string) (*Project, error)
	GetProjects(ctx context.Context) ([]Project, error)
	GetProjectRuns(ctx context.Context, id uuid.UUID) ([]Run, error)
}

type SupervisorRequestStore interface {
	CreateSupervisionRequest(ctx context.Context, request SupervisionRequest) (*uuid.UUID, error)
	GetSupervisionRequest(ctx context.Context, id uuid.UUID) (*SupervisionRequest, error)
	// GetSupervisionRequests(ctx context.Context) ([]SupervisionRequest, error)
	GetSupervisionRequestsForStatus(ctx context.Context, status Status) ([]SupervisionRequest, error)
}

type SupervisorResultsStore interface {
	// GetSupervisionResults(ctx context.Context, id uuid.UUID) ([]*SupervisionResult, error)
	CreateSupervisionResult(ctx context.Context, result SupervisionResult, requestId uuid.UUID) (*uuid.UUID, error)
}

type ToolRequestStore interface {
	CreateToolRequestGroup(ctx context.Context, toolId uuid.UUID, request ToolRequestGroup) (*ToolRequestGroup, error)
}

type ToolStore interface {
	CreateTool(ctx context.Context, runId uuid.UUID, attributes map[string]interface{}, name string, description string, ignoredAttributes []string) (uuid.UUID, error)
	GetTool(ctx context.Context, id uuid.UUID) (*Tool, error)
	GetToolFromValues(ctx context.Context, attributes map[string]interface{}, name string, description string, ignoredAttributes []string) (*Tool, error)
	GetRunTools(ctx context.Context, id uuid.UUID) ([]Tool, error)
	GetProjectTools(ctx context.Context, id uuid.UUID) ([]Tool, error)
}

type SupervisorStore interface {
	GetSupervisor(ctx context.Context, id uuid.UUID) (*Supervisor, error)
	GetSupervisorFromValues(ctx context.Context, code string, name string, desc string, t SupervisorType, attributes map[string]interface{}) (*Supervisor, error)
	GetSupervisors(ctx context.Context, projectId uuid.UUID) ([]Supervisor, error)
	CreateSupervisor(ctx context.Context, supervisor Supervisor) (uuid.UUID, error)
	CreateToolSupervisorChain(ctx context.Context, toolId uuid.UUID, chain ChainRequest) (*uuid.UUID, error)
	GetToolSupervisorChains(ctx context.Context, toolId uuid.UUID) ([]ToolSupervisorChain, error)
}

type RunStore interface {
	CreateRun(ctx context.Context, run Run) (uuid.UUID, error)
	GetRun(ctx context.Context, id uuid.UUID) (*Run, error)
	GetRuns(ctx context.Context, projectId uuid.UUID) ([]Run, error)
	GetProjectRuns(ctx context.Context, id uuid.UUID) ([]Run, error)
}
