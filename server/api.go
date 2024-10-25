package sentinel

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type Server struct {
	Hub   *Hub
	Store Store
}

func InitAPI(store Store) {
	// Initialize the WebSocket hub
	hub := NewHub(store)
	go hub.Run()

	// Create an instance of your ServerInterface implementation
	server := Server{
		Hub:   hub,
		Store: store,
	}

	// Generate the API handler using the generated code
	apiHandler := Handler(server)

	// Wrap the API handler with the CORS middleware
	corsHandler := enableCorsMiddleware(apiHandler)

	mux := http.NewServeMux()

	// Register the wrapped API handler under the /api/ path
	mux.Handle("/api/", corsHandler)

	// Register the WebSocket handler separately
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	// Start the server on the port specified
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

func (s Server) GetSwaggerDocs(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "swagger-ui/index.html")
}

func (s Server) GetOpenAPI(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "openapi.yaml")
}

// SubmitReview handles the POST /api/review/human endpoint
func (s Server) SubmitReviewHuman(w http.ResponseWriter, r *http.Request) {
	apiReviewHandler(s.Hub, w, r)
}

// GetReviewLLM handles the POST /api/review/llm endpoint
func (s Server) SubmitReviewLLM(w http.ResponseWriter, r *http.Request) {
	apiReviewLLMHandler(w, r)
}

func (s Server) GetLLMExplanation(w http.ResponseWriter, r *http.Request) {
	apiLLMExplanationHandler(w, r)
}

func (s Server) GetReviewStatus(w http.ResponseWriter, r *http.Request, id string) {
	apiReviewStatusHandler(w, r, id)
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

// RegisterProject handles the POST /api/project endpoint
func (s Server) RegisterProject(w http.ResponseWriter, r *http.Request) {
	apiRegisterProjectHandler(w, r)
}

// GetProjectById handles the GET /api/project/{id} endpoint
func (s Server) GetProjectById(w http.ResponseWriter, r *http.Request, id string) {
	apiGetProjectByIdHandler(w, r, id)
}

// GetProjects handles the GET /api/project endpoint
func (s Server) GetProjects(w http.ResponseWriter, r *http.Request) {
	apiGetProjectsHandler(w, r)
}

func enableCorsMiddleware(handler http.Handler) http.Handler {
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
		handler.ServeHTTP(w, r)
	})
}
