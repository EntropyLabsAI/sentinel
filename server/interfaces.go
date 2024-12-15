package asteroid

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
	TaskStore
	ChatStore
}

type SupervisionStore interface {
	// Requests
	CreateSupervisionRequest(ctx context.Context, request SupervisionRequest, chainId uuid.UUID, toolCallId uuid.UUID) (*uuid.UUID, error)
	GetSupervisionRequest(ctx context.Context, id uuid.UUID) (*SupervisionRequest, error)
	GetSupervisionRequestsForStatus(ctx context.Context, status Status) ([]SupervisionRequest, error)

	// Results
	GetSupervisionResultFromRequestID(ctx context.Context, requestId uuid.UUID) (*SupervisionResult, error)
	CreateSupervisionResult(ctx context.Context, result SupervisionResult, requestId uuid.UUID) (*uuid.UUID, error)

	// Statuses
	CreateSupervisionStatus(ctx context.Context, requestID uuid.UUID, status SupervisionStatus) error

	// Util
	CountSupervisionRequests(ctx context.Context, status Status) (int, error)

	// GetSupervisionRequests(ctx context.Context) ([]SupervisionRequest, error)
	// GetSupervisionStatusesForRequest(ctx context.Context, requestId uuid.UUID) ([]SupervisionStatus, error)

	GetChainExecutionSupervisionRequests(ctx context.Context, chainExecutionId uuid.UUID) ([]SupervisionRequest, error)
	GetSupervisionRequestStatus(ctx context.Context, requestId uuid.UUID) (*SupervisionStatus, error)

	GetExecutionFromChainId(ctx context.Context, chainId uuid.UUID) (*uuid.UUID, error)
	GetChainExecution(ctx context.Context, executionId uuid.UUID) (*uuid.UUID, *uuid.UUID, error)
	GetChainExecutionFromChainAndToolCall(ctx context.Context, chainId uuid.UUID, toolCallId uuid.UUID) (*uuid.UUID, error)
	GetChainExecutionsFromToolCall(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error)
	GetChainExecutionState(ctx context.Context, executionId uuid.UUID) (*ChainExecutionState, error)
}

type TaskStore interface {
	CreateTask(ctx context.Context, task Task) (*uuid.UUID, error)
	GetTask(ctx context.Context, id uuid.UUID) (*Task, error)
	GetProjectTasks(ctx context.Context, projectId uuid.UUID) ([]Task, error)
}

type ProjectStore interface {
	CreateProject(ctx context.Context, project Project) error
	GetProject(ctx context.Context, id uuid.UUID) (*Project, error)
	GetProjectFromName(ctx context.Context, name string) (*Project, error)
	GetProjects(ctx context.Context) ([]Project, error)
}

type ToolRequestStore interface {
	// CreateToolRequestGroup(ctx context.Context, toolId uuid.UUID, request ToolRequestGroup) (*ToolRequestGroup, error)
	// GetRunRequestGroups(ctx context.Context, runId uuid.UUID, includeArgs bool) ([]ToolRequestGroup, error)
	// GetRequestGroup(ctx context.Context, id uuid.UUID, includeArgs bool) (*ToolRequestGroup, error)
	// CreateToolCall(ctx context.Context, toolCallId uuid.UUID, request ToolRequest) (*uuid.UUID, error)
	GetToolCall(ctx context.Context, id uuid.UUID) (*AsteroidToolCall, error)
	GetToolCallFromCallId(ctx context.Context, id string) (*AsteroidToolCall, error)
}

type ToolStore interface {
	CreateTool(ctx context.Context, runId uuid.UUID, attributes map[string]interface{}, name string, description string, ignoredAttributes []string, code string) (*Tool, error)
	GetTool(ctx context.Context, id uuid.UUID) (*Tool, error)
	// GetToolFromValues(ctx context.Context, attributes map[string]interface{}, name string, description string, ignoredAttributes []string) (*Tool, error)
	GetRunTools(ctx context.Context, id uuid.UUID) ([]Tool, error)
	GetProjectTools(ctx context.Context, id uuid.UUID) ([]Tool, error)
	GetToolFromNameAndRunId(ctx context.Context, name string, runId uuid.UUID) (*Tool, error)
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
	GetRuns(ctx context.Context, taskId uuid.UUID) ([]Run, error)
	GetTaskRuns(ctx context.Context, taskId uuid.UUID) ([]Run, error)
	UpdateRunStatus(ctx context.Context, runId uuid.UUID, status Status) error
	UpdateRunResult(ctx context.Context, runId uuid.UUID, result string) error
}

type ChatStore interface {
	CreateChatRequest(
		ctx context.Context,
		runId uuid.UUID,
		request []byte,
		response []byte,
		choices []AsteroidChoice,
		format ChatFormat,
		requestMessages []AsteroidMessage,
	) (*uuid.UUID, error)
	// GetMessagesForRun(ctx context.Context, runId uuid.UUID, includeInvalidated bool) ([]AsteroidMessage, error)
	GetChat(ctx context.Context, runId uuid.UUID, index int) ([]byte, []byte, ChatFormat, error)
	GetMessage(ctx context.Context, id uuid.UUID) (*AsteroidMessage, error)
	UpdateMessage(ctx context.Context, id uuid.UUID, message AsteroidMessage) error
	GetRunChatCount(ctx context.Context, runId uuid.UUID) (int, error)
}

type AsteroidConverter interface {
	ToAsteroidMessages(ctx context.Context, requestData, responseData []byte, runId uuid.UUID) ([]AsteroidMessage, error)
	ToAsteroidChoices(ctx context.Context, responseData []byte, runId uuid.UUID) ([]AsteroidChoice, error)
	ValidateB64EncodedRequest(encodedData string) ([]byte, error)
	ValidateB64EncodedResponse(encodedData string) ([]byte, error)
}
