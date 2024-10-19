package sentinel

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type Server struct {
	Hub *Hub
}

func InitAPI() {
	// Initialize the WebSocket hub
	hub := NewHub()
	go hub.Run()

	// Create an instance of your ServerInterface implementation
	server := Server{
		Hub: hub,
	}

	// Generate the API handler using the generated code
	apiHandler := Handler(server)

	// Wrap the API handler with the CORS middleware
	corsHandler := enableCorsMiddleware(apiHandler)

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register the wrapped API handler under the /api/ path
	mux.Handle("/api/", corsHandler)

	mux.HandleFunc("/api/docs", serveSwaggerUI)

	mux.HandleFunc("/api/openapi.yaml", serveOpenAPI)

	// Register the WebSocket handler separately
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	// Start the server, default to port 8080 if APPROVAL_WEBSERVER_PORT is not set
	port := os.Getenv("APPROVAL_WEBSERVER_PORT")
	if port == "" {
		log.Fatal("APPROVAL_WEBSERVER_PORT not set, failing out")
	}

	log.Printf("Server started on port %s", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
	if err != nil {
		log.Fatal("Error listening and serving: ", err)
	}
}

// SubmitReview handles the POST /api/review/human endpoint
func (s Server) SubmitReview(w http.ResponseWriter, r *http.Request) {
	apiReviewHandler(s.Hub, w, r)
}

// GetReviewLLM handles the POST /api/review/llm endpoint
func (s Server) SubmitReviewLLM(w http.ResponseWriter, r *http.Request) {
	apiReviewLLMHandler(w, r)
}

func (s Server) GetLLMExplanation(w http.ResponseWriter, r *http.Request) {
	apiLLMExplanationHandler(w, r)
}

// GetReviewResult handles the GET /api/review/status endpoint
func (s Server) GetReviewResult(w http.ResponseWriter, r *http.Request, params GetReviewResultParams) {
	apiReviewStatusHandler(w, r)
}

// GetHubStats handles the GET /api/stats endpoint
func (s Server) GetHubStats(w http.ResponseWriter, r *http.Request) {
	apiStatsHandler(s.Hub, w, r)
}

// GetReviewLLMResult handles the GET /api/review/llm/list endpoint
func (s Server) GetLLMReviews(w http.ResponseWriter, r *http.Request) {
	apiGetLLMReviews(w, r)
}

// SetLLMPrompt handles the POST /api/llm/prompt endpoint
func (s Server) SetLLMPrompt(w http.ResponseWriter, r *http.Request) {
	apiSetLLMPromptHandler(w, r)
}

// GetLLMPrompt handles the GET /api/review/llm/prompt endpoint
func (s Server) GetLLMPrompt(w http.ResponseWriter, r *http.Request) {
	apiGetLLMPromptHandler(w, r)
}

func serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Serving Swagger UI\n")
	http.ServeFile(w, r, "swagger-ui/index.html")
}

func serveOpenAPI(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "openapi.yaml")
}

func enableCorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Handle OPTIONS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
