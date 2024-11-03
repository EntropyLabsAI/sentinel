import asyncio
from typing import Any, Callable, List, Optional
from functools import wraps

from .config import supervision_config, SupervisionDecisionType, SupervisionDecision
from ..mocking.policies import MockPolicy
from ..utils.utils import create_random_value
from .llm_sampling import sample_from_llm
from entropy_labs.api.sentinel_api_client_helper import create_execution, get_supervisors_for_tool, send_supervision_request, send_review_result
import random
from uuid import UUID

def supervise(
    mock_policy: Optional[MockPolicy] = None,
    mock_responses: Optional[List[Any]] = None,
    supervision_functions: Optional[List[Callable]] = None,
    ignored_attributes: Optional[List[str]] = None
):
    def decorator(func):
        # Add the function and supervision functions to the registry within the context
        supervision_config.context.add_supervised_function(func, supervision_functions, ignored_attributes)

        @wraps(func)
        def wrapper(*args, **kwargs):
            supervision_context = supervision_config.context
            run_id = supervision_config.run_id
            client = supervision_config.client  # Get the Sentinel API client

            print(f"\n--- Supervision ---")
            print(f"Function Name: {func.__qualname__}")
            print(f"Description: {func.__doc__}")
            print(f"Arguments: {args}, {kwargs.keys()}")

            # Retrieve the tool_id from the registry
            entry = supervision_context.get_supervised_function_entry(func)
            if entry:
                tool_id = entry['tool_id']
                ignored_attributes = entry['ignored_attributes']
            else:
                raise Exception(f"Tool ID for function {func.__name__} not found in the registry.")
            
            execution_id = create_execution(tool_id, run_id, client)
            supervisors_list = get_supervisors_for_tool(tool_id, run_id, client)

            # Determine effective mock policy
            effective_mock_policy = (
                supervision_config.global_mock_policy
                if supervision_config.override_local_policy
                else (mock_policy or supervision_config.global_mock_policy)
            )
            # Handle mocking
            if effective_mock_policy != MockPolicy.NO_MOCK:
                mock_result = handle_mocking(func, effective_mock_policy, mock_responses, *args, **kwargs)
                print(f"Mocking function execution. Mock result: {mock_result}")
                return mock_result
            # Supervision logic
            if not supervisors_list:
                print(f"No supervisors found for function {func.__name__}. Executing function.")
                return func(*args, **kwargs)

            for supervisor in supervisors_list:
                # We send supervision request to the API
                review_id = send_supervision_request(supervisor, func, supervision_context, execution_id, tool_id, *args, **kwargs)

                decision = None
                supervisor_func = supervision_config.get_supervisor_by_id(supervisor.id)
                if supervisor_func is None:
                    print(f"No local supervisor function found for ID {supervisor.id}. Skipping.")
                    return None  # Continue to next supervisor

                # Execute supervisor function
                decision = call_supervisor_function(supervisor_func, func, supervision_context, review_id=review_id, ignored_attributes=ignored_attributes, tool_kwargs=kwargs)
                print(f"Supervisor decision: {decision.decision}")

                # We send the decision to the API
                send_review_result(
                        review_id=review_id,
                        execution_id=execution_id,
                        run_id=run_id,
                        tool_id=tool_id,
                        supervisor_id=supervisor.id,
                        decision=decision,
                        client=client,
                        *args,
                        **kwargs
                )
                # Handle the decision
                if decision.decision == SupervisionDecisionType.APPROVE:
                    return func(*args, **kwargs)
                elif decision.decision == SupervisionDecisionType.REJECT:
                    return f"Execution of {func.__qualname__} was rejected. Explanation: {decision.explanation}"
                elif decision.decision == SupervisionDecisionType.ESCALATE:
                    continue
                elif decision.decision == SupervisionDecisionType.MODIFY:
                    print(f"Modified. Executing modified function.")
                    modified_args = decision.modified.get('args', args)
                    modified_kwargs = decision.modified.get('kwargs', kwargs)
                    return func(*modified_args, **modified_kwargs)
                elif decision.decision == SupervisionDecisionType.TERMINATE: #TODO: Agent termination does not work for all agents
                    return f"Execution of {func.__qualname__} was terminated. Explanation: {decision.explanation}"
                else:
                    print(f"Unknown decision: {decision.decision}. Cancelling execution.")
                    return f"Execution of {func.__qualname__} was cancelled due to an unknown supervision decision."
            # If all supervisors escalated without approval
            print(f"All supervisors escalated without approval. Cancelling {func.__qualname__} execution.")
            return f"Execution of {func.__qualname__} was cancelled after all supervision levels."
        return wrapper
    return decorator


def call_supervisor_function(supervisor_func, func, supervision_context, review_id: UUID, ignored_attributes: List[str], tool_kwargs: dict[str, Any]):
    if asyncio.iscoroutinefunction(supervisor_func):
        decision = asyncio.run(supervisor_func(
            func, supervision_context=supervision_context, review_id=review_id, ignored_attributes=ignored_attributes, tool_kwargs=tool_kwargs
        ))
    else:
        decision = supervisor_func(
            func, supervision_context=supervision_context, review_id=review_id, ignored_attributes=ignored_attributes, tool_kwargs=tool_kwargs
        )
    return decision

def handle_supervision_decision(decision: SupervisionDecision, func: Callable, *args, **kwargs):
    print(f"Supervisor decision: {decision.decision}")
    if decision.decision == SupervisionDecisionType.APPROVE:
        print(f"Approved. Executing function.")
        return func(*args, **kwargs)
    elif decision.decision == SupervisionDecisionType.REJECT:
        print(f"Rejected. Cancelling execution.")
        return f"Execution of {func.__name__} was rejected. Explanation: {decision.explanation}"
    elif decision.decision == SupervisionDecisionType.ESCALATE:
        print(f"Escalated. Moving to next supervisor.")
        return None  # Continue to next supervisor
    elif decision.decision == SupervisionDecisionType.MODIFY:
        print(f"Modified. Executing modified function.")
        # Assuming modified data is in decision.modified
        modified_args = decision.modified.get('args', args)
        modified_kwargs = decision.modified.get('kwargs', kwargs)
        return func(*modified_args, **modified_kwargs)
    elif decision.decision == SupervisionDecisionType.TERMINATE:
        print(f"Terminated. Cancelling execution.")
        return f"Execution of {func.__name__} was terminated. Explanation: {decision.explanation}"
    else:
        print(f"Unknown decision: {decision.decision}. Cancelling execution.")
        return f"Execution of {func.__name__} was cancelled due to an unknown supervision decision."

def handle_mocking(func, mock_policy, mock_responses, *args, **kwargs):
    """Handle different mock policies."""
    if mock_policy == MockPolicy.NO_MOCK:
        return func(*args, **kwargs)
    elif mock_policy == MockPolicy.SAMPLE_LIST:
        if mock_responses:
            return random.choice(mock_responses)
        else:
            raise ValueError("No mock responses provided for SAMPLE_LIST policy")
    elif mock_policy == MockPolicy.SAMPLE_RANDOM:
        tool_return_type = func.__annotations__.get('return', None)
        if tool_return_type:
            return create_random_value(tool_return_type)
        else:
            raise ValueError("No return type specified for the function")
    elif mock_policy == MockPolicy.SAMPLE_PREVIOUS_CALLS:
        try:
            return supervision_config.get_mock_response(func.__name__)  # TODO: Make sure this works
        except ValueError as e:
            print(f"Warning: {str(e)}. Falling back to actual function execution.")
            return func(*args, **kwargs)
    elif mock_policy == MockPolicy.SAMPLE_LLM:
        return sample_from_llm(func, *args, **kwargs)
    else:
        raise ValueError(f"Unsupported mock policy: {mock_policy}")
