package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sentinel "github.com/entropylabsai/sentinel/server"
	"github.com/google/uuid"
	"github.com/lib/pq"
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
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting project: %w", err)
	}

	return &project, nil
}

func (s *PostgresqlStore) GetProjectFromName(ctx context.Context, name string) (*sentinel.Project, error) {
	query := `
		SELECT id, name, created_at
		FROM project
		WHERE name = $1`

	var project sentinel.Project
	err := s.db.QueryRowContext(ctx, query, name).Scan(&project.Id, &project.Name, &project.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting project: %w", err)
	}

	return &project, nil
}

func (s *PostgresqlStore) GetToolFromValues(ctx context.Context, attributes map[string]interface{}, name string, description string, ignoredAttributes []string) (*sentinel.Tool, error) {
	query := `
		SELECT id, name, description, attributes, ignored_attributes
		FROM tool
		WHERE name = $1
		AND description = $2
		AND attributes = $3
		AND ignored_attributes = $4`

	attrJSON, err := json.Marshal(attributes)
	if err != nil {
		return nil, fmt.Errorf("error marshalling attributes: %w", err)
	}

	var tool sentinel.Tool
	var attributesJSON []byte
	var toolIgnoredAttributes []string
	err = s.db.QueryRowContext(ctx, query, name, description, attrJSON, pq.Array(ignoredAttributes)).Scan(
		&tool.Id,
		&tool.Name,
		&tool.Description,
		&attributesJSON,
		pq.Array(&toolIgnoredAttributes),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting tool from values: %w", err)
	}

	// Parse the JSON attributes if they exist
	if len(attributesJSON) > 0 {
		var attrs map[string]interface{}
		if err := json.Unmarshal(attributesJSON, &attrs); err != nil {
			return nil, fmt.Errorf("error parsing tool attributes: %w", err)
		}
		tool.Attributes = &attrs
	}

	tool.IgnoredAttributes = &toolIgnoredAttributes
	return &tool, nil
}

func (s *PostgresqlStore) GetSupervisorFromValues(
	ctx context.Context,
	code string,
	name string,
	desc string,
	t sentinel.SupervisorType,
	attributes map[string]interface{},
) (*sentinel.Supervisor, error) {
	query := `
		SELECT id, code, name, description, type, created_at
		FROM supervisor
		WHERE code = $1
		AND name = $2
		AND description = $3
		AND type = $4
		AND attributes = $5`

	attrJSON, err := json.Marshal(attributes)
	if err != nil {
		return nil, fmt.Errorf("error marshalling attributes: %w", err)
	}

	var supervisor sentinel.Supervisor
	err = s.db.QueryRowContext(
		ctx, query, code, name, desc, t, attrJSON,
	).Scan(
		&supervisor.Id, &supervisor.Code, &supervisor.Name, &supervisor.Description, &supervisor.Type, &supervisor.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting supervisor by values: %w", err)
	}

	// Just use the attributes passed in since we know they match, saves us having to fiddle with converting.
	supervisor.Attributes = attributes

	return &supervisor, nil
}

func (s *PostgresqlStore) GetProjects(ctx context.Context) ([]sentinel.Project, error) {
	query := `
		SELECT id, name, created_at
		FROM project
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error listing projects: %w", err)
	}
	defer rows.Close()

	projects := make([]sentinel.Project, 0)
	for rows.Next() {
		var project sentinel.Project
		if err := rows.Scan(&project.Id, &project.Name, &project.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning project: %w", err)
		}
		projects = append(projects, project)
	}

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
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error getting project runs: %w", err)
	}
	defer rows.Close()

	runs := make([]sentinel.Run, 0)
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

func (s *PostgresqlStore) getChainsForTool(ctx context.Context, toolId uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT chain_id
		FROM chain_tool ct
		WHERE tool_id = $1`

	rows, err := s.db.QueryContext(ctx, query, toolId)
	if err != nil {
		return nil, fmt.Errorf("error getting chains: %w", err)
	}
	defer rows.Close()

	chainIds := make([]uuid.UUID, 0)
	for rows.Next() {
		var chainId uuid.UUID
		if err := rows.Scan(&chainId); err != nil {
			return nil, fmt.Errorf("error scanning chain ID: %w", err)
		}
		chainIds = append(chainIds, chainId)
	}

	return chainIds, nil
}

func (s *PostgresqlStore) CreateToolSupervisorChain(ctx context.Context, toolId uuid.UUID, chain sentinel.ChainRequest) (*uuid.UUID, error) {
	ids := *chain.SupervisorIds
	if ids == nil {
		return nil, fmt.Errorf("supervisor IDs are required to make a chain of supervisors")
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Create new chain
	chainId := uuid.New()
	query := `
		INSERT INTO chain (id)
		VALUES ($1)`

	_, err = tx.ExecContext(ctx, query, chainId)
	if err != nil {
		return nil, fmt.Errorf("error creating chain: %w", err)
	}

	// Link chain to tool
	query = `
		INSERT INTO chain_tool (tool_id, chain_id)
		VALUES ($1, $2)`

	_, err = tx.ExecContext(ctx, query, toolId, chainId)
	if err != nil {
		return nil, fmt.Errorf("error linking tool to chain: %w", err)
	}

	// Add chain_supervisor entries for each supervisor
	query = `
		INSERT INTO chain_supervisor (chain_id, supervisor_id, position_in_chain)
		VALUES ($1, $2, $3)`

	for i, supervisorId := range ids {
		_, err = tx.ExecContext(ctx, query, chainId, supervisorId, i)
		if err != nil {
			return nil, fmt.Errorf("error adding supervisor to chain: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return &chainId, nil
}

func (s *PostgresqlStore) GetToolSupervisorChains(ctx context.Context, toolId uuid.UUID) ([]sentinel.ToolSupervisorChain, error) {
	chainIds, err := s.getChainsForTool(ctx, toolId)
	if err != nil {
		return nil, fmt.Errorf("error getting chains for tool: %w", err)
	}

	chains := make([]sentinel.ToolSupervisorChain, 0)
	for _, chainId := range chainIds {
		query := `
			SELECT s.id, s.name, s.description, s.type, s.attributes, s.created_at, s.code
			FROM chain_supervisor cs
			INNER JOIN supervisor s ON cs.supervisor_id = s.id
			WHERE cs.chain_id = $1`

		rows, err := s.db.QueryContext(ctx, query, chainId)
		if err != nil {
			return nil, fmt.Errorf("error getting tool supervisor chains: %w", err)
		}
		defer rows.Close()

		supervisors := make([]sentinel.Supervisor, 0)
		for rows.Next() {
			// Parse out the attributes bytes into json
			var attributesJSON []byte
			var supervisor sentinel.Supervisor
			if err := rows.Scan(
				&supervisor.Id,
				&supervisor.Name,
				&supervisor.Description,
				&supervisor.Type,
				&attributesJSON,
				&supervisor.CreatedAt,
				&supervisor.Code,
			); err != nil {
				return nil, fmt.Errorf("error scanning supervisor: %w", err)
			}

			// Parse theSupervisorTypeNoSupervisor JSON attributes if they exist
			var attributes map[string]interface{}
			if len(attributesJSON) > 0 {
				if err := json.Unmarshal(attributesJSON, &attributes); err != nil {
					return nil, fmt.Errorf("error parsing supervisor attributes: %w", err)
				}
				supervisor.Attributes = attributes
			}

			supervisors = append(supervisors, supervisor)
		}

		chains = append(chains, sentinel.ToolSupervisorChain{
			ChainId:     chainId,
			Supervisors: supervisors,
		})
	}

	return chains, nil
}

func (s *PostgresqlStore) CreateToolRequestGroup(ctx context.Context, toolId uuid.UUID, request sentinel.ToolRequestGroup) (*sentinel.ToolRequestGroup, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	groupId := uuid.New()
	trg := sentinel.ToolRequestGroup{
		Id:           &groupId,
		ToolRequests: make([]sentinel.ToolRequest, 0, len(request.ToolRequests)),
	}

	// Create a new requestgroup
	query := `
		INSERT INTO requestgroup (id)
		VALUES ($1)`
	_, err = tx.ExecContext(ctx, query, groupId)
	if err != nil {
		return nil, fmt.Errorf("error creating requestgroup: %w", err)
	}

	// For each tool request, create a tool request
	for _, toolRequest := range request.ToolRequests {
		id := uuid.New()
		toolRequest.Id = &id
		toolRequest.RequestgroupId = &groupId
		err = s.createToolRequest(ctx, tx, toolRequest)
		if err != nil {
			return nil, fmt.Errorf("error creating tool request: %w", err)
		}
		trg.ToolRequests = append(trg.ToolRequests, toolRequest)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return &trg, nil
}

func (s *PostgresqlStore) GetRequestGroup(ctx context.Context, id uuid.UUID) (*sentinel.ToolRequestGroup, error) {
	query := `
		SELECT tr.id, tr.tool_id, tr.message_id, tr.arguments, tr.task_state, tr.requestgroup_id, m.role, m.content
		FROM toolrequest tr
		INNER JOIN message m ON tr.message_id = m.id
		WHERE tr.requestgroup_id = $1`

	toolRequestGroup := sentinel.ToolRequestGroup{
		Id:           &id,
		ToolRequests: make([]sentinel.ToolRequest, 0),
	}

	rows, err := s.db.QueryContext(ctx, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting request group: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var toolRequest sentinel.ToolRequest
		var taskStateJSON []byte
		if err := rows.Scan(
			&toolRequest.Id,
			&toolRequest.ToolId,
			&toolRequest.Arguments,
			&taskStateJSON,
			&toolRequest.RequestgroupId,
			&toolRequest.Message.Role,
			&toolRequest.Message.Content,
		); err != nil {
			return nil, fmt.Errorf("error scanning tool request: %w", err)
		}

		// Parse the task state JSON if it exists
		if len(taskStateJSON) > 0 {
			if err := json.Unmarshal(taskStateJSON, &toolRequest.TaskState); err != nil {
				return nil, fmt.Errorf("error parsing task state: %w", err)
			}
		}

		toolRequestGroup.ToolRequests = append(toolRequestGroup.ToolRequests, toolRequest)
	}

	return &toolRequestGroup, nil
}

func (s *PostgresqlStore) GetRunRequestGroups(ctx context.Context, runId uuid.UUID) ([]sentinel.ToolRequestGroup, error) {
	return nil, nil
}

func (s *PostgresqlStore) getChainExecution(ctx context.Context, chainExecutionId uuid.UUID) (*uuid.UUID, error) {
	query := `
		SELECT 1
		FROM chainexecution
		WHERE id = $1`

	err := s.db.QueryRowContext(ctx, query, chainExecutionId).Scan()
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting chain execution: %w", err)
	}

	return &chainExecutionId, nil
}

func (s *PostgresqlStore) CreateSupervisionRequest(ctx context.Context, request sentinel.SupervisionRequest) (*uuid.UUID, error) {
	// Sanity check that we're recording this against a valid chain execution group that already exists
	execution, err := s.getChainExecution(ctx, request.ChainexecutionId)
	if err != nil {
		return nil, fmt.Errorf("error getting execution: %w", err)
	}
	if execution == nil {
		return nil, fmt.Errorf("execution not found for supervision request: %s", request.ChainexecutionId)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		INSERT INTO supervisionrequest (id, supervisor_id, position_in_chain, chainexecution_id)
		VALUES ($1, $2, $3, $4)`

	requestID := uuid.New()
	_, err = tx.ExecContext(
		ctx, query, requestID, request.SupervisorId, request.PositionInChain, request.ChainexecutionId,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating supervision request: %w", err)
	}

	status := sentinel.SupervisionStatus{
		Status:    sentinel.Pending,
		CreatedAt: time.Now(),
	}

	// Store a supervisor status pending
	err = s.createSupervisionStatus(ctx, requestID, status, tx)
	if err != nil {
		return nil, fmt.Errorf("error creating supervisor status: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return &requestID, nil
}

func (s *PostgresqlStore) createMessage(ctx context.Context, tx *sql.Tx, message sentinel.Message) (*uuid.UUID, error) {
	query := `
		INSERT INTO message (id, role, content)
		VALUES ($1, $2, $3)`

	id := uuid.New()
	_, err := tx.ExecContext(ctx, query, id, message.Role, message.Content)
	if err != nil {
		return nil, fmt.Errorf("error creating message: %w", err)
	}

	return &id, nil
}

func (s *PostgresqlStore) createToolRequest(ctx context.Context, tx *sql.Tx, request sentinel.ToolRequest) error {

	query := `
		INSERT INTO toolrequest (id, tool_id, message_id, arguments, task_state, requestgroup_id)
		VALUES ($1, $2, $3, $4, $5, $6)`

	messageID, err := s.createMessage(ctx, tx, request.Message)
	if err != nil {
		return fmt.Errorf("error creating message: %w", err)
	}

	taskStateJSON, err := json.Marshal(request.TaskState)
	if err != nil {
		return fmt.Errorf("error marshalling task state: %w", err)
	}

	argumentsJSON, err := json.Marshal(request.Arguments)
	if err != nil {
		return fmt.Errorf("error marshalling tool request arguments: %w", err)
	}

	_, err = tx.ExecContext(
		ctx, query, request.Id, request.ToolId, messageID, argumentsJSON, taskStateJSON, request.RequestgroupId,
	)
	if err != nil {
		return fmt.Errorf("error creating tool request: %w", err)
	}

	return nil
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

func (s *PostgresqlStore) GetTool(ctx context.Context, id uuid.UUID) (*sentinel.Tool, error) {
	query := `
		SELECT id, run_id, name, description, attributes, ignored_attributes
		FROM tool
		WHERE id = $1`

	var tool sentinel.Tool
	var attributesJSON []byte
	var ignoredAttributes []string
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&tool.Id,
		&tool.RunId,
		&tool.Name,
		&tool.Description,
		&attributesJSON,
		pq.Array(&ignoredAttributes),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting tool: %w", err)
	}

	// Parse the JSON attributes if they exist
	if len(attributesJSON) > 0 {
		var attributes map[string]interface{}
		if err := json.Unmarshal(attributesJSON, &attributes); err != nil {
			return nil, fmt.Errorf("error parsing tool attributes: %w", err)
		}
		tool.Attributes = &attributes
	}

	tool.IgnoredAttributes = &ignoredAttributes
	return &tool, nil
}

func (s *PostgresqlStore) GetProjectTools(ctx context.Context, projectId uuid.UUID) ([]sentinel.Tool, error) {
	query := `
		SELECT id FROM run WHERE project_id = $1`

	rows, err := s.db.QueryContext(ctx, query, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project tools: %w", err)
	}
	defer rows.Close()

	tools := make([]sentinel.Tool, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(
			&id,
		); err != nil {
			return nil, fmt.Errorf("error scanning tool: %w", err)
		}

		runTools, err := s.GetRunTools(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("error getting run tools for project: %w", err)
		}

		tools = append(tools, runTools...)
	}

	return tools, nil
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
	if err != nil {
		return 0, fmt.Errorf("error counting supervision requests: %w", err)
	}

	return count, nil
}

func (s *PostgresqlStore) CreateSupervisionResult(ctx context.Context, result sentinel.SupervisionResult, requestId uuid.UUID) (*uuid.UUID, error) {
	query := `
		INSERT INTO supervisionresult (id, supervisionrequest_id, created_at, decision, reasoning)
		VALUES ($1, $2, $3, $4, $5)`

	id := uuid.New()
	_, err := s.db.ExecContext(ctx, query, id, requestId, result.CreatedAt, result.Decision, result.Reasoning)
	if err != nil {
		return nil, fmt.Errorf("error creating supervision result: %w", err)
	}

	return &id, nil
}

func (s *PostgresqlStore) GetSupervisionRequestsForStatus(ctx context.Context, status sentinel.Status) ([]sentinel.SupervisionRequest, error) {
	// Get IDs of supervision requests with the given status (excluding client supervisors)
	query := `
		SELECT sr.id
		FROM supervisionrequest sr
		JOIN supervisor s ON s.id = sr.supervisor_id
		JOIN (
				SELECT supervisionrequest_id, MAX(id) as latest_status_id
				FROM supervisionrequest_status
				GROUP BY supervisionrequest_id
		) latest ON sr.id = latest.supervisionrequest_id
		JOIN supervisionrequest_status srs ON srs.id = latest.latest_status_id
		WHERE s.type != $1 AND srs.status = $2
	`
	rows, err := s.db.QueryContext(ctx, query, sentinel.ClientSupervisor, status)
	if err != nil {
		return nil, fmt.Errorf("error getting supervision request IDs: %w", err)
	}
	defer rows.Close()

	var requestIds []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("error scanning supervision request ID: %w", err)
		}
		requestIds = append(requestIds, id)
	}

	// If there are no requests, return an empty list
	if len(requestIds) == 0 {
		return nil, nil
	}

	// Get full details for each supervision request
	requests := make([]sentinel.SupervisionRequest, 0, len(requestIds))
	for _, id := range requestIds {
		request, err := s.GetSupervisionRequest(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("error getting supervision request %s: %w", id, err)
		}
		if request != nil {
			requests = append(requests, *request)
		}
	}

	if len(requests) != len(requestIds) {
		return nil, fmt.Errorf("number of requests (%d) does not match number of request IDs (%d) for status %s", len(requests), len(requestIds), status)
	}

	return requests, nil
}

func (s *PostgresqlStore) GetSupervisionRequest(ctx context.Context, id uuid.UUID) (*sentinel.SupervisionRequest, error) {
	query := `
		SELECT id, supervisor_id, position_in_chain, chainexecution_id
		FROM supervisionrequest
		WHERE id = $1`

	var request sentinel.SupervisionRequest
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&request.Id,
		&request.SupervisorId,
		&request.PositionInChain,
		&request.ChainexecutionId,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting supervision request: %w", err)
	}

	return &request, nil
}

func (s *PostgresqlStore) CreateSupervisor(ctx context.Context, supervisor sentinel.Supervisor) (uuid.UUID, error) {
	// Try to find an existing supervisor with the same values
	if existingSupervisor, err := s.GetSupervisorFromValues(
		ctx, supervisor.Code, supervisor.Name, supervisor.Description, supervisor.Type, supervisor.Attributes,
	); err != nil {
		return uuid.UUID{}, fmt.Errorf("error getting existing supervisor during create supervisor: %w", err)
	} else if existingSupervisor != nil {
		return *existingSupervisor.Id, nil
	}

	id := uuid.New()

	attributes, err := json.Marshal(supervisor.Attributes)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error marshalling supervisor attributes: %w", err)
	}

	query := `
		INSERT INTO supervisor (id, description, name, created_at, type, code, attributes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = s.db.ExecContext(ctx, query, id, supervisor.Description, supervisor.Name, supervisor.CreatedAt, supervisor.Type, supervisor.Code, attributes)
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

func (s *PostgresqlStore) CreateTool(
	ctx context.Context,
	runId uuid.UUID,
	attributes map[string]interface{},
	name string,
	description string,
	ignoredAttributes []string,
) (uuid.UUID, error) {
	// Convert attributes to JSON if it's not already
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error marshaling tool attributes: %w", err)
	}

	if ignoredAttributes == nil {
		ignoredAttributes = []string{}
	}

	id := uuid.New()
	query := `
		INSERT INTO tool (id, run_id, name, description, attributes, ignored_attributes)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = s.db.ExecContext(ctx, query,
		id,
		runId,
		name,
		description,
		attributesJSON, // Use the JSON-encoded attributes
		pq.Array(ignoredAttributes),
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error creating tool: %w", err)
	}

	return id, nil
}

func (s *PostgresqlStore) GetRun(ctx context.Context, id uuid.UUID) (*sentinel.Run, error) {
	query := `
		SELECT id, project_id, created_at
		FROM run
		WHERE id = $1`

	var run sentinel.Run
	err := s.db.QueryRowContext(ctx, query, id).Scan(&run.Id, &run.ProjectId, &run.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting run: %w", err)
	}

	return &run, nil
}

func (s *PostgresqlStore) GetRunTools(ctx context.Context, runId uuid.UUID) ([]sentinel.Tool, error) {
	query := `
		SELECT tool.id, tool.run_id, tool.name, tool.description, tool.attributes, COALESCE(tool.ignored_attributes, '{}') as ignored_attributes
		FROM tool 
		WHERE run_id = $1`

	rows, err := s.db.QueryContext(ctx, query, runId)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error getting run tools: %w", err)
	}
	defer rows.Close()

	tools := make([]sentinel.Tool, 0)
	for rows.Next() {
		var tool sentinel.Tool
		var attributesJSON []byte
		var ignoredAttributes []string

		if err := rows.Scan(
			&tool.Id,
			&tool.RunId,
			&tool.Name,
			&tool.Description,
			&attributesJSON,
			pq.Array(&ignoredAttributes),
		); err != nil {
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
		tool.IgnoredAttributes = &ignoredAttributes

		tools = append(tools, tool)
	}

	return tools, nil
}

func (s *PostgresqlStore) GetSupervisor(ctx context.Context, id uuid.UUID) (*sentinel.Supervisor, error) {
	query := `
		SELECT id, description, name, created_at, type, attributes
		FROM supervisor
		WHERE id = $1`

	var supervisor sentinel.Supervisor
	var attributesJSON []byte
	err := s.db.QueryRowContext(ctx, query, id).Scan(&supervisor.Id, &supervisor.Description, &supervisor.Name, &supervisor.CreatedAt, &supervisor.Type, &attributesJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting supervisor: %w", err)
	}

	// Parse the JSON attributes if they exist
	if len(attributesJSON) > 0 {
		if err := json.Unmarshal(attributesJSON, &supervisor.Attributes); err != nil {
			return nil, fmt.Errorf("error parsing supervisor attributes: %w", err)
		}
	}

	return &supervisor, nil
}

func (s *PostgresqlStore) GetSupervisors(ctx context.Context, projectId uuid.UUID) ([]sentinel.Supervisor, error) {
	query := `
		SELECT s.id, s.description, s.name, s.code, s.created_at, s.type, s.attributes
		FROM supervisor s 
		INNER JOIN chain_supervisor cs ON s.id = cs.supervisor_id
		INNER JOIN chain c ON cs.chain_id = c.id
		INNER JOIN chain_tool ct ON c.id = ct.chain_id
		INNER JOIN tool t ON ct.tool_id = t.id
		INNER JOIN run r ON t.run_id = r.id
		WHERE r.project_id = $1`

	rows, err := s.db.QueryContext(ctx, query, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting supervisors: %w", err)
	}
	defer rows.Close()

	supervisors := make([]sentinel.Supervisor, 0)
	for rows.Next() {
		var supervisor sentinel.Supervisor
		var attributesJSON []byte
		if err := rows.Scan(
			&supervisor.Id,
			&supervisor.Description,
			&supervisor.Name,
			&supervisor.Code,
			&supervisor.CreatedAt,
			&supervisor.Type,
			&attributesJSON,
		); err != nil {
			return nil, fmt.Errorf("error scanning supervisor: %w", err)
		}

		// Parse the JSON attributes if they exist
		if len(attributesJSON) > 0 {
			if err := json.Unmarshal(attributesJSON, &supervisor.Attributes); err != nil {
				return nil, fmt.Errorf("error parsing supervisor attributes: %w", err)
			}
		}

		supervisors = append(supervisors, supervisor)
	}

	return supervisors, nil
}

func (s *PostgresqlStore) GetSupervisionStatusesForRequest(ctx context.Context, requestId uuid.UUID) ([]sentinel.SupervisionStatus, error) {
	query := `
        SELECT ss.id, ss.supervisionrequest_id, ss.status, ss.created_at
        FROM supervisionrequest_status ss
        WHERE ss.supervisionrequest_id = $1`

	rows, err := s.db.QueryContext(ctx, query, requestId)
	if err != nil {
		return nil, fmt.Errorf("error getting supervision statuses for request %s: %w", requestId, err)
	}
	defer rows.Close()

	var statuses []sentinel.SupervisionStatus
	for rows.Next() {
		var status sentinel.SupervisionStatus
		if err := rows.Scan(&status.Id, &status.SupervisionRequestId, &status.Status, &status.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning supervision status: %w", err)
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

func (s *PostgresqlStore) GetSupervisionResultsForChainExecution(ctx context.Context, executionId uuid.UUID) ([]sentinel.SupervisionResult, error) {
	query := `
        SELECT sr.id, sr.supervisionrequest_id, sr.created_at, sr.decision, sr.reasoning
        FROM supervisionresult sr
        INNER JOIN supervisionrequest sreq ON sr.supervisionrequest_id = sreq.id
        WHERE sreq.chainexecution_id = $1`

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

func (s *PostgresqlStore) GetSupervisionStatusesForChainExecution(ctx context.Context, executionId uuid.UUID) ([]sentinel.SupervisionStatus, error) {
	query := `
        SELECT ss.id, ss.supervisionrequest_id, ss.status, ss.created_at
        FROM supervisionrequest_status ss
        INNER JOIN supervisionrequest sr ON ss.supervisionrequest_id = sr.id
        WHERE sr.chainexecution_id = $1`

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
