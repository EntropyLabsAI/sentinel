"""
Wrapper for the OpenAI client to intercept requests and responses.
"""

import openai
from sentinel.api.client import APIClient
from sentinel.config import settings

# Store the original method
_original_completion_create = openai.chat.completions.create

def patch_openai():
    """
    Patches the OpenAI Completion.create method to intercept requests and responses.
    """
    openai.chat.completions.create = wrapped_completion_create

def wrapped_completion_create(*args, **kwargs):
    """
    Wrapped method for openai.Completion.create.
    """
    print("wrapped_completion_create")
    request_data = {
        "prompt": kwargs.get("prompt", "Test prompt"),
        "model": kwargs.get("model", "gpt-4o-mini"),
        "temperature": kwargs.get("temperature", 1.0),
        "max_tokens": kwargs.get("max_tokens", 256),
        "additional_params": {k: v for k, v in kwargs.items() if k not in ["prompt", "model", "temperature", "max_tokens"]}
    }

    submit_request(request_data)

    # Call the original method
    response = _original_completion_create(*args, **kwargs)

    submit_response(response)

    return response

def submit_request(request_data):
    client = APIClient(api_key=settings.api_key)
    client.post("/requests", json=request_data)

def submit_response(response_data):
    client = APIClient(api_key=settings.api_key)
    client.post("/responses", json=response_data)
