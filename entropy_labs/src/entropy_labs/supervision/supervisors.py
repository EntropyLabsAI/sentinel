from typing import Callable, Optional, Protocol, Any
from .config import (
    SupervisionDecision,
    SupervisionDecisionType,
    SupervisionContext,
    PREFERRED_LLM_MODEL,
)
import inspect
import json
from openai import OpenAI
from inspect_ai.tool import ToolCall
from entropy_labs.sentinel_api_client.sentinel_api_client.client import Client
from uuid import UUID

client = OpenAI()

class Supervisor(Protocol):
    """
    Protocol for supervisor functions.
    Defines the expected signature for supervisor callables.
    """

    def __call__(
        self,
        func: Callable,
        supervision_context: SupervisionContext,
        ignored_attributes: list[str],
        tool_kwargs: dict[str, Any],
        review_id: Optional[UUID],
        **kwargs
    ) -> SupervisionDecision:
        """
        Makes a decision based on the function, its arguments,
        and the supervision context.

        Args:
            func (Callable): The function being supervised.
            supervision_context (SupervisionContext): Additional context.
            ignored_attributes (List[str]): Attributes to ignore.
            tool_kwargs (dict[str, Any]): Keyword arguments for the function.
        Returns:
            SupervisionDecision: The decision made by the supervisor.
        """
        ...

DEFAULT_SYSTEM_PROMPT = (
    "You are an AI safety reviewer. Your task is to evaluate the following function call "
    "and decide whether it should be approved, rejected, escalated, terminated, or modified. "
    "Provide your decision along with an explanation."
)

def llm_supervisor(
    instructions: str,
    supervisor_name: Optional[str] = None,
    description: Optional[str] = None,
    openai_model: str = PREFERRED_LLM_MODEL,
    system_prompt: Optional[str] = None,
    include_context: bool = False
) -> Supervisor:
    """
    Create a supervisor function that uses an LLM to make a supervision decision.
    """
    if system_prompt is None:
        system_prompt = DEFAULT_SYSTEM_PROMPT

    def supervisor(
        func: Callable,
        supervision_context: SupervisionContext,
        ignored_attributes: list[str],
        tool_kwargs: dict[str, Any],
        **kwargs
    ) -> SupervisionDecision:
        """
        LLM supervisor that makes a decision based on the function call, its arguments, and the supervision instructions.
        """
        # Extract function details
        func_name = func.__qualname__
        func_description = func.__doc__ or "No description available."
        try:
            func_implementation = inspect.getsource(func)
        except OSError:
            func_implementation = "Source code not available."
            
        # Prepare tool arguments string
        tool_kwargs_str = ", ".join(
            [f"{k}={v}" for k, v in tool_kwargs.items() if k not in ignored_attributes] +
            [f"{k}=<value hidden> - Assume the value is correct" for k in ignored_attributes]
        )
        if not tool_kwargs_str:
            tool_kwargs_str = "The function does not require any arguments."

        # Prepare the assistant's instructions
        instructions_content = f"""
Instructions:
{instructions}

Function Name:
{func_name}

Function Description:
{func_description}

Function Implementation:
{func_implementation}

Arguments passed to the function: 
{tool_kwargs_str}
"""

        if include_context and supervision_context:
            # Convert SupervisionContext into a textual description
            context_description = supervision_context.to_text()
            instructions_content += f"\nContext of the function call:\n{context_description}\n"

        # Prepare messages
        messages = [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": instructions_content.strip()},
        ]

        # Define the function schema for SupervisionDecision
        supervision_decision_schema = SupervisionDecision.model_json_schema()

        # Prepare the function definition for the OpenAI API
        functions = [
            {
                "name": "supervision_decision",
                "description": "Make a supervision decision for the given function call.",
                "parameters": supervision_decision_schema,
            }
        ]

        try:
            # Call the OpenAI API using sample_from_llm_function_call
            completion = client.chat.completions.create(
                model=openai_model,
                messages=messages,
                functions=functions,
                function_call={"name": "supervision_decision"},
            )

            # Extract the function call arguments from the response
            message = completion.choices[0].message
            if message.function_call:
                response_args = message.function_call.arguments
                response_data = json.loads(response_args)
            else:
                raise ValueError("No valid function call in assistant's response.")

            # Validate the response against the schema
            decision = SupervisionDecision(**response_data)
            return decision

        except Exception as e:
            print(f"Error during LLM supervision: {str(e)}")
            return SupervisionDecision(
                decision=SupervisionDecisionType.ESCALATE,
                explanation=f"Error during LLM supervision: {str(e)}",
                modified=None
            )
    supervisor.__name__ = supervisor_name if supervisor_name else llm_supervisor.__name__
    supervisor.__doc__ = description if description else supervisor.__doc__
    supervisor.supervisor_attributes = {"instructions": instructions, "openai_model": openai_model, "system_prompt": system_prompt, "include_context": include_context}
    return supervisor


def human_supervisor(
    backend_api_endpoint: Optional[str] = None,
    timeout: int = 300,
    n: int = 1,
) -> Supervisor:
    """
    Create a supervisor function that requires human approval via backend API or CLI.

    Args:
        backend_api_endpoint (Optional[str]): Endpoint for backend API for human supervision.
        timeout (int): Timeout in seconds for waiting for the human decision.
        n (int): Number of approvals required.

    Returns:
        Supervisor: A supervisor function that implements human supervision.
    """
    async def supervisor(
        func: Callable,
        supervision_context: SupervisionContext,
        ignored_attributes: list[str],
        tool_kwargs: dict[str, Any],
        review_id: UUID,
        **kwargs
    ) -> SupervisionDecision:
        """
        Human supervisor that requests approval via backend API or CLI.

        Args:
            func (Callable): The function being supervised.
            supervision_context (SupervisionContext): Additional context.
            *args: Positional arguments for the function.
            **kwargs: Keyword arguments for the function.

        Returns:
            SupervisionDecision: The decision made by the supervisor.
        """
        from .common import human_supervisor_wrapper
        from .config import supervision_config

        # Create TaskState from supervision_context
        task_state = supervision_context.to_task_state()

        # Create a ToolCall object representing the function call
        tool_call = ToolCall(
            id="tool_id",  # Use an appropriate ID if available
            function=func.__qualname__,
            arguments=tool_kwargs,
            type='function'
        )

        # Initialize client if needed
        client = supervision_config.client  # Assuming supervision_config is accessible

        # Get the human supervision decision
        supervisor_decision = await human_supervisor_wrapper(
            task_state=task_state,
            call=tool_call,
            backend_api_endpoint=backend_api_endpoint,
            timeout=timeout,
            use_inspect_ai=False,
            n=n,
            review_id=review_id,
            client=client
        )

        return supervisor_decision

    supervisor.__name__ = human_supervisor.__name__
    supervisor.supervisor_attributes = {"backend_api_endpoint": backend_api_endpoint, "timeout": timeout, "n": n}
    return supervisor


def auto_approve_supervisor() -> Supervisor:
    """Creates a supervisor that automatically approves any input."""
    def supervisor(
        func: Callable,
        supervision_context: SupervisionContext,
        ignored_attributes: list[str],
        tool_kwargs: dict[str, Any],
        **kwargs
    ) -> SupervisionDecision:
        return SupervisionDecision(
            decision=SupervisionDecisionType.APPROVE,
            explanation="No supervisor found for this function. It's automatically approved.",
            modified=None
        )
    supervisor.__name__ = auto_approve_supervisor.__name__
    supervisor.supervisor_attributes = {}
    return supervisor
