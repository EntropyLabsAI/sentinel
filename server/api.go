package sentinel

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
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

func (s Server) CreateReview(w http.ResponseWriter, r *http.Request) {
	apiReviewHandler(s.Hub, w, r, s.Store)
}

func (s Server) CreateRun(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiCreateRunHandler(w, r, id, s.Store)
}

func (s Server) CreateTool(w http.ResponseWriter, r *http.Request) {
	apiCreateToolHandler(w, r, s.Store)
}

func (s Server) GetProject(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetProjectByIdHandler(w, r, id, s.Store)
}

func (s Server) GetLLMExplanation(w http.ResponseWriter, r *http.Request) {
	apiLLMExplanationHandler(w, r)
}

func (s Server) GetReviewStatus(w http.ResponseWriter, r *http.Request, id string) {
	apiReviewStatusHandler(w, r, id)
}

func (s Server) GetHubStats(w http.ResponseWriter, r *http.Request) {
	apiStatsHandler(s.Hub, w, r)
}

func (s Server) GetLLMReviews(w http.ResponseWriter, r *http.Request) {
	apiGetLLMReviews(w, r)
}

func (s Server) CreateProject(w http.ResponseWriter, r *http.Request) {
	apiRegisterProjectHandler(w, r, s.Store)
}

func (s Server) GetProjectById(w http.ResponseWriter, r *http.Request, id string) {
	apiGetProjectByIdHandler(w, r, id, s.Store)
}

func (s Server) GetProjects(w http.ResponseWriter, r *http.Request) {
	apiGetProjectsHandler(w, r, s.Store)
}

func (s Server) GetProjectTools(w http.ResponseWriter, r *http.Request, id string) {
	apiGetProjectToolsHandler(w, r, id, s.Store)
}

func (s Server) RegisterProjectTool(w http.ResponseWriter, r *http.Request, id string) {
	apiRegisterProjectToolHandler(w, r, id, s.Store)
}

func (s Server) AssignSupervisor(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiAssignSupervisorHandler(w, r, id, s.Store)
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
