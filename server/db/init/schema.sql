-- First drop tables in reverse dependency order
DROP TABLE IF EXISTS supervisionresult CASCADE;
DROP TABLE IF EXISTS supervisionrequest_status CASCADE;
DROP TABLE IF EXISTS supervisionrequest CASCADE;
DROP TABLE IF EXISTS chainexecution CASCADE;
DROP TABLE IF EXISTS toolrequest CASCADE;
DROP TABLE IF EXISTS chain_tool CASCADE;
DROP TABLE IF EXISTS chain_supervisor CASCADE;
DROP TABLE IF EXISTS message CASCADE;
DROP TABLE IF EXISTS user_project CASCADE;
DROP TABLE IF EXISTS tool CASCADE;
DROP TABLE IF EXISTS run CASCADE;
DROP TABLE IF EXISTS chain CASCADE;
DROP TABLE IF EXISTS supervisor CASCADE;
DROP TABLE IF EXISTS requestgroup CASCADE;
DROP TABLE IF EXISTS project CASCADE;
DROP TABLE IF EXISTS sentinel_user CASCADE;

-- Create tables in dependency order (tables with no foreign keys first)
CREATE TABLE sentinel_user (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

CREATE TABLE project (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT DEFAULT '' UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE requestgroup (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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

CREATE TABLE chain (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE run (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES project(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tool (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID REFERENCES run(id),
    name VARCHAR DEFAULT '',
    description TEXT DEFAULT '',
    attributes JSONB DEFAULT '{}' NOT NULL,
    ignored_attributes TEXT[] DEFAULT '{}' NOT NULL
);

CREATE TABLE message (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role TEXT DEFAULT 'user' CHECK (role IN ('system', 'user', 'assistant')),
    content TEXT DEFAULT ''
);

CREATE TABLE user_project (
    user_id UUID REFERENCES sentinel_user(id),
    project_id UUID REFERENCES project(id),
    PRIMARY KEY (user_id, project_id)
);

CREATE TABLE chain_supervisor (
    supervisor_id UUID REFERENCES supervisor(id),
    chain_id UUID REFERENCES chain(id),
    position_in_chain INTEGER,
    PRIMARY KEY (supervisor_id, chain_id)
);

CREATE TABLE chain_tool (
    tool_id UUID REFERENCES tool(id),
    chain_id UUID REFERENCES chain(id),
    PRIMARY KEY (tool_id, chain_id)
);

CREATE TABLE toolrequest (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tool_id UUID REFERENCES tool(id),
    message_id UUID REFERENCES message(id) NULL,
    arguments JSONB DEFAULT '{}' NOT NULL,
    task_state JSONB DEFAULT '{}' NOT NULL,
    requestgroup_id UUID REFERENCES requestgroup(id) NULL
);

CREATE TABLE chainexecution (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requestgroup_id UUID REFERENCES requestgroup(id),
    chain_id UUID REFERENCES chain(id),
    supervisor_id UUID REFERENCES supervisor(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE supervisionrequest (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chainexecution_id UUID REFERENCES chainexecution(id),
    supervisor_id UUID REFERENCES supervisor(id),
    position_in_chain INTEGER
);

CREATE TABLE supervisionrequest_status (
    id SERIAL PRIMARY KEY,
    supervisionrequest_id UUID REFERENCES supervisionrequest(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status TEXT DEFAULT 'pending' CHECK (status IN ('timeout', 'pending', 'completed', 'failed', 'assigned'))
);

CREATE TABLE supervisionresult (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supervisionrequest_id UUID REFERENCES supervisionrequest(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    decision TEXT DEFAULT 'reject' CHECK (decision IN ('approve', 'reject', 'terminate', 'modify', 'escalate')),
    reasoning TEXT DEFAULT '',
    chosen_toolrequest_id UUID REFERENCES toolrequest(id) NULL
);
