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

	hub := NewHub(store, humanReviewChan)
	go hub.Run()

	processor := NewProcessor(store, humanReviewChan)
	go processor.Start(context.Background())

	server := Server{
		Hub:   hub,
		Store: store,
	}

	apiHandler := Handler(server)
	corsHandler := enableCorsMiddleware(apiHandler)

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", corsHandler))
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

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

func (s Server) CreateProject(w http.ResponseWriter, r *http.Request) {
	apiCreateProjectHandler(w, r, s.Store)
}

func (s Server) GetProjects(w http.ResponseWriter, r *http.Request) {
	apiGetProjectsHandler(w, r, s.Store)
}

func (s Server) GetProject(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetProjectHandler(w, r, id, s.Store)
}

func (s Server) CreateTask(w http.ResponseWriter, r *http.Request, projectId uuid.UUID) {
	apiCreateTaskHandler(w, r, projectId, s.Store)
}

func (s Server) CreateRun(w http.ResponseWriter, r *http.Request, taskId uuid.UUID) {
	apiCreateRunHandler(w, r, taskId, s.Store)
}

func (s Server) GetProjectTasks(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetProjectTasksHandler(w, r, id, s.Store)
}

func (s Server) GetRun(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetRunHandler(w, r, id, s.Store)
}

func (s Server) GetTaskRuns(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetTaskRunsHandler(w, r, id, s.Store)
}

func (s Server) GetTask(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetTaskHandler(w, r, id, s.Store)
}

func (s Server) GetRunTools(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetRunToolsHandler(w, r, id, s.Store)
}

func (s Server) CreateRunTool(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiCreateRunToolHandler(w, r, id, s.Store)
}

func (s Server) CreateSupervisor(w http.ResponseWriter, r *http.Request, projectId uuid.UUID) {
	apiCreateSupervisorHandler(w, r, projectId, s.Store)
}

func (s Server) GetSupervisors(w http.ResponseWriter, r *http.Request, projectId uuid.UUID) {
	apiGetSupervisorsHandler(w, r, projectId, s.Store)
}

func (s Server) GetSupervisor(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetSupervisorHandler(w, r, id, s.Store)
}

func (s Server) CreateToolSupervisorChains(w http.ResponseWriter, r *http.Request, toolId uuid.UUID) {
	apiCreateToolSupervisorChainsHandler(w, r, toolId, s.Store)
}

func (s Server) GetToolSupervisorChains(w http.ResponseWriter, r *http.Request, toolId uuid.UUID) {
	apiGetToolSupervisorChainsHandler(w, r, toolId, s.Store)
}

func (s Server) GetToolCall(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetToolCallHandler(w, r, id, s.Store)
}

func (s Server) GetProjectTools(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetProjectToolsHandler(w, r, id, s.Store)
}

func (s Server) GetTool(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	apiGetToolHandler(w, r, id, s.Store)
}

func (s Server) CreateSupervisionRequest(w http.ResponseWriter, r *http.Request, toolCallId uuid.UUID, chainId uuid.UUID, supervisorId uuid.UUID) {
	apiCreateSupervisionRequestHandler(w, r, toolCallId, chainId, supervisorId, s.Store)
}

func (s Server) GetSupervisionRequestStatus(w http.ResponseWriter, r *http.Request, supervisionRequestId uuid.UUID) {
	apiGetSupervisionRequestStatusHandler(w, r, supervisionRequestId, s.Store)
}

func (s Server) CreateSupervisionResult(w http.ResponseWriter, r *http.Request, supervisionRequestId uuid.UUID) {
	apiCreateSupervisionResultHandler(w, r, supervisionRequestId, s.Store)
}

func (s Server) GetSupervisionResult(w http.ResponseWriter, r *http.Request, supervisionRequestId uuid.UUID) {
	apiGetSupervisionResultHandler(w, r, supervisionRequestId, s.Store)
}

func (s Server) GetHubStats(w http.ResponseWriter, r *http.Request) {
	apiGetHubStatsHandler(w, r, s.Hub)
}

func (s Server) GetSupervisionReviewPayload(w http.ResponseWriter, r *http.Request, supervisionRequestId uuid.UUID) {
	apiGetSupervisionReviewPayloadHandler(w, r, supervisionRequestId, s.Store)
}

func (s Server) GetToolCallStatus(w http.ResponseWriter, r *http.Request, toolCallId uuid.UUID) {
	apiGetToolCallStatusHandler(w, r, toolCallId, s.Store)
}

func (s Server) GetToolCallState(w http.ResponseWriter, r *http.Request, toolCallId string) {
	apiGetToolCallStateHandler(w, r, toolCallId, s.Store)
}

func (s Server) GetRunStatus(w http.ResponseWriter, r *http.Request, runId uuid.UUID) {
	apiGetRunStatusHandler(w, r, runId, s.Store)
}

func (s Server) UpdateRunStatus(w http.ResponseWriter, r *http.Request, runId uuid.UUID) {
	apiUpdateRunStatusHandler(w, r, runId, s.Store)
}

func (s Server) UpdateRunResult(w http.ResponseWriter, r *http.Request, runId uuid.UUID) {
	apiUpdateRunResultHandler(w, r, runId, s.Store)
}

func (s Server) CreateNewChat(w http.ResponseWriter, r *http.Request, runId uuid.UUID) {
	apiCreateNewChatHandler(w, r, runId, s.Store)
}

func (s Server) GetRunMessages(w http.ResponseWriter, r *http.Request, runId uuid.UUID) {
	apiGetRunMessagesHandler(w, r, runId, s.Store)
}

func enableCorsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}
