-- First drop tables in reverse dependency order
DROP TABLE IF EXISTS supervisionresult CASCADE;
DROP TABLE IF EXISTS supervisionrequest_status CASCADE;
DROP TABLE IF EXISTS supervisionrequest CASCADE;
DROP TABLE IF EXISTS chainexecution CASCADE;
DROP TABLE IF EXISTS toolrequest_group CASCADE;
DROP TABLE IF EXISTS toolrequest CASCADE;
DROP TABLE IF EXISTS tool_supervisorchain CASCADE;
DROP TABLE IF EXISTS supervisorchain_supervisor CASCADE;
DROP TABLE IF EXISTS message CASCADE;
DROP TABLE IF EXISTS user_project CASCADE;
DROP TABLE IF EXISTS tool CASCADE;
DROP TABLE IF EXISTS run CASCADE;
DROP TABLE IF EXISTS supervisorchain CASCADE;
DROP TABLE IF EXISTS supervisor CASCADE;
DROP TABLE IF EXISTS "group" CASCADE;
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

CREATE TABLE "group" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
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

CREATE TABLE supervisorchain (
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

CREATE TABLE supervisorchain_supervisor (
    supervisor_id UUID REFERENCES supervisor(id),
    supervisorchain_id UUID REFERENCES supervisorchain(id),
    PRIMARY KEY (supervisor_id, supervisorchain_id)
);

CREATE TABLE tool_supervisorchain (
    tool_id UUID REFERENCES tool(id),
    supervisorchain_id UUID REFERENCES supervisorchain(id),
    PRIMARY KEY (tool_id, supervisorchain_id)
);

CREATE TABLE toolrequest (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tool_id UUID REFERENCES tool(id),
    run_id UUID REFERENCES run(id),
    message_id UUID REFERENCES message(id) NULL,
    arguments JSONB DEFAULT '{}' NOT NULL
);

CREATE TABLE toolrequest_group (
    group_id UUID REFERENCES "group"(id),
    toolrequest_id UUID REFERENCES toolrequest(id),
    PRIMARY KEY (group_id, toolrequest_id)
);

CREATE TABLE chainexecution (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID REFERENCES run(id),
    toolrequest_group_id UUID REFERENCES toolrequest_group(id),
    supervisorchain_id UUID REFERENCES supervisorchain(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE supervisionrequest (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chainexecution_id UUID REFERENCES chainexecution(id),
    supervisor_id UUID REFERENCES supervisor(id),
    position_in_chain INTEGER,
    task_state JSONB DEFAULT '{}' NOT NULL
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
    reasoning TEXT DEFAULT ''
);
