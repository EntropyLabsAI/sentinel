from typing import Callable, Optional, Protocol
from .config import (
    SupervisionDecision,
    SupervisionDecisionType,
    SupervisionContext,
    PREFERRED_LLM_MODEL,
)
import inspect
import json
from openai import OpenAI

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
        *args,
        **kwargs
    ) -> SupervisionDecision:
        """
        Makes a decision based on the function, its arguments,
        and the supervision context.

        Args:
            func (Callable): The function being supervised.
            supervision_context (SupervisionContext): Additional context.
            *args: Positional arguments for the function.
            **kwargs: Keyword arguments for the function.

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
        *args,
        **kwargs
    ) -> SupervisionDecision:
        # Extract function details
        func_name = func.__name__
        func_description = func.__doc__ or "No description available."
        try:
            func_implementation = inspect.getsource(func)
        except OSError:
            func_implementation = "Source code not available."

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

Function Arguments:
Positional Arguments: {args}
Keyword Arguments: {kwargs}
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
    return supervisor
