package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	sentinel "github.com/entropylabsai/sentinel/server"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type PostgresqlStore struct {
	db *sql.DB
}

// Check if PostgresqlStore implements sentinel.Store
var _ sentinel.Store = &PostgresqlStore{}

// NewPostgresqlStore creates a new PostgreSQL store
func NewPostgresqlStore(connStr string) (*PostgresqlStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return &PostgresqlStore{db: db}, nil
}

// Close closes the database connection
func (s *PostgresqlStore) Close() error {
	return s.db.Close()
}

// ProjectStore implementation
func (s *PostgresqlStore) CreateProject(ctx context.Context, project sentinel.Project) error {
	query := `
		INSERT INTO project (id, name, created_at)
		VALUES ($1, $2, $3)`

	fmt.Printf("Creating project: %+v\n", project)

	_, err := s.db.ExecContext(ctx, query, project.Id, project.Name, project.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creating project: %w", err)
	}

	return nil
}

func (s *PostgresqlStore) GetProject(ctx context.Context, id uuid.UUID) (*sentinel.Project, error) {
	query := `
		SELECT id, name, created_at
		FROM project
		WHERE id = $1`

	var project sentinel.Project
	err := s.db.QueryRowContext(ctx, query, id).Scan(&project.Id, &project.Name, &project.CreatedAt)
	if err == sql.ErrNoRows {
		log.Printf("no rows found for project ID: %s\n", id)
		return nil, fmt.Errorf("no rows found for project ID: %s", id)
	}
	if err != nil {
		log.Printf("error getting project: %v\n", err)
		return nil, fmt.Errorf("error getting project: %w", err)
	}

	log.Printf("found project: %s\n", project.Name) // Add this line to see what we got
	return &project, nil
}

func (s *PostgresqlStore) CreateExecution(ctx context.Context, runId uuid.UUID, toolId uuid.UUID) (uuid.UUID, error) {

	// First check if both the run and tool exist
	run, err := s.GetRun(ctx, runId)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error getting run: %w", err)
	}
	if run == nil {
		return uuid.UUID{}, fmt.Errorf("run not found: %s", runId)
	}

	tool, err := s.GetTool(ctx, toolId)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error getting tool: %w", err)
	}
	if tool == nil {
		return uuid.UUID{}, fmt.Errorf("tool not found: %s", toolId)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	id := uuid.New()
	status := sentinel.Pending
	query := `
		INSERT INTO execution (id, run_id, tool_id, created_at)
		VALUES ($1, $2, $3, $4)`

	_, err = tx.ExecContext(ctx, query, id, runId, toolId, time.Now())
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating execution: %w", err)
	}

	query = `
		INSERT INTO execution_status (execution_id, status, created_at)
		VALUES ($1, $2, $3)`
	_, err = tx.ExecContext(ctx, query, id, status, time.Now())
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating execution status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return uuid.UUID{}, fmt.Errorf("error committing transaction: %w", err)
	}

	return id, nil
}

func (s *PostgresqlStore) GetExecution(ctx context.Context, id uuid.UUID) (*sentinel.Execution, error) {
	query := `
		SELECT id, run_id, tool_id, created_at
		FROM execution
		WHERE id = $1`

	var execution sentinel.Execution
	err := s.db.QueryRowContext(ctx, query, id).Scan(&execution.Id, &execution.RunId, &execution.ToolId, &execution.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("error getting execution: %w", err)
	}

	return &execution, nil
}

func (s *PostgresqlStore) GetRunExecutions(ctx context.Context, runId uuid.UUID) ([]sentinel.Execution, error) {
	query := `
		SELECT e.id, e.run_id, e.tool_id, e.created_at
		FROM execution e
		WHERE run_id = $1`

	rows, err := s.db.QueryContext(ctx, query, runId)
	if err != nil {
		return nil, fmt.Errorf("error getting run executions: %w", err)
	}
	defer rows.Close()

	var executions []sentinel.Execution
	for rows.Next() {
		var execution sentinel.Execution
		if err := rows.Scan(&execution.Id, &execution.RunId, &execution.ToolId, &execution.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning execution: %w", err)
		}
		executions = append(executions, execution)
	}

	// For each execution, get the status. There will be a list of statuses, we want the latest one
	for i, execution := range executions {
		status, err := s.GetExecutionStatus(ctx, execution.Id)
		if err != nil {
			return nil, fmt.Errorf("error getting execution status: %w", err)
		}
		executions[i].Status = &status
	}

	return executions, nil
}

func (s *PostgresqlStore) GetExecutionStatus(ctx context.Context, id uuid.UUID) (sentinel.Status, error) {
	query := `
		SELECT status
		FROM execution_status
		WHERE execution_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	var status sentinel.Status
	err := s.db.QueryRowContext(ctx, query, id).Scan(&status)
	if err != nil {
		return sentinel.Status(""), fmt.Errorf("error getting execution status: %w", err)
	}

	return status, nil
}

func (s *PostgresqlStore) GetProjects(ctx context.Context) ([]sentinel.Project, error) {
	query := `
		SELECT id, name, created_at
		FROM project
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listing projects: %w", err)
	}
	defer rows.Close()

	var projects []sentinel.Project
	for rows.Next() {
		var project sentinel.Project
		if err := rows.Scan(&project.Id, &project.Name, &project.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning project: %w", err)
		}
		projects = append(projects, project)
	}

	// If no rows were found, projects will be empty slice
	return projects, nil
}

func (s *PostgresqlStore) GetRuns(ctx context.Context, projectId uuid.UUID) ([]sentinel.Run, error) {
	query := `
		SELECT id, project_id, created_at
		FROM run
		WHERE project_id = $1`

	rows, err := s.db.QueryContext(ctx, query, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting runs: %w", err)
	}
	defer rows.Close()

	var runs []sentinel.Run
	for rows.Next() {
		var run sentinel.Run
		if err := rows.Scan(&run.Id, &run.ProjectId, &run.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning run: %w", err)
		}
		runs = append(runs, run)
	}

	// If no rows were found, runs will be empty slice
	return runs, nil
}

func (s *PostgresqlStore) GetProjectRuns(ctx context.Context, id uuid.UUID) ([]sentinel.Run, error) {
	query := `
		SELECT id, project_id, created_at
		FROM run
		WHERE project_id = $1`

	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("error getting project runs: %w", err)
	}
	defer rows.Close()

	var runs []sentinel.Run
	for rows.Next() {
		var run sentinel.Run
		if err := rows.Scan(&run.Id, &run.ProjectId, &run.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning run: %w", err)
		}
		runs = append(runs, run)
	}

	// If no rows were found, runs will be empty slice
	return runs, nil
}

func (s *PostgresqlStore) CreateSupervisionRequest(ctx context.Context, request sentinel.SupervisionRequest) (uuid.UUID, error) {
	if request.SupervisorId == nil {
		return uuid.UUID{}, fmt.Errorf("can't create supervision request without a supervisor ID")
	}

	execution, err := s.GetExecution(ctx, request.ExecutionId)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error getting execution: %w", err)
	}
	if execution == nil {
		return uuid.UUID{}, fmt.Errorf("execution not found: %s", request.ExecutionId)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Store the llm_messages first and keep track of the IDs
	var messageIDs []uuid.UUID
	for _, message := range request.Messages {
		messageID := uuid.New()
		query := `
			INSERT INTO llm_message (id, role, content)
			VALUES ($1, $2, $3)`

		_, err = tx.ExecContext(ctx, query, messageID, message.Role, message.Content)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("error creating llm message: %w", err)
		}
		messageIDs = append(messageIDs, messageID)
	}

	// Store the supervisor request
	taskStateJSON, err := json.Marshal(request.TaskState)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error marshalling task state: %w", err)
	}

	query := `
		INSERT INTO supervisionrequest (id, execution_id, task_state, supervisor_id)
		VALUES ($1, $2, $3, $4)`

	requestID := uuid.New()
	_, err = tx.ExecContext(
		ctx, query, requestID, request.ExecutionId, taskStateJSON, request.SupervisorId,
	)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating supervision request: %w", err)
	}

	// Store the tool requests
	for i, toolRequest := range request.ToolRequests {
		toolRequestID := uuid.New()

		// Marshal the Arguments map to JSON
		argumentsJSON, err := json.Marshal(toolRequest.Arguments)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("error marshalling tool request arguments: %w", err)
		}

		query := `
			INSERT INTO toolrequest (id, supervisionrequest_id, tool_id, message_id, arguments)
			VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.ExecContext(
			ctx, query, toolRequestID, requestID, toolRequest.ToolId, messageIDs[i], argumentsJSON,
		)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("error creating tool request: %w", err)
		}
	}

	status := sentinel.SupervisionStatus{Status: sentinel.Pending, CreatedAt: time.Now()}

	// Store a supervisor status pending
	err = s.createSupervisionStatus(ctx, requestID, status, tx)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating supervisor status: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error committing transaction: %w", err)
	}

	return requestID, nil
}

func (s *PostgresqlStore) CreateSupervisionStatus(ctx context.Context, requestID uuid.UUID, status sentinel.SupervisionStatus) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	err = s.createSupervisionStatus(ctx, requestID, status, tx)
	if err != nil {
		return fmt.Errorf("error creating supervisor status: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (s *PostgresqlStore) createSupervisionStatus(ctx context.Context, requestID uuid.UUID, status sentinel.SupervisionStatus, tx *sql.Tx) error {
	query := `
		INSERT INTO supervisionrequest_status (supervisionrequest_id, status, created_at)
		VALUES ($1, $2, $3)`

	_, err := tx.ExecContext(ctx, query, requestID, status.Status, status.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creating supervisor status: %w", err)
	}

	return nil
}

func (s *PostgresqlStore) GetSupervisionRequest(ctx context.Context, id uuid.UUID) (*sentinel.SupervisionRequest, error) {
	query := `
		SELECT sr.id, sr.execution_id, e.run_id, sr.supervisor_id, sr.task_state, ss.id, ss.status, ss.created_at
		FROM supervisionrequest sr
		INNER JOIN execution e ON sr.execution_id = e.id
		INNER JOIN supervisionrequest_status ss ON sr.id = ss.supervisionrequest_id
		WHERE sr.id = $1
		ORDER BY ss.created_at DESC
		LIMIT 1`

	var supervisorRequest sentinel.SupervisionRequest
	var status sentinel.SupervisionStatus
	var taskStateJSON []byte // Add temporary variable for JSON data
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&supervisorRequest.Id,
		&supervisorRequest.ExecutionId,
		&supervisorRequest.RunId,
		&supervisorRequest.SupervisorId,
		&taskStateJSON,
		&status.Id,
		&status.Status,
		&status.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting supervisor: %w", err)
	}

	// Parse the JSON task state
	if err := json.Unmarshal(taskStateJSON, &supervisorRequest.TaskState); err != nil {
		return nil, fmt.Errorf("error parsing task state: %w", err)
	}

	supervisorRequest.Status = &status

	// Get the tool requests
	toolRequests, err := s.GetSupervisionToolRequests(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting tool requests: %w", err)
	}
	supervisorRequest.ToolRequests = toolRequests

	// Get the messages
	messages, err := s.GetSupervisionMessages(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting messages: %w", err)
	}
	supervisorRequest.Messages = messages

	return &supervisorRequest, nil
}

func (s *PostgresqlStore) GetSupervisionMessages(ctx context.Context, id uuid.UUID) ([]sentinel.LLMMessage, error) {
	query := `
		SELECT m.id, m.role, m.content
		FROM llm_message m
		INNER JOIN toolrequest tr ON m.id = tr.message_id
		WHERE tr.supervisionrequest_id = $1`

	rows, err := s.db.QueryContext(ctx, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no messages found for supervisor request %s, there should be at least one", id)
	} else if err != nil {
		return nil, fmt.Errorf("error getting messages: %w", err)
	}
	defer rows.Close()

	var messages []sentinel.LLMMessage
	for rows.Next() {
		var message sentinel.LLMMessage
		if err := rows.Scan(&message.Id, &message.Role, &message.Content); err != nil {
			return nil, fmt.Errorf("error scanning message: %w", err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}

func (s *PostgresqlStore) CreateSupervisionResult(ctx context.Context, result sentinel.SupervisionResult) error {
	query := `
		INSERT INTO supervisionresult (id, supervisionrequest_id, created_at, decision, reasoning, toolrequest_id)
		VALUES ($1, $2, $3, $4, $5, $6)`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(
		ctx, query, result.Id, result.SupervisionRequestId, result.CreatedAt, result.Decision, result.Reasoning, result.Toolrequest.Id,
	)
	if err != nil {
		return fmt.Errorf("error creating supervisor result: %w", err)
	}

	// Log the supervisionrequest_status entry for the supervision request
	rs := sentinel.SupervisionStatus{Status: sentinel.Completed, CreatedAt: time.Now()}
	err = s.createSupervisionStatus(ctx, result.SupervisionRequestId, rs, tx)
	if err != nil {
		return fmt.Errorf("error creating supervisor status: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (s *PostgresqlStore) GetSupervisionResults(ctx context.Context, id uuid.UUID) ([]*sentinel.SupervisionResult, error) {
	query := `
		SELECT sr.id, sr.supervisionrequest_id, sr.created_at, sr.decision, sr.reasoning, 
		sr.toolrequest_id, tr.tool_id, tr.message_id, tr.arguments
		FROM supervisionresult sr
		LEFT JOIN toolrequest tr ON sr.toolrequest_id = tr.id
		WHERE sr.supervisionrequest_id = $1`

	var tr sentinel.ToolRequest

	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("error getting supervisor results: %w", err)
	}
	defer rows.Close()

	var results []*sentinel.SupervisionResult
	for rows.Next() {
		var result sentinel.SupervisionResult
		if err := rows.Scan(
			&result.Id, &result.SupervisionRequestId, &result.CreatedAt, &result.Decision, &result.Reasoning,
			&tr.Id, &tr.ToolId, &tr.MessageId, &tr.Arguments,
		); err != nil {
			return nil, fmt.Errorf("error scanning supervisor result: %w", err)
		}
		result.Toolrequest = &tr
		results = append(results, &result)
	}

	return results, nil

}

func (s *PostgresqlStore) UpdateSupervisionRequest(ctx context.Context, supervisorRequest sentinel.SupervisionRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	if err = tx.Rollback(); err != nil {
		return fmt.Errorf("error rolling back transaction: %w", err)
	}

	// Update supervisor request
	query1 := `
		UPDATE supervisionrequest 
		SET task_state = $1
		WHERE id = $2`

	_, err = tx.ExecContext(ctx, query1, supervisorRequest.TaskState, supervisorRequest.Id)
	if err != nil {
		return fmt.Errorf("error updating supervisor request: %w", err)
	}

	// Insert new status
	query2 := `
		INSERT INTO supervisionrequest_status (id, supervisionrequest_id, created_at, status)
		VALUES ($1, $2, CURRENT_TIMESTAMP, $3)`

	_, err = tx.ExecContext(ctx, query2, supervisorRequest.Id, supervisorRequest.Id, supervisorRequest.Status.Status)
	if err != nil {
		return fmt.Errorf("error updating supervisor status: %w", err)
	}

	return tx.Commit()
}

func (s *PostgresqlStore) GetSupervisionRequests(ctx context.Context) ([]sentinel.SupervisionRequest, error) {
	// Get a list of all supervision request IDs then pass them to GetSupervisionRequest
	query := `SELECT id FROM supervisionrequest`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting supervision requests: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("error scanning supervision request ID: %w", err)
		}
		ids = append(ids, id)
	}

	var supervisorRequests []sentinel.SupervisionRequest
	for _, id := range ids {
		supervisorRequest, err := s.GetSupervisionRequest(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("error getting supervision request: %w", err)
		}
		supervisorRequests = append(supervisorRequests, *supervisorRequest)
	}

	return supervisorRequests, nil
}

func (s *PostgresqlStore) CountSupervisionRequests(ctx context.Context, status sentinel.Status) (int, error) {
	query := `
        SELECT COUNT(*)
        FROM (
            SELECT DISTINCT ON (sr.id) sr.id
            FROM supervisionrequest sr
            JOIN supervisionrequest_status ss ON sr.id = ss.supervisionrequest_id
            WHERE NOT EXISTS (
                SELECT 1 
                FROM supervisionrequest_status newer
                WHERE newer.supervisionrequest_id = sr.id 
                AND newer.created_at > ss.created_at
            )
            AND ss.status = $1
        ) as latest_requests`

	var count int
	err := s.db.QueryRowContext(ctx, query, status).Scan(&count)
	return count, err
}

// ProjectToolStore implementation
func (s *PostgresqlStore) GetProjectTools(ctx context.Context, id uuid.UUID) ([]sentinel.Tool, error) {
	fmt.Printf("Stub: GetProjectTools called with project ID: %s\n", id)
	return []sentinel.Tool{}, nil
}

func (s *PostgresqlStore) CreateProjectTool(ctx context.Context, id uuid.UUID, tool sentinel.Tool) error {
	fmt.Printf("Stub: CreateProjectTool called with project ID: %s and tool ID: %s\n", id, tool.Id)
	return nil
}

func (s *PostgresqlStore) GetTool(ctx context.Context, id uuid.UUID) (*sentinel.Tool, error) {
	query := `
		SELECT id, name, attributes, description
		FROM tool
		WHERE id = $1`

	var tool sentinel.Tool
	var attributesJSON []byte
	err := s.db.QueryRowContext(ctx, query, id).Scan(&tool.Id, &tool.Name, &attributesJSON, &tool.Description)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting tool: %w", err)
	}

	// Initialize the attributes map
	t := make(map[string]interface{})

	// Parse the JSON attributes
	if len(attributesJSON) > 0 {
		if err := json.Unmarshal(attributesJSON, &t); err != nil {
			return nil, fmt.Errorf("error parsing tool attributes: %w", err)
		}
	}
	tool.Attributes = &t
	return &tool, nil
}

func (s *PostgresqlStore) GetTools(ctx context.Context) ([]sentinel.Tool, error) {
	query := `
		SELECT id, name, attributes, description
		FROM tool`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting tools: %w", err)
	}
	defer rows.Close()

	var tools []sentinel.Tool
	for rows.Next() {
		var tool sentinel.Tool
		var attributesJSON []byte
		if err := rows.Scan(&tool.Id, &tool.Name, &attributesJSON, &tool.Description); err != nil {
			return nil, fmt.Errorf("error scanning tool: %w", err)
		}

		// Initialize the attributes map
		t := make(map[string]interface{})

		// Parse the JSON attributes if they exist
		if len(attributesJSON) > 0 {
			if err := json.Unmarshal(attributesJSON, &t); err != nil {
				return nil, fmt.Errorf("error parsing tool attributes: %w", err)
			}
		}
		tool.Attributes = &t

		tools = append(tools, tool)
	}

	return tools, nil
}

func (s *PostgresqlStore) GetPendingSupervisionRequests(ctx context.Context) ([]sentinel.SupervisionRequest, error) {
	query := `
        SELECT DISTINCT ON (sr.id) 
            sr.id, sr.execution_id, e.run_id, sr.task_state, 
            ss.status, ss.created_at, sr.supervisor_id
        FROM supervisionrequest sr
				INNER JOIN execution e ON sr.execution_id = e.id
        JOIN supervisionrequest_status ss ON sr.id = ss.supervisionrequest_id
        WHERE ss.status = $1
        AND NOT EXISTS (
            SELECT 1 
            FROM supervisionrequest_status newer
            WHERE newer.supervisionrequest_id = sr.id 
            AND newer.created_at > ss.created_at
        )
        ORDER BY sr.id, ss.created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, sentinel.Pending)
	if err != nil {
		return nil, fmt.Errorf("error getting pending supervision requests: %w", err)
	}
	defer rows.Close()

	var supervisorRequests []sentinel.SupervisionRequest
	for rows.Next() {
		var supervisorRequest sentinel.SupervisionRequest
		var status sentinel.SupervisionStatus
		var taskStateJSON []byte
		if err := rows.Scan(
			&supervisorRequest.Id,
			&supervisorRequest.ExecutionId,
			&supervisorRequest.RunId,
			&taskStateJSON,
			&status.Status,
			&status.CreatedAt,
			&supervisorRequest.SupervisorId,
		); err != nil {
			return nil, fmt.Errorf("error scanning supervisor: %w", err)
		}

		if err := json.Unmarshal(taskStateJSON, &supervisorRequest.TaskState); err != nil {
			return nil, fmt.Errorf("error parsing task state: %w", err)
		}

		supervisorRequests = append(supervisorRequests, supervisorRequest)
	}

	return supervisorRequests, nil
}

func (s *PostgresqlStore) GetSupervisorFromToolID(ctx context.Context, id uuid.UUID) (*sentinel.Supervisor, error) {
	query := `
		SELECT s.id, s.description, s.created_at, s.type
		FROM supervisor s
		INNER JOIN tool_supervisor ts ON s.id = ts.supervisor_id
		INNER JOIN tool t ON ts.tool_id = t.id
		WHERE t.id = $1`

	var supervisor sentinel.Supervisor
	err := s.db.QueryRowContext(ctx, query, id).Scan(&supervisor.Id, &supervisor.Description, &supervisor.CreatedAt, &supervisor.Type)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting supervisor: %w", err)
	}
	return &supervisor, nil
}

func (s *PostgresqlStore) CreateSupervisor(ctx context.Context, supervisor sentinel.Supervisor) (uuid.UUID, error) {
	if supervisor.Code == nil {
		return uuid.UUID{}, fmt.Errorf("can't create supervisor, code is required")
	}

	var id uuid.UUID

	// Try and find a supervisor with the same code
	query := `
		SELECT id
		FROM supervisor
		WHERE code = $1`

	var existingSupervisorId uuid.UUID
	err := s.db.QueryRowContext(ctx, query, supervisor.Code).Scan(&existingSupervisorId)
	if err != nil && err != sql.ErrNoRows {
		return uuid.UUID{}, fmt.Errorf("error checking if supervisor already exists: %w", err)
	}

	// If the supervisor already exists, just use the existing ID, else create a new one
	if existingSupervisorId != uuid.Nil {
		return existingSupervisorId, nil
	}

	id = uuid.New()

	query = `
		INSERT INTO supervisor (id, description, created_at, type, code)
		VALUES ($1, $2, $3, $4, $5)`

	_, err = s.db.ExecContext(ctx, query, id, supervisor.Description, supervisor.CreatedAt, supervisor.Type, supervisor.Code)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating supervisor: %w", err)
	}

	return id, nil
}

func (s *PostgresqlStore) CreateRun(ctx context.Context, run sentinel.Run) (uuid.UUID, error) {
	// First check if the project exists
	p, err := s.GetProject(ctx, run.ProjectId)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error getting project: %w", err)
	}
	if p == nil {
		return uuid.UUID{}, fmt.Errorf("project not found: %s", run.ProjectId)
	}

	id := uuid.New()

	query := `
		INSERT INTO run (id, project_id, created_at)
		VALUES ($1, $2, $3)`

	_, err = s.db.ExecContext(ctx, query, id, run.ProjectId, run.CreatedAt)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating run: %w", err)
	}
	return id, nil
}

func (s *PostgresqlStore) CreateTool(ctx context.Context, tool sentinel.Tool) (uuid.UUID, error) {
	id := uuid.New()

	if tool.Name == "" || tool.Description == "" || tool.Attributes == nil {
		return uuid.UUID{}, fmt.Errorf("can't create tool, tool name, description, and attributes are required. Values: %+v", tool)
	}

	attributes, err := json.Marshal(tool.Attributes)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error marshalling tool attributes: %w", err)
	}

	query := `
		INSERT INTO tool (id, name, description, attributes)
		VALUES ($1, $2, $3, $4)`

	_, err = s.db.ExecContext(ctx, query, id, tool.Name, tool.Description, attributes)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating tool: %w", err)
	}

	return id, nil
}

func (s *PostgresqlStore) GetSupervisionToolRequests(ctx context.Context, id uuid.UUID) ([]sentinel.ToolRequest, error) {
	query := `
		SELECT id, supervisionrequest_id, tool_id, message_id, arguments
		FROM toolrequest
		WHERE supervisionrequest_id = $1`

	rows, err := s.db.QueryContext(ctx, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no tool requests found for supervisor request %s, there should be at least one", id)
	} else if err != nil {
		return nil, fmt.Errorf("error getting supervisor tool requests: %w", err)
	}
	defer rows.Close()

	var toolRequests []sentinel.ToolRequest
	for rows.Next() {
		var toolRequest sentinel.ToolRequest
		var argumentsJSON []byte
		if err := rows.Scan(&toolRequest.Id, &toolRequest.SupervisionRequestId, &toolRequest.ToolId, &toolRequest.MessageId, &argumentsJSON); err != nil {
			return nil, fmt.Errorf("error scanning tool request: %w", err)
		}

		// Parse the JSON arguments
		if err := json.Unmarshal(argumentsJSON, &toolRequest.Arguments); err != nil {
			return nil, fmt.Errorf("error parsing tool request arguments: %w", err)
		}

		toolRequests = append(toolRequests, toolRequest)
	}
	return toolRequests, nil
}

func (s *PostgresqlStore) GetRun(ctx context.Context, id uuid.UUID) (*sentinel.Run, error) {
	query := `
		SELECT id, project_id, created_at
		FROM run
		WHERE id = $1`

	var run sentinel.Run
	err := s.db.QueryRowContext(ctx, query, id).Scan(&run.Id, &run.ProjectId, &run.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting run: %w", err)
	}

	return &run, nil
}

func (s *PostgresqlStore) AssignSupervisorsToTool(ctx context.Context, runId uuid.UUID, toolID uuid.UUID, supervisorIds []uuid.UUID) error {
	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, supervisorID := range supervisorIds {
		// Check if the supervisor exists
		supervisor, err := s.GetSupervisor(ctx, supervisorID)
		if err != nil {
			return fmt.Errorf("error getting supervisor: %w", err)
		}
		if supervisor == nil {
			return fmt.Errorf("supervisor %s not found", supervisorID)
		}

		// Check if the tool exists
		tool, err := s.GetTool(ctx, toolID)
		if err != nil {
			return fmt.Errorf("error getting tool: %w", err)
		}
		if tool == nil {
			return fmt.Errorf("tool %s not found", toolID)
		}

		// Check if the supervisor is already assigned to the tool
		query := `
		SELECT 1
		FROM run_tool_supervisor
		WHERE run_id = $1 AND tool_id = $2 AND supervisor_id = $3`

		var exists bool
		err = tx.QueryRowContext(ctx, query, runId, toolID, supervisorID).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("error checking if supervisor is already assigned to tool: %w", err)
		}

		if exists {
			return fmt.Errorf("supervisor %s is already assigned to tool %s", supervisorID, toolID)
		}

		t := time.Now()
		query = `
		INSERT INTO run_tool_supervisor (run_id, tool_id, supervisor_id, created_at)
		VALUES ($1, $2, $3, $4)`

		// If the supervisor is not assigned to the tool, assign them
		_, err = tx.ExecContext(ctx, query, runId, toolID, supervisorID, t)
		if err != nil {
			return fmt.Errorf("error assigning supervisor to tool: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (s *PostgresqlStore) GetRunTools(ctx context.Context, runId uuid.UUID) ([]sentinel.Tool, error) {
	query := `
		SELECT tool.id, tool.name, tool.description, tool.attributes
		FROM run_tool_supervisor
		INNER JOIN tool ON run_tool_supervisor.tool_id = tool.id
		WHERE run_tool_supervisor.run_id = $1`

	rows, err := s.db.QueryContext(ctx, query, runId)
	if err != nil {
		return nil, fmt.Errorf("error getting run tools: %w", err)
	}
	defer rows.Close()

	var tools []sentinel.Tool
	for rows.Next() {
		var tool sentinel.Tool
		var attributesJSON []byte
		if err := rows.Scan(&tool.Id, &tool.Name, &tool.Description, &attributesJSON); err != nil {
			return nil, fmt.Errorf("error scanning tool: %w", err)
		}

		// Initialize the attributes map
		t := make(map[string]interface{})

		// Parse the JSON attributes if they exist
		if len(attributesJSON) > 0 {
			if err := json.Unmarshal(attributesJSON, &t); err != nil {
				return nil, fmt.Errorf("error parsing tool attributes: %w", err)
			}
		}
		tool.Attributes = &t

		tools = append(tools, tool)
	}

	return tools, nil
}

// GetRunToolSupervisors returns an ordered list of supervisors assigned to a tool for a run. The list is ordered by ID of the run_tool_supervisor record, most recent first.
func (s *PostgresqlStore) GetRunToolSupervisors(ctx context.Context, runId uuid.UUID, toolId uuid.UUID) ([]sentinel.Supervisor, error) {
	query := `
		SELECT supervisor.id, supervisor.description, supervisor.created_at, supervisor.type, rts.created_at
		FROM run_tool_supervisor rts
		INNER JOIN supervisor ON rts.supervisor_id = supervisor.id
		WHERE rts.run_id = $1 AND rts.tool_id = $2
		ORDER BY rts.id DESC`

	rows, err := s.db.QueryContext(ctx, query, runId, toolId)
	if err != nil {
		return nil, fmt.Errorf("error getting run tools: %w", err)
	}
	defer rows.Close()

	// TODO this is returning the created_at of the run_tool_supervisor record, not the supervisor
	var supervisors []sentinel.Supervisor
	for rows.Next() {
		var supervisor sentinel.Supervisor
		if err := rows.Scan(
			&supervisor.Id,
			&supervisor.Description,
			&supervisor.CreatedAt,
			&supervisor.Type,
			&supervisor.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning supervisor: %w", err)
		}
		supervisors = append(supervisors, supervisor)
	}

	return supervisors, nil
}

func (s *PostgresqlStore) GetSupervisor(ctx context.Context, id uuid.UUID) (*sentinel.Supervisor, error) {
	query := `
		SELECT id, description, created_at, type
		FROM supervisor
		WHERE id = $1`

	var supervisor sentinel.Supervisor
	err := s.db.QueryRowContext(ctx, query, id).Scan(&supervisor.Id, &supervisor.Description, &supervisor.CreatedAt, &supervisor.Type)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting supervisor: %w", err)
	}

	return &supervisor, nil
}

func (s *PostgresqlStore) GetSupervisors(ctx context.Context) ([]sentinel.Supervisor, error) {
	query := `
		SELECT id, description, created_at, type
		FROM supervisor`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting supervisors: %w", err)
	}
	defer rows.Close()

	var supervisors []sentinel.Supervisor
	for rows.Next() {
		var supervisor sentinel.Supervisor
		if err := rows.Scan(&supervisor.Id, &supervisor.Description, &supervisor.CreatedAt, &supervisor.Type); err != nil {
			return nil, fmt.Errorf("error scanning supervisor: %w", err)
		}
		supervisors = append(supervisors, supervisor)
	}

	return supervisors, nil
}

func (s *PostgresqlStore) GetSupervisionRequestsForExecution(ctx context.Context, executionId uuid.UUID) ([]sentinel.SupervisionRequest, error) {
	query := `
        SELECT sr.id, sr.execution_id, sr.supervisor_id, sr.task_state
        FROM supervisionrequest sr
        WHERE sr.execution_id = $1`

	rows, err := s.db.QueryContext(ctx, query, executionId)
	if err != nil {
		return nil, fmt.Errorf("error getting supervision requests: %w", err)
	}
	defer rows.Close()

	var requests []sentinel.SupervisionRequest
	for rows.Next() {
		var request sentinel.SupervisionRequest
		var taskStateJSON []byte

		err := rows.Scan(
			&request.Id,
			&request.ExecutionId,
			&request.SupervisorId,
			&taskStateJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning supervision request: %w", err)
		}

		// Parse the task state JSON
		if err := json.Unmarshal(taskStateJSON, &request.TaskState); err != nil {
			return nil, fmt.Errorf("error parsing task state: %w", err)
		}

		requests = append(requests, request)
	}

	return requests, nil
}

func (s *PostgresqlStore) GetSupervisionResultsForExecution(ctx context.Context, executionId uuid.UUID) ([]sentinel.SupervisionResult, error) {
	query := `
        SELECT sr.id, sr.supervisionrequest_id, sr.created_at, sr.decision, sr.reasoning
        FROM supervisionresult sr
        INNER JOIN supervisionrequest sreq ON sr.supervisionrequest_id = sreq.id
        WHERE sreq.execution_id = $1`

	rows, err := s.db.QueryContext(ctx, query, executionId)
	if err != nil {
		return nil, fmt.Errorf("error getting supervision results: %w", err)
	}
	defer rows.Close()

	var results []sentinel.SupervisionResult
	for rows.Next() {
		var result sentinel.SupervisionResult
		err := rows.Scan(
			&result.Id,
			&result.SupervisionRequestId,
			&result.CreatedAt,
			&result.Decision,
			&result.Reasoning,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning supervision result: %w", err)
		}

		results = append(results, result)
	}

	return results, nil
}

func (s *PostgresqlStore) GetSupervisionStatusesForExecution(ctx context.Context, executionId uuid.UUID) ([]sentinel.SupervisionStatus, error) {
	query := `
        SELECT ss.id, ss.supervisionrequest_id, ss.status, ss.created_at
        FROM supervisionrequest_status ss
        INNER JOIN supervisionrequest sr ON ss.supervisionrequest_id = sr.id
        WHERE sr.execution_id = $1`

	rows, err := s.db.QueryContext(ctx, query, executionId)
	if err != nil {
		return nil, fmt.Errorf("error getting supervision statuses: %w", err)
	}
	defer rows.Close()

	var statuses []sentinel.SupervisionStatus
	for rows.Next() {
		var status sentinel.SupervisionStatus
		err := rows.Scan(
			&status.Id,
			&status.SupervisionRequestId,
			&status.Status,
			&status.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning supervision status: %w", err)
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}
