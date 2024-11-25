"""
Handles project registration with Sentinel.
"""

from typing import Optional, Dict
from sentinel.api.client import APIClient
from sentinel.config import settings
from requests.exceptions import RequestException

def register_project(name: str, metadata: Optional[Dict] = None) -> str:
    """
    Registers a new project with Sentinel.

    Args:
        name (str): Name of the project.
        metadata (Optional[Dict]): Additional metadata for the project.

    Returns:
        str: The newly created project ID.

    Raises:
        ValueError: If the project name already exists or invalid metadata format.
        ConnectionError: For network errors.
    """
    client = APIClient(api_key=settings.api_key)
    payload = {
        "name": name,
        "metadata": metadata or {}
    }

    try:
        response = client.post("/projects", json=payload)
        response.raise_for_status()
        project_id = response.json().get("project_id")
        return project_id
    except RequestException as e:
        raise ValueError(f"Failed to register project: {e}")
