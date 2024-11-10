from uuid import UUID, uuid4
from entropy_labs.sentinel_api_client.sentinel_api_client.client import Client
from entropy_labs.sentinel_api_client.sentinel_api_client.api.project.create_project import (
    sync_detailed as create_project_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.api.run.create_project_run import (
    sync_detailed as create_run_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.create_project_body import CreateProjectBody
from entropy_labs.api.sentinel_api_client_helper import register_tools_and_supervisors
import logging
from entropy_labs.supervision.config import supervision_config
from typing import Optional, List, Callable
import yaml
import fnmatch

def register_project(project_name: str, entropy_labs_backend_url: str) -> UUID:
    """
    Registers a new project using the Sentinel API.

    Args:
        project_name (str): The name of the project to create.
        entropy_labs_backend_url (str): The URL of the Entropy Labs backend.

    Returns:
        UUID: The project ID.
    """
    client = Client(base_url=entropy_labs_backend_url)
    # Set the client in the supervision config
    supervision_config.client = client

    # Create new project
    project_data = CreateProjectBody(name=project_name)

    response = create_project_sync_detailed(client=client, body=project_data)
    if (
        response.status_code in [200, 201]
        and response.parsed is not None
    ):
        # Add the project using the new method
        if isinstance(response.parsed, UUID):
            supervision_config.add_project(project_name, response.parsed)
            return response.parsed
        else:
            raise Exception("Unexpected response type. Expected UUID.")
    else:
        raise Exception(f"Failed to create project. Status code: {response.status_code}")


def register_task(project_id: UUID, task_name: str) -> UUID:
    """
    Registers a new task under a project using the Sentinel API.

    Args:
        project_id (UUID): The ID of the project.
        task_name (str): The name of the task.

    Returns:
        UUID: The task ID.
    """
    # Retrieve project by ID
    project = supervision_config.get_project_by_id(project_id)
    if not project:
        raise ValueError(f"Project with ID '{project_id}' not found in supervision config.")
    project_name = project.project_name

    # Generate a new task ID (replace with API call if available)
    task_id = uuid4()  # TODO: Implement with backend API

    # Add the task to the project
    supervision_config.add_task(project_name, task_name, task_id)
    return task_id


def create_run(project_id: UUID, task_id: UUID, run_name: Optional[str] = None, entropy_labs_backend_url: Optional[str] = None,
               tools: Optional[List[Callable]] = None) -> UUID:
    """
    Creates a new run for a task under a project using the Sentinel API.

    Args:
        project_id (UUID): The ID of the project.
        task_id (UUID): The ID of the task.
        run_name (Optional[str]): The name of the run.
        entropy_labs_backend_url (Optional[str]): The URL of the Entropy Labs backend.
        tools (Optional[List[Callable]]): The tools to register. If None, all tools with @supervise() decorators are registered.

    Returns:
        UUID: The run ID.
    """
    if run_name is None:
        run_name = f"run-{uuid4()}" #TODO: Have fun run names
    
    if entropy_labs_backend_url is None:
        client = supervision_config.client
        if client is None:
            raise Exception("Client not set. Please provide the entropy_labs_backend_url or set the client in the supervision config.")
    else:
        client = Client(base_url=entropy_labs_backend_url)
        supervision_config.client = client

    # Retrieve project and task by IDs
    project = supervision_config.get_project_by_id(project_id)
    if not project:
        raise ValueError(f"Project with ID '{project_id}' not found in supervision config.")
    project_name = project.project_name

    task = supervision_config.get_task_by_id(task_id)
    if not task:
        raise ValueError(f"Task with ID '{task_id}' not found in supervision config.")
    if task.task_name not in project.tasks:
        raise ValueError(f"Task '{task.task_name}' does not belong to project '{project_name}'.")
    task_name = task.task_name

    # Create the run using the API
    response = create_run_sync_detailed(project_id=project_id, client=client)  # TODO: Add task_id when API supports it

    if (
        response.status_code in [200, 201]
        and response.parsed is not None
    ):
        run_id = response.parsed

        # Automatically register tools and supervisors
        register_tools_and_supervisors(run_id, tools)
        logging.info(f"Tools and supervisors registered for run {run_id}")

        # Add the run to the task
        supervision_config.add_run(project_name, task_name, run_name, run_id)

        return run_id
    else:
        raise Exception(f"Failed to create run. Status code: {response.status_code}")

        
def register_inspect_approvals(run_id: UUID, approval_file: str):
    """
    Reads the inspect approval YAML file and registers the approvals and tools for the run.
    
    Args:
        run_id (UUID): The ID of the run.
        approval_file (str): The path to the inspect approval YAML file.
    """
    from inspect_ai._util.registry import registry_find
    
    # Read the approval file
    with open(approval_file, 'r') as file:
        approvals = yaml.safe_load(file)

    client = supervision_config.client
    if client is None:
        raise Exception("Client not set in the supervision config. Please set the client before calling this function.")

    supervision_context = supervision_config.context

    supervised_tools = {}
    # For each approver in the approval file
    for approver in approvals.get('approvers', []):
        # Get the tools pattern (may be a wildcard)
        tools_pattern = approver.get('tools', '*')

        # Find tools from the registry matching the pattern
        tools = registry_find(lambda x: x.type == "tool")

        # Filter tools based on tools_pattern
        matching_tools = []
        for tool in tools:
            tool_name = tool.__registry_info__.name
            if fnmatch.fnmatch(tool_name, tools_pattern):
                matching_tools.append(tool)

        # For each matching tool, add supervised function to the supervision context
        for tool in matching_tools:
            # The tool is a function decorated with @tool in inspect_ai
            func = tool

            # Get the supervisor function (approval function)
            approver_name = approver.get('name')
            approval_funcs = registry_find(lambda x: x.type == "approver" and x.name == approver_name)

            if not approval_funcs:
                logging.warning(f"Approval function '{approver_name}' not found in the registry.")
                continue

            approval_func = approval_funcs[0]

            # Configure the approval function with attributes from the approval file
            supervisor_attributes = {k: v for k, v in approver.items() if k not in ['name', 'tools']}
            approval_func_initialised = approval_func(**supervisor_attributes)

            if func not in supervised_tools:
                supervised_tools[func] = [approval_func_initialised]
            else:
                supervised_tools[func].append(approval_func_initialised)

    for tool_func in supervised_tools:
        supervision_functions = supervised_tools[tool_func]
        supervision_context.add_supervised_function(
            func=tool_func,
            supervision_functions=[supervision_functions],
            ignored_attributes=[]
        )

    # Register the tools and supervisors with the Sentinel API
    register_tools_and_supervisors(run_id=run_id)
    
    