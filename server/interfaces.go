package sentinel

import (
	"context"

	"github.com/google/uuid"
)

// Store defines the interface for all storage operations
type Store interface {
	ProjectStore
	SupervisorRequestStore
	ProjectToolStore
	ToolStore
	SupervisorStore
	RunStore
	SupervisorResultsStore
	SupervisorStatusStore
	ExecutionStore
}

type SupervisorStatusStore interface {
	CreateSupervisionStatus(ctx context.Context, requestID uuid.UUID, status SupervisionStatus) error
	CountSupervisionRequests(ctx context.Context, status Status) (int, error)
}

type ExecutionStore interface {
	CreateExecution(ctx context.Context, runId uuid.UUID, toolId uuid.UUID) (uuid.UUID, error)
	GetExecution(ctx context.Context, id uuid.UUID) (*Execution, error)
	GetRunExecutions(ctx context.Context, runId uuid.UUID) ([]Execution, error)
	GetSupervisionRequestsForExecution(ctx context.Context, executionId uuid.UUID) ([]SupervisionRequest, error)
	GetSupervisionResultsForExecution(ctx context.Context, executionId uuid.UUID) ([]SupervisionResult, error)
	GetSupervisionStatusesForExecution(ctx context.Context, executionId uuid.UUID) ([]SupervisionStatus, error)
}

type ProjectStore interface {
	CreateProject(ctx context.Context, project Project) error
	GetProject(ctx context.Context, id uuid.UUID) (*Project, error)
	GetProjects(ctx context.Context) ([]Project, error)
	GetProjectRuns(ctx context.Context, id uuid.UUID) ([]Run, error)
}

type SupervisorRequestStore interface {
	CreateSupervisionRequest(ctx context.Context, request SupervisionRequest) (uuid.UUID, error)
	GetSupervisionRequest(ctx context.Context, id uuid.UUID) (*SupervisionRequest, error)
	GetSupervisionRequests(ctx context.Context) ([]SupervisionRequest, error)
	UpdateSupervisionRequest(ctx context.Context, supervisorRequest SupervisionRequest) error
	GetPendingSupervisionRequests(ctx context.Context) ([]SupervisionRequest, error)
}

type SupervisorResultsStore interface {
	GetSupervisionResults(ctx context.Context, id uuid.UUID) ([]*SupervisionResult, error)
	CreateSupervisionResult(ctx context.Context, result SupervisionResult) error
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
	GetSupervisionToolRequests(ctx context.Context, id uuid.UUID) ([]ToolRequest, error)
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
