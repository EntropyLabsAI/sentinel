package sentinel

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type db struct {
	db *sql.DB
}

func (d *db) Close() error {
	return d.db.Close()
}

func (d *db) GetProject(id string) (*Project, error) {
	var project Project
	err := d.db.QueryRow(`
		SELECT p.id, p.name 
		FROM projects p 
		WHERE p.id = ?`, id).Scan(&project.Id, &project.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %v", err)
	}

	// Get associated tools
	rows, err := d.db.Query(`
		SELECT t.id, t.name, t.description, pt.attributes
		FROM tools t
		JOIN project_tools pt ON t.id = pt.tool_id
		WHERE pt.project_id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get project tools: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tool Tool
		var attributes string
		if err := rows.Scan(&tool.Name, &tool.Description, &attributes); err != nil {
			return nil, fmt.Errorf("failed to scan tool: %v", err)
		}
		if err := json.Unmarshal([]byte(attributes), &tool.Attributes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tool attributes: %v", err)
		}
		project.Tools = append(project.Tools, tool)
	}

	return &project, nil
}

func (d *db) RegisterProject(project *Project) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Insert project
	_, err = tx.Exec("INSERT INTO projects (id, name) VALUES (?, ?)",
		project.Id, project.Name)
	if err != nil {
		return fmt.Errorf("failed to insert project: %v", err)
	}

	// Insert tools and project_tools relationships
	for _, tool := range project.Tools {
		var toolId int64
		// Insert tool if it doesn't exist
		result, err := tx.Exec(`
			INSERT INTO tools (name, description) 
			VALUES (?, ?) 
			ON CONFLICT(name) DO UPDATE SET description=?`,
			tool.Name, tool.Description, tool.Description)
		if err != nil {
			return fmt.Errorf("failed to insert/update tool: %v", err)
		}

		toolId, err = result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get tool id: %v", err)
		}

		// Marshal tool attributes to JSON
		attributes, err := json.Marshal(tool.Attributes)
		if err != nil {
			return fmt.Errorf("failed to marshal tool attributes: %v", err)
		}

		// Insert project_tools relationship
		_, err = tx.Exec(`
			INSERT INTO project_tools (project_id, tool_id, attributes) 
			VALUES (?, ?, ?)`,
			project.Id, toolId, string(attributes))
		if err != nil {
			return fmt.Errorf("failed to insert project_tool: %v", err)
		}
	}

	return tx.Commit()
}

func (d *db) GetProjects() ([]*Project, error) {
	var projects []*Project

	rows, err := d.db.Query("SELECT * FROM projects")
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.Id, &project.Name); err != nil {
			return nil, fmt.Errorf("failed to scan project: %v", err)
		}
		projects = append(projects, &project)
	}

	return projects, nil
}

func newDB() (*db, error) {
	d, err := sql.Open("sqlite3", "./data/sentinel.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	return &db{
		db: d,
	}, nil
}

func (d *db) RegisterAgent(projectId string, agent *Agent) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Insert agent
	_, err = tx.Exec("INSERT INTO agents (id) VALUES (?)", agent.Id)
	if err != nil {
		return fmt.Errorf("failed to insert agent: %v", err)
	}

	// Create project_agents relationship
	_, err = tx.Exec(`
		INSERT INTO project_agents (project_id, agent_id) 
		VALUES (?, ?)`,
		projectId, agent.Id)
	if err != nil {
		return fmt.Errorf("failed to insert project_agent: %v", err)
	}

	return tx.Commit()
}

func (d *db) GetAgents(projectId string) ([]*Agent, error) {
	rows, err := d.db.Query(`
		SELECT a.id 
		FROM agents a
		JOIN project_agents pa ON a.id = pa.agent_id
		WHERE pa.project_id = ?`, projectId)
	if err != nil {
		return nil, fmt.Errorf("failed to get agents: %v", err)
	}
	defer rows.Close()

	var agents []*Agent
	for rows.Next() {
		var agent Agent
		if err := rows.Scan(&agent.Id); err != nil {
			return nil, fmt.Errorf("failed to scan agent: %v", err)
		}
		agents = append(agents, &agent)
	}

	return agents, nil
}

func (d *db) StoreReview(review *Review) error {
	requestJson, err := json.Marshal(review.Request)
	if err != nil {
		return fmt.Errorf("failed to marshal review request: %v", err)
	}

	_, err = d.db.Exec(`
		INSERT INTO reviews (id, agent_id, project_id, request, review_type, status)
		VALUES (?, ?, ?, ?, ?, ?)`,
		review.Id, review.Request.AgentId, "", requestJson, "human", "queued")
	if err != nil {
		return fmt.Errorf("failed to store review: %v", err)
	}

	return nil
}

func (d *db) StoreReviewResult(result *ReviewResult) error {
	toolChoiceJson, err := json.Marshal(result.ToolChoice)
	if err != nil {
		return fmt.Errorf("failed to marshal tool choice: %v", err)
	}

	_, err = d.db.Exec(`
		INSERT INTO review_results (id, review_id, decision, reasoning, tool_choice)
		VALUES (?, ?, ?, ?, ?)`,
		result.Id, result.Id, result.Decision, result.Reasoning, string(toolChoiceJson))
	if err != nil {
		return fmt.Errorf("failed to store review result: %v", err)
	}

	// Update review status
	_, err = d.db.Exec(`
		UPDATE reviews 
		SET status = 'completed', result_id = ? 
		WHERE id = ?`,
		result.Id, result.Id)
	if err != nil {
		return fmt.Errorf("failed to update review status: %v", err)
	}

	return nil
}

func (d *db) GetReviewStatus(id string) (*ReviewStatusResponse, error) {
	var status ReviewStatusResponse
	var dbStatus string
	err := d.db.QueryRow(`
		SELECT id, status 
		FROM reviews 
		WHERE id = ?`, id).Scan(&status.Id, &dbStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to get review status: %v", err)
	}

	status.Status = Status(dbStatus)
	return &status, nil
}

func (d *db) GetLLMReviews() ([]*ReviewResult, error) {
	rows, err := d.db.Query(`
		SELECT rr.id, rr.decision, rr.reasoning, rr.tool_choice
		FROM review_results rr
		JOIN reviews r ON r.result_id = rr.id
		WHERE r.review_type = 'llm'`)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM reviews: %v", err)
	}
	defer rows.Close()

	var reviews []*ReviewResult
	for rows.Next() {
		var review ReviewResult
		var toolChoiceJson string
		if err := rows.Scan(&review.Id, &review.Decision, &review.Reasoning, &toolChoiceJson); err != nil {
			return nil, fmt.Errorf("failed to scan review: %v", err)
		}
		if err := json.Unmarshal([]byte(toolChoiceJson), &review.ToolChoice); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tool choice: %v", err)
		}
		reviews = append(reviews, &review)
	}

	return reviews, nil
}
