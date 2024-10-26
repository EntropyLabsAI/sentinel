package sentinel

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
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
		CreatedAt: time.Now().Unix(),
	}

	// Store the project in the global projects map
	err = store.CreateProject(ctx, project)
	if err != nil {
		http.Error(w, "Failed to register project", http.StatusInternalServerError)
		return
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(project)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiCreateRunHandler handles the POST /api/run endpoint
func apiCreateRunHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store RunStore) {
	ctx := r.Context()

	log.Printf("received new run request")

	run := Run{
		Id:        uuid.New(),
		ProjectId: id,
		CreatedAt: time.Now().Unix(),
	}

	err := store.CreateRun(ctx, run)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(run)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiCreateToolHandler handles the POST /api/tool endpoint
func apiCreateToolHandler(w http.ResponseWriter, r *http.Request, store ToolStore) {
	ctx := r.Context()

	var request ToolCreate
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	t := time.Now().Unix()

	tool := Tool{
		Id:        uuid.New(),
		Name:      request.Name,
		CreatedAt: &t,
	}

	err = store.CreateTool(ctx, tool)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(tool)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// apiReviewHandler receives review requests via the HTTP API
func apiReviewHandler(hub *Hub, w http.ResponseWriter, r *http.Request, store Store) {
	ctx := r.Context()

	var request ReviewRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := uuid.New()

	review := Review{
		Id:        id,
		RunId:     request.RunId,
		TaskState: request.TaskState,
		Status: &ReviewStatus{
			Status:    Pending,
			CreatedAt: time.Now().Unix(),
		},
	}

	// Handle the review depending on the type of supervisor
	supervisor, err := store.GetSupervisorFromToolID(ctx, request.ToolChoices[0].Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch supervisor.Type {
	case SupervisorTypeHuman:
		if err := processHumanReview(ctx, hub, review, store); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case SupervisorTypeLlm:
		// if err := processLLMReview(hub, w, r, store); err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
	default:
		http.Error(w, "Invalid supervisor type", http.StatusBadRequest)
		return
	}

	response := ReviewStatus{
		Id:        id,
		Status:    Pending,
		CreatedAt: time.Now().Unix(),
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

func processHumanReview(ctx context.Context, hub *Hub, review Review, store Store) error {
	// Add the review request to the human review queue
	hub.ReviewChan <- review

	log.Printf("received new review request ID %s via API.", review.Id)

	// Create a channel for this review request
	responseChan := make(chan ReviewResult)
	reviewChannels.Store(review.Id, responseChan)

	// Start a goroutine to wait for the response
	go func() {
		select {
		case response := <-responseChan:
			// Store the completed review
			completedHumanReviews.Store(response.Id, response)
			reviewChannels.Delete(response.Id)
			log.Printf("review ID %s completed with decision: %s.", response.Id, response.Decision)
		case <-time.After(reviewTimeout):

			reviewStatus := ReviewStatus{
				Id:        review.Id,
				Status:    Timeout,
				CreatedAt: time.Now().Unix(),
			}

			// Timeout occurred
			completedHumanReviews.Store(review.Id, reviewStatus)
			reviewChannels.Delete(review.Id)
			log.Printf("review ID %s timed out.", review.Id)
		}
	}()

	return nil
}

// apiGetReviewHandler handles the GET /api/review/{id} endpoint
func apiGetReviewHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	ctx := r.Context()

	review, err := store.GetReview(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

// apiGetReviewResultsHandler handles the GET /api/review/{id}/results endpoint
func apiGetReviewResultsHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	ctx := r.Context()

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
func apiReviewStatusHandler(w http.ResponseWriter, _ *http.Request, reviewID string) {
	// Use the reviewID directly
	if _, ok := reviewChannels.Load(reviewID); ok {
		// There's a pending review
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(map[string]string{"status": "pending"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Check if there's a stored response for this review
		if response, ok := completedHumanReviews.Load(reviewID); ok {
			log.Printf("status request for review ID %s: completed\n", reviewID)

			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// Review not found
			w.WriteHeader(http.StatusNotFound)
			err := json.NewEncoder(w).Encode(map[string]string{"status": "not_found"})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}

// apiLLMExplanationHandler receives a code snippet and returns an explanation and a danger score by calling an LLM
func apiLLMExplanationHandler(w http.ResponseWriter, r *http.Request) {
	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

func apiReviewLLMHandler(w http.ResponseWriter, r *http.Request, store Store) {
	id := uuid.New().String()

	log.Printf("received new LLM review request, ID: %s", id)

	// Parse the request body to get the same input as /api/review
	var reviewRequest ReviewRequest
	err := json.NewDecoder(r.Body).Decode(&reviewRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO allow LLM reviewer to handle multiple tool choice options
	if len(reviewRequest.ToolChoices) != 1 {
		http.Error(w, "Invalid number of tool choices provided for LLM review", http.StatusBadRequest)
		return
	}

	toolChoice := reviewRequest.ToolChoices[0]

	// Call the LLM to evaluate the tool_choice
	llmReasoning, decision, err := callLLMForReview(r.Context(), toolChoice, store)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error calling LLM: %v", err), http.StatusInternalServerError)
		return
	}

	resultID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing UUID: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := ReviewResult{
		Id:          resultID,
		Decision:    decision,
		Toolrequest: &toolChoice,
		Reasoning:   llmReasoning,
	}

	// Store the completed LLM review
	completedLLMReviews.Store(id, response)

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
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

// getLLMResponse is a helper function that interacts with the OpenAI API and returns the LLM response.
func getLLMResponse(ctx context.Context, messages []openai.ChatCompletionMessage, model string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(apiKey)

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		},
	)

	if err != nil {
		return "", fmt.Errorf("error creating LLM chat completion: %v", err)
	}

	return resp.Choices[0].Message.Content, nil
}

// apiGetLLMReviews returns all LLM reviews
func apiGetLLMReviews(w http.ResponseWriter, _ *http.Request) {
	reviews := make([]ReviewResult, 0)

	completedLLMReviews.Range(func(key, value any) bool {
		reviews = append(reviews, value.(ReviewResult))
		return true
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(reviews)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiGetProjectsHandler returns all projects
func apiGetProjectsHandler(w http.ResponseWriter, r *http.Request, store ProjectStore) {
	ctx := r.Context()

	projects, err := store.ListProjects(ctx)
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

	// Retrieve the project from the projects map
	if project, err := store.GetProject(ctx, id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(project)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}
}

func apiGetProjectToolsHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	ctx := r.Context()

	// Check if the project exists
	if _, err := store.GetProject(ctx, id); err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Get the tools for the project
	tools, err := store.GetProjectTools(ctx, id)
	if err != nil {
		http.Error(w, "Failed to get project tools", http.StatusInternalServerError)
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

func apiRegisterProjectToolHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	ctx := r.Context()

	var request struct {
		Tool Tool `json:"tool"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err = store.CreateProjectTool(ctx, id, request.Tool); err != nil {
		http.Error(w, "Failed to register project tool", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func apiAssignSupervisorHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	return
}
