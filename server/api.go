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
	log.Println("Initializing API v1")

	humanReviewChan := make(chan SupervisionRequest, 100)

	// Initialize the WebSocket hub
	hub := NewHub(store, humanReviewChan)
	go hub.Run()

	// Start the processor which will pick up reviews from the DB and send them to the humanReviewChan
	processor := NewProcessor(store, humanReviewChan)
	go processor.Start(context.Background())

	server := Server{
		Hub:   hub,
		Store: store,
	}

	// Generate the API handler using the generated code
	apiHandler := Handler(server)

	// Wrap the API handler with the CORS middleware
	corsHandler := enableCorsMiddleware(apiHandler)

	mux := http.NewServeMux()

	// Register the wrapped API handler under the /api/v1/ path
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", corsHandler))

	// Register the WebSocket handler separately
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	// Start the server on the port specified
	port := os.Getenv("APPROVAL_WEBSERVER_PORT")
	if port == "" {
		log.Fatal("APPROVAL_WEBSERVER_PORT not set, failing out")
	}

	log.Printf("Server v1 started on port %s", port)
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

// CreateTask
func (s Server) CreateTask(w http.ResponseWriter, r *http.Request, projectId uuid.UUID) {
	apiCreateTaskHandler(w, r, projectId, s.Store)
}

// CreateRun
func (s Server) CreateRun(w http.ResponseWriter, r *http.Request, taskId uuid.UUID) {
	apiCreateRunHandler(w, r, taskId, s.Store)
}

// GetProjectTasks
func (s Server) GetProjectTasks(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetProjectTasksHandler(w, r, id, s.Store)
}

// GetRun
func (s Server) GetRun(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetRunHandler(w, r, id, s.Store)
}

// GetTaskRuns
func (s Server) GetTaskRuns(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetTaskRunsHandler(w, r, id, s.Store)
}

// GetTask
func (s Server) GetTask(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetTaskHandler(w, r, id, s.Store)
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

// // CreateToolRequestGroup
// func (s Server) CreateToolRequestGroup(w http.ResponseWriter, r *http.Request, toolId uuid.UUID) {
// 	apiCreateToolRequestGroupHandler(w, r, toolId, s.Store)
// }

// // GetToolRequest
// func (s Server) GetToolRequest(w http.ResponseWriter, r *http.Request, toolRequestId uuid.UUID) {
// 	apiGetToolRequestHandler(w, r, toolRequestId, s.Store)
// }

// GetToolCall
func (s Server) GetToolCall(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetToolCallHandler(w, r, id, s.Store)
}

// GetRunRequestGroups
// func (s Server) GetRunRequestGroups(w http.ResponseWriter, r *http.Request, runId uuid.UUID) {
// 	apiGetRunRequestGroupsHandler(w, r, runId, s.Store)
// }

// GetRequestGroup
// func (s Server) GetRequestGroup(w http.ResponseWriter, r *http.Request, requestGroupId uuid.UUID) {
// 	apiGetRequestGroupHandler(w, r, requestGroupId, s.Store)
// }

// GetProjectTools
func (s Server) GetProjectTools(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetProjectToolsHandler(w, r, id, s.Store)
}

// GetTool
func (s Server) GetTool(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetToolHandler(w, r, id, s.Store)
}

// CreateSupervisionRequest
func (s Server) CreateSupervisionRequest(w http.ResponseWriter, r *http.Request, toolCallId uuid.UUID, chainId uuid.UUID, supervisorId uuid.UUID) {
	apiCreateSupervisionRequestHandler(w, r, toolCallId, chainId, supervisorId, s.Store)
}

// GetSupervisionRequestStatus
func (s Server) GetSupervisionRequestStatus(w http.ResponseWriter, r *http.Request, supervisionRequestId uuid.UUID) {
	apiGetSupervisionRequestStatusHandler(w, r, supervisionRequestId, s.Store)
}

// CreateSupervisionResult
func (s Server) CreateSupervisionResult(w http.ResponseWriter, r *http.Request, supervisionRequestId uuid.UUID) {
	apiCreateSupervisionResultHandler(w, r, supervisionRequestId, s.Store)
}

// GetSupervisionResult
func (s Server) GetSupervisionResult(w http.ResponseWriter, r *http.Request, supervisionRequestId uuid.UUID) {
	apiGetSupervisionResultHandler(w, r, supervisionRequestId, s.Store)
}

// GetHubStats
func (s Server) GetHubStats(w http.ResponseWriter, r *http.Request) {
	apiGetHubStatsHandler(w, r, s.Hub)
}

// GetRunState
func (s Server) GetRunState(w http.ResponseWriter, r *http.Request, runId uuid.UUID) {
	apiGetRunStateHandler(w, r, runId, s.Store)
}

// CreateToolRequest
// func (s Server) CreateToolRequest(w http.ResponseWriter, r *http.Request, requestGroupId uuid.UUID) {
// 	apiCreateToolRequestHandler(w, r, requestGroupId, s.Store)
// }

// GetSupervisionReviewPayload
func (s Server) GetSupervisionReviewPayload(w http.ResponseWriter, r *http.Request, supervisionRequestId uuid.UUID) {
	apiGetSupervisionReviewPayloadHandler(w, r, supervisionRequestId, s.Store)
}

// GetRequestGroupStatus
// func (s Server) GetRequestGroupStatus(w http.ResponseWriter, r *http.Request, requestGroupId uuid.UUID) {
// 	apiGetRequestGroupStatusHandler(w, r, requestGroupId, s.Store)
// }

// GetToolCallStatus
func (s Server) GetToolCallStatus(w http.ResponseWriter, r *http.Request, toolCallId uuid.UUID) {
	apiGetToolCallStatusHandler(w, r, toolCallId, s.Store)
}

// GetRunStatus
func (s Server) GetRunStatus(w http.ResponseWriter, r *http.Request, runId uuid.UUID) {
	apiGetRunStatusHandler(w, r, runId, s.Store)
}

// UpdateRunStatus
func (s Server) UpdateRunStatus(w http.ResponseWriter, r *http.Request, runId uuid.UUID) {
	apiUpdateRunStatusHandler(w, r, runId, s.Store)
}

// UpdateRunResult
func (s Server) UpdateRunResult(w http.ResponseWriter, r *http.Request, runId uuid.UUID) {
	apiUpdateRunResultHandler(w, r, runId, s.Store)
}

func enableCorsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
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

func (s Server) CreateNewChat(w http.ResponseWriter, r *http.Request, runId uuid.UUID) {
	apiCreateNewChatHandler(w, r, runId, s.Store)
}
