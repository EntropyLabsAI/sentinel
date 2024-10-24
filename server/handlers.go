package sentinel

import (
	"context"
	"database/sql"
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

// reviewChannels maps a reviews ID to the channel configured to receive the reviewer's response
var reviewChannels = &sync.Map{}

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
func apiRegisterProjectHandler(w http.ResponseWriter, r *http.Request) {
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
		Id:    id,
		Name:  request.Name,
		Tools: request.Tools,
	}

	db, err := newDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Store the project in the database
	if err := db.RegisterProject(&project); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := RegisterProjectResponse{
		Id: id,
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiHumanReviewHandler handles human review requests
func apiHumanReviewHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	fmt.Println("received new human review request")
	var request ReviewRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmt.Printf("request: %+v\n", request)

	// Generate review ID
	id := uuid.New().String()
	review := Review{
		Id:      id,
		Request: request,
	}

	fmt.Printf("review: %+v\n", review)

	// Store the review in the database
	db, err := newDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	fmt.Printf("storing review in database\n")

	if err := db.StoreReview(&review); err != nil {
		http.Error(w, fmt.Sprintf("failed to store review: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Printf("adding review to hub queue\n")

	// Add the review request to the hub queue
	hub.ReviewChan <- review

	fmt.Printf("review added to hub queue\n")

	log.Printf("received new review request ID %s via API.", review.Id)

	response := ReviewStatusResponse{
		Id:     id,
		Status: Queued,
	}

	fmt.Printf("sending response\n")

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiReviewStatusHandler checks the status of a review request
func apiReviewStatusHandler(w http.ResponseWriter, _ *http.Request, reviewID string) {
	db, err := newDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	status, err := db.GetReviewStatus(reviewID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			err := json.NewEncoder(w).Encode(map[string]string{"status": "not_found"})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

// apiReviewLLMHandler handles LLM review requests
func apiReviewLLMHandler(w http.ResponseWriter, r *http.Request) {
	id := uuid.New().String()
	log.Printf("received new LLM review request, ID: %s", id)

	var reviewRequest ReviewRequest
	err := json.NewDecoder(r.Body).Decode(&reviewRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(reviewRequest.ToolChoices) != 1 {
		http.Error(w, "Invalid number of tool choices provided for LLM review", http.StatusBadRequest)
		return
	}

	llmReasoning, decision, err := callLLMForReview(r.Context(), reviewRequest.ToolChoices[0])
	if err != nil {
		http.Error(w, fmt.Sprintf("Error calling LLM: %v", err), http.StatusInternalServerError)
		return
	}

	response := ReviewResult{
		Id:         id,
		Decision:   decision,
		ToolChoice: reviewRequest.ToolChoices[0],
		Reasoning:  llmReasoning,
	}

	// Store the review result in the database
	db, err := newDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	if err := db.StoreReviewResult(&response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
	db, err := newDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	reviews, err := db.GetLLMReviews()
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

// apiGetProjectsHandler returns all projects
func apiGetProjectsHandler(w http.ResponseWriter, _ *http.Request) {
	db, err := newDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	projectList, err := db.GetProjects()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p := make([]Project, 0)
	for _, project := range projectList {
		p = append(p, *project)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(p)
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
func apiGetProjectByIdHandler(w http.ResponseWriter, _ *http.Request, id string) {
	db, err := newDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	project, err := db.GetProject(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(project)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiRegisterAgentHandler handles the POST /api/project/{id}/agent endpoint
func apiRegisterAgentHandler(w http.ResponseWriter, r *http.Request, projectId string) {
	var agent Agent
	err := json.NewDecoder(r.Body).Decode(&agent)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate agent ID if not provided
	if agent.Id == "" {
		agent.Id = uuid.New().String()
	}

	db, err := newDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	if err := db.RegisterAgent(projectId, &agent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(agent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiGetAgentsHandler handles the GET /api/project/{id}/agent endpoint
func apiGetAgentsHandler(w http.ResponseWriter, _ *http.Request, projectId string) {
	db, err := newDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	agents, err := db.GetAgents(projectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(agents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
