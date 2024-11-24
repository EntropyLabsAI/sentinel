from entropy_labs.supervision import supervise
from entropy_labs.api import register_project, create_run, register_task, submit_run_status, submit_run_result, Status
from entropy_labs.supervision.supervisors import human_supervisor, llm_supervisor, Supervisor
from entropy_labs.supervision.config import (
    SupervisionDecision,
    SupervisionDecisionType,
    SupervisionContext,
    get_supervision_context,
)
from __future__ import annotations

from anthropic import Anthropic
from anthropic.types import ToolParam, MessageParam

client = Anthropic()

user_message: MessageParam = {
    "role": "user",
    "content": "What is the weather in SF?",
}
tools: list[ToolParam] = [
    {
        "name": "get_weather",
        "description": "Get the weather for a specific location",
        "input_schema": {
            "type": "object",
            "properties": {"location": {"type": "string"}},
        },
    }
]

message = client.messages.create(
    model="claude-3-opus-20240229",
    max_tokens=1024,
    messages=[user_message],
    tools=tools,
)
print(f"Initial response: {message.model_dump_json(indent=2)}")

assert message.stop_reason == "tool_use"

tool = next(c for c in message.content if c.type == "tool_use")
response = client.messages.create(
    model="claude-3-opus-20240229",
    max_tokens=1024,
    messages=[
        user_message,
        {"role": message.role, "content": message.content},
        {
            "role": "user",
            "content": [
                {
                    "type": "tool_result",
                    "tool_use_id": tool.id,
                    "content": [{"type": "text", "text": "The weather is 73f"}],
                }
            ],
        },
    ],
    tools=tools,
)
print(f"\nFinal response: {response.model_dump_json(indent=2)}")

# Register project, task, and run with Entropy Labs
entropy_labs_backend_url = "http://localhost:8080"

# Entropy Labs backend needs to be running
project_id = register_project(
    project_name="Email Assistant",
    entropy_labs_backend_url=entropy_labs_backend_url
)
task_id = register_task(project_id=project_id, task_name="Email Assistant")
run_id = create_run(project_id=project_id, task_id=task_id)





