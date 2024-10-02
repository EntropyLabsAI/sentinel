package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Initialize the WebSocket hub
	hub := NewHub()
	go hub.Run()

	// Set up HTTP routes
	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	serveTemplate(w, r)
	// })
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	http.HandleFunc("/api/review", func(w http.ResponseWriter, r *http.Request) {
		apiReviewHandler(hub, w, r)
	})
	http.HandleFunc("/api/review/status", apiReviewStatusHandler)

	// Serve static files
	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start the server
	log.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
