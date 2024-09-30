package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Map to store channels for synchronizing review requests and responses
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
	reviewRequest.ID = generateUniqueID()

	// Add the review request to the queue
	hub.Review <- reviewRequest

	// Respond immediately with 200 OK
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(map[string]string{"status": "queued", "id": reviewRequest.ID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func generateUniqueID() string {
	// Implement a method to generate a unique ID, e.g., using UUID
	return "unique-id-" + strconv.FormatInt(time.Now().UnixNano(), 10)
}
