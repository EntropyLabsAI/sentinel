package sentinel

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
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
	var request RegisterProjectRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate a new Project ID
	id := uuid.New().String()

	// Create the Project struct
	project := Project{
		Id:   id,
		Name: request.Name,
	}

	// Store the project in the global projects map
	err = store.CreateProject(ctx, project)
	if err != nil {
		http.Error(w, "Failed to register project", http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := map[string]string{
		"id": id,
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiReviewHandler receives review requests via the HTTP API
func apiReviewHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	var request ReviewRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert the request to a Review by adding an ID
	id := uuid.New().String()

	review := Review{
		Id:      id,
		Request: request,
	}

	// Add the review request to the queue
	hub.ReviewChan <- review

	log.Printf("received new review request ID %s via API.", review.Id)

	// Create a channel for this review request
	responseChan := make(chan ReviewResult)
	reviewChannels.Store(id, responseChan)

	// Start a goroutine to wait for the response
	go func() {
		select {
		case response := <-responseChan:
			// Store the completed review
			completedHumanReviews.Store(response.Id, response)
			reviewChannels.Delete(response.Id)
			log.Printf("review ID %s completed with decision: %s.", response.Id, response.Decision)
		case <-time.After(reviewTimeout):

			reviewStatus := ReviewStatusResponse{
				Status: Timeout,
				Id:     review.Id,
			}

			// Timeout occurred
			completedHumanReviews.Store(review.Id, reviewStatus)
			reviewChannels.Delete(review.Id)
			log.Printf("review ID %s timed out.", review.Id)
		}
	}()

	response := ReviewStatusResponse{
		Id:     id,
		Status: Queued,
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

func apiReviewLLMHandler(w http.ResponseWriter, r *http.Request) {
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

	// Call the LLM to evaluate the tool_choice
	llmReasoning, decision, err := callLLMForReview(r.Context(), reviewRequest.ToolChoices[0])
	if err != nil {
		http.Error(w, fmt.Sprintf("Error calling LLM: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := ReviewResult{
		Id:         id,
		Decision:   decision,
		ToolChoice: reviewRequest.ToolChoices[0],
		Reasoning:  llmReasoning,
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

func apiSetLLMPromptHandler(w http.ResponseWriter, r *http.Request) {
	var request LLMPrompt

	// Decode the request body
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "api: invalid request body", http.StatusBadRequest)
		return
	}

	// Sanitize the prompt.
	sanitizedPrompt := sanitizePrompt(request.Prompt)
	if sanitizedPrompt == "" {
		http.Error(w, "api: invalid prompt", http.StatusBadRequest)
		return
	}

	// Update the global prompt
	llmReviewPrompt = sanitizedPrompt

	response := LLMPromptResponse{
		Status:  "success",
		Message: "LLM prompt updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Sanitize the prompt to prevent XSS
func sanitizePrompt(prompt string) string {
	// Remove any HTML tags
	cleanPrompt := stripHTMLTags(prompt)
	// Trim whitespace
	cleanPrompt = strings.TrimSpace(cleanPrompt)
	return cleanPrompt
}

func stripHTMLTags(input string) string {
	// Simple regex to remove HTML tags
	re := regexp.MustCompile(`<.*?>`)
	return re.ReplaceAllString(input, "")
}

// Add this to your handlers.go file

func apiGetLLMPromptHandler(w http.ResponseWriter, _ *http.Request) {
	response := struct {
		Prompt string `json:"prompt"`
	}{
		Prompt: llmReviewPrompt,
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiGetProjectByIdHandler handles the GET /api/project/{id} endpoint
func apiGetProjectByIdHandler(w http.ResponseWriter, r *http.Request, id string, store ProjectStore) {
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

func apiGetProjectToolsHandler(w http.ResponseWriter, r *http.Request, id string, store Store) {
	ctx := r.Context()

	// Check if the project exists
	if _, err := store.GetProject(context.Background(), id); err != nil {
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

func apiRegisterProjectToolHandler(w http.ResponseWriter, r *http.Request, id string, store Store) {
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
