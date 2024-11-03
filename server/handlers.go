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
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// apiRegisterProjectHandler handles the POST /api/project/register endpoint
func apiRegisterProjectHandler(w http.ResponseWriter, r *http.Request, store ProjectStore) {
	ctx := r.Context()

	log.Printf("received new project registration request")
	var request ProjectCreate
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
		respondJSON(w, existingProject)
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
	respondJSON(w, project)
}

// apiCreateRunHandler handles the POST /api/run endpoint
func apiCreateRunHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store RunStore) {
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

	run.Id = runID

	log.Printf("created run with ID: %s", run.Id)

	respondJSON(w, run)
}

// apiGetProjectRunsHandler handles the GET /api/project/{id}/runs endpoint
func apiGetProjectRunsHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store RunStore) {
	ctx := r.Context()

	runs, err := store.GetProjectRuns(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, runs)
}

func apiGetRunsHandler(w http.ResponseWriter, r *http.Request, projectId uuid.UUID, store RunStore) {
	ctx := r.Context()

	runs, err := store.GetRuns(ctx, projectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, runs)
}

// apiGetRunHandler handles the GET /api/run/{id} endpoint
func apiGetRunHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store RunStore) {
	ctx := r.Context()

	run, err := store.GetRun(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if run == nil {
		http.Error(w, "Run not found", http.StatusNotFound)
		return
	}

	respondJSON(w, run)
}

// apiGetRunToolSupervisorsHandler handles the GET /api/runs/{runId}/tools/{toolId}/supervisors endpoint
func apiGetRunToolSupervisorsHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, toolId uuid.UUID, store SupervisorStore) {
	ctx := r.Context()

	supervisors, err := store.GetRunToolSupervisors(ctx, runId, toolId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if supervisors == nil {
		http.Error(w, "Supervisors not found", http.StatusNotFound)
		return
	}

	respondJSON(w, supervisors)
}

func apiCreateToolHandler(w http.ResponseWriter, r *http.Request, store ToolStore) {
	ctx := r.Context()

	var request Tool
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	var existingTool *Tool
	if request.Attributes != nil && request.Name != "" && request.Description != "" && request.IgnoredAttributes != nil {
		found, err := store.GetToolFromValues(ctx, *request.Attributes, request.Name, request.Description, *request.IgnoredAttributes)
		if err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "error trying to locate an existing tool", err.Error())
			return
		}
		if found != nil {
			existingTool = found
		}
	}

	if existingTool != nil {
		respondJSON(w, existingTool)
		return
	}

	toolId, err := store.CreateTool(ctx, request)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error creating tool", err.Error())
		return
	}

	request.Id = &toolId

	respondJSON(w, request)
}

// apiCreateRunToolSupervisorsHandler handles the POST /api/runs/{runId}/tools/{toolId}/supervisors endpoint
func apiCreateRunToolSupervisorsHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, toolId uuid.UUID, store SupervisorStore) {
	ctx := r.Context()

	var request SupervisorChainAssignment
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	if len(request) == 0 {
		sendErrorResponse(w, http.StatusBadRequest, "No supervisors provided", "")
		return
	}

	err = store.AssignSupervisorsToTool(ctx, runId, toolId, request)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error assigning supervisors to tool", err.Error())
		return
	}

	respondJSON(w, nil)
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

func apiCreateSupervisorHandler(w http.ResponseWriter, r *http.Request, store SupervisorStore) {
	ctx := r.Context()

	var request Supervisor
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	// Create new supervisor
	supervisorId, err := store.CreateSupervisor(ctx, request)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error creating supervisor", err.Error())
		return
	}

	request.Id = &supervisorId
	respondJSON(w, request)
}

func apiGetSupervisorsHandler(w http.ResponseWriter, r *http.Request, params GetSupervisorsParams, store SupervisorStore) {
	ctx := r.Context()

	supervisors, err := store.GetSupervisors(ctx, params.ProjectId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervisors", err.Error())
		return
	}

	respondJSON(w, supervisors)
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

// Get the tools for a project
func apiGetToolsHandler(w http.ResponseWriter, r *http.Request, params GetToolsParams, store ToolStore) {
	ctx := r.Context()

	tools, err := store.GetProjectTools(ctx, params.ProjectId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting tools", err.Error())
		return
	}

	respondJSON(w, tools)
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

// apiCreateSupervisionRequestHandler receives supervisor requests via the HTTP API
func apiCreateSupervisionRequestHandler(w http.ResponseWriter, r *http.Request, store Store) {
	ctx := r.Context()

	t := time.Now()

	var request SupervisionRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	if len(request.ToolRequests) == 0 || len(request.Messages) == 0 {
		sendErrorResponse(w, http.StatusBadRequest, "No tool choices or messages provided", "")
		return
	}

	if len(request.ToolRequests) != len(request.Messages) {
		sendErrorResponse(w, http.StatusBadRequest, "Number of tool choices and messages must be the same", "")
		return
	}

	execution, err := store.GetExecution(ctx, request.ExecutionId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting execution when creating supervision request", err.Error())
		return
	}

	if execution == nil {
		sendErrorResponse(w, http.StatusNotFound, "Execution not found", "")
		return
	}

	// We don't support n requests for different tools yet. They must be the same tool.
	toolID := request.ToolRequests[0].ToolId
	for _, toolRequest := range request.ToolRequests {
		if toolRequest.ToolId != toolID {
			sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Agent submitted %d samples, some of which were not the same tool choice", len(request.ToolRequests)), "")
			return
		}
	}

	// Check the tool exists
	tool, err := store.GetTool(ctx, toolID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting tool", err.Error())
		return
	}

	if tool == nil {
		sendErrorResponse(w, http.StatusNotFound, "Tool not found", "")
		return
	}

	// Get the supervisors that are supposed to be supervising this tool for this run
	supervisorChains, err := store.GetRunToolSupervisors(ctx, request.RunId, toolID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervisors for tool", err.Error())
		return
	}

	// Check supervisor ID exists in the lists of supervisors for this tool/run combination
	exists := false
	for _, chain := range supervisorChains {
		for _, supervisor := range chain.Supervisors {
			if supervisor.Id.String() == request.SupervisorId.String() {
				exists = true
				break
			}
		}
	}

	if !exists {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Trying to do supervision with supervisor %s that is not associated with tool %s for run %s. All supervisors must be registered to a tool for a run before the run is started.", request.SupervisorId, toolID, request.RunId), "")
		return
	}

	// Store the supervisor in the database
	reviewID, err := store.CreateSupervisionRequest(ctx, request)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error creating supervision request", err.Error())
		return
	}

	response := SupervisionStatus{
		SupervisionRequestId: &reviewID,
		Status:               Pending,
		CreatedAt:            t,
	}

	// Respond immediately with 200 OK.
	// The client will receive and ID they can use to poll the status of their supervisor
	respondJSON(w, response)
}

func apiGetExecutionHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	ctx := r.Context()

	execution, err := store.GetExecution(ctx, id)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting execution", err.Error())
		return
	}

	respondJSON(w, execution)
}

func apiGetRunExecutionsHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, store Store) {
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

	executions, err := store.GetRunExecutions(ctx, runId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting run executions", err.Error())
		return
	}

	respondJSON(w, executions)
}

func apiCreateExecutionHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, store Store) {
	ctx := r.Context()

	var request struct {
		ToolId uuid.UUID `json:"toolId"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	executionId, err := store.CreateExecution(ctx, runId, request.ToolId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error creating execution", err.Error())
		return
	}

	execution, err := store.GetExecution(ctx, executionId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting execution", err.Error())
		return
	}

	if execution == nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Something went wrong, execution not found", "")
		return
	}

	respondJSON(w, execution)
}

// apiGetSupervisionRequestHandler handles the GET /api/supervisor/{id} endpoint
func apiGetSupervisionRequestHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	ctx := r.Context()

	supervisor, err := store.GetSupervisionRequest(ctx, id)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervisor", err.Error())
		return
	}

	if supervisor == nil {
		sendErrorResponse(w, http.StatusNotFound, "Supervisor not found", "")
		return
	}

	respondJSON(w, supervisor)
}

// apiGetSupervisionRequestsHandler handles the GET /api/supervisor endpoint
func apiGetSupervisionRequestsHandler(w http.ResponseWriter, r *http.Request, _ GetSupervisionRequestsParams, store Store) {
	ctx := r.Context()

	reviews, err := store.GetSupervisionRequests(ctx)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervision requests", err.Error())
		return
	}

	respondJSON(w, reviews)
}

// apiGetSupervisionResultsHandler handles the GET /api/supervisor/{id}/results endpoint
func apiGetSupervisionResultsHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	ctx := r.Context()

	// First check if supervisor exists
	supervisor, err := store.GetSupervisionRequest(ctx, id)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervisor", err.Error())
		return
	}

	if supervisor == nil {
		sendErrorResponse(w, http.StatusNotFound, "Supervisor not found", "")
		return
	}

	results, err := store.GetSupervisionResults(ctx, id)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervision results", err.Error())
		return
	}

	respondJSON(w, results)
}

// apiCreateSupervisionResultHandler handles the POST /api/supervisor/{id}/results endpoint
func apiCreateSupervisionResultHandler(w http.ResponseWriter, r *http.Request, _ uuid.UUID, store Store) {
	ctx := r.Context()

	var request CreateSupervisionResult
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	fmt.Printf("CreateSupervisionResultHandler: %v\n", request)

	supervisorChains, err := store.GetRunToolSupervisors(ctx, request.RunId, request.ToolId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervisors for tool", err.Error())
		return
	}

	if len(supervisorChains) == 0 {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("No supervisors found for run %s and tool %s", request.RunId, request.ToolId), "")
		return
	}

	// Check supervisor ID exists in the lists of supervisors for this tool/run combination
	exists := false
	for _, chain := range supervisorChains {
		for _, supervisor := range chain.Supervisors {
			if supervisor.Id.String() == request.SupervisorId.String() {
				exists = true
				break
			}
		}
	}

	if !exists {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Supervisor %s not associated with tool %s for run %s", request.SupervisorId, request.ToolId, request.RunId), "")
		return
	}

	err = store.CreateSupervisionResult(ctx, request.SupervisionResult)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error creating supervision result", err.Error())
		return
	}

	respondJSON(w, request.SupervisionResult)
}

// apiSupervisionStatusHandler checks the status of a supervisor request
func apiSupervisionStatusHandler(w http.ResponseWriter, r *http.Request, reviewID uuid.UUID, store Store) {
	ctx := r.Context()
	// Use the reviewID directly
	supervisor, err := store.GetSupervisionRequest(ctx, reviewID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervisor", err.Error())
		return
	}

	if supervisor == nil {
		sendErrorResponse(w, http.StatusNotFound, "Supervisor not found", "")
		return
	}

	respondJSON(w, supervisor.Status)
}

// apiGetReviewToolRequestsHandler handles the GET /api/supervisor/{id}/toolrequests endpoint
func apiGetReviewToolRequestsHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	ctx := r.Context()

	// First check if supervisor exists
	supervisor, err := store.GetSupervisionRequest(ctx, id)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervisor", err.Error())
		return
	}

	if supervisor == nil {
		sendErrorResponse(w, http.StatusNotFound, "Supervisor not found", "")
		return
	}

	results, err := store.GetSupervisionToolRequests(ctx, id)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting supervisor tool requests", err.Error())
		return
	}

	respondJSON(w, results)
}

func apiStatsHandler(hub *Hub, w http.ResponseWriter, _ *http.Request) {
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

// apiGetProjectByIdHandler handles the GET /api/project/{id} endpoint
func apiGetProjectByIdHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store ProjectStore) {
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

// apiLLMExplanationHandler receives a code snippet and returns an explanation and a danger score by calling an LLM
func apiLLMExplanationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		Text string `json:"text"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	explanation, score, err := getExplanationFromLLM(ctx, request.Text)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to get explanation from LLM", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"explanation": explanation, "score": score})
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to get explanation from LLM", err.Error())
		return
	}
}

// apiGetExecutionSupervisionsHandler handles the GET /api/executions/{executionId}/supervisions endpoint
func apiGetExecutionSupervisionsHandler(w http.ResponseWriter, r *http.Request, executionId uuid.UUID, store Store) {
	ctx := r.Context()

	// First check if execution exists
	execution, err := store.GetExecution(ctx, executionId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting execution", err.Error())
		return
	}

	if execution == nil {
		sendErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Execution not found for ID %s", executionId), "")
		return
	}

	supervisions, err := store.GetExecutionSupervisions(ctx, executionId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting execution supervisions", err.Error())
		return
	}

	// Combine the data into ExecutionSupervisions response
	response := &ExecutionSupervisions{
		ExecutionId:  executionId,
		Supervisions: supervisions,
	}

	respondJSON(w, response)
}
