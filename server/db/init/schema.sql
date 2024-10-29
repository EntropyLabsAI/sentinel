DROP TABLE IF EXISTS supervisionresult CASCADE;
DROP TABLE IF EXISTS toolrequest CASCADE;
DROP TABLE IF EXISTS supervisionrequest_status CASCADE;
DROP TABLE IF EXISTS supervisionrequest CASCADE;
DROP TABLE IF EXISTS run_tool_supervisor CASCADE;
DROP TABLE IF EXISTS execution_status CASCADE;
DROP TABLE IF EXISTS llm_supervisor CASCADE;
DROP TABLE IF EXISTS code_supervisor CASCADE;
DROP TABLE IF EXISTS execution CASCADE;
DROP TABLE IF EXISTS run CASCADE;
DROP TABLE IF EXISTS user_project CASCADE;
DROP TABLE IF EXISTS llm_message CASCADE;
DROP TABLE IF EXISTS supervisor CASCADE;
DROP TABLE IF EXISTS project CASCADE;
DROP TABLE IF EXISTS tool CASCADE;
DROP TABLE IF EXISTS sentinel_user CASCADE;

CREATE TABLE sentinel_user (
    id UUID PRIMARY KEY
);

CREATE TABLE tool (
    id UUID PRIMARY KEY,
    name VARCHAR,
    description TEXT,
    attributes JSONB
);

CREATE TABLE project (
    id UUID PRIMARY KEY,
    name TEXT,
    created_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE supervisor (
    id UUID PRIMARY KEY,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE,
    type TEXT CHECK (type in ('human_supervisor', 'client_supervisor')),
    code TEXT
);

CREATE TABLE llm_message (
    id UUID PRIMARY KEY,
    role TEXT CHECK (role IN ('system', 'user', 'assistant')),
    content TEXT
);

CREATE TABLE user_project (
    user_id UUID,
    project_id UUID,
    PRIMARY KEY (user_id, project_id),
    FOREIGN KEY (user_id) REFERENCES sentinel_user(id),
    FOREIGN KEY (project_id) REFERENCES project(id)
);

CREATE TABLE run (
    id UUID PRIMARY KEY,
    project_id UUID,
    created_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (project_id) REFERENCES project(id)
);

CREATE TABLE execution (
    id UUID PRIMARY KEY,
    run_id UUID,
    created_at TIMESTAMP WITH TIME ZONE,
    tool_id UUID,
    FOREIGN KEY (run_id) REFERENCES run(id),
    FOREIGN KEY (tool_id) REFERENCES tool(id)
);

CREATE TABLE llm_supervisor (
    supervisor_id UUID PRIMARY KEY,
    prompt TEXT,
    FOREIGN KEY (supervisor_id) REFERENCES supervisor(id)
);

CREATE TABLE code_supervisor (
    supervisor_id UUID PRIMARY KEY,
    code UUID,
    FOREIGN KEY (supervisor_id) REFERENCES supervisor(id)
);

CREATE TABLE execution_status (
    id SERIAL PRIMARY KEY,
    execution_id UUID,
    created_at TIMESTAMP WITH TIME ZONE,
    status TEXT CHECK (status IN ('pending', 'completed', 'failed')),
    FOREIGN KEY (execution_id) REFERENCES execution(id)
);

CREATE TABLE run_tool_supervisor (
    id SERIAL PRIMARY KEY,
    tool_id UUID,
    run_id UUID,
    supervisor_id UUID,
    created_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (tool_id) REFERENCES tool(id),
    FOREIGN KEY (run_id) REFERENCES run(id),
    FOREIGN KEY (supervisor_id) REFERENCES supervisor(id)
);

CREATE TABLE supervisionrequest (
    id UUID PRIMARY KEY,
    execution_id UUID,
    supervisor_id UUID,
    task_state JSONB,
    FOREIGN KEY (execution_id) REFERENCES execution(id),
    FOREIGN KEY (supervisor_id) REFERENCES supervisor(id)
);

CREATE TABLE supervisionrequest_status (
    id SERIAL PRIMARY KEY,
    supervisionrequest_id UUID,
    created_at TIMESTAMP WITH TIME ZONE,
    status TEXT CHECK (status IN ('timeout', 'pending', 'completed', 'failed', 'assigned')),
    FOREIGN KEY (supervisionrequest_id) REFERENCES supervisionrequest(id)
);

CREATE TABLE toolrequest (
    id UUID PRIMARY KEY,
    tool_id UUID,
    supervisionrequest_id UUID,
    message_id UUID,
    arguments JSONB,
    FOREIGN KEY (supervisionrequest_id) REFERENCES supervisionrequest(id),
    FOREIGN KEY (tool_id) REFERENCES tool(id),
    FOREIGN KEY (message_id) REFERENCES llm_message(id)
);

CREATE TABLE supervisionresult (
    id UUID PRIMARY KEY,
    supervisionrequest_id UUID,
    created_at TIMESTAMP WITH TIME ZONE,
    decision TEXT CHECK (decision IN ('approve', 'reject', 'terminate', 'modify', 'escalate')),
    toolrequest_id UUID,
    reasoning TEXT,
    FOREIGN KEY (supervisionrequest_id) REFERENCES supervisionrequest(id),
    FOREIGN KEY (toolrequest_id) REFERENCES toolrequest(id)
);
