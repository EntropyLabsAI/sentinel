package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

var completedReviews = &sync.Map{}
var reviewChannels = &sync.Map{}

// Timeout duration for waiting for the reviewer to respond
const reviewTimeout = 5 * time.Minute

// serveTemplate renders the index.html template
func serveTemplate(w http.ResponseWriter, _ *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiReviewHandler receives review requests via the HTTP API
func apiReviewHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	var reviewRequest ReviewRequest
	err := json.NewDecoder(r.Body).Decode(&reviewRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate a unique ID for the review request
	reviewRequest.ID = uuid.New().String()

	// Add the review request to the queue
	hub.ReviewChan <- reviewRequest
	log.Printf("Received new review request ID %s via API.", reviewRequest.ID)

	// Create a channel for this review request
	responseChan := make(chan ReviewerResponse)
	reviewChannels.Store(reviewRequest.ID, responseChan)

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
			completedReviews.Store(reviewRequest.ID, map[string]string{
				"status": "timeout",
				"id":     reviewRequest.ID,
			})
			reviewChannels.Delete(reviewRequest.ID)
			log.Printf("Review ID %s timed out.", reviewRequest.ID)
		}
	}()

	// Respond immediately with 200 OK
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{"status": "queued", "id": reviewRequest.ID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// apiReviewStatusHandler checks the status of a review request
func apiReviewStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the review ID from the query parameters
	reviewID := r.URL.Query().Get("id")
	if reviewID == "" {
		http.Error(w, "Missing review ID", http.StatusBadRequest)
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
