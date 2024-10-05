package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
		Send: make(chan ReviewRequest, 3), // Buffer size of 3 reviews
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

	fmt.Printf("Received new review request: %v", request)

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
		fmt.Printf("Review ID %s is pending", reviewID)
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(map[string]string{"status": "pending"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Check if there's a stored response for this review
		if response, ok := completedReviews.Load(reviewID); ok {
			fmt.Printf("Review ID %s has a stored response: %v", reviewID, response)
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

	explanation, err := getExplanationFromLLM(request.Text)
	if err != nil {
		http.Error(w, "Failed to get explanation from LLM", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"explanation": explanation})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getExplanationFromLLM(text string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(apiKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistant that explains technical concepts.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Please explain the following text: " + text,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

// Add this new helper function
// TODO: do this in a more secure way
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
