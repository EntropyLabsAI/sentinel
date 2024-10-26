package sentinel

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

var completedHumanReviews = &sync.Map{}
var completedLLMReviews = &sync.Map{}

// reviewChannels maps a reviews ID to the channel configured to receive the reviewer's response
var reviewChannels = &sync.Map{}

// Timeout duration for waiting for the reviewer to respond
const reviewTimeout = 1440 * time.Minute

// serveWs upgrades the HTTP connection to a WebSocket connection and registers the client with the hub
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	client := &Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan Review),
	}
	hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
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
func apiGetRunHandler(w http.ResponseWriter, r *http.Request, projectId uuid.UUID, id uuid.UUID, store RunStore) {
	ctx := r.Context()

	run, err := store.GetRun(ctx, projectId, id)
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

// apiGetToolSupervisorsHandler handles the GET /api/tools/{id}/supervisors endpoint
func apiGetToolSupervisorsHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store SupervisorStore) {
	ctx := r.Context()

	supervisors, err := store.GetSupervisorFromToolID(ctx, id)
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

// apiAssignSupervisorToToolHandler handles the POST /api/tools/{id}/supervisors endpoint
func apiAssignSupervisorToToolHandler(w http.ResponseWriter, r *http.Request, _ uuid.UUID, id uuid.UUID, store Store) {
	ctx := r.Context()

	// Decode the request body
	var request struct {
		SupervisorId uuid.UUID `json:"supervisorId"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.SupervisorId == uuid.Nil {
		http.Error(w, "Supervisor ID is required", http.StatusBadRequest)
		return
	}

	log.Printf("received new supervisor assignment request for tool ID: %s and supervisor ID: %s", id, request.SupervisorId)

	err = store.AssignSupervisorToTool(ctx, request.SupervisorId, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// apiCreateToolHandler handles the POST /api/tool endpoint
func apiCreateRunToolHandler(w http.ResponseWriter, r *http.Request, _ uuid.UUID, runId uuid.UUID, store ToolStore) {
	ctx := r.Context()

	var request ToolCreate
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	t := time.Now()

	tool := Tool{
		Id:        uuid.Nil,
		Name:      request.Name,
		CreatedAt: &t,
	}

	toolId, err := store.CreateRunTool(ctx, runId, tool)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tool.Id = toolId

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(tool)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

func apiGetSupervisorsHandler(w http.ResponseWriter, r *http.Request, store SupervisorStore) {
	ctx := r.Context()

	supervisors, err := store.GetSupervisors(ctx)
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

func apiGetToolsHandler(w http.ResponseWriter, r *http.Request, store ToolStore) {
	ctx := r.Context()

	tools, err := store.GetTools(ctx)
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

func apiGetRunToolsHandler(w http.ResponseWriter, r *http.Request, projectId uuid.UUID, id uuid.UUID, store Store) {
	ctx := r.Context()

	// First check if run exists
	run, err := store.GetRun(ctx, projectId, id)
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

// apiReviewHandler receives review requests via the HTTP API
func apiReviewHandler(hub *Hub, w http.ResponseWriter, r *http.Request, store Store) {
	ctx := r.Context()

	t := time.Now()

	var request ReviewRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	review := Review{
		RunId:     request.RunId,
		TaskState: request.TaskState,
		Status: &ReviewStatus{
			Status:    Pending,
			CreatedAt: t,
		},
	}

	// Store the review in the database
	reviewID, err := store.CreateReview(ctx, review)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	review.Id = reviewID

	if len(request.ToolRequests) == 0 {
		http.Error(w, "No tool choices provided", http.StatusBadRequest)
		return
	}

	toolChoiceID := request.ToolRequests[0].Id

	for _, toolRequest := range request.ToolRequests {
		if toolRequest.Id != toolChoiceID {
			http.Error(w, fmt.Sprintf("Agent submitted %d samples, some of which were not the same tool choice", len(request.ToolRequests)), http.StatusBadRequest)
			return
		}
	}

	// Handle the review depending on the type of supervisor
	supervisor, err := store.GetSupervisorFromToolID(ctx, toolChoiceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch supervisor.Type {
	case Human:
		if err := processHumanReview(ctx, hub, review, store); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case Llm:
		err := processLLMReview(ctx, request, store)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, "Invalid supervisor type", http.StatusBadRequest)
		return
	}

	response := ReviewStatus{
		Id:        review.Id,
		Status:    Pending,
		CreatedAt: t,
	}

	// Respond immediately with 200 OK.
	// The client will receive and ID they can use to poll the status of their review
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiGetReviewHandler handles the GET /api/review/{id} endpoint
func apiGetReviewHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	ctx := r.Context()

	review, err := store.GetReview(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if review == nil {
		http.Error(w, "Review not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(review)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiGetReviewsHandler handles the GET /api/reviews endpoint
func apiGetReviewsHandler(w http.ResponseWriter, r *http.Request, _ GetReviewsParams, store Store) {
	ctx := r.Context()

	reviews, err := store.GetReviews(ctx)
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

// apiGetReviewResultsHandler handles the GET /api/review/{id}/results endpoint
func apiGetReviewResultsHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	ctx := r.Context()

	// First check if review exists
	review, err := store.GetReview(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if review == nil {
		http.Error(w, "Review not found", http.StatusNotFound)
		return
	}

	results, err := store.GetReviewResults(ctx, id)
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

// apiReviewStatusHandler checks the status of a review request
func apiReviewStatusHandler(w http.ResponseWriter, r *http.Request, reviewID uuid.UUID, store Store) {
	ctx := r.Context()
	// Use the reviewID directly
	review, err := store.GetReview(ctx, reviewID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if review == nil {
		http.Error(w, "Review not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(review.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiGetReviewToolRequestsHandler handles the GET /api/review/{id}/toolrequests endpoint
func apiGetReviewToolRequestsHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	ctx := r.Context()

	// First check if review exists
	review, err := store.GetReview(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if review == nil {
		http.Error(w, "Review not found", http.StatusNotFound)
		return
	}

	results, err := store.GetReviewToolRequests(ctx, id)
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
	stats := hub.getStats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(stats)
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
