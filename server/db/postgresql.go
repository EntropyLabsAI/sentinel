package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	asteroid "github.com/asteroidai/asteroid/server"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type PostgresqlStore struct {
	db *sql.DB
}

// Check if PostgresqlStore implements asteroid.Store
var _ asteroid.Store = &PostgresqlStore{}

// NewPostgresqlStore creates a new PostgreSQL store
func NewPostgresqlStore() (*PostgresqlStore, error) {
	var db *sql.DB
	var err error

	if os.Getenv("ENVIRONMENT") == "production" {
		// Production: Use Cloud SQL connector
		db, err = connectWithCloudSQL()
		if err != nil {
			return nil, fmt.Errorf("error connecting to Cloud SQL: %w", err)
		}
	} else {
		// Development: Use direct connection
		connStr := os.Getenv("DATABASE_URL")
		if connStr == "" {
			return nil, fmt.Errorf("DATABASE_URL is not set")
		}

		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("error opening database: %w", err)
		}
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return &PostgresqlStore{db: db}, nil
}

// connectWithCloudSQL handles Cloud SQL connection in production
func connectWithCloudSQL() (*sql.DB, error) {
	mustGetenv := func(k string) string {
		v := os.Getenv(k)
		if v == "" {
			log.Fatalf("Fatal Error: %s environment variable not set.\n", k)
		}
		return v
	}

	var (
		dbUser                 = mustGetenv("DB_USER")
		dbPwd                  = mustGetenv("DB_PASS")
		dbName                 = mustGetenv("DB_NAME")
		instanceConnectionName = mustGetenv("INSTANCE_CONNECTION_NAME")
		usePrivate             = os.Getenv("PRIVATE_IP")
	)

	dsn := fmt.Sprintf("user=%s password=%s database=%s", dbUser, dbPwd, dbName)
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	var opts []cloudsqlconn.Option
	if usePrivate != "" {
		opts = append(opts, cloudsqlconn.WithDefaultDialOptions(cloudsqlconn.WithPrivateIP()))
	}

	d, err := cloudsqlconn.NewDialer(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	config.DialFunc = func(ctx context.Context, network, instance string) (net.Conn, error) {
		return d.Dial(ctx, instanceConnectionName)
	}

	dbURI := stdlib.RegisterConnConfig(config)
	dbPool, err := sql.Open("pgx", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	return dbPool, nil
}

// Close closes the database connection
func (s *PostgresqlStore) Close() error {
	return s.db.Close()
}

// ProjectStore implementation
func (s *PostgresqlStore) CreateProject(ctx context.Context, project asteroid.Project) error {
	query := `
		INSERT INTO project (id, name, created_at, run_result_tags)
		VALUES ($1, $2, $3, $4)`

	_, err := s.db.ExecContext(ctx, query, project.Id, project.Name, project.CreatedAt, pq.Array(project.RunResultTags))
	if err != nil {
		return fmt.Errorf("error creating project: %w", err)
	}

	return nil
}

func (s *PostgresqlStore) GetProject(ctx context.Context, id uuid.UUID) (*asteroid.Project, error) {
	query := `
		SELECT id, name, created_at, run_result_tags
		FROM project
		WHERE id = $1`

	var project asteroid.Project
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&project.Id,
		&project.Name,
		&project.CreatedAt,
		pq.Array(&project.RunResultTags),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting project: %w", err)
	}

	return &project, nil
}

func (s *PostgresqlStore) GetProjectFromName(ctx context.Context, name string) (*asteroid.Project, error) {
	query := `
		SELECT id, name, created_at, run_result_tags
		FROM project
		WHERE name = $1`

	var project asteroid.Project
	err := s.db.QueryRowContext(ctx, query, name).Scan(
		&project.Id,
		&project.Name,
		&project.CreatedAt,
		pq.Array(&project.RunResultTags),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting project: %w", err)
	}

	return &project, nil
}

func (s *PostgresqlStore) GetToolFromNameAndRunId(ctx context.Context, name string, runId uuid.UUID) (*asteroid.Tool, error) {
	query := `
		SELECT id, name, description, attributes, ignored_attributes, code
		FROM tool
		WHERE name = $1
		AND run_id = $2`

	var tool asteroid.Tool
	var attributesJSON []byte
	var toolIgnoredAttributes []string
	err := s.db.QueryRowContext(ctx, query, name, runId).Scan(
		&tool.Id,
		&tool.Name,
		&tool.Description,
		&attributesJSON,
		pq.Array(&toolIgnoredAttributes),
		&tool.Code,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting tool from name: %w", err)
	}

	// Parse the JSON attributes if they exist
	if len(attributesJSON) > 0 {
		var attrs map[string]interface{}
		if err := json.Unmarshal(attributesJSON, &attrs); err != nil {
			return nil, fmt.Errorf("error parsing tool attributes: %w", err)
		}
		tool.Attributes = attrs
	}

	tool.IgnoredAttributes = &toolIgnoredAttributes

	return &tool, nil
}

func (s *PostgresqlStore) GetToolFromValues(ctx context.Context, attributes map[string]interface{}, name string, description string, ignoredAttributes []string, code string) (*asteroid.Tool, error) {
	query := `
		SELECT id, name, description, attributes, ignored_attributes, code
		FROM tool
		WHERE name = $1
		AND description = $2
		AND attributes = $3
		AND ignored_attributes = $4
		AND code = $5`
	attrJSON, err := json.Marshal(attributes)
	if err != nil {
		return nil, fmt.Errorf("error marshalling attributes: %w", err)
	}

	var tool asteroid.Tool
	var attributesJSON []byte
	var toolIgnoredAttributes []string
	err = s.db.QueryRowContext(ctx, query, name, description, attrJSON, pq.Array(ignoredAttributes)).Scan(
		&tool.Id,
		&tool.Name,
		&tool.Description,
		&attributesJSON,
		pq.Array(&toolIgnoredAttributes),
		&tool.Code,
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
		tool.Attributes = attrs
	}

	tool.IgnoredAttributes = &toolIgnoredAttributes
	return &tool, nil
}

func (s *PostgresqlStore) GetSupervisorFromValues(
	ctx context.Context,
	code string,
	name string,
	desc string,
	t asteroid.SupervisorType,
	attributes map[string]interface{},
) (*asteroid.Supervisor, error) {
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

	var supervisor asteroid.Supervisor
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

func (s *PostgresqlStore) GetProjects(ctx context.Context) ([]asteroid.Project, error) {
	query := `
		SELECT id, name, created_at, run_result_tags
		FROM project
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error listing projects: %w", err)
	}
	defer rows.Close()

	projects := make([]asteroid.Project, 0)
	for rows.Next() {
		var project asteroid.Project
		if err := rows.Scan(
			&project.Id,
			&project.Name,
			&project.CreatedAt,
			pq.Array(&project.RunResultTags),
		); err != nil {
			return nil, fmt.Errorf("error scanning project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func (s *PostgresqlStore) GetRuns(ctx context.Context, taskId uuid.UUID) ([]asteroid.Run, error) {
	query := `
		SELECT id, task_id, created_at, status, result
		FROM run
		WHERE task_id = $1`

	rows, err := s.db.QueryContext(ctx, query, taskId)
	if err != nil {
		return nil, fmt.Errorf("error getting runs: %w", err)
	}
	defer rows.Close()

	var runs []asteroid.Run
	for rows.Next() {
		var run asteroid.Run
		if err := rows.Scan(&run.Id, &run.TaskId, &run.CreatedAt, &run.Status, &run.Result); err != nil {
			return nil, fmt.Errorf("error scanning run: %w", err)
		}
		runs = append(runs, run)
	}

	// If no rows were found, runs will be empty slice
	return runs, nil
}

func (s *PostgresqlStore) GetTaskRuns(ctx context.Context, taskId uuid.UUID) ([]asteroid.Run, error) {
	query := `
		SELECT id, task_id, created_at, status, result
		FROM run
		WHERE task_id = $1`

	rows, err := s.db.QueryContext(ctx, query, taskId)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error getting task runs: %w", err)
	}
	defer rows.Close()

	runs := make([]asteroid.Run, 0)
	for rows.Next() {
		var run asteroid.Run
		if err := rows.Scan(&run.Id, &run.TaskId, &run.CreatedAt, &run.Status, &run.Result); err != nil {
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

func (s *PostgresqlStore) CreateSupervisorChain(ctx context.Context, toolId uuid.UUID, chain asteroid.ChainRequest) (*uuid.UUID, error) {
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

func (s *PostgresqlStore) GetSupervisorChain(ctx context.Context, chainId uuid.UUID) (*asteroid.SupervisorChain, error) {
	// Order by the position column in chain_supervisor table
	query := `
		SELECT s.id, s.name, s.description, s.type, s.attributes, s.created_at, s.code
		FROM chain_supervisor cs
		INNER JOIN supervisor s ON cs.supervisor_id = s.id
		WHERE cs.chain_id = $1
		ORDER BY cs.position_in_chain ASC`

	rows, err := s.db.QueryContext(ctx, query, chainId)
	if err != nil {
		return nil, fmt.Errorf("error getting tool supervisor chain: %w", err)
	}
	defer rows.Close()

	supervisors := make([]asteroid.Supervisor, 0)
	for rows.Next() {
		// Parse out the attributes bytes into json
		var attributesJSON []byte
		var supervisor asteroid.Supervisor
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

		// Parse the JSON attributes if they exist
		var attributes map[string]interface{}
		if len(attributesJSON) > 0 {
			if err := json.Unmarshal(attributesJSON, &attributes); err != nil {
				return nil, fmt.Errorf("error parsing supervisor attributes: %w", err)
			}
			supervisor.Attributes = attributes
		}

		supervisors = append(supervisors, supervisor)
	}

	return &asteroid.SupervisorChain{
		ChainId:     chainId,
		Supervisors: supervisors,
	}, nil
}

func (s *PostgresqlStore) GetSupervisorChains(ctx context.Context, toolId uuid.UUID) ([]asteroid.SupervisorChain, error) {
	chainIds, err := s.getChainsForTool(ctx, toolId)
	if err != nil {
		return nil, fmt.Errorf("error getting chains for tool: %w", err)
	}

	chains := make([]asteroid.SupervisorChain, 0)
	for _, chainId := range chainIds {
		chain, err := s.GetSupervisorChain(ctx, chainId)
		if err != nil {
			return nil, fmt.Errorf("error getting tool supervisor chain: %w", err)
		}
		chains = append(chains, *chain)
	}

	return chains, nil
}

func (s *PostgresqlStore) GetToolCallFromCallId(ctx context.Context, id string) (*asteroid.AsteroidToolCall, error) {
	query := `
		SELECT id, call_id, created_at, tool_id, tool_call_data
		FROM toolcall
		WHERE call_id = $1`

	var toolCall asteroid.AsteroidToolCall
	var toolCallDataJSON []byte
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&toolCall.Id,
		&toolCall.CallId,
		&toolCall.CreatedAt,
		&toolCall.ToolId,
		&toolCallDataJSON,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting tool request: %w", err)
	}

	args := string(toolCallDataJSON)
	toolCall.Arguments = &args
	toolCall.CallId = &id

	return &toolCall, nil
}

func (s *PostgresqlStore) GetToolCall(ctx context.Context, id uuid.UUID) (*asteroid.AsteroidToolCall, error) {
	query := `
		SELECT id, call_id, created_at, tool_id, tool_call_data
		FROM toolcall
		WHERE id = $1`

	var toolCall asteroid.AsteroidToolCall
	var toolCallDataJSON []byte
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&toolCall.Id,
		&toolCall.CallId,
		&toolCall.CreatedAt,
		&toolCall.ToolId,
		&toolCallDataJSON,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting tool request: %w", err)
	}

	args := string(toolCallDataJSON)
	toolCall.Arguments = &args

	return &toolCall, nil
}

func (s *PostgresqlStore) GetChainExecutionsFromToolCall(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	query := `
			SELECT id FROM chainexecution WHERE toolcall_id = $1`

	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("error getting chain executions from tool call ID: %w", err)
	}
	defer rows.Close()

	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(
			&id,
		); err != nil {
			return nil, fmt.Errorf("error scanning chain execution ID: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (s *PostgresqlStore) GetExecutionFromChainId(ctx context.Context, chainId uuid.UUID) (*uuid.UUID, error) {
	query := `
		SELECT id
		FROM chainexecution
		WHERE chain_id = $1`

	var id uuid.UUID
	err := s.db.QueryRowContext(ctx, query, chainId).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting execution from chain ID: %w", err)
	}

	return &id, nil
}

func (s *PostgresqlStore) createChainExecutions(ctx context.Context, tx *sql.Tx, toolCallId uuid.UUID, chainIds []uuid.UUID) error {
	for _, chainId := range chainIds {
		_, err := s.createChainExecution(ctx, chainId, toolCallId, tx)
		if err != nil {
			return fmt.Errorf("error creating chain execution: %w", err)
		}
	}

	return nil
}

func (s *PostgresqlStore) createChainExecution(
	ctx context.Context,
	chainId uuid.UUID,
	toolCallId uuid.UUID,
	tx *sql.Tx,
) (*uuid.UUID, error) {
	query := `
		INSERT INTO chainexecution (id, chain_id, toolcall_id)
		VALUES ($1, $2, $3)`

	id := uuid.New()
	_, err := tx.ExecContext(ctx, query, id, chainId, toolCallId)
	if err != nil {
		return nil, fmt.Errorf("error creating chain execution: %w", err)
	}

	return &id, nil
}

func (s *PostgresqlStore) CreateSupervisionRequest(
	ctx context.Context,
	request asteroid.SupervisionRequest,
	chainId uuid.UUID,
	toolCallId uuid.UUID,
) (*uuid.UUID, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Sanity check that we're recording this against a valid chain execution group that already exists
	if request.ChainexecutionId == nil && request.PositionInChain == 0 {
		// These chain executions should be initialised when the tool call is created
		ceId, err := s.getChainExecutionForToolCall(ctx, chainId, toolCallId, tx)
		if err != nil {
			return nil, fmt.Errorf("error creating chain execution: %w", err)
		}

		if ceId == nil {
			return nil, fmt.Errorf("chain execution not found for tool call %s and chain %s", toolCallId, chainId)
		}

		request.ChainexecutionId = ceId
	} else if request.ChainexecutionId == nil && request.PositionInChain > 0 {
		return nil, fmt.Errorf("chain execution ID is required when creating a supervision request for a non-zero position in the chain")
	}

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

	status := asteroid.SupervisionStatus{
		Status:    asteroid.Pending,
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

func (s *PostgresqlStore) getChainExecutionForToolCall(ctx context.Context, chainId uuid.UUID, toolCallId uuid.UUID, tx *sql.Tx) (*uuid.UUID, error) {
	query := `
		SELECT id 
		FROM chainexecution 
		WHERE chain_id = $1 AND toolcall_id = $2`

	var id uuid.UUID
	err := tx.QueryRowContext(ctx, query, chainId, toolCallId).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("error getting chain execution for tool call: %w", err)
	}

	return &id, nil
}

func (s *PostgresqlStore) CreateSupervisionStatus(ctx context.Context, requestID uuid.UUID, status asteroid.SupervisionStatus) error {
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

func (s *PostgresqlStore) createSupervisionStatus(ctx context.Context, requestID uuid.UUID, status asteroid.SupervisionStatus, tx *sql.Tx) error {
	query := `
		INSERT INTO supervisionrequest_status (supervisionrequest_id, status, created_at)
		VALUES ($1, $2, $3)`

	_, err := tx.ExecContext(ctx, query, requestID, status.Status, status.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creating supervisor status: %w", err)
	}

	return nil
}

func (s *PostgresqlStore) GetTool(ctx context.Context, id uuid.UUID) (*asteroid.Tool, error) {
	query := `
		SELECT id, run_id, name, description, attributes, ignored_attributes, code
		FROM tool
		WHERE id = $1`

	var tool asteroid.Tool
	var attributesJSON []byte
	var ignoredAttributes []string
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&tool.Id,
		&tool.RunId,
		&tool.Name,
		&tool.Description,
		&attributesJSON,
		pq.Array(&ignoredAttributes),
		&tool.Code,
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
		tool.Attributes = attributes
	}

	tool.IgnoredAttributes = &ignoredAttributes
	return &tool, nil
}

func (s *PostgresqlStore) GetProjectTools(ctx context.Context, projectId uuid.UUID) ([]asteroid.Tool, error) {
	query := `
		SELECT DISTINCT r.id
		FROM run r
		INNER JOIN task t ON t.id = r.task_id
		WHERE t.project_id = $1`

	rows, err := s.db.QueryContext(ctx, query, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project tools: %w", err)
	}
	defer rows.Close()

	tools := make([]asteroid.Tool, 0)
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

func (s *PostgresqlStore) GetProjectTasks(ctx context.Context, projectId uuid.UUID) ([]asteroid.Task, error) {
	query := `
		SELECT id, project_id, name, description, created_at
		FROM task
		WHERE project_id = $1`

	rows, err := s.db.QueryContext(ctx, query, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]asteroid.Task, 0)
	for rows.Next() {
		var task asteroid.Task
		if err := rows.Scan(&task.Id, &task.ProjectId, &task.Name, &task.Description, &task.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning task: %w", err)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (s *PostgresqlStore) CountSupervisionRequests(ctx context.Context, status asteroid.Status) (int, error) {
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

func (s *PostgresqlStore) CreateSupervisionResult(ctx context.Context, result asteroid.SupervisionResult, requestId uuid.UUID) (*uuid.UUID, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		INSERT INTO supervisionresult (id, supervisionrequest_id, created_at, decision, reasoning, toolcall_id)
		VALUES ($1, $2, $3, $4, $5, $6)`

	id := uuid.New()
	_, err = tx.ExecContext(
		ctx,
		query,
		id,
		requestId,
		result.CreatedAt,
		result.Decision,
		result.Reasoning,
		result.ToolcallId,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating supervision result: %w", err)
	}

	// Create a supervisionrequest_status
	err = s.createSupervisionStatus(ctx, requestId, asteroid.SupervisionStatus{
		Status:    asteroid.Completed,
		CreatedAt: result.CreatedAt,
	}, tx)
	if err != nil {
		return nil, fmt.Errorf("error creating supervision status for result: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return &id, nil
}

func (s *PostgresqlStore) GetSupervisionRequestsForStatus(ctx context.Context, status asteroid.Status) ([]asteroid.SupervisionRequest, error) {
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
	rows, err := s.db.QueryContext(ctx, query, asteroid.ClientSupervisor, status)
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
	requests := make([]asteroid.SupervisionRequest, 0, len(requestIds))
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

func (s *PostgresqlStore) GetSupervisionResultFromRequestID(ctx context.Context, requestId uuid.UUID) (*asteroid.SupervisionResult, error) {
	query := `
		SELECT id, supervisionrequest_id, created_at, decision, reasoning, toolcall_id
		FROM supervisionresult
		WHERE supervisionrequest_id = $1`

	var result asteroid.SupervisionResult
	err := s.db.QueryRowContext(ctx, query, requestId).Scan(
		&result.Id,
		&result.SupervisionRequestId,
		&result.CreatedAt,
		&result.Decision,
		&result.Reasoning,
		&result.ToolcallId,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting supervision result: %w", err)
	}

	return &result, nil
}

func (s *PostgresqlStore) GetSupervisionRequest(ctx context.Context, id uuid.UUID) (*asteroid.SupervisionRequest, error) {
	query := `
		SELECT id, supervisor_id, position_in_chain, chainexecution_id
		FROM supervisionrequest
		WHERE id = $1`

	var request asteroid.SupervisionRequest
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

	// Get the latest status for the request
	status, err := s.GetSupervisionRequestStatus(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting supervision request status: %w", err)
	}
	request.Status = status

	return &request, nil
}

func (s *PostgresqlStore) CreateSupervisor(ctx context.Context, supervisor asteroid.Supervisor) (uuid.UUID, error) {
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

func (s *PostgresqlStore) CreateTask(ctx context.Context, task asteroid.Task) (*uuid.UUID, error) {
	// First check if a task with the same values already exists
	query := `
		SELECT id 
		FROM task 
		WHERE project_id = $1 
		AND name = $2 
		AND description = $3`

	var existingId uuid.UUID
	err := s.db.QueryRowContext(ctx, query, task.ProjectId, task.Name, task.Description).Scan(&existingId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("error checking for existing task: %w", err)
	}
	if err == nil {
		// Task already exists, return its ID
		return &existingId, nil
	}

	// No existing task found, create a new one
	id := uuid.New()
	query = `
		INSERT INTO task (id, project_id, name, description, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err = s.db.ExecContext(ctx, query, id, task.ProjectId, task.Name, task.Description, task.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("error creating task: %w", err)
	}

	return &id, nil
}

func (s *PostgresqlStore) GetTask(ctx context.Context, id uuid.UUID) (*asteroid.Task, error) {
	query := `
		SELECT id, project_id, name, description, created_at
		FROM task
		WHERE id = $1`

	var task asteroid.Task
	err := s.db.QueryRowContext(ctx, query, id).Scan(&task.Id, &task.ProjectId, &task.Name, &task.Description, &task.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting task: %w", err)
	}
	return &task, nil
}

func (s *PostgresqlStore) CreateRun(ctx context.Context, run asteroid.Run) (uuid.UUID, error) {
	// First check if the task exists
	p, err := s.GetTask(ctx, run.TaskId)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error getting task: %w", err)
	}
	if p == nil {
		return uuid.UUID{}, fmt.Errorf("task not found: %s", run.TaskId)
	}

	id := uuid.New()

	query := `
		INSERT INTO run (id, task_id, created_at, status)
		VALUES ($1, $2, $3, $4)`

	_, err = s.db.ExecContext(ctx, query, id, run.TaskId, run.CreatedAt, asteroid.Pending)
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
	code string,
) (*asteroid.Tool, error) {
	// Convert attributes to JSON if it's not already
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return nil, fmt.Errorf("error marshaling tool attributes: %w", err)
	}

	if ignoredAttributes == nil {
		ignoredAttributes = []string{}
	}

	id := uuid.New()
	query := `
		INSERT INTO tool (id, run_id, name, description, attributes, ignored_attributes, code)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = s.db.ExecContext(ctx, query,
		id,
		runId,
		name,
		description,
		attributesJSON, // Use the JSON-encoded attributes
		pq.Array(ignoredAttributes),
		code,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating tool: %w", err)
	}

	tool := asteroid.Tool{
		Id:                &id,
		RunId:             runId,
		Name:              name,
		Description:       description,
		Attributes:        attributes,
		IgnoredAttributes: &ignoredAttributes,
		Code:              code,
	}

	return &tool, nil
}

func (s *PostgresqlStore) GetRun(ctx context.Context, id uuid.UUID) (*asteroid.Run, error) {
	query := `
		SELECT id, task_id, created_at, status, result
		FROM run
		WHERE id = $1`

	var run asteroid.Run
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&run.Id,
		&run.TaskId,
		&run.CreatedAt,
		&run.Status,
		&run.Result,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting run: %w", err)
	}

	return &run, nil
}

func (s *PostgresqlStore) GetRunTools(ctx context.Context, runId uuid.UUID) ([]asteroid.Tool, error) {
	query := `
		SELECT tool.id, tool.run_id, tool.name, tool.description, tool.attributes, COALESCE(tool.ignored_attributes, '{}') as ignored_attributes, tool.code
		FROM tool 
		WHERE run_id = $1`

	rows, err := s.db.QueryContext(ctx, query, runId)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error getting run tools: %w", err)
	}
	defer rows.Close()

	tools := make([]asteroid.Tool, 0)
	for rows.Next() {
		var tool asteroid.Tool
		var attributesJSON []byte
		var ignoredAttributes []string

		if err := rows.Scan(
			&tool.Id,
			&tool.RunId,
			&tool.Name,
			&tool.Description,
			&attributesJSON,
			pq.Array(&ignoredAttributes),
			&tool.Code,
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
		tool.Attributes = t
		tool.IgnoredAttributes = &ignoredAttributes

		tools = append(tools, tool)
	}

	return tools, nil
}

func (s *PostgresqlStore) GetSupervisor(ctx context.Context, id uuid.UUID) (*asteroid.Supervisor, error) {
	query := `
		SELECT id, description, name, created_at, type, attributes
		FROM supervisor
		WHERE id = $1`

	var supervisor asteroid.Supervisor
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

func (s *PostgresqlStore) GetSupervisors(ctx context.Context, projectId uuid.UUID) ([]asteroid.Supervisor, error) {
	query := `
		SELECT s.id, s.description, s.name, s.code, s.created_at, s.type, s.attributes
		FROM supervisor s 
		INNER JOIN chain_supervisor cs ON s.id = cs.supervisor_id
		INNER JOIN chain c ON cs.chain_id = c.id
		INNER JOIN chain_tool ct ON c.id = ct.chain_id
		INNER JOIN tool t ON ct.tool_id = t.id
		INNER JOIN run r ON t.run_id = r.id
		WHERE r.task_id IN (
			SELECT id 
			FROM task 
			WHERE project_id = $1
		)`

	rows, err := s.db.QueryContext(ctx, query, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting supervisors: %w", err)
	}
	defer rows.Close()

	supervisors := make([]asteroid.Supervisor, 0)
	for rows.Next() {
		var supervisor asteroid.Supervisor
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

func (s *PostgresqlStore) GetSupervisionStatusesForRequest(ctx context.Context, requestId uuid.UUID) ([]asteroid.SupervisionStatus, error) {
	query := `
        SELECT ss.id, ss.supervisionrequest_id, ss.status, ss.created_at
        FROM supervisionrequest_status ss
        WHERE ss.supervisionrequest_id = $1`

	rows, err := s.db.QueryContext(ctx, query, requestId)
	if err != nil {
		return nil, fmt.Errorf("error getting supervision statuses for request %s: %w", requestId, err)
	}
	defer rows.Close()

	var statuses []asteroid.SupervisionStatus
	for rows.Next() {
		var status asteroid.SupervisionStatus
		if err := rows.Scan(&status.Id, &status.SupervisionRequestId, &status.Status, &status.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning supervision status: %w", err)
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

func (s *PostgresqlStore) GetSupervisionResultsForChainExecution(ctx context.Context, executionId uuid.UUID) ([]asteroid.SupervisionResult, error) {
	query := `
        SELECT sr.id, sr.supervisionrequest_id, sr.created_at, sr.decision, sr.reasoning, sr.toolcall_id
        FROM supervisionresult sr
        INNER JOIN supervisionrequest sreq ON sr.supervisionrequest_id = sreq.id
        WHERE sreq.chainexecution_id = $1`

	rows, err := s.db.QueryContext(ctx, query, executionId)
	if err != nil {
		return nil, fmt.Errorf("error getting supervision results: %w", err)
	}
	defer rows.Close()

	var results []asteroid.SupervisionResult
	for rows.Next() {
		var result asteroid.SupervisionResult
		err := rows.Scan(
			&result.Id,
			&result.SupervisionRequestId,
			&result.CreatedAt,
			&result.Decision,
			&result.Reasoning,
			&result.ToolcallId,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning supervision result: %w", err)
		}

		results = append(results, result)
	}

	return results, nil
}

func (s *PostgresqlStore) GetSupervisionStatusesForChainExecution(ctx context.Context, executionId uuid.UUID) ([]asteroid.SupervisionStatus, error) {
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

	var statuses []asteroid.SupervisionStatus
	for rows.Next() {
		var status asteroid.SupervisionStatus
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

// GetChainExecutionSupervisionRequests gets all supervision requests for a specific chain execution
func (s *PostgresqlStore) GetChainExecutionSupervisionRequests(ctx context.Context, chainExecutionId uuid.UUID) ([]asteroid.SupervisionRequest, error) {
	query := `
        SELECT sr.id, sr.supervisor_id, sr.chainexecution_id, position_in_chain
        FROM supervisionrequest sr
        JOIN chainexecution ce ON sr.chainexecution_id = ce.id
        WHERE ce.id = $1
        ORDER BY sr.id ASC`

	rows, err := s.db.QueryContext(ctx, query, chainExecutionId)
	if err != nil {
		return nil, fmt.Errorf("error getting chain execution supervision requests: %w", err)
	}
	defer rows.Close()

	requests := make([]asteroid.SupervisionRequest, 0)
	for rows.Next() {
		var request asteroid.SupervisionRequest
		if err := rows.Scan(
			&request.Id,
			&request.SupervisorId,
			&request.ChainexecutionId,
			&request.PositionInChain,
		); err != nil {
			return nil, fmt.Errorf("error scanning supervision request: %w", err)
		}

		status, err := s.GetSupervisionRequestStatus(ctx, *request.Id)
		if err != nil {
			return nil, fmt.Errorf("error trying to get request status during chain execution query: %w", err)
		}

		if status != nil {
			request.Status = status
		}

		requests = append(requests, request)
	}

	return requests, nil
}

// GetSupervisionRequestStatus gets the latest status for a supervision request
func (s *PostgresqlStore) GetSupervisionRequestStatus(ctx context.Context, requestId uuid.UUID) (*asteroid.SupervisionStatus, error) {
	query := `
        SELECT ss.id, ss.supervisionrequest_id, ss.status, ss.created_at
        FROM supervisionrequest_status ss
        WHERE ss.supervisionrequest_id = $1
        ORDER BY ss.created_at DESC
        LIMIT 1`

	var status asteroid.SupervisionStatus
	err := s.db.QueryRowContext(ctx, query, requestId).Scan(
		&status.Id,
		&status.SupervisionRequestId,
		&status.Status,
		&status.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting supervision request status: %w", err)
	}

	return &status, nil
}

func (s *PostgresqlStore) GetChainExecution(ctx context.Context, executionId uuid.UUID) (*uuid.UUID, *uuid.UUID, error) {
	query := `SELECT chain_id, toolcall_id FROM chainexecution WHERE id = $1`

	var chainId, toolCallId uuid.UUID
	err := s.db.QueryRowContext(ctx, query, executionId).Scan(&chainId, &toolCallId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("error getting chain ID from execution ID: %w", err)
	}

	return &chainId, &toolCallId, nil
}

// GetChainExecutionState returns the chain state for a given chain execution ID
func (s *PostgresqlStore) GetChainExecutionState(ctx context.Context, executionId uuid.UUID) (*asteroid.ChainExecutionState, error) {
	// First, get the chain execution record
	var chainExecution asteroid.ChainExecution
	err := s.db.QueryRowContext(ctx, `
        SELECT id, toolcall_id, chain_id, created_at
        FROM chainexecution
        WHERE id = $1
				ORDER BY id ASC
    `, executionId).Scan(
		&chainExecution.Id,
		&chainExecution.ToolcallId,
		&chainExecution.ChainId,
		&chainExecution.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get chain execution: %w", err)
	}

	// Get the supervisor chain
	supervisorChain, err := s.GetSupervisorChain(ctx, chainExecution.ChainId)
	if err != nil {
		return nil, fmt.Errorf("failed to get supervisor chain: %w", err)
	}

	// Get all supervision requests for this chain execution
	supervisionRequests, err := s.GetChainExecutionSupervisionRequests(ctx, chainExecution.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get supervision requests: %w", err)
	}

	// For each supervision request, get the latest status and result
	var supervisionRequestStates []asteroid.SupervisionRequestState
	for _, request := range supervisionRequests {
		// Get the latest status
		var status asteroid.SupervisionStatus
		err := s.db.QueryRowContext(ctx, `
            SELECT id, supervisionrequest_id, created_at, status FROM supervisionrequest_status WHERE supervisionrequest_id = $1 ORDER BY created_at DESC LIMIT 1
        `, request.Id).Scan(
			&status.Id,
			&status.SupervisionRequestId,
			&status.CreatedAt,
			&status.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get supervision request status: %w", err)
		}

		// Get the result, if any
		result := &asteroid.SupervisionResult{}
		err = s.db.QueryRowContext(ctx, `
            SELECT id, supervisionrequest_id, created_at, decision, reasoning, toolcall_id
            FROM supervisionresult
            WHERE supervisionrequest_id = $1
        `, request.Id).Scan(
			&result.Id,
			&result.SupervisionRequestId,
			&result.CreatedAt,
			&result.Decision,
			&result.Reasoning,
			&result.ToolcallId,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				result = nil // No result yet
			} else {
				return nil, fmt.Errorf("failed to get supervision result: %w", err)
			}
		}

		supervisionRequestState := asteroid.SupervisionRequestState{
			SupervisionRequest: request,
			Status:             status,
			Result:             result,
		}
		supervisionRequestStates = append(supervisionRequestStates, supervisionRequestState)
	}

	// Build and return the ChainExecutionState
	state := asteroid.ChainExecutionState{
		Chain:               *supervisorChain,
		ChainExecution:      chainExecution,
		SupervisionRequests: supervisionRequestStates,
	}

	return &state, nil
}

// GetChainExecutionFromChainAndToolCall gets the chain execution ID for a given chain ID and tool call ID
func (s *PostgresqlStore) GetChainExecutionFromChainAndToolCall(ctx context.Context, chainId uuid.UUID, toolCallId uuid.UUID) (*uuid.UUID, error) {
	query := `
        SELECT id FROM chainexecution
        WHERE chain_id = $1 
				AND toolcall_id = $2
    `

	var executionId uuid.UUID
	err := s.db.QueryRowContext(ctx, query, chainId, toolCallId).Scan(&executionId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get chain execution from chain and tool call: %w", err)
	}

	return &executionId, nil
}

func (s *PostgresqlStore) UpdateRunStatus(ctx context.Context, runId uuid.UUID, status asteroid.Status) error {
	query := `UPDATE run SET status = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, status, runId)
	if err != nil {
		return fmt.Errorf("error updating run status: %w", err)
	}

	return nil
}

func (s *PostgresqlStore) UpdateRunResult(ctx context.Context, runId uuid.UUID, result string) error {
	query := `
		UPDATE run SET result = $1 WHERE id = $2
	`
	_, err := s.db.ExecContext(ctx, query, result, runId)
	if err != nil {
		return fmt.Errorf("error creating run result: %w", err)
	}

	return nil
}

func (s *PostgresqlStore) CreateChatRequest(
	ctx context.Context,
	runId uuid.UUID,
	request []byte,
	response []byte,
	choices []asteroid.AsteroidChoice,
	format string,
	requestMessages []asteroid.AsteroidMessage,
) (*uuid.UUID, error) {
	if len(request) == 0 {
		return nil, fmt.Errorf("request is empty")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		INSERT INTO chat (request_data, response_data, run_id, format)
		VALUES ($1, $2, $3, $4) RETURNING id
	`
	var id uuid.UUID
	err = tx.QueryRowContext(ctx, query, request, response, runId, format).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("error creating chat entry: %w", err)
	}

	// Store the choices
	err = s.createChatChoices(ctx, tx, id, choices, requestMessages)
	if err != nil {
		return nil, fmt.Errorf("error creating chat choices: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return &id, nil
}

func (s *PostgresqlStore) createChatRequestMessages(
	ctx context.Context,
	tx *sql.Tx,
	choiceId uuid.UUID,
	requestMessages []asteroid.AsteroidMessage,
) error {
	// For each message, store it in the DB
	for _, message := range requestMessages {
		query := `
			INSERT INTO msg (id, choice_id, msg_data)
			VALUES ($1, $2, $3)
		`
		msgData, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("error marshalling message data: %w", err)
		}
		_, err = tx.ExecContext(ctx, query, message.Id, choiceId, msgData)
		if err != nil {
			return fmt.Errorf("error creating chat message: %w", err)
		}
	}

	return nil
}

func (s *PostgresqlStore) GetMessage(ctx context.Context, id uuid.UUID) (*asteroid.AsteroidMessage, error) {
	query := `
		SELECT msg_data FROM msg WHERE id = $1
	`
	var msgData []byte
	err := s.db.QueryRowContext(ctx, query, id).Scan(&msgData)
	if err != nil {
		return nil, fmt.Errorf("error getting message: %w", err)
	}

	var message asteroid.AsteroidMessage
	err = json.Unmarshal(msgData, &message)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling message: %w", err)
	}

	return &message, nil
}

func (s *PostgresqlStore) UpdateMessage(ctx context.Context, id uuid.UUID, message asteroid.AsteroidMessage) error {
	query := `
		UPDATE msg SET msg_data = $1 WHERE id = $2	
	`
	msgData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshalling message data: %w", err)
	}
	_, err = s.db.ExecContext(ctx, query, msgData, id)
	if err != nil {
		return fmt.Errorf("error updating message: %w", err)
	}
	return nil
}

func (s *PostgresqlStore) GetChat(
	ctx context.Context,
	runId uuid.UUID,
	index int,
) ([]byte, []byte, error) {
	query := `
		SELECT request_data, response_data
		FROM chat
		WHERE run_id = $1
		ORDER BY created_at DESC
		LIMIT 1 OFFSET $2
	`

	var requestData, responseData []byte
	err := s.db.QueryRowContext(ctx, query, runId, index).Scan(&requestData, &responseData)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting message: %w", err)
	}

	return requestData, responseData, nil
}

func (s *PostgresqlStore) GetRunChatCount(ctx context.Context, runId uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM chat 
		WHERE run_id = $1
	`

	var count int
	err := s.db.QueryRowContext(ctx, query, runId).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting chat count: %w", err)
	}

	return count, nil
}

func (s *PostgresqlStore) createChatChoices(
	ctx context.Context,
	tx *sql.Tx,
	chatId uuid.UUID,
	choices []asteroid.AsteroidChoice,
	requestMessages []asteroid.AsteroidMessage,
) error {
	// Store the choices in the DB
	for _, choice := range choices {
		query := `
			INSERT INTO choice (id, chat_id, choice_data)
			VALUES ($1, $2, $3)
		`

		choiceData, err := json.Marshal(choice)
		if err != nil {
			return fmt.Errorf("error marshalling choice data: %w", err)
		}

		_, err = tx.ExecContext(ctx, query, choice.AsteroidId, chatId, choiceData)
		if err != nil {
			return fmt.Errorf("error creating chat choice: %w", err)
		}

		// Convert the AsteroidId to a uuid
		choiceId, err := uuid.Parse(choice.AsteroidId)
		if err != nil {
			return fmt.Errorf("error parsing AsteroidId: %w", err)
		}

		// Store the request messages which are unique to the request that generated this choice
		err = s.createChatRequestMessages(ctx, tx, choiceId, requestMessages)
		if err != nil {
			return fmt.Errorf("error creating chat request messages: %w", err)
		}

		fmt.Printf("Choice message: %+v\n", choice.Message)
		// Store the message
		query = `
			INSERT INTO msg (id, choice_id, msg_data)
			VALUES ($1, $2, $3)
		`
		messageData, err := json.Marshal(choice.Message)
		if err != nil {
			return fmt.Errorf("error marshalling message data: %w", err)
		}

		msgId := choice.Message.Id
		if msgId == nil {
			return fmt.Errorf("message ID is nil")
		}

		_, err = tx.ExecContext(ctx, query, *msgId, choice.AsteroidId, messageData)
		if err != nil {
			return fmt.Errorf("error creating chat message: %w", err)
		}

		if choice.Message.ToolCalls != nil {
			// Store the tool calls
			err = s.createToolCalls(ctx, tx, *msgId, *choice.Message.ToolCalls)
			if err != nil {
				return fmt.Errorf("error creating tool calls: %w", err)
			}
		}
	}

	return nil
}

func (s *PostgresqlStore) createToolCalls(
	ctx context.Context,
	tx *sql.Tx,
	msgId uuid.UUID,
	toolCalls []asteroid.AsteroidToolCall,
) error {
	// Store the tool calls in the DB
	for _, toolCall := range toolCalls {
		query := `
			INSERT INTO toolcall (id, call_id, msg_id, tool_call_data, tool_id)
			VALUES ($1, $2, $3, $4, $5)
		`
		toolCallData, err := json.Marshal(toolCall)
		if err != nil {
			return fmt.Errorf("error marshalling tool call data: %w", err)
		}
		_, err = tx.ExecContext(ctx, query, toolCall.Id, toolCall.CallId, msgId, toolCallData, toolCall.ToolId)
		if err != nil {
			return fmt.Errorf("error creating tool call: %w", err)
		}

		// Get the chains configured for this tool
		chains, err := s.getChainsForTool(ctx, toolCall.ToolId)
		if err != nil {
			return fmt.Errorf("error getting chains configured for tool call: %w", err)
		}

		// Init the chain executions
		err = s.createChainExecutions(ctx, tx, toolCall.Id, chains)
		if err != nil {
			return fmt.Errorf("error creating chain executions: %w", err)
		}
	}

	return nil
}
