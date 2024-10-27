import requests
import logging
import asyncio
from typing import List, Any, Dict
from inspect_ai.solver import TaskState
from entropy_labs.supervision.config import SupervisionDecision, SupervisionDecisionType
from inspect_ai.solver._task_state import state_jsonable
from entropy_labs.supervision.inspect_ai._config import SLEEP_TIME, FRONTEND_URL
from inspect_ai.tool import ToolCall
from entropy_labs.supervision.inspect_ai.utils import (
    generate_tool_call_change_explanation,
    generate_tool_call_suggestions,
    chat_message_jsonable
)
from rich.console import Console
from inspect_ai.util._console import input_screen

# Create an asyncio.Lock to prevent concurrent console access
_console_lock = asyncio.Lock()

def prepare_payload(agent_id: str, task_state: TaskState, last_messages: List[Any], tool_options: List[Any]) -> Dict[str, Any]:
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

    state_json = state_jsonable(task_state)
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
            status_response = requests.get(f"{approval_api_endpoint}/api/review/status/{review_id}")
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


async def get_human_supervision_decision_api(
    backend_api_endpoint: str,
    agent_id: str,
    task_state: TaskState,
    call: ToolCall,
    n: int = 1,
    timeout: int = 300,
    use_inspect_ai: bool = False
) -> SupervisionDecision:
    """Get the supervision decision from the backend API."""

    # Generate tool call suggestions
    if use_inspect_ai:
        last_messages, tool_options = await generate_tool_call_suggestions(
            task_state=task_state, n=n, call=call)
    else:
        if n > 1:
            logging.warning("n>1 is not supported for human approval outside of inspect_ai, using n=1 instead.")
        last_messages, tool_options = await generate_tool_call_suggestions(
            task_state=task_state, n=1, call=call)
    payload = prepare_payload(agent_id, task_state, last_messages, tool_options)

    if use_inspect_ai:
        async with _console_lock:
            with input_screen(width=None) as console:
                # Try setting record=True, if possible
                console.record = True  # Note: Only if console.record is settable
                return await _get_supervision_decision(
                    console, backend_api_endpoint, payload, timeout, call, tool_options
                )
    else:
        console = Console(record=True)
        return await _get_supervision_decision(
            console, backend_api_endpoint, payload, timeout, call, tool_options
        )

async def _get_supervision_decision(
    console: Console,
    backend_api_endpoint: str,
    payload: Dict[str, Any],
    timeout: int,
    call: ToolCall,
    tool_options: List[Any]
) -> SupervisionDecision:
    """Helper function to get supervision decision and handle exceptions."""
    try:
        # Send review request to the backend API
        response = requests.post(f"{backend_api_endpoint}/api/review/human", json=payload)
        response.raise_for_status()

        review_id = response.json().get("id")
        if not review_id:
            message = "Failed to get review ID from initial response"
            console.print(f"[bold red]{message}[/bold red]")
            return SupervisionDecision(
                decision=SupervisionDecisionType.ESCALATE,
                explanation=message
            )

        # Process supervision decision
        return await _process_supervision_decision(
            console, backend_api_endpoint, review_id, timeout, call, tool_options
        )
    except requests.RequestException as e:
        message = f"Error communicating with remote approver: {str(e)}"
        console.print(f"[bold red]{message}[/bold red]")
        return SupervisionDecision(
            decision=SupervisionDecisionType.ESCALATE,
            explanation=message
        )
    except TimeoutError:
        message = "Timed out waiting for human approval"
        console.print(f"[bold red]{message}[/bold red]")
        return SupervisionDecision(
            decision=SupervisionDecisionType.ESCALATE,
            explanation=message
        )

async def _process_supervision_decision(
    console: Console,
    backend_api_endpoint: str,
    review_id: str,
    timeout: int,
    call: ToolCall,
    tool_options: List[Any]
) -> SupervisionDecision:
    """Helper function to process the supervision decision."""
    _display_review_sent_message(console, backend_api_endpoint, review_id)
    # Notify the user that we're waiting for approval
    console.print(f"[bold yellow]Waiting for human approval...[/bold yellow]")

    # Poll for status updates
    status_data = await poll_status(backend_api_endpoint, review_id, timeout)

    decision = status_data["decision"]
    explanation = status_data.get("explanation", "Human provided no additional explanation.")
    selected_index = status_data.get("selected_index")
    console.print(
        f"Received decision: {decision}, explanation: {explanation}, selected_index: {selected_index}"
    )

    if decision == "modify":
        modified_tool_call_data = status_data["tool_choice"]
        explanation = generate_tool_call_change_explanation(call, modified_tool_call_data)
        modified_tool_call = ToolCall(**modified_tool_call_data)
        console.print(f"Modified tool call data: {modified_tool_call_data}")
    elif selected_index is not None and 0 <= selected_index < len(tool_options):
        modified_tool_call = ToolCall(**tool_options[selected_index])
        console.print(f"Selected modified tool call from options at index {selected_index}")
    else:
        modified_tool_call = None
        console.print("No modified tool call provided.")

    console.print(
        f"Returning SupervisionDecision with decision: {decision}, explanation: {explanation}"
    )

    return SupervisionDecision(
        decision=decision,
        explanation=explanation,
        modified=modified_tool_call
    )

def _display_review_sent_message(console: Console, backend_api_endpoint: str, review_id: str):
    """Helper function to display the review sent message."""
    message = (
        f"[bold green]Review has been sent to the server for human approval.[/bold green]\n"
        f"You can view the review at: {FRONTEND_URL}/supervisor/human\n"
        f"Review ID: {review_id}"
    )
    console.print(message)

async def get_llm_supervision_decision_api(backend_api_endpoint: str, agent_id: str, task_state:      TaskState, call: ToolCall, n: int = 1, timeout: int = 300):
    """ Get the supervision decision from the backend API """
    #TODO: This will probably be deprecated soon

    logging.info(f"Generating {n} tool call suggestions for LLM review")

    last_messages, tool_options = await generate_tool_call_suggestions(task_state=task_state, n=n, call=call)
    payload = prepare_payload(agent_id, task_state, last_messages, tool_options)

    assert len(payload['tool_choices']) == 1, "Only one tool call is supported for LLM approval"
    assert len(payload['last_messages']) == 1, "Only one message is supported for LLM approval"

    try:
        response = requests.post(f'{backend_api_endpoint}/api/review/llm', json=payload)
        response.raise_for_status()

        review_result = response.json()

        if "decision" in review_result:
            decision = review_result["decision"]
            explanation = review_result.get("reasoning", "LLM provided no additional explanation.")

            if decision == "modify":
                modified_tool_call_data = review_result["tool_choice"]
                explanation = generate_tool_call_change_explanation(call, modified_tool_call_data)
                modified_tool_call = ToolCall(**modified_tool_call_data)
            else:
                modified_tool_call = None

            return SupervisionDecision(decision=decision, explanation=explanation, modified=modified_tool_call)

        return SupervisionDecision(
            decision=SupervisionDecisionType.ESCALATE,
            explanation=f"Unexpected response from LLM review endpoint: {review_result}",
        )

    except requests.RequestException as e:
        logging.error(f"Error communicating with LLM approver: {str(e)}")
        return SupervisionDecision(decision=SupervisionDecisionType.ESCALATE, explanation=f"Error communicating with LLM approver: {str(e)}")
