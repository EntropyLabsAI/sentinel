from inspect_ai.tool import ToolCall
from inspect_ai.model import ChatMessageAssistant, ChatMessage
from inspect_ai.tool._tool_call import ToolCall
from inspect_ai.model import get_model, Model
from inspect_ai.solver import TaskState
from el.supervision.config import SupervisionDecision
from inspect_ai.approval import Approval
from pydantic_core import to_jsonable_python
from typing import Any
from copy import deepcopy
import logging
from inspect_ai.model import get_model, Model
from typing import List, Dict,  Any, Tuple
import asyncio
import requests
from ._config import SLEEP_TIME
from inspect_ai.solver._task_state import state_jsonable

def tool_jsonable(tool_call: ToolCall | None = None) -> dict[str, Any] | None:
    if tool_call is None:
        return None

    return {
        "id": tool_call.id,
        "function": tool_call.function,
        "arguments": tool_call.arguments,
        "type": tool_call.type,
    }

def chat_message_jsonable(message: ChatMessage) -> dict[str, Any]:
    def as_jsonable(value: Any) -> Any:
        return to_jsonable_python(value, exclude_none=True, fallback=lambda _x: None)

    message_data = {
        "role": message.role,
        "content": message.content,
        "source": message.source,
    }

    if isinstance(message, ChatMessageAssistant):
        message_data["tool_calls"] = [tool_jsonable(call) for call in message.tool_calls if call is not None] if message.tool_calls else None

    jsonable = as_jsonable(message_data)
    return deepcopy(jsonable)


def generate_tool_call_change_explanation(original_call: ToolCall, modified_data: dict) -> str:
    """
    Generate a detailed explanation of changes made to a tool call.

    Args:
        original_call (ToolCall): The original tool call.
        modified_data (dict): The modified tool call data.

    Returns:
        str: A formatted explanation of the changes.
    """
    explanation = "Human changed the tool call:\n"
    explanation += f"From: {original_call.function}({', '.join(f'{k}={v}' for k, v in original_call.arguments.items())})\n"
    explanation += f"To: {modified_data['function']}("
    if 'arguments' in modified_data:
        explanation += ', '.join(f"{k}='{v}'" for k, v in modified_data['arguments'].items())
    explanation += ")\n"
    explanation += f"ID: {modified_data.get('id', 'N/A')}\n"
    explanation += f"Type: {modified_data.get('type', 'N/A')}"
    return explanation



async def generate_tool_call_suggestions(state: TaskState, n: int) -> Tuple[List[Any], List[Any]]:
    """
    Generate N tool call suggestions.

    Args:
        state (TaskState): The current state of the task.
        n (int): The number of tool call suggestions to generate.

    Returns:
        Tuple[List[Any], List[Any]]: A tuple containing lists of last messages and tool options.
    """
    model: Model = get_model()
    message_without_last_message = deepcopy(state.messages)
    last_messages = [message_without_last_message[-1]]
    tool_options = [tool_jsonable(state.messages[-1].tool_calls[0])] if hasattr(state.messages[-1], 'tool_calls') and state.messages[-1].tool_calls else [None]
    message_without_last_message.pop()

    for _ in range(n - 1):
        output = await model.generate(message_without_last_message, tools=state.tools)
        last_messages.append(output.message)
        tool_options.append(tool_jsonable(output.message.tool_calls[0]) if output.message.tool_calls else None)
    
    return last_messages, tool_options

def prepare_payload(agent_id: str, state: TaskState, last_messages: List[Any], tool_options: List[Any]) -> Dict[str, Any]:
    """
    Prepare the payload for API requests.

    Args:
        agent_id (str): The ID of the agent.
        state (TaskState): The current state of the task.
        last_messages (List[Any]): List of last messages.
        tool_options (List[Any]): List of tool options.

    Returns:
        Dict[str, Any]: The prepared payload.
    """
    last_messages_json = [chat_message_jsonable(message) for message in last_messages if message is not None]

    state_json = state_jsonable(state)
    state_json['tool_choice'] = None  # TODO: Fix this

    return {
        "agent_id": agent_id,
        "task_state": state_json,
        "tool_choices": tool_options,
        "last_messages": last_messages_json,
    }

async def poll_status(approval_api_endpoint: str, review_id: str, timeout: int) -> Dict[str, Any]:
    """
    Poll the status endpoint until we get a response.

    Args:
        approval_api_endpoint (str): The URL of the remote approval API.
        review_id (str): The ID of the review to poll for.
        timeout (int): The maximum time to wait for a response, in seconds.

    Returns:
        Dict[str, Any]: The status data received from the API.
    """
    max_attempts = timeout // SLEEP_TIME
    logging.info(f"Waiting for approval for review {review_id}")
    for _ in range(int(max_attempts)):
        try:
            status_response = requests.get(f"{approval_api_endpoint}/api/review/status?id={review_id}")
            status_response.raise_for_status()

            status_data = status_response.json()
            logging.debug(f"Status data: {status_data}")

            if status_data.get("status") == "pending":
                await asyncio.sleep(SLEEP_TIME)  # Wait before polling again
                continue

            if "decision" in status_data:
                return status_data

        except requests.RequestException as poll_error:
            logging.error(f"Error polling status: {poll_error}")
            await asyncio.sleep(SLEEP_TIME)  # Wait before retrying
            continue

    raise TimeoutError("Timed out waiting for approval")
