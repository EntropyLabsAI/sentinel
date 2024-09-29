package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"sync"
	"time"
)

// Map to store channels for synchronizing review requests and responses
var reviewChannels = &sync.Map{}

// Timeout duration for waiting for the reviewer to respond
const reviewTimeout = 5 * time.Minute

// serveTemplate renders the index.html template
func serveTemplate(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// apiReviewHandler receives review requests via the HTTP API
func apiReviewHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	var reviewRequest ReviewRequest
	err := json.NewDecoder(r.Body).Decode(&reviewRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create a channel to wait for the reviewer's response
	responseChan := make(chan ReviewerResponse)
	// Store the channel in the map with the request ID as the key
	reviewChannels.Store(reviewRequest.ID, responseChan)
	defer reviewChannels.Delete(reviewRequest.ID) // Clean up after we're done

	// Send the review request to the frontend via WebSocket
	hub.Broadcast <- reviewRequest

	// Wait for the reviewer's response or timeout
	select {
	case reviewerResponse := <-responseChan:
		// Send the response back to the HTTP client
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(reviewerResponse)
	case <-time.After(reviewTimeout):
		// Timeout reached, send an error response
		http.Error(w, "Timeout waiting for reviewer response", http.StatusGatewayTimeout)
	}
}
