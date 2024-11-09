package sentinel

import (
	"context"
	"encoding/json"
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

// sendErrorResponse writes an error response with specific status code and message
func sendErrorResponse(w http.ResponseWriter, status int, message string, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := ErrorResponse{
		Error:   message,
		Details: &details,
	}
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding error response: %v", err)
	}
}

func InitAPI(store Store) {
	humanReviewChan := make(chan SupervisionRequest, 100)

	// Initialize the WebSocket hub
	hub := NewHub(store, humanReviewChan)
	go hub.Run()

	// Start the processor which will pick up reviews from the DB and send them to the humanReviewChan
	processor := NewProcessor(store, humanReviewChan)
	go processor.Start(context.Background())

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
	// err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
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

// CreateProject
func (s Server) CreateProject(w http.ResponseWriter, r *http.Request) {
	apiCreateProjectHandler(w, r, s.Store)
}

// GetProjects
func (s Server) GetProjects(w http.ResponseWriter, r *http.Request) {
	apiGetProjectsHandler(w, r, s.Store)
}

// GetProjectById
func (s Server) GetProject(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetProjectHandler(w, r, id, s.Store)
}

// CreateProjectRun
func (s Server) CreateProjectRun(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiCreateProjectRunHandler(w, r, id, s.Store)
}

// GetProjectRuns
func (s Server) GetProjectRuns(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetProjectRunsHandler(w, r, id, s.Store)
}

// GetRunTools
func (s Server) GetRunTools(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetRunToolsHandler(w, r, id, s.Store)
}

// CreateRunTool
func (s Server) CreateRunTool(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiCreateRunToolHandler(w, r, id, s.Store)
}

// CreateSupervisor
func (s Server) CreateSupervisor(w http.ResponseWriter, r *http.Request, projectId uuid.UUID) {
	apiCreateSupervisorHandler(w, r, projectId, s.Store)
}

// GetSupervisors
func (s Server) GetSupervisors(w http.ResponseWriter, r *http.Request, projectId uuid.UUID) {
	apiGetSupervisorsHandler(w, r, projectId, s.Store)
}

func (s Server) GetSupervisor(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetSupervisorHandler(w, r, id, s.Store)
}

// CreateToolSupervisorChains
func (s Server) CreateToolSupervisorChains(w http.ResponseWriter, r *http.Request, toolId uuid.UUID) {
	apiCreateToolSupervisorChainsHandler(w, r, toolId, s.Store)
}

// GetToolSupervisorChains
func (s Server) GetToolSupervisorChains(w http.ResponseWriter, r *http.Request, toolId uuid.UUID) {
	apiGetToolSupervisorChainsHandler(w, r, toolId, s.Store)
}

// CreateToolRequestGroup
func (s Server) CreateToolRequestGroup(w http.ResponseWriter, r *http.Request, toolId uuid.UUID) {
	apiCreateToolRequestGroupHandler(w, r, toolId, s.Store)
}

// GetRunRequestGroups
func (s Server) GetRunRequestGroups(w http.ResponseWriter, r *http.Request, runId uuid.UUID) {
	apiGetRunRequestGroupsHandler(w, r, runId, s.Store)
}

// GetProjectTools
func (s Server) GetProjectTools(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetToolsHandler(w, r, id, s.Store)
}

// GetTool
func (s Server) GetTool(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetToolHandler(w, r, id, s.Store)
}

// CreateSupervisionRequest
func (s Server) CreateSupervisionRequest(w http.ResponseWriter, r *http.Request, requestGroupId uuid.UUID, chainId uuid.UUID, supervisorId uuid.UUID) {
	apiCreateSupervisionRequestHandler(w, r, requestGroupId, chainId, supervisorId, s.Store)
}

// CreateSupervisionResult
func (s Server) CreateSupervisionResult(w http.ResponseWriter, r *http.Request, supervisionRequestId uuid.UUID) {
	apiCreateSupervisionResultHandler(w, r, supervisionRequestId, s.Store)
}

// GetHubStats
func (s Server) GetHubStats(w http.ResponseWriter, r *http.Request) {
	apiGetHubStatsHandler(w, r, s.Hub)
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
