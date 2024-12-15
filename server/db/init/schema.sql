-- First drop tables in reverse dependency order
DROP TABLE IF EXISTS msg CASCADE;
DROP TABLE IF EXISTS choice CASCADE;
DROP TABLE IF EXISTS chat CASCADE;
DROP TABLE IF EXISTS supervisionresult CASCADE;
DROP TABLE IF EXISTS supervisionrequest_status CASCADE;
DROP TABLE IF EXISTS supervisionrequest CASCADE;
DROP TABLE IF EXISTS chainexecution CASCADE;
DROP TABLE IF EXISTS toolcall CASCADE;
DROP TABLE IF EXISTS chain_tool CASCADE;
DROP TABLE IF EXISTS chain_supervisor CASCADE;
DROP TABLE IF EXISTS user_project CASCADE;
DROP TABLE IF EXISTS tool CASCADE;
DROP TABLE IF EXISTS run CASCADE;
DROP TABLE IF EXISTS chain CASCADE;
DROP TABLE IF EXISTS supervisor CASCADE;
DROP TABLE IF EXISTS project CASCADE;
DROP TABLE IF EXISTS asteroid_user CASCADE;
DROP TABLE IF EXISTS task CASCADE;

-- Create tables in dependency order (tables with no foreign keys first)
CREATE TABLE asteroid_user (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE
);

CREATE TABLE project (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT DEFAULT '' UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    run_result_tags TEXT[] DEFAULT '{"success", "failure"}' NOT NULL
);

CREATE TABLE supervisor (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT DEFAULT '',
    description TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    type TEXT DEFAULT 'no_supervisor' CHECK (type in ('human_supervisor', 'client_supervisor', 'no_supervisor', 'chat_supervisor')),
    code TEXT DEFAULT '',
    attributes JSONB DEFAULT '{}' NOT NULL
);

CREATE TABLE chain (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE task (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES project(id),
    name TEXT DEFAULT '',
    description TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE run (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID REFERENCES task(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed')) NOT NULL,
    result TEXT DEFAULT ''
);

CREATE TABLE tool (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID REFERENCES run(id),
    name VARCHAR DEFAULT '',
    description TEXT DEFAULT '',
    attributes JSONB DEFAULT '{}' NOT NULL,
    ignored_attributes TEXT[] DEFAULT '{}' NOT NULL,
    code TEXT DEFAULT ''
);

CREATE TABLE user_project (
    user_id UUID REFERENCES asteroid_user(id),
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

CREATE TABLE chat (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    request_data JSONB DEFAULT '{}' NOT NULL,
    response_data JSONB DEFAULT '{}' NOT NULL,
    run_id UUID REFERENCES run(id) NOT NULL,
    format TEXT CHECK (format IN ('openai', 'anthropic')) NOT NULL
);

CREATE TABLE choice (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id UUID REFERENCES chat(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    choice_data JSONB DEFAULT '{}' NOT NULL
);


CREATE TABLE msg (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    choice_id UUID REFERENCES choice(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    msg_data JSONB DEFAULT '{}' NOT NULL
);

CREATE TABLE toolcall (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    call_id TEXT DEFAULT '' NOT NULL,
    tool_id UUID REFERENCES tool(id),
    msg_id UUID REFERENCES msg(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    tool_call_data JSONB DEFAULT '{}' NOT NULL
);

CREATE TABLE chainexecution (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    toolcall_id UUID REFERENCES toolcall(id),
    chain_id UUID REFERENCES chain(id),
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
    toolcall_id UUID REFERENCES toolcall(id) NULL
);

