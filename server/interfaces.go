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
	SupervisionStore
}

type SupervisionStore interface {
	// Requests
	CreateSupervisionRequest(ctx context.Context, request SupervisionRequest) (*uuid.UUID, error)
	GetSupervisionRequest(ctx context.Context, id uuid.UUID) (*SupervisionRequest, error)
	GetSupervisionRequestsForStatus(ctx context.Context, status Status) ([]SupervisionRequest, error)

	// Results
	GetSupervisionResult(ctx context.Context, id uuid.UUID) (*SupervisionResult, error)
	CreateSupervisionResult(ctx context.Context, result SupervisionResult, requestId uuid.UUID) (*uuid.UUID, error)

	// Statuses
	CreateSupervisionStatus(ctx context.Context, requestID uuid.UUID, status SupervisionStatus) error

	// Util
	CountSupervisionRequests(ctx context.Context, status Status) (int, error)

	// GetSupervisionRequests(ctx context.Context) ([]SupervisionRequest, error)
	// GetSupervisionStatusesForRequest(ctx context.Context, requestId uuid.UUID) ([]SupervisionStatus, error)

	// New methods from Claude
	GetChainSupervisionRequests(ctx context.Context, chainId uuid.UUID) ([]SupervisionRequest, error)
	GetSupervisionRequestStatus(ctx context.Context, requestId uuid.UUID) (*SupervisionStatus, error)

	GetExecutionFromChainId(ctx context.Context, chainId uuid.UUID) (*uuid.UUID, error)
	CreateChainExecution(ctx context.Context, chainId uuid.UUID, requestGroupId uuid.UUID) (*uuid.UUID, error)
}

type ProjectStore interface {
	CreateProject(ctx context.Context, project Project) error
	GetProject(ctx context.Context, id uuid.UUID) (*Project, error)
	GetProjectFromName(ctx context.Context, name string) (*Project, error)
	GetProjects(ctx context.Context) ([]Project, error)
}

type ToolRequestStore interface {
	CreateToolRequestGroup(ctx context.Context, toolId uuid.UUID, request ToolRequestGroup) (*ToolRequestGroup, error)
	GetRunRequestGroups(ctx context.Context, runId uuid.UUID) ([]ToolRequestGroup, error)
	GetRequestGroup(ctx context.Context, id uuid.UUID) (*ToolRequestGroup, error)
	CreateToolRequest(ctx context.Context, requestGroupId uuid.UUID, request ToolRequest) (*uuid.UUID, error)
	GetToolRequest(ctx context.Context, id uuid.UUID) (*ToolRequest, error)
}

type ToolStore interface {
	CreateTool(ctx context.Context, runId uuid.UUID, attributes map[string]interface{}, name string, description string, ignoredAttributes []string) (uuid.UUID, error)
	GetTool(ctx context.Context, id uuid.UUID) (*Tool, error)
	GetToolFromValues(ctx context.Context, attributes map[string]interface{}, name string, description string, ignoredAttributes []string) (*Tool, error)
	GetRunTools(ctx context.Context, id uuid.UUID) ([]Tool, error)
	GetProjectTools(ctx context.Context, id uuid.UUID) ([]Tool, error)
}

type SupervisorStore interface {
	CreateSupervisor(ctx context.Context, supervisor Supervisor) (uuid.UUID, error)
	GetSupervisor(ctx context.Context, id uuid.UUID) (*Supervisor, error)
	GetSupervisorFromValues(ctx context.Context, code string, name string, desc string, t SupervisorType, attributes map[string]interface{}) (*Supervisor, error)
	GetSupervisors(ctx context.Context, projectId uuid.UUID) ([]Supervisor, error)
	CreateSupervisorChain(ctx context.Context, toolId uuid.UUID, chain ChainRequest) (*uuid.UUID, error)
	GetSupervisorChains(ctx context.Context, toolId uuid.UUID) ([]SupervisorChain, error)
	GetSupervisorChain(ctx context.Context, id uuid.UUID) (*SupervisorChain, error)
}

type RunStore interface {
	CreateRun(ctx context.Context, run Run) (uuid.UUID, error)
	GetRun(ctx context.Context, id uuid.UUID) (*Run, error)
	GetRuns(ctx context.Context, projectId uuid.UUID) ([]Run, error)
	GetProjectRuns(ctx context.Context, id uuid.UUID) ([]Run, error)
}
