from inspect_ai.tool import ToolCall
from inspect_ai.model import ChatMessageAssistant, ChatMessage
from pydantic_core import to_jsonable_python
from typing import Any
from copy import deepcopy

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
