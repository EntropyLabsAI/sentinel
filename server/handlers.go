package sentinel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
)

// respondJSON writes a JSON response with status 200 OK
func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func apiCreateProjectHandler(w http.ResponseWriter, r *http.Request, store ProjectStore) {
	ctx := r.Context()

	var request struct {
		Name          string   `json:"name"`
		RunResultTags []string `json:"run_result_tags"`
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	existingProject, err := store.GetProjectFromName(ctx, request.Name)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Error getting project", err.Error())
		return
	}

	if existingProject != nil {
		respondJSON(w, existingProject.Id.String(), http.StatusOK)
		return
	}

	// Generate a new Project ID
	id := uuid.New()

	// Create the Project struct
	project := Project{
		Id:            id,
		Name:          request.Name,
		RunResultTags: request.RunResultTags,
		CreatedAt:     time.Now(),
	}

	// Store the project in the global projects map
	err = store.CreateProject(ctx, project)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to register project", err.Error())
		return
	}

	// Send the response
	respondJSON(w, id.String(), http.StatusCreated)
}

func apiCreateTaskHandler(w http.ResponseWriter, r *http.Request, projectId uuid.UUID, store Store) {
	ctx := r.Context()

	var request struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	task := Task{
		ProjectId:   projectId,
		Name:        request.Name,
		Description: &request.Description,
		CreatedAt:   time.Now(),
	}

	id, err := store.CreateTask(ctx, task)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to create task", err.Error())
		return
	}

	respondJSON(w, id.String(), http.StatusCreated)
}

func apiGetTaskHandler(w http.ResponseWriter, r *http.Request, taskId uuid.UUID, store Store) {
	ctx := r.Context()

	task, err := store.GetTask(ctx, taskId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Error getting task", err.Error())
		return
	}

	if task == nil {
		sendErrorResponse(w, http.StatusNotFound, "Task not found", "")
		return
	}

	respondJSON(w, task, http.StatusOK)
}

func apiGetProjectTasksHandler(w http.ResponseWriter, r *http.Request, projectId uuid.UUID, store Store) {
	ctx := r.Context()

	tasks, err := store.GetProjectTasks(ctx, projectId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Error getting project tasks", err.Error())
		return
	}

	respondJSON(w, tasks, http.StatusOK)
}

func apiCreateRunHandler(w http.ResponseWriter, r *http.Request, taskId uuid.UUID, store Store) {
	ctx := r.Context()

	run := Run{
		Id:        uuid.New(),
		TaskId:    taskId, // Changed from ProjectId to TaskId
		CreatedAt: time.Now(),
	}

	runID, err := store.CreateRun(ctx, run)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Error creating run", err.Error())
		return
	}

	respondJSON(w, runID, http.StatusCreated)
}

func apiGetRunHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, store Store) {
	ctx := r.Context()

	run, err := store.GetRun(ctx, runId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Error getting run", err.Error())
		return
	}

	if run == nil {
		sendErrorResponse(w, http.StatusNotFound, "Run not found", "")
		return
	}

	respondJSON(w, run, http.StatusOK)
}

func apiGetTaskRunsHandler(w http.ResponseWriter, r *http.Request, taskId uuid.UUID, store Store) {
	ctx := r.Context()

	runs, err := store.GetTaskRuns(ctx, taskId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Error getting task runs", err.Error())
		return
	}

	respondJSON(w, runs, http.StatusOK)
}

func apiCreateRunToolHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, store Store) {
	ctx := r.Context()

	// Check that the run exists
	run, err := store.GetRun(ctx, runId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting run", err.Error())
		return
	}

	if run == nil {
		sendErrorResponse(w, http.StatusNotFound, "Run not found", "")
		return
	}

	var t struct {
		Attributes        map[string]interface{} `json:"attributes"`
		Name              string                 `json:"name"`
		Description       string                 `json:"description"`
		IgnoredAttributes []string               `json:"ignored_attributes"`
		Code              string                 `json:"code"`
	}
	err = json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	// TODO revive this logic so that we can share tools across runs, requires schema changes
	// var existingTool *Tool
	// if t.Attributes != nil && t.Name != "" && t.Description != "" && t.IgnoredAttributes != nil {
	// 	found, err := store.GetToolFromValues(ctx, t.Attributes, t.Name, t.Description, t.IgnoredAttributes)
	// 	if err != nil {
	// 		sendErrorResponse(w, http.StatusInternalServerError, "error trying to locate an existing tool", err.Error())
	// 		return
	// 	}
	// 	if found != nil {
	// 		existingTool = found
	// 	}
	// }

	// if existingTool != nil {
	// 	w.WriteHeader(http.StatusOK)
	// 	id := existingTool.Id.String()
	// 	respondJSON(w, id)
	// 	return
	// }

	toolId, err := store.CreateTool(ctx, runId, t.Attributes, t.Name, t.Description, t.IgnoredAttributes, t.Code)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error creating tool", err.Error())
		return
	}

	respondJSON(w, toolId, http.StatusCreated)
}

func apiGetSupervisorHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store SupervisorStore) {
	ctx := r.Context()

	supervisor, err := store.GetSupervisor(ctx, id)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervisor", err.Error())
		return
	}

	respondJSON(w, supervisor, http.StatusOK)
}

func apiCreateToolSupervisorChainsHandler(w http.ResponseWriter, r *http.Request, toolId uuid.UUID, store SupervisorStore) {
	ctx := r.Context()

	var request []ChainRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "invalid JSON format", err.Error())
		return
	}

	// TODO do we want to return the chains here?
	chainIds := make([]uuid.UUID, 0)
	for _, chain := range request {
		chainId, err := store.CreateSupervisorChain(ctx, toolId, chain)
		if err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "error creating supervisor chain", err.Error())
			return
		}
		chainIds = append(chainIds, *chainId)
	}

	respondJSON(w, chainIds, http.StatusCreated)
}

func apiCreateSupervisorHandler(w http.ResponseWriter, r *http.Request, _ uuid.UUID, store SupervisorStore) {
	ctx := r.Context()

	var request Supervisor
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "invalid JSON format", err.Error())
		return
	}

	// Create new supervisor
	supervisorId, err := store.CreateSupervisor(ctx, request)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error creating supervisor", err.Error())
		return
	}

	respondJSON(w, supervisorId, http.StatusCreated)
}

func apiGetToolSupervisorChainsHandler(w http.ResponseWriter, r *http.Request, toolId uuid.UUID, store Store) {
	ctx := r.Context()

	// First check if tool exists
	tool, err := store.GetTool(ctx, toolId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting tool", err.Error())
		return
	}

	if tool == nil {
		sendErrorResponse(w, http.StatusNotFound, "Tool not found", "")
		return
	}

	chains, err := store.GetSupervisorChains(ctx, toolId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting tool supervisor chains", err.Error())
		return
	}

	respondJSON(w, chains, http.StatusOK)
}

func apiGetRunToolsHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	ctx := r.Context()

	// First check if run exists
	run, err := store.GetRun(ctx, id)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting run", err.Error())
		return
	}

	if run == nil {
		sendErrorResponse(w, http.StatusNotFound, "Run not found", "")
		return
	}

	tools, err := store.GetRunTools(ctx, id)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting run tools", err.Error())
		return
	}

	respondJSON(w, tools, http.StatusOK)
}

func apiCreateToolRequestGroupHandler(w http.ResponseWriter, r *http.Request, toolId uuid.UUID, store ToolRequestStore) {
	ctx := r.Context()

	var request ToolRequestGroup
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	trg, err := store.CreateToolRequestGroup(ctx, toolId, request)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error creating tool request group", err.Error())
		return
	}

	respondJSON(w, trg, http.StatusCreated)
}

func apiGetRequestGroupHandler(w http.ResponseWriter, r *http.Request, requestGroupId uuid.UUID, store Store) {
	ctx := r.Context()

	requestGroup, err := store.GetRequestGroup(ctx, requestGroupId, true)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting request group", err.Error())
		return
	}

	respondJSON(w, requestGroup, http.StatusOK)
}

func apiGetToolRequestHandler(w http.ResponseWriter, r *http.Request, toolRequestId uuid.UUID, store ToolRequestStore) {
	ctx := r.Context()

	toolRequest, err := store.GetToolRequest(ctx, toolRequestId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting tool request", err.Error())
		return
	}

	if toolRequest == nil {
		sendErrorResponse(w, http.StatusNotFound, "Tool request not found", "")
		return
	}

	respondJSON(w, toolRequest, http.StatusOK)
}

func apiCreateToolRequestHandler(w http.ResponseWriter, r *http.Request, requestGroupId uuid.UUID, store ToolRequestStore) {
	ctx := r.Context()

	var request ToolRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	toolRequestId, err := store.CreateToolRequest(ctx, requestGroupId, request)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error creating tool request", err.Error())
		return
	}

	respondJSON(w, toolRequestId, http.StatusCreated)
}

func apiGetSupervisionResultHandler(w http.ResponseWriter, r *http.Request, supervisionRequestId uuid.UUID, store Store) {
	ctx := r.Context()

	// Check that the supervision request exists
	supervisionRequest, err := store.GetSupervisionRequest(ctx, supervisionRequestId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervision request", err.Error())
		return
	}

	if supervisionRequest == nil {
		sendErrorResponse(w, http.StatusNotFound, "Supervision request not found", "")
		return
	}

	supervisionResult, err := store.GetSupervisionResultFromRequestID(ctx, supervisionRequestId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervision result", err.Error())
		return
	}

	respondJSON(w, supervisionResult, http.StatusOK)
}

func apiGetRunRequestGroupsHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, store Store) {
	ctx := r.Context()

	run, err := store.GetRun(ctx, runId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting run", err.Error())
		return
	}

	if run == nil {
		sendErrorResponse(w, http.StatusNotFound, "Run not found", "")
		return
	}

	requestGroups, err := store.GetRunRequestGroups(ctx, runId, true)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting run request groups", err.Error())
		return
	}

	respondJSON(w, requestGroups, http.StatusOK)
}

func apiGetSupervisorsHandler(w http.ResponseWriter, r *http.Request, projectId uuid.UUID, store Store) {
	ctx := r.Context()

	// First check if project exists
	project, err := store.GetProject(ctx, projectId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting project", err.Error())
		return
	}

	if project == nil {
		sendErrorResponse(w, http.StatusNotFound, "Project not found", "")
		return
	}

	supervisors, err := store.GetSupervisors(ctx, projectId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervisors", err.Error())
		return
	}

	respondJSON(w, supervisors, http.StatusOK)
}

func apiGetProjectToolsHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store ToolStore) {
	ctx := r.Context()

	tools, err := store.GetProjectTools(ctx, id)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting project tools", err.Error())
		return
	}

	respondJSON(w, tools, http.StatusOK)
}

func apiGetToolHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store ToolStore) {
	ctx := r.Context()

	tool, err := store.GetTool(ctx, id)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting tool", err.Error())
		return
	}

	if tool == nil {
		sendErrorResponse(w, http.StatusNotFound, "Tool not found", "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(tool)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error encoding tool", err.Error())
		return
	}
}

func apiGetSupervisionRequestStatusHandler(w http.ResponseWriter, r *http.Request, reviewID uuid.UUID, store Store) {
	ctx := r.Context()
	status, err := store.GetSupervisionRequestStatus(ctx, reviewID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervisor", err.Error())
		return
	}

	respondJSON(w, status, http.StatusOK)
}

func apiCreateSupervisionRequestHandler(
	w http.ResponseWriter,
	r *http.Request,
	requestGroupId uuid.UUID,
	chainId uuid.UUID,
	supervisorId uuid.UUID,
	store Store,
) {
	ctx := r.Context()

	var request SupervisionRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	// Check that the request, chain and supervisor exist
	requestGroup, err := store.GetRequestGroup(ctx, requestGroupId, false)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting request group", err.Error())
		return
	}

	if requestGroup == nil {
		sendErrorResponse(w, http.StatusNotFound, "Request group not found", "")
		return
	}

	if len(requestGroup.ToolRequests) > 1 {
		sendErrorResponse(w, http.StatusBadRequest, "Request group must contain only one tool request", "")
		return
	}

	chain, err := store.GetSupervisorChain(ctx, chainId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervisor chain", err.Error())
		return
	}

	if chain == nil {
		sendErrorResponse(w, http.StatusNotFound, "Supervisor chain not found", "")
		return
	}

	supervisor, err := store.GetSupervisor(ctx, supervisorId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervisor", err.Error())
		return
	}

	if supervisor == nil {
		sendErrorResponse(w, http.StatusNotFound, "Supervisor not found", "")
		return
	}

	// Check that the supervisor is associated with the tool/request/chain
	found := false
	pos := -1
	for i, chainSupervisor := range chain.Supervisors {
		if chainSupervisor.Id.String() == supervisorId.String() {
			found = true
			pos = i
			break
		}
	}

	if !found {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Supervisor %s not associated with chain %s", supervisorId, chainId), "")
		return
	}

	if pos != request.PositionInChain {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Supervisor %s is not in the correct position in chain %s", supervisorId, chainId), "")
		return
	}

	// Check that the chainexecution entry exists
	foundExecutionId, err := store.GetChainExecutionFromChainAndRequestGroup(ctx, chainId, requestGroupId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting execution from chain ID", err.Error())
		return
	}

	// Sanity check: this is the first supervisor in the chain, we shouldn't have already created a chain execution
	if foundExecutionId != nil && pos == 0 {
		sendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("chain execution already exists for chain %s, yet supervisor is position 0. Curious.", chainId), "")
		return
	}

	if foundExecutionId == nil && pos > 0 {
		sendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("no ongoing chain execution found for chain %s, yet supervisor allegedly not a position 0. Curious.", chainId), "")
		return
	}

	if request.ChainexecutionId != nil && foundExecutionId != nil {
		if *request.ChainexecutionId != *foundExecutionId {
			sendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("chain execution ID mismatch for chain %s, request group %s, and supervisor %s", chainId, requestGroupId, supervisorId), "")
			return
		}
	}

	if foundExecutionId != nil && request.ChainexecutionId == nil {
		request.ChainexecutionId = foundExecutionId
	}

	// Store the supervision in the database
	reviewID, err := store.CreateSupervisionRequest(ctx, request, chainId, requestGroupId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error creating supervision request", err.Error())
		return
	}

	respondJSON(w, reviewID, http.StatusCreated)
}

func apiCreateSupervisionResultHandler(
	w http.ResponseWriter,
	r *http.Request,
	supervisionRequestId uuid.UUID,
	store Store,
) {
	ctx := r.Context()

	var result SupervisionResult
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	if result.Decision == Modify || result.Decision == Approve {
		if result.ChosenToolrequestId == nil {
			sendErrorResponse(w, http.StatusBadRequest, "Chosen tool request ID is required if you wish to modify or approve a given tool request", "")
			return
		}

		toolRequest, err := store.GetToolRequest(ctx, *result.ChosenToolrequestId)
		if err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "error getting tool request", err.Error())
			return
		}

		if toolRequest == nil {
			sendErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Tool request %s not found", *result.ChosenToolrequestId), "")
			return
		}
	}

	// Check that the group, chain and supervisor, and request exist
	id, err := store.CreateSupervisionResult(ctx, result, supervisionRequestId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error creating supervision result", err.Error())
		return
	}

	respondJSON(w, id, http.StatusCreated)
}

func apiGetHubStatsHandler(w http.ResponseWriter, _ *http.Request, hub *Hub) {
	stats, err := hub.getStats()
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting stats", err.Error())
		return
	}

	respondJSON(w, stats, http.StatusOK)
}

// apiGetProjectsHandler returns all projects
func apiGetProjectsHandler(w http.ResponseWriter, r *http.Request, store ProjectStore) {
	ctx := r.Context()

	projects, err := store.GetProjects(ctx)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting projects", err.Error())
		return
	}

	respondJSON(w, projects, http.StatusOK)
}

func apiGetProjectHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store ProjectStore) {
	ctx := r.Context()

	project, err := store.GetProject(ctx, id)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting project", err.Error())
		return
	}

	if project == nil {
		sendErrorResponse(w, http.StatusNotFound, "Project not found", "")
		return
	}

	respondJSON(w, project, http.StatusOK)
}

func apiGetRunStateHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, store Store) {
	ctx := r.Context()

	// First verify the run exists
	run, err := store.GetRun(ctx, runId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting run", err.Error())
		return
	}
	if run == nil {
		sendErrorResponse(w, http.StatusNotFound, "Run not found", "")
		return
	}

	// Get all request groups for this run
	requestGroups, err := store.GetRunRequestGroups(ctx, runId, false)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting request groups", err.Error())
		return
	}

	// Build the run state
	runState := make([]RunExecution, 0)

	// For each request group
	for _, requestGroup := range requestGroups {
		execution := RunExecution{
			RequestGroup: requestGroup,
			Chains:       make([]ChainExecutionState, 0),
		}

		// Get all tools from this request group
		for _, toolRequest := range requestGroup.ToolRequests {
			// Get all chains for this tool
			chains, err := store.GetSupervisorChains(ctx, toolRequest.ToolId)
			if err != nil {
				sendErrorResponse(w, http.StatusInternalServerError, "error getting chains", err.Error())
				return
			}

			// For each chain
			for _, chain := range chains {
				chainState := ChainExecutionState{
					Chain:               chain,
					SupervisionRequests: make([]SupervisionRequestState, 0),
				}

				// Get the chain execution from the chain ID + request group ID
				chainExecutionId, err := store.GetChainExecutionFromChainAndRequestGroup(ctx, chain.ChainId, *requestGroup.Id)
				if err != nil {
					sendErrorResponse(w, http.StatusInternalServerError, "error getting chain execution", err.Error())
					return
				}

				var supervisionRequests []SupervisionRequest

				// Get all supervision requests for this chain
				if chainExecutionId != nil {
					supervisionRequests, err = store.GetChainExecutionSupervisionRequests(ctx, *chainExecutionId)
					if err != nil {
						sendErrorResponse(w, http.StatusInternalServerError, "error getting supervision requests", err.Error())
						return
					}
				}

				// For each supervision request
				for _, request := range supervisionRequests {
					// Get the status
					status, err := store.GetSupervisionRequestStatus(ctx, *request.Id)
					if err != nil {
						sendErrorResponse(w, http.StatusInternalServerError, "error getting supervision status", err.Error())
						return
					}

					// Get the result if it exists
					var result *SupervisionResult
					if status.Status == "completed" {
						result, err = store.GetSupervisionResultFromRequestID(ctx, *request.Id)
						if err != nil {
							sendErrorResponse(w, http.StatusInternalServerError, "error getting supervision result", err.Error())
							return
						}
					}

					requestState := SupervisionRequestState{
						SupervisionRequest: request,
						Status:             *status,
						Result:             result,
					}

					chainState.SupervisionRequests = append(chainState.SupervisionRequests, requestState)
				}

				execution.Chains = append(execution.Chains, chainState)
			}
		}

		// TODO This seems super inefficient as it loads a bunch of duplicate stuff
		status, err := getRequestGroupStatus(ctx, *requestGroup.Id, store)
		if err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "error getting request group status", err.Error())
			return
		}

		execution.Status = status

		runState = append(runState, execution)
	}

	respondJSON(w, runState, http.StatusOK)
}

func apiGetSupervisionReviewPayloadHandler(w http.ResponseWriter, r *http.Request, supervisionRequestId uuid.UUID, store Store) {
	ctx := r.Context()

	// Get the supervision request
	supervisionRequest, err := store.GetSupervisionRequest(ctx, supervisionRequestId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervision request", err.Error())
		return
	}
	if supervisionRequest == nil {
		sendErrorResponse(w, http.StatusNotFound, "Supervision request not found", "")
		return
	}

	// Get the chain execution
	_, requestGroupId, err := store.GetChainExecution(ctx, *supervisionRequest.ChainexecutionId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting chain execution", err.Error())
		return
	}

	// Get the request group
	requestGroup, err := store.GetRequestGroup(ctx, *requestGroupId, true)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting request group", err.Error())
		return
	}

	// Get the chain state (all supervision requests and results for this chain execution)
	chainState, err := store.GetChainExecutionState(ctx, *supervisionRequest.ChainexecutionId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting chain state", err.Error())
		return
	}

	if len(requestGroup.ToolRequests) == 0 {
		sendErrorResponse(w, http.StatusInternalServerError, "request group has no tool requests", "")
		return
	}

	if requestGroup.ToolRequests[0].Id == nil {
		sendErrorResponse(w, http.StatusInternalServerError, "tool request ID is required", "")
		return
	}

	// Get the tool to find the run ID
	tool, err := store.GetTool(ctx, requestGroup.ToolRequests[0].ToolId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting tool", err.Error())
		return
	}

	if tool == nil {
		sendErrorResponse(w, http.StatusInternalServerError, "can't find run ID from tool", "")
		return
	}

	// Build the review payload
	reviewPayload := ReviewPayload{
		SupervisionRequest: *supervisionRequest,
		ChainState:         *chainState,
		RequestGroup:       *requestGroup,
		RunId:              tool.RunId,
	}

	respondJSON(w, reviewPayload, http.StatusOK)
}

// determineChainStatus checks if a supervision chain has completed
func determineChainStatus(requests []SupervisionRequestState, totalSupervisors int) Status {
	// No supervision requests means chain hasn't started
	if len(requests) == 0 {
		return Pending
	}

	// Find the last completed supervision in the chain
	var lastCompleted *SupervisionRequestState
	var highestPosition int = -1

	for _, req := range requests {
		if req.Status.Status == Completed {
			if req.SupervisionRequest.PositionInChain > highestPosition {
				highestPosition = req.SupervisionRequest.PositionInChain
				lastCompleted = &req
			}
		}
	}

	// No completed supervisions yet
	if lastCompleted == nil {
		return Pending
	}

	// Key change: If the last completed supervision approved the chain is complete
	// regardless of position
	if lastCompleted.Result != nil && lastCompleted.Result.Decision != Escalate {
		return Completed
	}

	// If we didn't get an approval, we need to check if we reached the end
	// and got a different valid completion (like Reject)
	if highestPosition == totalSupervisors-1 {
		if lastCompleted.Result != nil {
			return Completed
		}
	}

	return Pending
}

func getRequestGroupStatus(ctx context.Context, requestGroupId uuid.UUID, store Store) (Status, error) {
	chainExecutions, err := store.GetChainExecutionsFromRequestGroup(ctx, requestGroupId)
	if err != nil {
		return Pending, fmt.Errorf("error getting chain executions: %w", err)
	}

	// Track status for each chain execution
	executionStatuses := make([]Status, 0, len(chainExecutions))

	for _, execution := range chainExecutions {
		state, err := store.GetChainExecutionState(ctx, execution)
		if err != nil {
			return Pending, fmt.Errorf("error getting chain state: %w", err)
		}

		status := determineChainStatus(state.SupervisionRequests, len(state.Chain.Supervisors))
		executionStatuses = append(executionStatuses, status)
	}

	// Request group is complete only if all chains are complete
	status := Pending
	if allChainsComplete(executionStatuses) {
		status = Completed
	}

	return status, nil
}

func apiGetRequestGroupStatusHandler(w http.ResponseWriter, r *http.Request, requestGroupId uuid.UUID, store Store) {
	ctx := r.Context()

	status, err := getRequestGroupStatus(ctx, requestGroupId, store)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting request group status", err.Error())
		return
	}

	respondJSON(w, status, http.StatusOK)
}

// allChainsComplete checks if all supervision chains have completed
func allChainsComplete(statuses []Status) bool {
	if len(statuses) == 0 {
		return false
	}

	for _, status := range statuses {
		if status != Completed {
			return false
		}
	}
	return true
}

func apiGetRunStatusHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, store Store) {
	ctx := r.Context()

	run, err := store.GetRun(ctx, runId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting run status", err.Error())
		return
	}

	if run == nil {
		sendErrorResponse(w, http.StatusNotFound, "Run not found", "")
		return
	}

	respondJSON(w, run.Status, http.StatusOK)
}

func apiUpdateRunStatusHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, store Store) {
	ctx := r.Context()

	var status Status
	if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "error decoding run status", err.Error())
		return
	}

	// Update the run status
	err := store.UpdateRunStatus(ctx, runId, status)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error updating run status", err.Error())
		return
	}

	respondJSON(w, nil, http.StatusNoContent)
}

func apiUpdateRunResultHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, store Store) {
	ctx := r.Context()

	var result UpdateRunResultJSONBody
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "error decoding run result", err.Error())
		return
	}

	if result.Result == nil {
		sendErrorResponse(w, http.StatusBadRequest, "result is required", "")
		return
	}

	run, err := store.GetRun(ctx, runId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting run", err.Error())
		return
	}
	if run == nil {
		sendErrorResponse(w, http.StatusNotFound, "Run not found", "")
		return
	}

	task, err := store.GetTask(ctx, run.TaskId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting task", err.Error())
		return
	}
	if task == nil {
		sendErrorResponse(w, http.StatusNotFound, "Task not found", "")
		return
	}

	// Get the project's run result tags and check if the result is valid
	project, err := store.GetProject(ctx, task.ProjectId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting project", err.Error())
		return
	}
	if project == nil {
		sendErrorResponse(w, http.StatusNotFound, "Project not found", "")
		return
	}

	if !slices.Contains(project.RunResultTags, *result.Result) {
		sendErrorResponse(w, http.StatusBadRequest, "invalid run result", "")
		return
	}

	// Create the run result
	err = store.UpdateRunResult(ctx, runId, *result.Result)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error creating run result", err.Error())
		return
	}

	respondJSON(w, nil, http.StatusCreated)
}
