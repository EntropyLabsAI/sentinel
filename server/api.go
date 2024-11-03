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

func (s Server) CreateSupervisionRequest(w http.ResponseWriter, r *http.Request) {
	apiCreateSupervisionRequestHandler(w, r, s.Store)
}

func (s Server) GetSupervisionRequest(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetSupervisionRequestHandler(w, r, id, s.Store)
}

func (s Server) GetSupervisionRequests(w http.ResponseWriter, r *http.Request, params GetSupervisionRequestsParams) {
	apiGetSupervisionRequestsHandler(w, r, params, s.Store)
}

func (s Server) CreateSupervisionResult(w http.ResponseWriter, r *http.Request, supervisorRequestId uuid.UUID) {
	apiCreateSupervisionResultHandler(w, r, supervisorRequestId, s.Store)
}

func (s Server) CreateRun(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiCreateRunHandler(w, r, id, s.Store)
}

func (s Server) CreateTool(w http.ResponseWriter, r *http.Request) {
	apiCreateToolHandler(w, r, s.Store)
}

func (s Server) GetProjectRuns(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetProjectRunsHandler(w, r, id, s.Store)
}

func (s Server) GetRuns(w http.ResponseWriter, r *http.Request, projectId uuid.UUID) {
	apiGetRunsHandler(w, r, projectId, s.Store)
}

func (s Server) GetRun(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetRunHandler(w, r, id, s.Store)
}

func (s Server) GetRunTools(w http.ResponseWriter, r *http.Request, runId uuid.UUID) {
	apiGetRunToolsHandler(w, r, runId, s.Store)
}

func (s Server) CreateRunToolSupervisors(w http.ResponseWriter, r *http.Request, toolId uuid.UUID, runId uuid.UUID) {
	apiCreateRunToolSupervisorsHandler(w, r, toolId, runId, s.Store)
}

func (s Server) GetTool(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetToolHandler(w, r, id, s.Store)
}

func (s Server) GetTools(w http.ResponseWriter, r *http.Request, params GetToolsParams) {
	apiGetToolsHandler(w, r, params, s.Store)
}

func (s Server) GetRunToolSupervisors(w http.ResponseWriter, r *http.Request, runId uuid.UUID, toolId uuid.UUID) {
	apiGetRunToolSupervisorsHandler(w, r, runId, toolId, s.Store)
}

func (s Server) GetProject(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetProjectByIdHandler(w, r, id, s.Store)
}

func (s Server) GetSupervisionResults(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetSupervisionResultsHandler(w, r, id, s.Store)
}

func (s Server) GetReviewToolRequests(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetReviewToolRequestsHandler(w, r, id, s.Store)
}

func (s Server) GetSupervisor(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetSupervisorHandler(w, r, id, s.Store)
}

func (s Server) GetSupervisors(w http.ResponseWriter, r *http.Request, params GetSupervisorsParams) {
	apiGetSupervisorsHandler(w, r, params, s.Store)
}

func (s Server) CreateSupervisor(w http.ResponseWriter, r *http.Request) {
	apiCreateSupervisorHandler(w, r, s.Store)
}

func (s Server) GetSupervisionStatus(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiSupervisionStatusHandler(w, r, id, s.Store)
}

func (s Server) GetHubStats(w http.ResponseWriter, r *http.Request) {
	apiStatsHandler(s.Hub, w, r)
}

func (s Server) CreateProject(w http.ResponseWriter, r *http.Request) {
	apiRegisterProjectHandler(w, r, s.Store)
}

func (s Server) GetProjects(w http.ResponseWriter, r *http.Request) {
	apiGetProjectsHandler(w, r, s.Store)
}

func (s Server) CreateExecution(w http.ResponseWriter, r *http.Request, runId uuid.UUID) {
	apiCreateExecutionHandler(w, r, runId, s.Store)
}

func (s Server) GetExecution(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetExecutionHandler(w, r, id, s.Store)
}

func (s Server) GetRunExecutions(w http.ResponseWriter, r *http.Request, runId uuid.UUID) {
	apiGetRunExecutionsHandler(w, r, runId, s.Store)
}

func (s Server) GetExecutionSupervisions(w http.ResponseWriter, r *http.Request, executionId uuid.UUID) {
	apiGetExecutionSupervisionsHandler(w, r, executionId, s.Store)
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

// func (s Server) GetProjectById(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
// 	apiGetProjectByIdHandler(w, r, id, s.Store)
// }

// func (s Server) RegisterProjectTool(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
// 	apiRegisterProjectToolHandler(w, r, id, s.Store)
// }

// func (s Server) GetProjectTools(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
// 	apiGetProjectToolsHandler(w, r, id, s.Store)
// }

// func (s Server) GetProjectTools(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
// 	apiGetProjectToolsHandler(w, r, id, s.Store)
// }

func (s Server) GetLLMExplanation(w http.ResponseWriter, r *http.Request) {
	apiLLMExplanationHandler(w, r)
}
