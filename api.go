package sentinel

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func InitAPI() {
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
	http.HandleFunc("/api/review/llm", func(w http.ResponseWriter, r *http.Request) {
		apiReviewLLMHandler(w, r)
	})

	// Start the server, default to port 8080 if APPROVAL_WEBSERVER_PORT is not set
	port := os.Getenv("APPROVAL_WEBSERVER_PORT")
	if port == "" {
		log.Fatal("APPROVAL_WEBSERVER_PORT not set, failing out")
	}

	log.Printf("Server started on APPROVAL_WEBSERVER_PORT=%s", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal("error listening and serving: ", err)
	}
}
