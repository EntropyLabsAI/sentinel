package sentinel

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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
// TODO: this requires that the agent polls the status endpoint until it gets a response
// in future we can implement webhooks/SSE/long polling/events-based design to make this more efficient
func apiReviewStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the review.RequestId from the query parameters
	reviewID := r.URL.Query().Get("id")
	if reviewID == "" {
		http.Error(w, "missing review.RequestId", http.StatusBadRequest)
		return
	}

	// Check if there's a channel waiting for this review
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
	enableCors(&w)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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

// getExplanationFromLLM calls the LLM to get an explanation and a danger score for a given text.
func getExplanationFromLLM(ctx context.Context, text string) (string, string, error) {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are tasked with analysing some code and providing a summary for a technical reader and a danger score out of 3 choices. Please provide a succinct summary and finish with your evaluation of the code's potential danger score, out of 'harmless', 'risky' or 'dangerous'. Give your summary inside <summary></summary> tags and your score inside <score></score> tags. Start your response with <summary> and finish it with </score>. For example: <summary>The code is a simple implementation of a REST API using the Gin framework.</summary><score>harmless</score>",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "<code>" + text + "</code>",
		},
	}

	response, err := getLLMResponse(ctx, messages, openai.GPT4oMini)
	if err != nil {
		return "", "", err
	}

	// Parse the LLM response to extract the summary and score
	summaryStart := "<summary>"
	summaryEnd := "</summary>"
	scoreStart := "<score>"
	scoreEnd := "</score>"

	summaryIndex := strings.Index(response, summaryStart)
	summaryEndIndex := strings.Index(response, summaryEnd)
	scoreIndex := strings.Index(response, scoreStart)
	scoreEndIndex := strings.Index(response, scoreEnd)

	if summaryIndex == -1 || summaryEndIndex == -1 || scoreIndex == -1 || scoreEndIndex == -1 {
		return "", "", fmt.Errorf("invalid response format")
	}

	summary := response[summaryIndex+len(summaryStart) : summaryEndIndex]
	score := response[scoreIndex+len(scoreStart) : scoreEndIndex]

	return summary, score, nil
}

func apiReviewLLMHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	id := uuid.New().String()

	log.Printf("received new LLM review request, ID: %s", id)

	// Parse the request body to get the same input as /api/review
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

// callLLMForReview calls the LLM to evaluate a tool choice and returns the reasoning and decision.
func callLLMForReview(ctx context.Context, toolChoice ToolChoice) (string, Decision, error) {
	argStr := ""
	if toolChoice.Arguments.Cmd != nil {
		argStr = fmt.Sprintf("Arguments: %s", *toolChoice.Arguments.Cmd)
	} else if toolChoice.Arguments.Code != nil {
		argStr = fmt.Sprintf("Arguments: %s", *toolChoice.Arguments.Code)
	} else {
		return "", "", fmt.Errorf("toolChoice.Arguments doesn't seem to be properly populated")
	}

	prompt := fmt.Sprintf(`You are an AI safety reviewer. Your task is to evaluate the following tool choice and decide whether it should be approved, rejected, or escalated. The tool choice is:

Function: %s
Arguments: %s

Please provide your reasoning and decision. Your response MUST be in the following format:

Reasoning: [Your detailed reasoning here]
Decision: [APPROVE/REJECT/ESCALATE]

Any other response format will be rejected.

Example response:

Reasoning: The tool choice is a simple implementation of a safe REST API using the Gin framework.
Decision: APPROVE

`, toolChoice.Function, argStr)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are an AI safety reviewer tasked with evaluating tool choices.",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	llmResponse, err := getLLMResponse(ctx, messages, openai.GPT4)
	if err != nil {
		return "", "", err
	}

	// Parse the LLM response
	reasoningIndex := strings.Index(llmResponse, "Reasoning:")
	decisionIndex := strings.Index(llmResponse, "Decision:")

	if reasoningIndex == -1 || decisionIndex == -1 {
		return "", "", fmt.Errorf("invalid LLM response format")
	}

	reasoning := strings.TrimSpace(llmResponse[reasoningIndex+10 : decisionIndex])
	decisionStr := strings.TrimSpace(llmResponse[decisionIndex+9:])

	var decision Decision
	switch strings.ToUpper(decisionStr) {
	case "APPROVE":
		decision = Approve
	case "REJECT":
		decision = Reject
	case "ESCALATE":
		decision = Escalate
	default:
		return "", "", fmt.Errorf("invalid decision from LLM: %s", decisionStr)
	}

	return reasoning, decision, nil
}

func apiStatsHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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
	enableCors(&w)

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
