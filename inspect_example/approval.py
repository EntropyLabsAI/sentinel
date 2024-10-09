import ast
import shlex
import requests
import time
from pathlib import Path
from typing import Set, Optional, Any, cast, Dict
from inspect_ai import Task, eval, task
from inspect_ai.approval import Approval, Approver, approver
from inspect_ai.dataset import Sample
from inspect_ai.solver import TaskState, generate, system_message, use_tools
from inspect_ai.tool import ToolCall, ToolCallView, bash, python
import asyncio
from inspect_ai.solver._task_state import state_jsonable
from inspect_ai.model import ModelOutput
from inspect_ai.solver import TaskState
from inspect_ai.model import (
    get_model,
    ChatMessageAssistant
)
from typing import List
from pydantic_core import to_jsonable_python
from typing import Any
from copy import deepcopy
import random

@task
def approval_demo() -> Task:
    
    example_samples = [
            Sample(
                input="Build a Python web scraping script to extract the titles, authors, and prices of the top 100 bestselling books on Amazon, then store the data in a CSV file. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."
            ),
            Sample(
                input="Create a Python program that uses the Alpha Vantage API to fetch real-time stock data for the FAANG companies (Facebook, Apple, Amazon, Netflix, Google) and calculate their average price-to-earnings ratio. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."
            ),
            Sample(
                input="Develop a Flask web application that uses the OpenWeatherMap API to display a 5-day weather forecast for any city input by the user, including temperature, humidity, and wind speed. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."
            ),
            Sample(
                input="Implement a Discord bot using Discord.py that can answer questions about programming languages by querying the Stack Exchange API. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."
            ),
            Sample(
                input="Build a Python script that uses the Twitter API to collect tweets containing a specific hashtag, then perform sentiment analysis using the NLTK library and visualize the results with matplotlib. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."
            ),
            Sample(
                input="Create a React Native mobile app that uses the Fixer.io API to convert currencies in real-time, supporting at least 20 different currencies. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."
            ),
            Sample(
                input="Develop a real-time collaborative text editor using Node.js, Socket.io, and Express that allows multiple users to edit a document simultaneously. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."
            ),
            Sample(
                input="Build a Flutter mobile app that uses the Google Maps API and Foursquare API to find and display the nearest coffee shops within a 5km radius of the user's current location. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."
            ),
        ]
    
    return Task(
        dataset=[random.choice(example_samples)],
        solver=[
            system_message(
                "You will be asked to complete a complex programming task. Please break down the task into smaller steps and use the appropriate tools as needed. Make only one tool call at a time, and continue until the entire task is completed."
            ),
            use_tools(bash(), python()),
            generate(),
        ],
        sandbox="docker",
    )


@approver
def bash_allowlist(
    allowed_commands: list[str],
    allow_sudo: bool = False,
    command_specific_rules: dict[str, list[str]] | None = None,
) -> Approver:
    """
    Create an approver that checks if a bash command is in an allowed list.

    Args:
        allowed_commands (List[str]): List of allowed bash commands.
        allow_sudo (bool, optional): Whether to allow sudo commands. Defaults to False.
        command_specific_rules (Optional[Dict[str, List[str]]], optional): Dictionary of command-specific rules. Defaults to None.

    Returns:
        Approver: A function that approves or rejects bash commands based on the allowed list and rules.
    """
    allowed_commands_set = set(allowed_commands)
    command_specific_rules = command_specific_rules or {}
    dangerous_chars = ["&", "|", ";", ">", "<", "`", "$", "(", ")"]

    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: TaskState | None = None,
    ) -> Approval:
        # evaluate the first argument no matter its name (for compatiblity
        # with a broader range of bash command executing tools)
        command = str(next(iter(call.arguments.values()))).strip()
        if not command:
            return Approval(decision="reject", explanation="Empty command")

        try:
            tokens = shlex.split(command)
        except ValueError as e:
            return Approval(
                decision="reject", explanation=f"Invalid command syntax: {str(e)}"
            )

        if any(char in command for char in dangerous_chars):
            return Approval(
                decision="reject",
                explanation=f"Command contains potentially dangerous characters: {', '.join(char for char in dangerous_chars if char in command)}",
            )

        base_command = tokens[0]

        # Handle sudo
        if base_command == "sudo":
            if not allow_sudo:
                return Approval(decision="reject", explanation="sudo is not allowed")
            if len(tokens) < 2:
                return Approval(decision="reject", explanation="Invalid sudo command")
            base_command = tokens[1]
            tokens = tokens[1:]

        if base_command not in allowed_commands_set:
            return Approval(
                decision="escalate",
                explanation=f"Command '{base_command}' is not in the allowed list. Allowed commands: {', '.join(allowed_commands_set)}",
            )

        # Check command-specific rules
        if base_command in command_specific_rules:
            allowed_subcommands = command_specific_rules[base_command]
            if len(tokens) > 1 and tokens[1] not in allowed_subcommands:
                return Approval(
                    decision="escalate",
                    explanation=f"{base_command} subcommand '{tokens[1]}' is not allowed. Allowed subcommands: {', '.join(allowed_subcommands)}",
                )

        return Approval(
            decision="approve", explanation=f"Command '{command}' is approved."
        )

    return approve


@approver
def python_allowlist(
    allowed_modules: list[str],
    allowed_functions: list[str],
    disallowed_builtins: Set[str] | None = None,
    sensitive_modules: Set[str] | None = None,
    allow_system_state_modification: bool = False,
) -> Approver:
    """
    Create an approver that checks if Python code uses only allowed modules and functions, and applies additional safety checks.

    Args:
        allowed_modules (list[str]): List of allowed Python modules.
        allowed_functions (list[str]): List of allowed built-in functions.
        disallowed_builtins (Set[str] | None): Set of disallowed built-in functions.
        sensitive_modules (Set[str] | None): Set of sensitive modules to be blocked.
        allow_system_state_modification (bool): Whether to allow modification of system state.

    Returns:
        Approver: A function that approves or rejects Python code based on the allowed list and rules.
    """
    allowed_modules_set = set(allowed_modules)
    allowed_functions_set = set(allowed_functions)
    disallowed_builtins = disallowed_builtins or {
        "eval",
        "exec",
        "compile",
        "__import__",
        "open",
        "input",
    }
    sensitive_modules = sensitive_modules or {
        "os",
        "sys",
        "subprocess",
        "socket",
        "requests",
    }

    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: TaskState | None = None,
    ) -> Approval:
        # evaluate the first argument no matter its name (for compatiblity
        # with a broader range of python code executing tools)
        code = str(next(iter(call.arguments.values()))).strip()
        if not code:
            return Approval(decision="reject", explanation="Empty code")

        try:
            tree = ast.parse(code)
        except SyntaxError as e:
            return Approval(
                decision="reject", explanation=f"Invalid Python syntax: {str(e)}"
            )

        for node in ast.walk(tree):
            if isinstance(node, ast.Import):
                for alias in node.names:
                    if alias.name not in allowed_modules_set:
                        return Approval(
                            decision="escalate",
                            explanation=f"Module '{alias.name}' is not in the allowed list. Allowed modules: {', '.join(allowed_modules_set)}",
                        )
                    if alias.name in sensitive_modules:
                        return Approval(
                            decision="escalate",
                            explanation=f"Module '{alias.name}' is considered sensitive and not allowed.",
                        )
            elif isinstance(node, ast.ImportFrom):
                if node.module not in allowed_modules_set:
                    return Approval(
                        decision="escalate",
                        explanation=f"Module '{node.module}' is not in the allowed list. Allowed modules: {', '.join(allowed_modules_set)}",
                    )
                if node.module in sensitive_modules:
                    return Approval(
                        decision="escalate",
                        explanation=f"Module '{node.module}' is considered sensitive and not allowed.",
                    )
            elif isinstance(node, ast.Call):
                if isinstance(node.func, ast.Name):
                    if node.func.id not in allowed_functions_set:
                        return Approval(
                            decision="escalate",
                            explanation=f"Function '{node.func.id}' is not in the allowed list. Allowed functions: {', '.join(allowed_functions_set)}",
                        )
                    if node.func.id in disallowed_builtins:
                        return Approval(
                            decision="escalate",
                            explanation=f"Built-in function '{node.func.id}' is not allowed for security reasons.",
                        )

            if not allow_system_state_modification:
                if isinstance(node, ast.Assign):
                    for target in node.targets:
                        if isinstance(target, ast.Attribute) and target.attr.startswith(
                            "__"
                        ):
                            return Approval(
                                decision="escalate",
                                explanation="Modification of system state (dunder attributes) is not allowed.",
                            )

        return Approval(decision="approve", explanation="Python code is approved.")

    return approve


def tool_jsonable(tool_call: ToolCall | None = None) -> dict[str, Any] | None:
    if tool_call is None:
        return None

    return {
        "id": tool_call.id,
        "function": tool_call.function,
        "arguments": tool_call.arguments,
        "type": tool_call.type,
    }

def assistant_message_jsonable(message: ChatMessageAssistant) -> dict[str, Any]:
    def as_jsonable(value: Any) -> Any:
        return to_jsonable_python(value, exclude_none=True, fallback=lambda _x: None)

    message_data = {
        "role": message.role,
        "content": message.content,
        "source": message.source,
        "tool_calls": message.tool_calls
    }

    jsonable = as_jsonable(message_data)
    return deepcopy(jsonable)


@approver
def human_api(approval_api_endpoint: str, agent_id: str, timeout: int = 300) -> Approver:
    """
    Create an approver that calls an external API to approve the tool call.

    Args:
        approval_api_endpoint (str): The URL of the remote approval API.
        agent_id (str): The ID of the agent making the request.
        timeout (int): The maximum time to wait for a response, in seconds. Defaults to 300 seconds (5 minutes).

    Returns:
        Approver: A function that sends approval requests to a remote API for human decision.
    """
    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: TaskState | None = None,
    ) -> Approval:
        print(f"Sending tool call request to user: {call.function} {call.arguments}")

        # Serialize the state and tool call to JSON-compatible format
        
        state_json = state_jsonable(state)
        state_json['tool_choice'] = None #TODO: Fix this
        tool_json = tool_jsonable(call)

        payload = {
            "agent_id": agent_id,
            "task_state": state_json,
            "tool_choice": tool_json,
        }

        try:
            response = requests.post(f'{approval_api_endpoint}/api/review', json=payload)
            response.raise_for_status()
            
            review_id = response.json().get("id")
            if not review_id:
                return Approval(
                    decision="escalate",
                    explanation="Failed to get review ID from initial response",
                )
            
            # Poll the status endpoint until we get a response
            sleep_time = 2
            max_attempts = timeout // sleep_time
            print(f"Waiting for human approval for review {review_id}")
            for _ in range(int(max_attempts)):
                try:
                    status_response = requests.get(f"{approval_api_endpoint}/api/review/status?id={review_id}")
                    status_response.raise_for_status()

                    status_data = status_response.json()
                    print(f"Status data: {status_data}")
                    
                    if status_data.get("status") == "pending":
                        await asyncio.sleep(sleep_time)  # Wait before polling again
                        continue

                    if "decision" in status_data:
                        decision = status_data["decision"]
                        explanation = status_data.get("explanation", "Human provided no explanation.")

                        # Check if there's a modified tool call
                        if "modified_tool_call" in status_data:
                            modified_tool_call_data = status_data["modified_tool_call"]
                            try:
                                modified_tool_call = ToolCall(**modified_tool_call_data)
                                return Approval(
                                    decision=decision,
                                    explanation=explanation,
                                    modified_tool_call=modified_tool_call,
                                )
                            except Exception as e:
                                return Approval(
                                    decision="escalate",
                                    explanation=f"Failed to parse modified tool call: {e}",
                                )
                        else:
                            return Approval(decision=decision, explanation=explanation)

                    return Approval(
                        decision="escalate",
                        explanation=f"Unexpected response from status endpoint: {status_data}",
                    )

                except requests.RequestException as poll_error:
                    print(f"Error polling status: {poll_error}")
                    await asyncio.sleep(sleep_time)  # Wait before retrying
                    continue

            return Approval(
                decision="escalate",
                explanation="Timed out waiting for human approval",
            )

        except requests.RequestException as e:
            return Approval(decision="escalate", explanation=f"Error communicating with remote approver: {str(e)}")

    return approve



@approver
def human_api_sample_n(approval_api_endpoint: str, agent_id: str, n: int = 5, timeout: int = 300) -> Approver:
    """
    Create an approver that generates N tool call suggestions and sends them to an external API for human selection.

    Args:
        approval_api_endpoint (str): The URL of the remote approval API.
        agent_id (str): The ID of the agent making the request.
        model (Model): The model to use for generating tool call suggestions.
        n (int): The number of tool call suggestions to generate. Defaults to 5.
        timeout (int): The maximum time to wait for a response, in seconds. Defaults to 300 seconds (5 minutes).

    Returns:
        Approver: A function that sends multiple tool call suggestions to a remote API for human decision.
    """
    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: TaskState | None = None,
    ) -> Approval:
        if state is None:
            return Approval(decision="escalate", explanation="TaskState is required for this approver.")

        print(f"Generating {n} tool call suggestions for user review")

        model = get_model()
        # model = state.model

        # Generate N tool call suggestions        
        message_without_last_message = deepcopy(state.messages)
        last_messages = [message_without_last_message[-1]]
        tool_options = [tool_jsonable(state.messages[-1].tool_calls[0])] if state.messages[-1].tool_calls[0] else [None]
        message_without_last_message.pop()
        
        for _ in range(n-1):
            model = get_model()
            output = await model.generate(message_without_last_message, tools=state.tools)
            last_messages.append(output.message)
            # output = await model.generate(state.messages, state.tools, state.tool_choice, state.config)
            tool_options.append(tool_jsonable(output.message.tool_calls[0]) if output.message.tool_calls else None)

        # Prepare the payload with multiple tool suggestions
        # tool_options = [tool_jsonable(suggestion.tool_calls[0]) if suggestion.tool_calls else None for suggestion in tool_suggestions]
        last_messages_json = [assistant_message_jsonable(message) for message in last_messages]
        # in case tool calls are None, remove the corresponding last_message
        last_messages_json = [message for message in last_messages_json if message is not None]
        state_json = state_jsonable(state)
        state_json['tool_choice'] = None  # TODO: Fix this

        payload = {
            "agent_id": agent_id,
            "task_state": state_json,
            "tool_options": tool_options,
            "last_messages": last_messages_json,
        }

        try:
            response = requests.post(f'{approval_api_endpoint}/api/review', json=payload)
            response.raise_for_status()
            
            review_id = response.json().get("id")
            if not review_id:
                return Approval(
                    decision="escalate",
                    explanation="Failed to get review ID from initial response",
                )
            
            # Poll the status endpoint until we get a response
            sleep_time = 2
            max_attempts = timeout // sleep_time
            print(f"Waiting for human approval for review {review_id}")
            for _ in range(int(max_attempts)):
                try:
                    status_response = requests.get(f"{approval_api_endpoint}/api/review/status?id={review_id}")
                    status_response.raise_for_status()

                    status_data = status_response.json()
                    print(f"Status data: {status_data}")
                    
                    if status_data.get("status") == "pending":
                        await asyncio.sleep(sleep_time)  # Wait before polling again
                        continue

                    if "decision" in status_data:
                        decision = status_data["decision"]
                        explanation = status_data.get("explanation", "Human provided no explanation.")
                        selected_index = status_data.get("selected_index")
                        print(f"Selected index: {selected_index}")
                        if selected_index is not None and 0 <= selected_index < len(tool_options):
                            selected_tool_call = ToolCall(**tool_options[selected_index])
                            return Approval(
                                decision=decision,
                                explanation=explanation,
                                modified_tool_call=selected_tool_call,
                            )
                        else:
                            return Approval(decision=decision, explanation=explanation)

                    return Approval(
                        decision="escalate",
                        explanation=f"Unexpected response from status endpoint: {status_data}",
                    )

                except requests.RequestException as poll_error:
                    print(f"Error polling status: {poll_error}")
                    await asyncio.sleep(sleep_time)  # Wait before retrying
                    continue

            return Approval(
                decision="escalate",
                explanation="Timed out waiting for human approval",
            )

        except requests.RequestException as e:
            return Approval(decision="escalate", explanation=f"Error communicating with remote approver: {str(e)}")

    return approve

if __name__ == "__main__":
    approval = (Path(__file__).parent / "approval.yaml").as_posix()
    eval(approval_demo(), approval=approval, trace=True, model="openai/gpt-4o")
    