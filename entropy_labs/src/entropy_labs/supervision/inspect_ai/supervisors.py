import logging
import requests
from typing import List, Dict, Optional, Set, Any, Tuple
from inspect_ai.approval import Approval, Approver, approver
from inspect_ai.solver import TaskState
from inspect_ai.tool import ToolCall, ToolCallView
from el.supervision.common import check_bash_command, check_python_code
from el.supervision.inspect_ai._utils import (
    generate_tool_call_change_explanation,
    generate_tool_call_suggestions,
    prepare_payload,
    poll_status
)
from ._config import DEFAULT_TIMEOUT, DEFAULT_SUGGESTIONS

@approver
def bash_approver(
    allowed_commands: List[str],
    allow_sudo: bool = False,
    command_specific_rules: Optional[Dict[str, List[str]]] = None,
) -> Approver:
    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: Optional[TaskState] = None,
    ) -> Approval:
        command = str(next(iter(call.arguments.values()))).strip()
        is_approved, explanation = check_bash_command(command, allowed_commands, allow_sudo, command_specific_rules)
        
        if is_approved:
            return Approval(decision="approve", explanation=explanation)
        else:
            return Approval(decision="escalate", explanation=explanation)

    return approve

@approver
def python_approver(
    allowed_modules: List[str],
    allowed_functions: List[str],
    disallowed_builtins: Optional[Set[str]] = None,
    sensitive_modules: Optional[Set[str]] = None,
    allow_system_state_modification: bool = False,
) -> Approver:
    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: Optional[TaskState] = None,
    ) -> Approval:
        code = str(next(iter(call.arguments.values()))).strip()
        is_approved, explanation = check_python_code(
            code, 
            allowed_modules, 
            allowed_functions, 
            disallowed_builtins, 
            sensitive_modules, 
            allow_system_state_modification
        )
        
        if is_approved:
            return Approval(decision="approve", explanation=explanation)
        else:
            return Approval(decision="escalate", explanation=explanation)

    return approve

@approver
def human_approver(approval_api_endpoint: str, agent_id: str, n: int = DEFAULT_SUGGESTIONS, timeout: int = DEFAULT_TIMEOUT) -> Approver:
    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: Optional[TaskState] = None,
    ) -> Approval:
        if state is None:
            return Approval(decision="escalate", explanation="TaskState is required for this approver.")

        logging.info(f"Generating {n} tool call suggestions for user review")

        last_messages, tool_options = await generate_tool_call_suggestions(state, n)
        payload = prepare_payload(agent_id, state, last_messages, tool_options)

        try:
            response = requests.post(f'{approval_api_endpoint}/api/review/human', json=payload)
            response.raise_for_status()

            review_id = response.json().get("id")
            if not review_id:
                return Approval(decision="escalate", explanation="Failed to get review ID from initial response")

            status_data = await poll_status(approval_api_endpoint, review_id, timeout)

            decision = status_data["decision"]
            explanation = status_data.get("explanation", "Human provided no additional explanation.")
            selected_index = status_data.get("selected_index")

            if decision == "modify":
                modified_tool_call_data = status_data["tool_choice"]
                explanation = generate_tool_call_change_explanation(call, modified_tool_call_data)
                modified_tool_call = ToolCall(**modified_tool_call_data)
            elif selected_index is not None and 0 <= selected_index < len(tool_options):
                modified_tool_call = ToolCall(**tool_options[selected_index])
            else:
                modified_tool_call = None

            return Approval(decision=decision, explanation=explanation, modified=modified_tool_call)

        except requests.RequestException as e:
            logging.error(f"Error communicating with remote approver: {str(e)}")
            return Approval(decision="escalate", explanation=f"Error communicating with remote approver: {str(e)}")
        except TimeoutError:
            return Approval(decision="escalate", explanation="Timed out waiting for human approval")

    return approve

@approver
def llm_approver(approval_api_endpoint: str, agent_id: str, n: int = DEFAULT_SUGGESTIONS, timeout: int = DEFAULT_TIMEOUT) -> Approver:
    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: Optional[TaskState] = None,
    ) -> Approval:
        if state is None:
            return Approval(decision="escalate", explanation="TaskState is required for this approver.")

        logging.info(f"Generating {n} tool call suggestions for LLM review")

        last_messages, tool_options = await generate_tool_call_suggestions(state, n)
        payload = prepare_payload(agent_id, state, last_messages, tool_options)

        assert len(payload['tool_choices']) == 1, "Only one tool call is supported for LLM approval"
        assert len(payload['last_messages']) == 1, "Only one message is supported for LLM approval"

        try:
            response = requests.post(f'{approval_api_endpoint}/api/review/llm', json=payload)
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

                return Approval(decision=decision, explanation=explanation, modified=modified_tool_call)

            return Approval(
                decision="escalate",
                explanation=f"Unexpected response from LLM review endpoint: {review_result}",
            )

        except requests.RequestException as e:
            logging.error(f"Error communicating with LLM approver: {str(e)}")
            return Approval(decision="escalate", explanation=f"Error communicating with LLM approver: {str(e)}")

    return approve
