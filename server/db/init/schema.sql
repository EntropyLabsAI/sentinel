DROP TABLE IF EXISTS supervisionresult CASCADE;
DROP TABLE IF EXISTS toolrequest CASCADE;
DROP TABLE IF EXISTS supervisionrequest_status CASCADE;
DROP TABLE IF EXISTS supervisionrequest CASCADE;
DROP TABLE IF EXISTS run_tool_supervisor CASCADE;
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
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

CREATE TABLE tool (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR DEFAULT '',
    description TEXT DEFAULT '',
    attributes JSONB DEFAULT '{}' NOT NULL,
    ignored_attributes TEXT[] DEFAULT '{}' NOT NULL
);

CREATE TABLE project (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT DEFAULT '' UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE supervisor (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT DEFAULT '',
    description TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    type TEXT DEFAULT 'no_supervisor' CHECK (type in ('human_supervisor', 'client_supervisor', 'no_supervisor')),
    code TEXT DEFAULT '',
    attributes JSONB DEFAULT '{}' NOT NULL
);

CREATE TABLE llm_message (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role TEXT DEFAULT 'user' CHECK (role IN ('system', 'user', 'assistant')),
    content TEXT DEFAULT ''
);

CREATE TABLE user_project (
    user_id UUID REFERENCES sentinel_user(id),
    project_id UUID REFERENCES project(id),
    PRIMARY KEY (user_id, project_id)
);

CREATE TABLE run (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES project(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE execution (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID REFERENCES run(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    tool_id UUID REFERENCES tool(id)
);

CREATE TABLE llm_supervisor (
    supervisor_id UUID PRIMARY KEY REFERENCES supervisor(id),
    prompt TEXT DEFAULT ''
);

CREATE TABLE code_supervisor (
    supervisor_id UUID PRIMARY KEY REFERENCES supervisor(id),
    code UUID DEFAULT gen_random_uuid()
);

CREATE TABLE run_tool_supervisor (
    id SERIAL PRIMARY KEY,
    tool_id UUID REFERENCES tool(id),
    run_id UUID REFERENCES run(id),
    supervisor_id UUID REFERENCES supervisor(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE supervisionrequest (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID REFERENCES execution(id),
    supervisor_id UUID REFERENCES supervisor(id),
    task_state JSONB DEFAULT '{}' NOT NULL
);

CREATE TABLE supervisionrequest_status (
    id SERIAL PRIMARY KEY,
    supervisionrequest_id UUID REFERENCES supervisionrequest(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status TEXT DEFAULT 'pending' CHECK (status IN ('timeout', 'pending', 'completed', 'failed', 'assigned'))
);

CREATE TABLE toolrequest (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tool_id UUID REFERENCES tool(id),
    supervisionrequest_id UUID REFERENCES supervisionrequest(id),
    message_id UUID REFERENCES llm_message(id),
    arguments JSONB DEFAULT '{}' NOT NULL
);

CREATE TABLE supervisionresult (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supervisionrequest_id UUID REFERENCES supervisionrequest(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    decision TEXT DEFAULT 'reject' CHECK (decision IN ('approve', 'reject', 'terminate', 'modify', 'escalate')),
    toolrequest_id UUID REFERENCES toolrequest(id),
    reasoning TEXT DEFAULT ''
);
