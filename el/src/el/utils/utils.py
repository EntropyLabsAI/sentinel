from typing import Any, get_origin, get_args, Callable, Dict, Any, Union, Optional
import random
import string
import inspect
import requests
import os
from ..supervision.config import SupervisionDecision, SupervisionDecisionType


def create_random_value(return_type: type) -> Any:
    origin = get_origin(return_type)
    args = get_args(return_type)

    if origin is None:
        if return_type == int:
            return random.randint(-1000, 1000)
        elif return_type == float:
            return random.uniform(-1000.0, 1000.0)
        elif return_type == str:
            return ''.join(random.choices(string.ascii_letters + string.digits, k=10))
        elif return_type == bool:
            return random.choice([True, False])
        else:
            raise ValueError(f"Unsupported simple type: {return_type}")
    elif origin is list:
        return [create_random_value(args[0]) for _ in range(random.randint(1, 5))]
    elif origin is dict:
        key_type, value_type = args
        return {create_random_value(key_type): create_random_value(value_type) for _ in range(random.randint(1, 5))}
    elif origin is Union:
        return create_random_value(random.choice(args))
    elif origin is Optional:
        return random.choice([None, create_random_value(args[0])])
    else:
        raise ValueError(f"Unsupported complex type: {return_type}")

def get_function_code(func: Callable) -> str:
    """Retrieve the source code of a function."""
    try:
        return inspect.getsource(func)
    except Exception as e:
        return f"Error retrieving source code: {str(e)}"

def prompt_user_input_or_api(data: Dict[str, Any]) -> SupervisionDecision:
    """Prompt the user for input via CLI or send data to a backend API for approval."""
    # Check if backend API endpoint is set
    # TODO: UPDATE THIS
    backend_api_endpoint = os.getenv('BACKEND_API_ENDPOINT') #TODO: Make this configurable
    if backend_api_endpoint:
        # Send the data to the backend API
        response = requests.post(f'{backend_api_endpoint}/api/approve', json=data)
        if response.status_code == 200:
            decision_str = response.json().get('decision', 'escalate').lower()
        else:
            decision_str = 'escalate'
    else:
        # Fallback to CLI input
        while True:
            decision_str = input(f"Do you approve execution of {data['function']} with arguments {data['args']} and {data['kwargs']}? (approve/reject/escalate): ").strip().lower()
            if decision_str in ['approve', 'reject', 'escalate']:
                break
            else:
                print("Invalid input. Please enter 'approve', 'reject', or 'escalate'.")

    return SupervisionDecision(decision=decision_str, explanation="User or API decision")
