from typing import Any, Callable, List, Optional
from functools import wraps
import random

from .config import supervision_config
from ..mocking.policies import MockPolicy
from ..utils.utils import create_random_value
from .config import SupervisionDecision, SupervisionDecisionType
from .llm_sampling import sample_from_llm

def supervise(
    mock_policy: Optional[MockPolicy] = None,
    mock_responses: Optional[List[Any]] = None,
    supervision_functions: Optional[List[Callable]] = None  # List of supervision functions
):
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            print(f"\n--- Supervision ---")
            print(f"Function Name: {func.__name__}")
            print(f"Description: {func.__doc__}")
            print(f"Arguments: {args}, {kwargs}")
            
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
            supervisors = supervision_config.get_supervision_functions(func.__name__, supervision_functions)
            if not supervisors:
                print(f"No supervisors found for function {func.__name__}. Executing function.")
                return func(*args, **kwargs)
            # Remove duplicates while preserving order
            supervisors = list(dict.fromkeys(supervisors))
            
            for supervisor in supervisors:
                decision = supervisor(func, *args, **kwargs)  # Pass the function and its arguments
                print(f"Supervisor {supervisor.__name__} decision: {decision.decision}")
                
                if decision.decision == "approve":
                    print(f"Approved by supervisor {supervisor.__name__}. Executing function.")
                    return func(*args, **kwargs)
                elif decision.decision == "reject":
                    print(f"Rejected by supervisor {supervisor.__name__}. Cancelling execution.")
                    return f"Execution of {func.__name__} was rejected by a supervisor. Explanation: {decision.explanation}"
                elif decision.decision == "escalate":
                    print(f"Escalated by supervisor {supervisor.__name__}. Moving to next supervisor.")
                    continue  # Move to the next supervisor
                elif decision.decision == "modify":
                    print(f"Modified by supervisor {supervisor.__name__}. Executing modified function.")
                    # TODO: Here we want to handle the modified data
                    return decision.modified
                else:
                    print(f"Unknown decision: {decision.decision}. Cancelling execution.")
                    return f"Execution of {func.__name__} was cancelled due to an unknown supervision decision."
            
            # If all supervisors escalated without approval
            print(f"All supervisors escalated without approval. Cancelling {func.__name__} execution.")
            return f"Execution of {func.__name__} was cancelled after all supervision levels."

        return wrapper
    return decorator

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
            return supervision_config.get_mock_response(func.__name__)
        except ValueError as e:
            print(f"Warning: {str(e)}. Falling back to actual function execution.")
            return func(*args, **kwargs)
    elif mock_policy == MockPolicy.SAMPLE_LLM:
        return sample_from_llm(func, *args, **kwargs)
    else:
        raise ValueError(f"Unsupported mock policy: {mock_policy}")
