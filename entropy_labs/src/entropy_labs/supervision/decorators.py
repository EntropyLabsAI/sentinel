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
from entropy_labs.sentinel_api_client.sentinel_api_client.models.supervisor_type import SupervisorType
def supervise(
    mock_policy: Optional[MockPolicy] = None,
    mock_responses: Optional[List[Any]] = None,
    supervision_functions: Optional[List[List[Callable]]] = None,
    ignored_attributes: Optional[List[str]] = None
):
    def decorator(func):
        # Add the function and supervision functions to the registry within the context
        supervision_config.context.add_supervised_function(func, supervision_functions, ignored_attributes)

        @wraps(func)
        def wrapper(*tool_args, **tool_kwargs):
            supervision_context = supervision_config.context
            run_id = supervision_config.run_id
            client = supervision_config.client  # Get the Sentinel API client

            print(f"\n--- Supervision ---")
            print(f"Function Name: {func.__qualname__}")
            print(f"Description: {func.__doc__}")
            print(f"Arguments: {tool_args}, {tool_kwargs.keys()}")

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
                mock_result = handle_mocking(func, effective_mock_policy, mock_responses, *tool_args, **tool_kwargs)
                print(f"Mocking function execution. Mock result: {mock_result}")
                return mock_result
            # Supervision logic
            if not supervisors_list:
                print(f"No supervisors found for function {func.__name__}. Executing function.")
                return func(*tool_args, **tool_kwargs)
            all_decisions = []
            for supervisor_chain in supervisors_list:
                # We send supervision request to the API
                supervisors = supervisor_chain.supervisors
                for supervisor in supervisors:
                    review_id = send_supervision_request(supervisor, func, supervision_context, execution_id, tool_id, tool_args=tool_args, tool_kwargs=tool_kwargs)

                    decision = None
                    supervisor_func = supervision_config.get_supervisor_by_id(supervisor.id)
                    if supervisor_func is None:
                        print(f"No local supervisor function found for ID {supervisor.id}. Skipping.")
                        return None  # Continue to next supervisor

                    # Execute supervisor function
                    decision = call_supervisor_function(supervisor_func, func, supervision_context, review_id=review_id, ignored_attributes=ignored_attributes, tool_args=tool_args, tool_kwargs=tool_kwargs)
                    print(f"Supervisor decision: {decision.decision}")

                    if supervisor.type != SupervisorType.HUMAN_SUPERVISOR:
                        # We send the decision to the API
                        send_review_result(
                            review_id=review_id,
                            execution_id=execution_id,
                            run_id=run_id,
                            tool_id=tool_id,
                            supervisor_id=supervisor.id,
                            decision=decision,
                            client=client,
                            tool_args=tool_args, #TODO: If modified, send modified args and kwargs
                            tool_kwargs=tool_kwargs
                    )
                    # Handle the decision
                    if decision.decision == SupervisionDecisionType.APPROVE:
                        all_decisions.append(decision)
                        break
                    elif decision.decision == SupervisionDecisionType.REJECT:
                        return f"Execution of {func.__qualname__} was rejected. Explanation: {decision.explanation}"
                    elif decision.decision == SupervisionDecisionType.ESCALATE:
                        continue
                        #continue  # Proceed to the next supervisor
                    elif decision.decision == SupervisionDecisionType.MODIFY:
                        all_decisions.append(decision)
                        break
                    elif decision.decision == SupervisionDecisionType.TERMINATE:
                        return f"Execution of {func.__qualname__} was terminated. Explanation: {decision.explanation}"
                    else:
                        print(f"Unknown decision: {decision.decision}. Cancelling execution.")
                        return f"Execution of {func.__qualname__} was cancelled due to an unknown supervision decision."

            # Check decisions and apply modifications if any
            if all(decision.decision in [SupervisionDecisionType.APPROVE, SupervisionDecisionType.MODIFY] for decision in all_decisions):
                print("All decisions approved or modified.")
                # Start with original arguments
                final_args = list(tool_args)
                final_kwargs = tool_kwargs.copy()

                # Apply modifications while respecting ignored_attributes
                for decision in all_decisions:
                    if decision.decision == SupervisionDecisionType.MODIFY and decision.modified:
                        # Update positional arguments
                        # if decision.modified.tool_args:
                        #     for idx, value in enumerate(decision.modified.tool_args):
                        #         param_name = func.__code__.co_varnames[idx]
                        #         if ignored_attributes and param_name in ignored_attributes:
                        #             continue
                        #         final_args[idx] = value
                        #TODO: Fix this - make it work
                        tool_args = decision.modified.tool_args
                        # Update keyword arguments
                        if decision.modified.tool_kwargs:
                            for key, value in decision.modified.tool_kwargs.items():
                                if ignored_attributes and key in ignored_attributes:
                                    continue
                                final_kwargs[key] = value
                        break

                # Call the function with modified arguments
                return func(*final_args, **final_kwargs)
            else:
                return "All supervisors escalated without approval."
        return wrapper
    return decorator


def call_supervisor_function(supervisor_func, func, supervision_context, review_id: UUID, ignored_attributes: List[str], tool_args: List[Any], tool_kwargs: dict[str, Any], decision: Optional[SupervisionDecision] = None):
    if asyncio.iscoroutinefunction(supervisor_func):
        decision = asyncio.run(supervisor_func(
            func, supervision_context=supervision_context, review_id=review_id, ignored_attributes=ignored_attributes, tool_args=tool_args, tool_kwargs=tool_kwargs, decision=decision
        ))
    else:
        decision = supervisor_func(
            func, supervision_context=supervision_context, review_id=review_id, ignored_attributes=ignored_attributes, tool_args=tool_args, tool_kwargs=tool_kwargs, decision=decision
        )
    return decision

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
