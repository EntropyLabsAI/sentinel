package database

import (
	"context"
	"database/sql"
	"fmt"

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
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting project: %w", err)
	}
	return &project, nil
}

func (s *PostgresqlStore) ListProjects(ctx context.Context) ([]sentinel.Project, error) {
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
	return projects, nil
}

// ReviewStore implementation
func (s *PostgresqlStore) CreateReview(ctx context.Context, review sentinel.Review) error {
	query := `
		INSERT INTO reviewrequest (id, run_id, task_state)
		VALUES ($1, $2, $3)`

	_, err := s.db.ExecContext(ctx, query, review.Id, review.RunId, review.TaskState)
	if err != nil {
		return fmt.Errorf("error creating review: %w", err)
	}
	return nil
}

func (s *PostgresqlStore) GetReview(ctx context.Context, id uuid.UUID) (*sentinel.Review, error) {
	query := `
		SELECT r.id, r.run_id, r.task_state, rs.status
		FROM reviewrequest r
		LEFT JOIN reviewrequest_status rs ON r.id = rs.reviewrequest_id
		WHERE r.id = $1
		ORDER BY rs.created_at DESC
		LIMIT 1`

	var review sentinel.Review
	var status sentinel.ReviewStatus
	err := s.db.QueryRowContext(ctx, query, id).Scan(&review.Id, &review.RunId, &review.TaskState, &status.Status)
	if err == sql.ErrNoRows {
		// No review found
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting review: %w", err)
	}

	if status.Status != "" {
		review.Status = &status
	}
	return &review, nil
}

func (s *PostgresqlStore) UpdateReview(ctx context.Context, review sentinel.Review) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	// Update review request
	query1 := `
		UPDATE reviewrequest 
		SET task_state = $1
		WHERE id = $2`

	_, err = tx.ExecContext(ctx, query1, review.TaskState, review.Id)
	if err != nil {
		return fmt.Errorf("error updating review: %w", err)
	}

	// Insert new status
	query2 := `
		INSERT INTO reviewrequest_status (id, reviewrequest_id, created_at, status)
		VALUES ($1, $2, CURRENT_TIMESTAMP, $3)`

	_, err = tx.ExecContext(ctx, query2, review.Id, review.Id, review.Status)
	if err != nil {
		return fmt.Errorf("error updating review status: %w", err)
	}

	return tx.Commit()
}

func (s *PostgresqlStore) ListHumanReviews(ctx context.Context) ([]sentinel.Review, error) {
	fmt.Println("Stub: ListHumanReviews called")
	return []sentinel.Review{}, nil
}

func (s *PostgresqlStore) ListLLMReviews(ctx context.Context) ([]sentinel.Review, error) {
	fmt.Println("Stub: ListLLMReviews called")
	return []sentinel.Review{}, nil
}

func (s *PostgresqlStore) DeleteReview(ctx context.Context, id uuid.UUID) error {
	fmt.Printf("Stub: DeleteReview called with ID: %s\n", id)
	return nil
}

func (s *PostgresqlStore) CountReviews(ctx context.Context) (int, error) {
	fmt.Println("Stub: CountReviews called")
	return 0, nil
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
		SELECT id, name, created_at
		FROM tool
		WHERE id = $1`

	var tool sentinel.Tool
	err := s.db.QueryRowContext(ctx, query, id).Scan(&tool.Id, &tool.Name, &tool.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting tool: %w", err)
	}

	return &tool, nil
}

func (s *PostgresqlStore) GetSupervisorFromToolID(ctx context.Context, id uuid.UUID) (*sentinel.Supervisor, error) {
	query := `
		SELECT id, description, created_at, type
		FROM supervisor
		INNER JOIN tool_supervisor ON supervisor.id = tool_supervisor.supervisor_id
		INNER JOIN tool ON tool_supervisor.tool_id = tool.id
		WHERE tool.id = $1`

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

func (s *PostgresqlStore) CreateRun(ctx context.Context, run sentinel.Run) error {
	query := `
		INSERT INTO run (id, project_id, created_at)
		VALUES ($1, $2, $3)`

	_, err := s.db.ExecContext(ctx, query, run.Id, run.ProjectId, run.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creating run: %w", err)
	}
	return nil
}

// ToolStore implementation
func (s *PostgresqlStore) CreateTool(ctx context.Context, tool sentinel.Tool) error {
	// First, check if the tool already exists
	query := `
		SELECT id
		FROM tool
		WHERE id = $1`

	var existingTool sentinel.Tool
	err := s.db.QueryRowContext(ctx, query, tool.Id).Scan(&existingTool.Id)
	if err == nil {
		// Check if the tool has changed
		if existingTool.Name != tool.Name || existingTool.CreatedAt != tool.CreatedAt {
			return fmt.Errorf("tool already exists but has changed")
		}
		return nil
	}

	// If the tool doesn't exist, create it
	query = `
		INSERT INTO tool (id, name, created_at)
		VALUES ($1, $2, $3)`

	_, err = s.db.ExecContext(ctx, query, tool.Id, tool.Name, tool.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creating tool: %w", err)
	}
	return nil
}
