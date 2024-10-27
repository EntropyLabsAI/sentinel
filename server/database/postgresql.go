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

// ReviewStore implementation
// Todo:
// store messages
// store tool requests
// store review status
// check result of review is stored somewhere
// ensure that the review status is updated 3 times (timeout, pending, completed)

// for _, toolRequest := range request.ToolRequests {
// err := store.CreateToolRequest(ctx, reviewID, toolRequest)
// if err != nil {
// 	http.Error(w, err.Error(), http.StatusInternalServerError)
// 	return
// }
// }
// for _, message := range request.Messages {
// 	err := store.CreateMessage(ctx, reviewID, message)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }

func (s *PostgresqlStore) CreateReviewRequest(ctx context.Context, request sentinel.ReviewRequest) (uuid.UUID, error) {
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

	// Store the review request
	taskStateJSON, err := json.Marshal(request.TaskState)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error marshalling task state: %w", err)
	}

	query := `
		INSERT INTO reviewrequest (id, run_id, task_state)
		VALUES ($1, $2, $3)`

	requestID := uuid.New()
	_, err = tx.ExecContext(ctx, query, requestID, request.RunId, taskStateJSON)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating review: %w", err)
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
			INSERT INTO toolrequest (id, reviewrequest_id, tool_id, message_id, arguments)
			VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.ExecContext(
			ctx, query, toolRequestID, requestID, toolRequest.ToolId, messageIDs[i], argumentsJSON,
		)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("error creating tool request: %w", err)
		}
	}

	status := sentinel.ReviewStatus{Status: sentinel.Pending, CreatedAt: time.Now()}

	// Store a review status pending
	err = s.createReviewStatus(ctx, requestID, status, tx)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating review status: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error committing transaction: %w", err)
	}

	return requestID, nil
}

func (s *PostgresqlStore) CreateReviewStatus(ctx context.Context, requestID uuid.UUID, status sentinel.ReviewStatus) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	return s.createReviewStatus(ctx, requestID, status, tx)
}

func (s *PostgresqlStore) createReviewStatus(ctx context.Context, requestID uuid.UUID, status sentinel.ReviewStatus, tx *sql.Tx) error {
	query := `
		INSERT INTO reviewrequest_status (id, reviewrequest_id, status, created_at)
		VALUES ($1, $2, $3, $4)`

	id := uuid.New()
	_, err := tx.ExecContext(ctx, query, id, requestID, status.Status, status.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creating review status: %w", err)
	}

	return nil
}

func (s *PostgresqlStore) GetReviewRequest(ctx context.Context, id uuid.UUID) (*sentinel.ReviewRequest, error) {
	query := `
		SELECT rr.id, rr.run_id, rr.task_state, rs.id, rs.status, rs.created_at
		FROM reviewrequest rr
		INNER JOIN reviewrequest_status rs ON rr.id = rs.reviewrequest_id
		WHERE rr.id = $1
		ORDER BY rs.created_at DESC
		LIMIT 1`

	var reviewRequest sentinel.ReviewRequest
	var status sentinel.ReviewStatus
	var taskStateJSON []byte // Add temporary variable for JSON data
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&reviewRequest.Id, &reviewRequest.RunId, &taskStateJSON, &status.Id, &status.Status, &status.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting review: %w", err)
	}

	// Parse the JSON task state
	if err := json.Unmarshal(taskStateJSON, &reviewRequest.TaskState); err != nil {
		return nil, fmt.Errorf("error parsing task state: %w", err)
	}

	reviewRequest.Status = &status

	// Get the tool requests
	toolRequests, err := s.GetReviewToolRequests(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting tool requests: %w", err)
	}
	reviewRequest.ToolRequests = toolRequests

	// Get the messages
	messages, err := s.GetReviewMessages(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting messages: %w", err)
	}
	reviewRequest.Messages = messages

	return &reviewRequest, nil
}

func (s *PostgresqlStore) GetReviewMessages(ctx context.Context, id uuid.UUID) ([]sentinel.LLMMessage, error) {
	query := `
		SELECT id, role, content
		FROM llm_message
		WHERE reviewrequest_id = $1`

	rows, err := s.db.QueryContext(ctx, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no messages found for review request %s, there should be at least one", id)
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

func (s *PostgresqlStore) CreateReviewResult(ctx context.Context, result sentinel.ReviewResult) error {
	query := `
		INSERT INTO reviewresult (id, reviewrequest_id, created_at, decision, reasoning, toolrequest_id)
		VALUES ($1, $2, $3, $4, $5, $6)`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(
		ctx, query, result.Id, result.ReviewRequestId, result.CreatedAt, result.Decision, result.Reasoning, result.Toolrequest.Id,
	)
	if err != nil {
		return fmt.Errorf("error creating review result: %w", err)
	}

	// Log the reviewrequest_status entry for the review request
	rs := sentinel.ReviewStatus{Status: sentinel.Completed, CreatedAt: time.Now()}
	err = s.createReviewStatus(ctx, result.ReviewRequestId, rs, tx)
	if err != nil {
		return fmt.Errorf("error creating review status: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (s *PostgresqlStore) GetReviewResults(ctx context.Context, id uuid.UUID) ([]*sentinel.ReviewResult, error) {
	query := `
		SELECT rr.id, rr.reviewrequest_id, rr.created_at, rr.decision, rr.reasoning, 
		rr.toolrequest_id, tr.tool_id, tr.message_id, tr.arguments
		FROM reviewresult rr
		LEFT JOIN toolrequest tr ON rr.toolrequest_id = tr.id
		WHERE rr.reviewrequest_id = $1`

	var tr sentinel.ToolRequest

	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("error getting review results: %w", err)
	}
	defer rows.Close()

	var results []*sentinel.ReviewResult
	for rows.Next() {
		var result sentinel.ReviewResult
		if err := rows.Scan(
			&result.Id, &result.ReviewRequestId, &result.CreatedAt, &result.Decision, &result.Reasoning,
			&tr.Id, &tr.ToolId, &tr.MessageId, &tr.Arguments,
		); err != nil {
			return nil, fmt.Errorf("error scanning review result: %w", err)
		}
		result.Toolrequest = &tr
		results = append(results, &result)
	}

	return results, nil

}

func (s *PostgresqlStore) UpdateReviewRequest(ctx context.Context, reviewRequest sentinel.ReviewRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	if err = tx.Rollback(); err != nil {
		return fmt.Errorf("error rolling back transaction: %w", err)
	}

	// Update review request
	query1 := `
		UPDATE reviewrequest 
		SET task_state = $1
		WHERE id = $2`

	_, err = tx.ExecContext(ctx, query1, reviewRequest.TaskState, reviewRequest.Id)
	if err != nil {
		return fmt.Errorf("error updating review request: %w", err)
	}

	// Insert new status
	query2 := `
		INSERT INTO reviewrequest_status (id, reviewrequest_id, created_at, status)
		VALUES ($1, $2, CURRENT_TIMESTAMP, $3)`

	_, err = tx.ExecContext(ctx, query2, reviewRequest.Id, reviewRequest.Id, reviewRequest.Status.Status)
	if err != nil {
		return fmt.Errorf("error updating review status: %w", err)
	}

	return tx.Commit()
}

func (s *PostgresqlStore) GetReviewRequests(ctx context.Context) ([]sentinel.ReviewRequest, error) {
	query := `
		SELECT rr.id, rr.run_id, rr.task_state, rs.id, rs.status, rs.created_at
		FROM reviewrequest rr
		LEFT JOIN reviewrequest_status rs ON rr.id = rs.reviewrequest_id
		ORDER BY rs.created_at DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting reviews: %w", err)
	}
	defer rows.Close()

	var reviewRequests []sentinel.ReviewRequest
	for rows.Next() {
		var reviewRequest sentinel.ReviewRequest
		var status sentinel.ReviewStatus
		var taskStateJSON []byte
		if err := rows.Scan(&reviewRequest.Id, &reviewRequest.RunId, &taskStateJSON, &status.Id, &status.Status, &status.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning review request: %w", err)
		}

		// Parse the JSON task state
		if err := json.Unmarshal(taskStateJSON, &reviewRequest.TaskState); err != nil {
			return nil, fmt.Errorf("error parsing task state: %w", err)
		}

		reviewRequest.Status = &status
		reviewRequests = append(reviewRequests, reviewRequest)
	}

	return reviewRequests, nil
}

func (s *PostgresqlStore) CountReviewRequests(ctx context.Context, status sentinel.ReviewStatusStatus) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM reviewrequest_status
		WHERE status = $1`

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

	// Parse the JSON attributes
	if err := json.Unmarshal(attributesJSON, &tool.Attributes); err != nil {
		return nil, fmt.Errorf("error parsing tool attributes: %w", err)
	}

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

		// Parse the JSON attributes
		if err := json.Unmarshal(attributesJSON, &tool.Attributes); err != nil {
			return nil, fmt.Errorf("error parsing tool attributes: %w", err)
		}

		tools = append(tools, tool)
	}

	return tools, nil
}

func (s *PostgresqlStore) GetPendingReviewRequests(ctx context.Context) ([]sentinel.ReviewRequest, error) {
	query := `
        SELECT DISTINCT ON (rr.id) 
            rr.id, rr.run_id, rr.task_state, 
            rs.status, rs.created_at
        FROM reviewrequest rr
        JOIN reviewrequest_status rs ON rr.id = rs.reviewrequest_id
        WHERE rs.status = $1
        AND NOT EXISTS (
            SELECT 1 
            FROM reviewrequest_status newer
            WHERE newer.reviewrequest_id = rr.id 
            AND newer.created_at > rs.created_at
        )
        ORDER BY rr.id, rs.created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, sentinel.Pending)
	if err != nil {
		return nil, fmt.Errorf("error getting pending reviews: %w", err)
	}
	defer rows.Close()

	var reviewRequests []sentinel.ReviewRequest
	for rows.Next() {
		var reviewRequest sentinel.ReviewRequest
		var status sentinel.ReviewStatus
		var taskStateJSON []byte
		if err := rows.Scan(&reviewRequest.Id, &reviewRequest.RunId, &taskStateJSON, &status.Status, &status.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning review: %w", err)
		}

		if err := json.Unmarshal(taskStateJSON, &reviewRequest.TaskState); err != nil {
			return nil, fmt.Errorf("error parsing task state: %w", err)
		}

		reviewRequests = append(reviewRequests, reviewRequest)
	}

	return reviewRequests, nil
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
	id := uuid.New()

	query := `
		INSERT INTO run (id, project_id, created_at)
		VALUES ($1, $2, $3)`

	_, err := s.db.ExecContext(ctx, query, id, run.ProjectId, run.CreatedAt)
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

func (s *PostgresqlStore) GetReviewToolRequests(ctx context.Context, id uuid.UUID) ([]sentinel.ToolRequest, error) {
	query := `
		SELECT id, reviewrequest_id, tool_id, message_id, arguments
		FROM toolrequest
		WHERE reviewrequest_id = $1`

	rows, err := s.db.QueryContext(ctx, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no tool requests found for review request %s, there should be at least one", id)
	} else if err != nil {
		return nil, fmt.Errorf("error getting review tool requests: %w", err)
	}
	defer rows.Close()

	var toolRequests []sentinel.ToolRequest
	for rows.Next() {
		var toolRequest sentinel.ToolRequest
		var argumentsJSON []byte
		if err := rows.Scan(&toolRequest.Id, &toolRequest.ReviewRequestId, &toolRequest.ToolId, &toolRequest.MessageId, &argumentsJSON); err != nil {
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
		if err := rows.Scan(&tool.Id, &tool.Name, &tool.Description, &tool.Attributes); err != nil {
			return nil, fmt.Errorf("error scanning tool: %w", err)
		}
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
