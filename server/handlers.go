package sentinel

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// respondJSON writes a JSON response with status 200 OK
func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	// If the header is not set, set it to StatusOK
	if w.Header().Get("Content-Type") == "" {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func apiCreateProjectHandler(w http.ResponseWriter, r *http.Request, store ProjectStore) {
	ctx := r.Context()

	log.Printf("received new project registration request")
	var request struct {
		Name string `json:"name"`
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	existingProject, err := store.GetProjectFromName(ctx, request.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting project: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	if existingProject != nil {
		w.WriteHeader(http.StatusOK)
		respondJSON(w, existingProject.Id.String())
		return
	}

	// Generate a new Project ID
	id := uuid.New()

	// Create the Project struct
	project := Project{
		Id:        id,
		Name:      request.Name,
		CreatedAt: time.Now(),
	}

	// Store the project in the global projects map
	err = store.CreateProject(ctx, project)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to register project, %v", err), http.StatusInternalServerError)
		return
	}

	// Send the response
	w.WriteHeader(http.StatusCreated)
	respondJSON(w, id.String())
}

func apiCreateProjectRunHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store RunStore) {
	ctx := r.Context()

	log.Printf("received new run request for project ID: %s", id)

	run := Run{
		Id:        uuid.Nil,
		ProjectId: id,
		CreatedAt: time.Now(),
	}

	runID, err := store.CreateRun(ctx, run)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("created run with ID: %s", run.Id)

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, runID)
}

func apiGetProjectRunsHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store RunStore) {
	ctx := r.Context()

	runs, err := store.GetProjectRuns(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, runs)
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
	// 	fmt.Printf("existing tool found with ID: %s", id)
	// 	respondJSON(w, id)
	// 	return
	// }

	toolId, err := store.CreateTool(ctx, runId, t.Attributes, t.Name, t.Description, t.IgnoredAttributes, t.Code)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error creating tool", err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, toolId)
}

func apiGetSupervisorHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store SupervisorStore) {
	ctx := r.Context()

	supervisor, err := store.GetSupervisor(ctx, id)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervisor", err.Error())
		return
	}

	respondJSON(w, supervisor)
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

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, chainIds)
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

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, supervisorId)
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

	respondJSON(w, chains)
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

	respondJSON(w, tools)
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

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, trg)
}

func apiGetRequestGroupHandler(w http.ResponseWriter, r *http.Request, requestGroupId uuid.UUID, store Store) {
	ctx := r.Context()

	requestGroup, err := store.GetRequestGroup(ctx, requestGroupId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting request group", err.Error())
		return
	}

	respondJSON(w, requestGroup)
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

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, toolRequestId)
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

	respondJSON(w, supervisionResult)
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

	requestGroups, err := store.GetRunRequestGroups(ctx, runId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting run request groups", err.Error())
		return
	}

	respondJSON(w, requestGroups)
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

	respondJSON(w, supervisors)
}

func apiGetProjectToolsHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store ToolStore) {
	ctx := r.Context()

	tools, err := store.GetProjectTools(ctx, id)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting project tools", err.Error())
		return
	}

	respondJSON(w, tools)
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

	respondJSON(w, status)
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
	requestGroup, err := store.GetRequestGroup(ctx, requestGroupId)
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

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, reviewID)
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

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, id)
}

func apiGetHubStatsHandler(w http.ResponseWriter, _ *http.Request, hub *Hub) {
	stats, err := hub.getStats()
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting stats", err.Error())
		return
	}

	respondJSON(w, stats)
}

// apiGetProjectsHandler returns all projects
func apiGetProjectsHandler(w http.ResponseWriter, r *http.Request, store ProjectStore) {
	ctx := r.Context()

	projects, err := store.GetProjects(ctx)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting projects", err.Error())
		return
	}

	respondJSON(w, projects)
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

	respondJSON(w, project)
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
	requestGroups, err := store.GetRunRequestGroups(ctx, runId)
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

		runState = append(runState, execution)
	}

	respondJSON(w, runState)
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
	requestGroup, err := store.GetRequestGroup(ctx, *requestGroupId)
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

	respondJSON(w, reviewPayload)
}

type SupervisorResultPosition struct {
	supervisor *Supervisor
	result     *SupervisionResult
	request    *SupervisionRequest
	position   *int
	status     *SupervisionStatus
}

func determineChainStatus(chainMap map[string]SupervisorResultPosition, totalSupervisors int) Status {
	// If we have no supervision requests yet, the chain is pending
	if len(chainMap) == 0 {
		return Pending
	}

	// Find the last completed supervision in the chain
	var lastCompletedPosition int = -1

	for _, srp := range chainMap {
		if srp.position != nil &&
			srp.status != nil &&
			srp.status.Status == Completed &&
			srp.result != nil {
			if *srp.position > lastCompletedPosition {
				lastCompletedPosition = *srp.position
			}
		}
	}

	// If we found no completed supervisions, chain is pending
	if lastCompletedPosition == -1 {
		return Pending
	}

	// If the last completed supervision was the final supervisor in the chain
	if lastCompletedPosition == totalSupervisors-1 {
		return Completed
	}

	// If we're here, we have completed supervisions but haven't reached the end
	return Pending
}

func apiGetRequestGroupStatusHandler(w http.ResponseWriter, r *http.Request, requestGroupId uuid.UUID, store Store) {
	ctx := r.Context()

	// Get all of the chain executions for this request group
	chainExecutions, err := store.GetChainExecutionsFromRequestGroup(ctx, requestGroupId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting request group", err.Error())
		return
	}

	fmt.Printf("Got these chain execution IDs for requestgroup_id %s: %v\n", requestGroupId.String(), chainExecutions)

	// For every chain of supervisors associated with this request group, we need to check the status
	executionStatuses := make([]Status, len(chainExecutions))

	// Get the chain execution state for this request group
	for _, execution := range chainExecutions {
		state, err := store.GetChainExecutionState(ctx, execution)
		if err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "error getting request group", err.Error())
			return
		}

		chainMap := map[string]SupervisorResultPosition{}

		// Iterate over every supervisor in the chain
		for _, supervisor := range state.Chain.Supervisors {
			var req SupervisionRequest
			var res SupervisionResult
			var pos int
			var sta SupervisionStatus
			// Find the request that corresponds to this supervisor (if one exists)
			for _, request := range state.SupervisionRequests {
				if request.SupervisionRequest.SupervisorId == *supervisor.Id {
					req = request.SupervisionRequest
					pos = request.SupervisionRequest.PositionInChain
					if request.Result != nil {
						res = *request.Result
					}
					sta = request.Status
				}
			}

			c := SupervisorResultPosition{
				// supervisor,
				nil,
				&res,
				&req,
				&pos,
				&sta,
			}

			chainMap[supervisor.Id.String()] = c
		}

		chainStatus := determineChainStatus(chainMap, len(state.Chain.Supervisors))

		executionStatuses = append(executionStatuses, chainStatus)
	}

	// Using the slice of chain execution statuses, compute the group's status
	completed := computeStatus(executionStatuses)

	groupStatus := Pending

	if completed {
		groupStatus = Completed
	}

	respondJSON(w, groupStatus)
}

// If every status is completed then the ToolRequestGroup has finished processing
// If any status is not completed then it's either (failed, pending, assigned, timeout) which means
// we have not finished processing
func computeStatus(statuses []Status) bool {
	fmt.Printf("Statuses in compute status: %v\n", statuses)
	fmt.Printf("number of statuses is: %d\n", len(statuses))

	var none Status
	var completedCount int64
	for _, s := range statuses {
		if !(s == Completed || s == none) {
			return false
		}
		if s == Completed {
			completedCount++
		}
	}

	return completedCount != 0
}

//
// for i, sr := range state.SupervisionRequests {
// fmt.Printf(">>> supervision requests: %v\n", sr.SupervisionRequest.Id)
// fmt.Printf(">>> supervision requests: %v\n", sr.SupervisionRequest.PositionInChain)
// if sr.SupervisionRequest.Status != nil {

// 	fmt.Printf(">>> supervision requests: %v\n", sr.SupervisionRequest.Status.Status)
// 	fmt.Printf(">>> supervision requests: %v\n", sr.SupervisionRequest.Status.CreatedAt)
// }
// fmt.Printf(">>> supervision requests: %v\n", sr.SupervisionRequest.SupervisorId)
// fmt.Printf("---> Querying status for supervision request: %s\n", sr.SupervisionRequest.Id)
// fmt.Printf("---> supervisionRequest %d latest status: %s (time %s)\n", i, sr.Status.Status, sr.Status.CreatedAt.String())
// 		if sr.SupervisionRequest.Status != nil {
// 			chainMap[sr.SupervisionRequest.PositionInChain] = sr.SupervisionRequest.Status.Status
// 		} else {
// 			chainMap[sr.SupervisionRequest.PositionInChain] = none
// 		}

// 		if i > x {
// 			x = i
// 		}
// 		fmt.Printf("chainMap is now %v\n", chainMap)
// 	}

// 	// Get the last status in the chain. This is our executionStatus
// 	lastStatus := chainMap[x]
// 	executionStatuses = append(executionStatuses, lastStatus)
// 	fmt.Printf("-> Execution state for %s: %s\n", execution.String(), lastStatus)
// }
