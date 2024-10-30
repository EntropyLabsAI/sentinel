package sentinel

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

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
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(existingProject)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error encoding existing project: %v", err.Error()), http.StatusInternalServerError)
			return
		}
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(project)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(run)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiGetProjectRunsHandler handles the GET /api/project/{id}/runs endpoint
func apiGetProjectRunsHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store RunStore) {
	ctx := r.Context()

	runs, err := store.GetProjectRuns(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(runs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func apiGetRunsHandler(w http.ResponseWriter, r *http.Request, projectId uuid.UUID, store RunStore) {
	ctx := r.Context()

	runs, err := store.GetRuns(ctx, projectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(runs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(run)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(supervisors)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func apiCreateToolHandler(w http.ResponseWriter, r *http.Request, store ToolStore) {
	ctx := r.Context()

	var request Tool
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var existingTool *Tool
	if request.Attributes != nil && request.Name != "" && request.Description != "" {
		found, err := store.GetToolFromValues(ctx, *request.Attributes, request.Name, request.Description)
		if err != nil {
			http.Error(w, fmt.Sprintf("error trying to locate an existing tool: %v", err), http.StatusInternalServerError)
			return
		}
		if found != nil {
			existingTool = found
		}
	}

	if existingTool != nil {
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(existingTool)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	toolId, err := store.CreateTool(ctx, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	request.Id = &toolId

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiCreateRunToolSupervisorsHandler handles the POST /api/runs/{runId}/tools/{toolId}/supervisors endpoint
func apiCreateRunToolSupervisorsHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, toolId uuid.UUID, store SupervisorStore) {
	ctx := r.Context()

	var request []uuid.UUID
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(request) == 0 {
		http.Error(w, "No supervisors provided", http.StatusBadRequest)
		return
	}

	err = store.AssignSupervisorsToTool(ctx, runId, toolId, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func apiGetSupervisorHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store SupervisorStore) {
	ctx := r.Context()

	supervisor, err := store.GetSupervisor(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(supervisor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func apiCreateSupervisorHandler(w http.ResponseWriter, r *http.Request, store SupervisorStore) {
	ctx := r.Context()

	var request Supervisor
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if the supervisor submitted for creation already exists in the db by looking up the name,code,type,desc
	var existingSupervisor *Supervisor
	if request.Code != nil && request.Name != "" && request.Description != "" && request.Type != "" {
		found, err := store.GetSupervisorFromValues(ctx, *request.Code, request.Name, request.Description, request.Type)
		if err != nil {
			http.Error(w, fmt.Sprintf("error trying to locate an existing supervisor: %v", err), http.StatusInternalServerError)
			return
		}
		if found != nil {
			existingSupervisor = found
		}
	}

	if existingSupervisor != nil {
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(existingSupervisor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	supervisorId, err := store.CreateSupervisor(ctx, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	request.Id = &supervisorId

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func apiGetSupervisorsHandler(w http.ResponseWriter, r *http.Request, params GetSupervisorsParams, store SupervisorStore) {
	ctx := r.Context()

	supervisors, err := store.GetSupervisors(ctx, params.ProjectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(supervisors)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func apiGetToolHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store ToolStore) {
	ctx := r.Context()

	tool, err := store.GetTool(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if tool == nil {
		http.Error(w, "Tool not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(tool)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get the tools for a project
func apiGetToolsHandler(w http.ResponseWriter, r *http.Request, params GetToolsParams, store ToolStore) {
	ctx := r.Context()

	tools, err := store.GetProjectTools(ctx, params.ProjectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(tools)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func apiGetRunToolsHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	ctx := r.Context()

	// First check if run exists
	run, err := store.GetRun(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if run == nil {
		http.Error(w, "Run not found", http.StatusNotFound)
		return
	}

	tools, err := store.GetRunTools(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(tools)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiCreateSupervisionRequestHandler receives supervisor requests via the HTTP API
func apiCreateSupervisionRequestHandler(w http.ResponseWriter, r *http.Request, store Store) {
	ctx := r.Context()

	t := time.Now()

	var request SupervisionRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(request.ToolRequests) == 0 || len(request.Messages) == 0 {
		http.Error(w, "No tool choices or messages provided", http.StatusBadRequest)
		return
	}

	if len(request.ToolRequests) != len(request.Messages) {
		http.Error(w, "Number of tool choices and messages must be the same", http.StatusBadRequest)
		return
	}

	execution, err := store.GetExecution(ctx, request.ExecutionId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting execution when creating supervision request: %v", err), http.StatusInternalServerError)
		return
	}

	if execution == nil {
		http.Error(w, "Execution not found", http.StatusNotFound)
		return
	}

	// We don't support n requests for different tools yet. They must be the same tool.
	toolID := request.ToolRequests[0].ToolId
	for _, toolRequest := range request.ToolRequests {
		if toolRequest.ToolId != toolID {
			http.Error(w, fmt.Sprintf("Agent submitted %d samples, some of which were not the same tool choice", len(request.ToolRequests)), http.StatusBadRequest)
			return
		}
	}

	// Check the tool exists
	tool, err := store.GetTool(ctx, toolID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if tool == nil {
		http.Error(w, "Tool not found", http.StatusNotFound)
		return
	}

	// Get the supervisors that are supposed to be supervising this tool for this run
	supervisors, err := store.GetRunToolSupervisors(ctx, request.RunId, toolID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check supervisor ID exists in the list of supervisors for this tool/run combination
	exists := false
	for _, supervisor := range supervisors {
		if supervisor.Id.String() == request.SupervisorId.String() {
			exists = true
			break
		}
	}

	if !exists {
		http.Error(w, fmt.Sprintf("Trying to do supervision with supervisor %s that is not associated with tool %s for run %s. All supervisors must be registered to a tool for a run before the run is started.", request.SupervisorId, toolID, request.RunId), http.StatusBadRequest)
		return
	}

	// Store the supervisor in the database
	reviewID, err := store.CreateSupervisionRequest(ctx, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := SupervisionStatus{
		SupervisionRequestId: &reviewID,
		Status:               Pending,
		CreatedAt:            t,
	}

	// Respond immediately with 200 OK.
	// The client will receive and ID they can use to poll the status of their supervisor
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func apiGetExecutionHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	ctx := r.Context()

	execution, err := store.GetExecution(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(execution)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func apiGetRunExecutionsHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, store Store) {
	ctx := r.Context()

	run, err := store.GetRun(ctx, runId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if run == nil {
		http.Error(w, "Run not found", http.StatusNotFound)
		return
	}

	executions, err := store.GetRunExecutions(ctx, runId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(executions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func apiCreateExecutionHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, store Store) {
	ctx := r.Context()

	var request struct {
		ToolId uuid.UUID `json:"toolId"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	executionId, err := store.CreateExecution(ctx, runId, request.ToolId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	execution, err := store.GetExecution(ctx, executionId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if execution == nil {
		http.Error(w, "Something went wrong, execution not found", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(execution)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiGetSupervisionRequestHandler handles the GET /api/supervisor/{id} endpoint
func apiGetSupervisionRequestHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	ctx := r.Context()

	supervisor, err := store.GetSupervisionRequest(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if supervisor == nil {
		http.Error(w, "Supervisor not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(supervisor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiGetSupervisionRequestsHandler handles the GET /api/supervisor endpoint
func apiGetSupervisionRequestsHandler(w http.ResponseWriter, r *http.Request, _ GetSupervisionRequestsParams, store Store) {
	ctx := r.Context()

	reviews, err := store.GetSupervisionRequests(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(reviews)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiGetSupervisionResultsHandler handles the GET /api/supervisor/{id}/results endpoint
func apiGetSupervisionResultsHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	ctx := r.Context()

	// First check if supervisor exists
	supervisor, err := store.GetSupervisionRequest(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if supervisor == nil {
		http.Error(w, "Supervisor not found", http.StatusNotFound)
		return
	}

	results, err := store.GetSupervisionResults(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiCreateSupervisionResultHandler handles the POST /api/supervisor/{id}/results endpoint
func apiCreateSupervisionResultHandler(w http.ResponseWriter, r *http.Request, _ uuid.UUID, store Store) {
	ctx := r.Context()

	var request CreateSupervisionResult
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	supervisors, err := store.GetRunToolSupervisors(ctx, request.RunId, request.ToolId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(supervisors) == 0 {
		http.Error(w, fmt.Sprintf("No supervisors found for run %s and tool %s", request.RunId, request.ToolId), http.StatusBadRequest)
		return
	}

	// Check if the supervisor is associated with the tool for the run
	found := false
	for _, supervisor := range supervisors {
		if supervisor.Id.String() == request.SupervisorId.String() {
			fmt.Printf("Supervisor %s found for run %s and tool %s", supervisor.Id, request.RunId, request.ToolId)
			found = true
			break
		}
	}

	if !found {
		http.Error(w, fmt.Sprintf("Supervisor %s not associated with tool %s for run %s", request.SupervisorId, request.ToolId, request.RunId), http.StatusBadRequest)
		return
	}

	err = store.CreateSupervisionResult(ctx, request.SupervisionResult)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// apiSupervisionStatusHandler checks the status of a supervisor request
func apiSupervisionStatusHandler(w http.ResponseWriter, r *http.Request, reviewID uuid.UUID, store Store) {
	ctx := r.Context()
	// Use the reviewID directly
	supervisor, err := store.GetSupervisionRequest(ctx, reviewID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if supervisor == nil {
		http.Error(w, "Supervisor not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(supervisor.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiGetReviewToolRequestsHandler handles the GET /api/supervisor/{id}/toolrequests endpoint
func apiGetReviewToolRequestsHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	ctx := r.Context()

	// First check if supervisor exists
	supervisor, err := store.GetSupervisionRequest(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if supervisor == nil {
		http.Error(w, "Supervisor not found", http.StatusNotFound)
		return
	}

	results, err := store.GetSupervisionToolRequests(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func apiStatsHandler(hub *Hub, w http.ResponseWriter, _ *http.Request) {
	stats, err := hub.getStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(stats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiGetProjectsHandler returns all projects
func apiGetProjectsHandler(w http.ResponseWriter, r *http.Request, store ProjectStore) {
	ctx := r.Context()

	projects, err := store.GetProjects(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(projects)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiGetProjectByIdHandler handles the GET /api/project/{id} endpoint
func apiGetProjectByIdHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store ProjectStore) {
	ctx := r.Context()

	project, err := store.GetProject(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if project == nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(project); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiLLMExplanationHandler receives a code snippet and returns an explanation and a danger score by calling an LLM
func apiLLMExplanationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		Text string `json:"text"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	explanation, score, err := getExplanationFromLLM(ctx, request.Text)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		http.Error(w, "Failed to get explanation from LLM", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"explanation": explanation, "score": score})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiGetExecutionSupervisionsHandler handles the GET /api/executions/{executionId}/supervisions endpoint
func apiGetExecutionSupervisionsHandler(w http.ResponseWriter, r *http.Request, executionId uuid.UUID, store Store) {
	ctx := r.Context()

	// First check if execution exists
	execution, err := store.GetExecution(ctx, executionId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if execution == nil {
		http.Error(w, fmt.Sprintf("Execution not found for ID %s", executionId), http.StatusNotFound)
		return
	}

	// Get all supervision requests for this execution
	requests, err := store.GetSupervisionRequestsForExecution(ctx, executionId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get all supervision results
	results, err := store.GetSupervisionResultsForExecution(ctx, executionId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get all supervision statuses
	statuses, err := store.GetSupervisionStatusesForExecution(ctx, executionId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Combine the data into ExecutionSupervisions response
	response := &ExecutionSupervisions{
		ExecutionId: executionId,
		Requests:    requests,
		Results:     results,
		Statuses:    statuses,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
