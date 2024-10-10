package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Initialize the WebSocket hub
	hub := NewHub()
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	http.HandleFunc("/api/review", func(w http.ResponseWriter, r *http.Request) {
		apiReviewHandler(hub, w, r)
	})
	http.HandleFunc("/api/review/status", func(w http.ResponseWriter, r *http.Request) {
		apiReviewStatusHandler(w, r)
	})
	http.HandleFunc("/api/explain", func(w http.ResponseWriter, r *http.Request) {
		apiExplainHandler(w, r)
	})
	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		apiHubStatsHandler(hub, w, r)
	})

	// Serve wstatic files
	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start the server
	port := os.Getenv("APPROVAL_WEBSERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server started on APPROVAL_WEBSERVER_PORT=%s", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal("error listening and serving: ", err)
	}
}
