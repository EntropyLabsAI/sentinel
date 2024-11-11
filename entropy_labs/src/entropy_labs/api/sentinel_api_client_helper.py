import asyncio
from entropy_labs.supervision.config import (
    SupervisionDecision,
    SupervisionDecisionType,
    supervision_config,
)
from langchain_core.tools.structured import StructuredTool
from entropy_labs.supervision.inspect_ai._config import FRONTEND_URL
from entropy_labs.supervision.supervisors import auto_approve_supervisor
from rich.console import Console
from inspect_ai.util._console import input_screen
import time
from entropy_labs.sentinel_api_client.sentinel_api_client.client import Client
from entropy_labs.sentinel_api_client.sentinel_api_client.api.supervision.get_supervision_request_status import (
    sync_detailed as get_supervision_status_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.api.supervision.get_supervision_result import (
    sync_detailed as get_supervision_result_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.supervision_result import (
    SupervisionResult
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.status import Status
from entropy_labs.sentinel_api_client.sentinel_api_client.types import UNSET
from entropy_labs.sentinel_api_client.sentinel_api_client.models.supervisor_type import (
    SupervisorType,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.api.tool.create_run_tool import (
    sync_detailed as create_tool_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.api.supervisor.create_supervisor import (
    sync_detailed as create_supervisor_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.supervisor import Supervisor
from entropy_labs.sentinel_api_client.sentinel_api_client.api.supervisor.create_tool_supervisor_chains import (
    sync_detailed as create_tool_supervisor_chains_sync_detailed,
)

from entropy_labs.sentinel_api_client.sentinel_api_client.api.supervisor.get_tool_supervisor_chains import (
    sync_detailed as get_tool_supervisor_chains_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.api.request_group.create_tool_request_group import (
    sync_detailed as create_tool_request_group_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.api.request_group.get_run_request_groups import (
    sync_detailed as get_request_groups_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.api.supervision.create_supervision_request import (
    sync_detailed as create_supervision_request_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.api.supervision.create_supervision_result import (
    sync_detailed as create_supervision_result_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.api.request_group.create_tool_request import (
    sync_detailed as create_tool_request_sync_detailed,
)

from entropy_labs.sentinel_api_client.sentinel_api_client.models.decision import Decision
from entropy_labs.sentinel_api_client.sentinel_api_client.models.tool_attributes import (
    ToolAttributes,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.tool_request import (
    ToolRequest,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.tool_request_group import (
    ToolRequestGroup,
)

from entropy_labs.sentinel_api_client.sentinel_api_client.models.create_run_tool_body import (
    CreateRunToolBody,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.create_run_tool_body_attributes import (
    CreateRunToolBodyAttributes,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.supervision_request import (
    SupervisionRequest,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.supervisor_attributes import (
    SupervisorAttributes,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.chain_request import (
    ChainRequest,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.supervisor_chain import SupervisorChain
from uuid import uuid4, UUID
from datetime import datetime, timezone
import inspect
from entropy_labs.utils.utils import get_function_code
from typing import List, Optional, Any, Callable

# Create an asyncio.Lock to prevent concurrent console access
_console_lock = asyncio.Lock()

def register_tools_and_supervisors(run_id: UUID, tools: Optional[List[Callable | StructuredTool]] = None):
    """
    Registers tools and supervisors with the backend API.
    """
    # Access the registries from the context
    supervision_context = supervision_config.context
    client = supervision_config.client
    project_id = list(supervision_config.projects.values())[0].project_id # TODO: Make sure this is correct, build better check
    
    if tools is None: #TODO: Make sure this is correct
        # If no tools are provided, register all tools and supervisors
        supervised_functions = supervision_context.supervised_functions_registry
    else:
        # If list of tools is provided, only register the tools and supervisors for the provided tools  
        supervised_functions = {}
        for tool in tools:
            # Check if tool is StructuredTool
            if isinstance(tool, StructuredTool):
                supervised_functions[tool.func.__qualname__] = supervision_context.supervised_functions_registry[tool.func.__qualname__]
            else:
                supervised_functions[tool.__qualname__] = supervision_context.supervised_functions_registry[tool.__qualname__]


    for tool_name, data in supervised_functions.items():
        supervision_functions = data['supervision_functions']
        ignored_attributes = data['ignored_attributes']
        func = data['function']
        
        # Add the run_id to the supervised function
        supervision_context.add_run_id_to_supervised_function(func, run_id)

        # Extract function arguments using inspect
        func_signature = inspect.signature(func)
        func_arguments = {
            param.name: str(param.annotation) if param.annotation is not param.empty else 'Any'
            for param in func_signature.parameters.values()
        }

        # Pass the extracted arguments to ToolAttributes.from_dict
        attributes = CreateRunToolBodyAttributes.from_dict(src_dict=func_arguments)

        # Register the tool
        tool_data = CreateRunToolBody(
            name=tool_name,
            description=str(func.__doc__) if func.__doc__ else tool_name,
            attributes=attributes,
            ignored_attributes=ignored_attributes
        )
        tool_response = create_tool_sync_detailed(
            client=client,
            run_id=run_id,
            body=tool_data
        )
        if (
            tool_response.status_code in [200, 201] and
            tool_response.parsed is not None
        ):
            # Update the tool_id in the registry 
            tool_id = tool_response.parsed
            supervision_context.update_tool_id(func, tool_id)
            print(f"Tool '{tool_name}' registered with ID: {tool_id}")
        else:
            raise Exception(f"Failed to register tool '{tool_name}'. Status code: {tool_response}")

        # Register supervisors and associate them with the tool
        supervisor_chain_ids: List[List[UUID]] = []
        if supervision_functions == []:
            supervisor_chain_ids.append([])
            supervisor_func = auto_approve_supervisor()
            supervisor_info: dict[str, Any] = {
                    'func': supervisor_func,
                    'name': getattr(supervisor_func, '__name__', 'supervisor_name'),
                    'description': getattr(supervisor_func, '__doc__', 'supervisor_description'),
                    'type': SupervisorType.NO_SUPERVISOR,
                    'code': get_function_code(supervisor_func),
                    'supervisor_attributes': getattr(supervisor_func, 'supervisor_attributes', {})
            }
            supervisor_id = register_supervisor(client, supervisor_info, project_id, supervision_context)
            supervisor_chain_ids[0] = [supervisor_id]
        else:
            for idx, supervisor_func_list in enumerate(supervision_functions):
                supervisor_chain_ids.append([])
                for supervisor_func in supervisor_func_list:
                    supervisor_info: dict[str, Any] = {
                        'func': supervisor_func,
                        'name': getattr(supervisor_func, '__name__', None) or 'supervisor_name',
                        'description': getattr(supervisor_func, '__doc__', None) or 'supervisor_description',
                        'type': SupervisorType.HUMAN_SUPERVISOR if getattr(supervisor_func, '__name__', 'supervisor_name') in ['human_supervisor', 'human_approver'] else SupervisorType.CLIENT_SUPERVISOR,
                        'code': get_function_code(supervisor_func),
                        'supervisor_attributes': getattr(supervisor_func, 'supervisor_attributes', {})
                    }
                    supervisor_id = register_supervisor(client, supervisor_info, project_id, supervision_context)
                    supervisor_chain_ids[idx].append(supervisor_id)

        # Ensure tool_id is a UUID before proceeding
        if tool_id is UNSET or not isinstance(tool_id, UUID):
            raise ValueError("Invalid tool_id: Expected UUID")

        print(f"Associating supervisors with tool '{tool_name}' for run ID {run_id}")
        if supervisor_chain_ids:
            chain_requests = [ChainRequest(supervisor_ids=supervisor_ids) for supervisor_ids in supervisor_chain_ids]
            association_response = create_tool_supervisor_chains_sync_detailed(
                tool_id=tool_id,
                client=client,
                body=chain_requests
            )
            if association_response.status_code in [200, 201]:
                print(f"Supervisors assigned to tool '{tool_name}' for run ID {run_id}")
            else:
                raise Exception(f"Failed to assign supervisors to tool '{tool_name}'. Response: {association_response}")
        else:
                print(f"No supervisors to assign to tool '{tool_name}'")



def wait_for_human_decision(supervision_request_id: UUID, client: Client, timeout: int = 300) -> Status:
    start_time = time.time()

    while True:
        try:
            response = get_supervision_status_sync_detailed(
                client=client,
                supervision_request_id=supervision_request_id
            )
            if response.status_code == 200 and response.parsed is not None:
                status = response.parsed.status
                if isinstance(status, Status) and status in [Status.FAILED, Status.COMPLETED, Status.TIMEOUT]:
                    # Map status to SupervisionDecision
                    print(f"Polling for human decision completed. Status: {status}")
                    return status
                else:
                    print("Waiting for human supervisor decision...")
            else:
                print(f"Unexpected response while polling for supervision status: {response}")
        except Exception as e:
            print(f"Error while polling for supervision status: {e}")

        if time.time() - start_time > timeout:
            print(f"Timed out waiting for human supervision decision. Timeout: {timeout} seconds")
            return Status.TIMEOUT

        time.sleep(5)  # Wait for 5 seconds before polling again


async def get_human_supervision_decision_api(
    supervision_request_id: UUID,
    client: Client,
    timeout: int = 300,
    use_inspect_ai: bool = False) -> SupervisionDecision:
    """Get the supervision decision from the backend API."""

    if use_inspect_ai:
        async with _console_lock:
            with input_screen(width=None) as console:
                console.record = True
                supervision_status = wait_for_human_decision(supervision_request_id=supervision_request_id, client=client, timeout=timeout)
    else:
        console = Console(record=True)
        supervision_status = wait_for_human_decision(supervision_request_id=supervision_request_id, client=client, timeout=timeout)
    
    # get supervision results
    if supervision_status == 'completed':
        # Get the decision from the API
        response = get_supervision_result_sync_detailed(
            client=client,
            supervision_request_id=supervision_request_id
        )
        if response.status_code == 200 and response.parsed:
            supervision_result = response.parsed
            return map_result_to_decision(supervision_result)
        else:
            return SupervisionDecision(
                decision=SupervisionDecisionType.ESCALATE,
                explanation="Failed to retrieve supervision results."
            )
    elif supervision_status == 'failed':
        return SupervisionDecision(decision=SupervisionDecisionType.ESCALATE,
                                   explanation="The human supervisor failed to provide a decision.")
    elif supervision_status == 'assigned':
        return SupervisionDecision(decision=SupervisionDecisionType.ESCALATE,
                                   explanation="The human supervisor is currently busy and has not yet provided a decision.")
    elif supervision_status == 'timeout':
        return SupervisionDecision(decision=SupervisionDecisionType.ESCALATE,
                                   explanation="The human supervisor did not provide a decision within the timeout period.")
    elif supervision_status == 'pending':
        return SupervisionDecision(decision=SupervisionDecisionType.APPROVE,
                                   explanation="The human supervisor has not yet provided a decision.")
    
    # Default return statement in case no conditions are met
    return SupervisionDecision(
        decision=SupervisionDecisionType.ESCALATE,
        explanation="Unexpected supervision status."
    )

def map_result_to_decision(result: SupervisionResult) -> SupervisionDecision:
    decision_map = {
        'approve': SupervisionDecisionType.APPROVE,
        'reject': SupervisionDecisionType.REJECT,
        'modify': SupervisionDecisionType.MODIFY,
        'escalate': SupervisionDecisionType.ESCALATE,
        'terminate': SupervisionDecisionType.TERMINATE
    }
    decision_type = decision_map.get(result.decision.value.lower(), SupervisionDecisionType.ESCALATE)
    modified_output = None
    if decision_type == SupervisionDecisionType.MODIFY and result.toolrequest is not UNSET:  #TODO: Make the modified output work
        modified_output = result.toolrequest  # Assuming toolrequest contains the modified output
    return SupervisionDecision(
        decision=decision_type,
        explanation=result.reasoning,
        modified=modified_output
    )

def _display_review_sent_message(console: Console, backend_api_endpoint: str, review_id: str):
    """Helper function to display the review sent message."""
    message = (
        f"[bold green]Review has been sent to the server for human approval.[/bold green]\n"
        f"You can view the review at: {FRONTEND_URL}/supervisor/human\n"
        f"Review ID: {review_id}"
    )
    console.print(message)


def create_tool_request_group(tool_id: UUID, tool_requests: List[ToolRequest], client: Client) -> Optional[ToolRequestGroup]:
    # Call the create_execution API
    tool_request_group = ToolRequestGroup(tool_requests=tool_requests)
    try:
        tool_request_group_response = create_tool_request_group_sync_detailed(
            tool_id=tool_id,
            client=client,
            body=tool_request_group
        )
        if (
            tool_request_group_response.status_code in [200, 201] and
            tool_request_group_response.parsed is not None
        ):
            tool_request_group = tool_request_group_response.parsed
            print(f"Tool request group created with ID: {tool_request_group.id}")
            return tool_request_group
        else:
            raise Exception(f"Failed to create tool request group. Response: {tool_request_group_response}")
    except Exception as e:
        print(f"Error creating tool request group: {e}")
    
    return None

def get_tool_request_groups(run_id: UUID, client: Client) -> List[ToolRequestGroup]:
    """
    Retrieve a list of request groups for the specified run ID.
    """
    request_groups: List[ToolRequestGroup] = []
    try:
        response = get_request_groups_sync_detailed(
            run_id=run_id,
            client=client,
        )
        if response.status_code == 200 and response.parsed:
            request_groups = response.parsed
            print(f"Retrieved {len(request_groups)} request groups for run ID {run_id}")
        elif response.status_code == 200:
            print(f"No request groups found for run ID {run_id}")
        else:
            print(f"Failed to retrieve request groups for run ID {run_id}. Response: {response}")
    except Exception as e:
        print(f"Error retrieving request groups: {e}")
    return request_groups

def get_supervisor_chains_for_tool(tool_id: UUID, client: Client) -> List[SupervisorChain]:
    """
    Retrieve the supervisor chains for a specific tool.
    """
    
    supervisors_list: List[SupervisorChain] = []
    try:
        supervisors_response = get_tool_supervisor_chains_sync_detailed(
            tool_id=tool_id,
            client=client,
        )
        if supervisors_response is not None and supervisors_response.parsed is not None:
            supervisors_list = supervisors_response.parsed  # List[SupervisorChain]
            print(f"Retrieved {len(supervisors_list)} supervisor chains from the API.")
        else:
            print("No supervisors found for this tool and run.")
    except Exception as e:
        print(f"Error retrieving supervisors: {e}")
    
    return supervisors_list


def send_supervision_request(supervisor_id: UUID, supervisor_chain_id: UUID, request_group_id: UUID, position_in_chain: int) -> UUID:
    client = supervision_config.client

    supervision_request = SupervisionRequest(
        position_in_chain=position_in_chain,
        supervisor_id=supervisor_id
    )

    # Send the SupervisionRequest to /api/reviews
    try:
        supervision_request_response = create_supervision_request_sync_detailed(
            client=client,
            request_group_id=request_group_id,
            chain_id=supervisor_chain_id,
            supervisor_id=supervisor_id,    
            body=supervision_request
        )
        if (
            supervision_request_response.status_code in [200, 201] and
            supervision_request_response.parsed is not None
        ):
            supervision_request_id = supervision_request_response.parsed
            print(f"Created supervision request with ID: {supervision_request_id}")
            if isinstance(supervision_request_id, UUID):
                return supervision_request_id
            else:
                raise ValueError("Invalid supervision request ID received.")
        else:
            raise Exception(f"Failed to create supervision request. Response: {supervision_request_response}")
    except Exception as e:
        print(f"Error creating supervision request: {e}")
        raise

def send_supervision_result(
    supervision_request_id: UUID,
    request_group_id: UUID,
    tool_id: UUID,
    supervisor_id: UUID,
    decision: SupervisionDecision,
    client: Client,
    tool_args: list[Any],
    tool_kwargs: dict[str, Any],
    tool_request: ToolRequest
):
    """
    Send the supervision result to the API.
    """
    # Map SupervisionDecisionType to Decision enum
    decision_mapping = {
        SupervisionDecisionType.APPROVE: Decision.APPROVE,
        SupervisionDecisionType.REJECT: Decision.REJECT,
        SupervisionDecisionType.MODIFY: Decision.MODIFY,
        SupervisionDecisionType.ESCALATE: Decision.ESCALATE,
        SupervisionDecisionType.TERMINATE: Decision.TERMINATE,
    }
    
    api_decision = decision_mapping.get(decision.decision)
    if not api_decision:
        raise ValueError(f"Unsupported decision type: {decision.decision}")
    
    if decision.modified is not None:
        pass
        # new_tool_request = copy.deepcopy(tool_request)
        # new_tool_request.arguments = decision.modified.tool_args
        # create_tool_request_sync_detailed(request_group_id=request_group_id, client=client, body=decision.modified)
        # TODO: create the new tool request with new tool id
    # Create the SupervisionResult object
    supervision_result = SupervisionResult(
        supervision_request_id=supervision_request_id,
        created_at=datetime.now(timezone.utc),
        decision=api_decision,
        reasoning=decision.explanation or "",
        chosen_toolrequest_id=tool_request.id
    )
    # Send the supervision result to the API
    try:
        response = create_supervision_result_sync_detailed(
            supervision_request_id=supervision_request_id,
            client=client,
            body=supervision_result
        )
        if response.status_code in [200, 201]:
            print(f"Successfully submitted supervision result for supervision request ID: {supervision_request_id}")
        else:
            raise Exception(f"Failed to submit supervision result. Response: {response}")
    except Exception as e:
        print(f"Error submitting supervision result: {e}")
        raise

def register_supervisor(client: Client, supervisor_info: dict, project_id: UUID, supervision_context) -> UUID:
    """Registers a single supervisor with the API and returns its ID."""
    supervisor_data = Supervisor(
        name=supervisor_info['name'],
        description=supervisor_info['description'],
        created_at=datetime.now(timezone.utc),
        type=supervisor_info['type'],
        code=supervisor_info['code'],
        attributes=SupervisorAttributes.from_dict(src_dict=supervisor_info['supervisor_attributes'])
    )
    
    supervisor_response = create_supervisor_sync_detailed(
        project_id=project_id,
        client=client,
        body=supervisor_data
    )
    
    if (
        supervisor_response.status_code in [200, 201] and
        supervisor_response.parsed is not None
    ):
        supervisor_id = supervisor_response.parsed
        supervision_context.add_supervisor_id(supervisor_info['name'], supervisor_id)
        
        if isinstance(supervisor_id, UUID):
            supervision_config.add_local_supervisor(supervisor_id, supervisor_info['func'])
        else:
            raise ValueError("Invalid supervisor_id: Expected UUID")
            
        print(f"Supervisor '{supervisor_info['name']}' registered with ID: {supervisor_id}")
        return supervisor_id
    else:
        raise Exception(f"Failed to register supervisor '{supervisor_info['name']}'. Response: {supervisor_response}")
    

def _serialize_arguments(tool_args: list[Any], tool_kwargs: dict[str, Any]) -> dict[str, Any]:
    """
    Helper function to serialize tool_args and tool_kwargs into a single arguments_dict.
    Non-serializable objects are converted to strings.
    """
    arguments_dict = {}

    for idx, arg in enumerate(tool_args):
        try:
            # Attempt to serialize the argument
            if isinstance(arg, (str, int, float, bool, dict, list)):
                arguments_dict[f'arg_{idx}'] = arg
            else:
                arguments_dict[f'arg_{idx}'] = str(arg)
        except Exception:
            arguments_dict[f'arg_{idx}'] = str(arg)

    for key, value in tool_kwargs.items():
        try:
            if isinstance(value, (str, int, float, bool, dict, list)):
                arguments_dict[key] = value
            else:
                arguments_dict[key] = str(value)
        except Exception:
            arguments_dict[key] = str(value)

    return arguments_dict