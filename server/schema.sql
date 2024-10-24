-- Drop existing tables if they exist
DROP TABLE IF EXISTS project_tools;
DROP TABLE IF EXISTS project_agents;
DROP TABLE IF EXISTS review_results;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS tools;
DROP TABLE IF EXISTS agents;
DROP TABLE IF EXISTS projects;

-- Create table for projects
CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

-- Create table for tools
CREATE TABLE tools (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    parameters JSON,
    description TEXT,
    source_code TEXT
);

-- Create table for project_tools to associate tools with projects
CREATE TABLE project_tools (
    project_id TEXT NOT NULL,
    tool_id INTEGER NOT NULL,
    PRIMARY KEY (project_id, tool_id),
    FOREIGN KEY (project_id) REFERENCES projects(id),
    FOREIGN KEY (tool_id) REFERENCES tools(id)
);

-- Create table for agents
CREATE TABLE agents (
    id TEXT PRIMARY KEY
);

-- Create table for project_agents to associate agents with projects
CREATE TABLE project_agents (
    project_id TEXT NOT NULL,
    agent_id TEXT NOT NULL,
    PRIMARY KEY (project_id, agent_id),
    FOREIGN KEY (project_id) REFERENCES projects(id),
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

-- Create table for reviews
CREATE TABLE reviews (
    id TEXT PRIMARY KEY,
    agent_id TEXT,
    project_id TEXT,
    request TEXT,     -- JSON-serialized ReviewRequest
    review_type TEXT, -- 'human' or 'llm'
    status TEXT,      -- 'queued', 'processing', 'completed', 'timeout'
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

-- Create table for review results
CREATE TABLE review_results (
    id TEXT PRIMARY KEY,
    review_id TEXT NOT NULL,
    decision TEXT,
    reasoning TEXT, 
    tool_choice TEXT, -- JSON-serialized ToolChoice
    FOREIGN KEY (review_id) REFERENCES reviews(id)
);
