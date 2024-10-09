package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"os"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

var completedReviews = &sync.Map{}
var reviewChannels = &sync.Map{}

// Timeout duration for waiting for the reviewer to respond
const reviewTimeout = 5 * time.Minute

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := &Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan ReviewRequest),
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

	fmt.Printf("Received new review request ID %s via API.\n", request.RequestID)
	fmt.Printf("Request tool choice: %+v\n", request.ToolChoices)
	fmt.Printf("Request tool message: %+v\n", request.LastMessages)

	// Generate a unique ID for the review request
	request.RequestID = uuid.New().String()

	// Add the review request to the queue
	hub.ReviewChan <- request
	log.Printf("Received new review request ID %s via API.", request.RequestID)

	// Create a channel for this review request
	responseChan := make(chan ReviewerResponse)
	reviewChannels.Store(request.RequestID, responseChan)

	// Start a goroutine to wait for the response
	go func() {
		select {
		case response := <-responseChan:
			// Store the completed review
			completedReviews.Store(response.ID, response)
			reviewChannels.Delete(response.ID)
			log.Printf("Review ID %s completed with decision: %s.", response.ID, response.Decision)
		case <-time.After(reviewTimeout):
			// Timeout occurred
			completedReviews.Store(request.RequestID, map[string]string{
				"status": "timeout",
				"id":     request.RequestID,
			})
			reviewChannels.Delete(request.RequestID)
			log.Printf("Review ID %s timed out.", request.RequestID)
		}
	}()

	// Respond immediately with 200 OK
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{"status": "queued", "id": request.RequestID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiReviewStatusHandler checks the status of a review request
func apiReviewStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the review.RequestID from the query parameters
	reviewID := r.URL.Query().Get("id")
	if reviewID == "" {
		http.Error(w, "Missing review.RequestID", http.StatusBadRequest)
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
		if response, ok := completedReviews.Load(reviewID); ok {
			fmt.Printf("Status request for review ID %s: completed\n", reviewID)
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

func apiExplainHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	enableCors(&w)

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

	explanation, score, err := getExplanationFromLLM(request.Text)
	if err != nil {
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

func getExplanationFromLLM(text string) (string, string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", "", fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(apiKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are tasked with analysing some code and providing a summary for a technical reader and a danger score out of 3 choices. Please provide a succint summary and finish with your evaluation of the code's potential danger score, out of 'harmless', 'risky' or 'dangerous'. Give your summary inside <summary></summary> tags and your score inside <score></score> tags. Start your response with <summary> and finish it with </score>. For example: <summary>The code is a simple implementation of a REST API using the Gin framework.</summary><score>harmless</score>",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "<code>" + text + "</code>",
				},
			},
		},
	)

	if err != nil {
		return "", "", fmt.Errorf("error creating chat completion: %v", err)
	}

	response := resp.Choices[0].Message.Content
	summaryStart := "<summary>"
	summaryEnd := "</summary>"
	scoreStart := "<score>"
	scoreEnd := "</score>"

	summaryIndex := strings.Index(response, summaryStart)
	summaryEndIndex := strings.Index(response, summaryEnd)
	scoreIndex := strings.Index(response, scoreStart)
	scoreEndIndex := strings.Index(response, scoreEnd)

	summary := response[summaryIndex+len(summaryStart) : summaryEndIndex]
	score := response[scoreIndex+len(scoreStart) : scoreEndIndex]

	return summary, score, nil
}

// Add this new helper function
// TODO: do this in a more secure way
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func apiHubStatsHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	enableCors(&w)

	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	stats := hub.getStats()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(stats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
