import asyncio
from entropy_labs.supervision.config import (
    SupervisionDecision,
    SupervisionDecisionType,
    supervision_config,
)
from entropy_labs.supervision.inspect_ai._config import FRONTEND_URL
from entropy_labs.supervision.supervisors import auto_approve_supervisor
from rich.console import Console
from inspect_ai.util._console import input_screen
import time
from entropy_labs.sentinel_api_client.sentinel_api_client.client import Client
from entropy_labs.sentinel_api_client.sentinel_api_client.api.reviews.get_supervision_status import (
    sync_detailed as get_supervision_status_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.api.reviews.get_supervision_results import (
    sync_detailed as get_supervision_results_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.supervision_result import (
    SupervisionResult,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.status import Status
from entropy_labs.sentinel_api_client.sentinel_api_client.types import UNSET
from entropy_labs.sentinel_api_client.sentinel_api_client.models.supervisor_type import (
    SupervisorType,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.api.tools.create_tool import (
    sync_detailed as create_tool_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.tool import Tool
from entropy_labs.sentinel_api_client.sentinel_api_client.api.supervisors.create_supervisor import (
    sync_detailed as create_supervisor_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.supervisor import Supervisor
from entropy_labs.sentinel_api_client.sentinel_api_client.api.supervisors.create_run_tool_supervisors import (
    sync_detailed as create_run_tool_supervisors_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.api.supervisors.get_run_tool_supervisors import (
    sync_detailed as get_run_tool_supervisors_sync,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.api.executions.create_execution import (
    sync_detailed as create_execution_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.api.reviews.create_supervision_request import (
    sync_detailed as create_supervision_request_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.api.reviews.create_supervision_result import (
    sync_detailed as create_supervision_result_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.api.projects.get_projects import (
    sync_detailed as get_projects_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.create_supervision_result import (
    CreateSupervisionResult,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.decision import Decision
from entropy_labs.sentinel_api_client.sentinel_api_client.models.tool_attributes import (
    ToolAttributes,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.tool_request import (
    ToolRequest,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.tool_request_arguments import (
    ToolRequestArguments,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.create_execution_body import (
    CreateExecutionBody,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.supervision_request import (
    SupervisionRequest,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.task_state import TaskState
from entropy_labs.sentinel_api_client.sentinel_api_client.models.supervisor_attributes import (
    SupervisorAttributes,
)
from uuid import uuid4, UUID
from datetime import datetime, timezone
import inspect
from entropy_labs.utils.utils import get_function_code
from typing import List, Optional, Any, Callable
from entropy_labs.sentinel_api_client.sentinel_api_client.api.executions.get_run_executions import (
    sync_detailed as get_executions_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.execution import Execution
from entropy_labs.sentinel_api_client.sentinel_api_client.api.supervisors.get_run_tool_supervisors import (
    sync_detailed as get_run_tool_supervisors_sync_detailed,
)
from entropy_labs.sentinel_api_client.sentinel_api_client.models.supervisor_chain import SupervisorChain


# Create an asyncio.Lock to prevent concurrent console access
_console_lock = asyncio.Lock()

def register_tools_and_supervisors(client: Client, run_id: UUID):
    # Access the registries from the context
    supervision_context = supervision_config.context

    for func, data in supervision_context.supervised_functions_registry.items():
        supervision_functions = data['supervision_functions']
        ignored_attributes = data['ignored_attributes']

        tool_name = func.__qualname__
        # Extract function arguments using inspect
        func_signature = inspect.signature(func)
        func_arguments = {
            param.name: str(param.annotation) if param.annotation is not param.empty else 'Any'
            for param in func_signature.parameters.values()
        }

        # Pass the extracted arguments to ToolAttributes.from_dict
        attributes = ToolAttributes.from_dict(src_dict=func_arguments)

        # Register the tool
        tool_data = Tool(
            name=tool_name,
            description=str(func.__doc__) if func.__doc__ else tool_name,
            attributes=attributes,
            ignored_attributes=ignored_attributes,
            created_at=datetime.now(timezone.utc),
        )
        tool_response = create_tool_sync_detailed(
            client=client,
            body=tool_data
        )
        if (
            tool_response.status_code == 200 and
            tool_response.parsed is not None and
            tool_response.parsed.id is not None
        ):
            # Update the tool_id in the registry 
            tool_id = tool_response.parsed.id
            supervision_context.update_tool_id(func, tool_id)
            print(f"Tool '{tool_name}' registered with ID: {tool_id}")
        else:
            raise Exception(f"Failed to register tool '{tool_name}'. Status code: {tool_response}")

        # Register supervisors and associate them with the tool
        supervisor_ids = []
        if supervision_functions == []:
            supervisor_func = auto_approve_supervisor()
            supervisor_info: dict[str, Any] = {
                    'func': supervisor_func,
                    'name': getattr(supervisor_func, '__name__', 'supervisor_name'),
                    'description': getattr(supervisor_func, '__doc__', 'supervisor_description'),
                    'type': SupervisorType.NO_SUPERVISOR,
                    'code': get_function_code(supervisor_func),
                    'supervisor_attributes': {}
            }
            supervisor_id = register_supervisor(client, supervisor_info, supervision_context)
            supervisor_ids.append([supervisor_id])
        else:
            for idx, supervisor_func_list in enumerate(supervision_functions):
                supervisor_ids.append([])
                for supervisor_func in supervisor_func_list:
                    supervisor_info: dict[str, Any] = {
                        'func': supervisor_func,
                        'name': getattr(supervisor_func, '__name__', None) or 'supervisor_name',
                        'description': getattr(supervisor_func, '__doc__', None) or 'supervisor_description',
                        'type': SupervisorType.HUMAN_SUPERVISOR if getattr(supervisor_func, '__name__', 'supervisor_name') in ['human_supervisor', 'human_approver'] else SupervisorType.CLIENT_SUPERVISOR,
                        'code': get_function_code(supervisor_func),
                        'supervisor_attributes': getattr(supervisor_func, 'supervisor_attributes', {})
                    }
                    supervisor_id = register_supervisor(client, supervisor_info, supervision_context)
                    supervisor_ids[idx].append(supervisor_id)

        # Ensure tool_id is a UUID before proceeding
        if tool_id is UNSET or not isinstance(tool_id, UUID):
            raise ValueError("Invalid tool_id: Expected UUID")

        # Associate supervisors with the tool for the given run
        if supervisor_ids:
            association_response = create_run_tool_supervisors_sync_detailed(
                run_id=run_id,
                tool_id=tool_id,
                client=client,
                body=supervisor_ids
            )
            if association_response.status_code == 200:
                print(f"Supervisors assigned to tool '{tool_name}' for run ID {run_id}")
            else:
                raise Exception(f"Failed to assign supervisors to tool '{tool_name}'. Response: {association_response}")
        else:
            print(f"No supervisors to assign to tool '{tool_name}'")



def wait_for_human_decision(review_id: UUID, client: Client, timeout: int = 300) -> Status:
    start_time = time.time()

    while True:
        try:
            response = get_supervision_status_sync_detailed(
                client=client,
                review_id=review_id
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
    review_id: UUID,
    client: Client,
    timeout: int = 300,
    use_inspect_ai: bool = False) -> SupervisionDecision:
    """Get the supervision decision from the backend API."""

    if use_inspect_ai:
        async with _console_lock:
            with input_screen(width=None) as console:
                console.record = True
                supervision_status = wait_for_human_decision(review_id=review_id, client=client, timeout=timeout)
    else:
        console = Console(record=True)
        supervision_status = wait_for_human_decision(review_id=review_id, client=client, timeout=timeout)
    
    # get supervision results
    if supervision_status == 'completed':
        # Get the decision from the API
        response = get_supervision_results_sync_detailed(
            client=client,
            review_id=review_id
        )
        if response.status_code == 200 and response.parsed:
            supervision_results = response.parsed
            latest_result = supervision_results[-1]  # Get the latest result
            decision = map_result_to_decision(latest_result)
            return decision
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
    if decision_type == SupervisionDecisionType.MODIFY and result.toolrequest is not UNSET:
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


def create_execution(tool_id: UUID, run_id: UUID, client: Client) -> Optional[UUID]:
    # Call the create_execution API
    create_execution_body = CreateExecutionBody(tool_id=tool_id)
    try:
        execution_response = create_execution_sync_detailed(
            run_id=run_id,
            client=client,
            body=create_execution_body
        )
        if (
            execution_response.status_code == 200 and
            execution_response.parsed is not None and
            execution_response.parsed.id is not None
        ):
            execution_id = execution_response.parsed.id
            print(f"Execution created with ID: {execution_id}")
            return execution_id
        else:
            raise Exception(f"Failed to create execution. Response: {execution_response}")
    except Exception as e:
        print(f"Error creating execution: {e}")
    
    return None

def get_executions(run_id: UUID, client: Client) -> List[Execution]:
    """
    Retrieve a list of executions for the specified run ID.
    """
    executions_list: List[Execution] = []
    try:
        response = get_executions_sync_detailed(
            run_id=run_id,
            client=client,
        )
        if response.status_code == 200 and response.parsed:
            executions_list = response.parsed
            print(f"Retrieved {len(executions_list)} executions for run ID {run_id}")
        elif response.status_code == 200:
            print(f"No executions found for run ID {run_id}")
        else:
            print(f"Failed to retrieve executions for run ID {run_id}. Response: {response}")
    except Exception as e:
        print(f"Error retrieving executions: {e}")
    return executions_list

def get_supervisors_for_tool(tool_id: UUID, run_id: UUID, client: Client) -> List[Supervisor]:
    # Get the list of supervisors using run_id and tool_id from the API
    supervisors_list: List[Supervisor] = []
    try:
        supervisors_response = get_run_tool_supervisors_sync(
            run_id=run_id,
            tool_id=tool_id,
            client=client,
        )
        if supervisors_response is not None and supervisors_response.parsed is not None:
            supervisors_list = supervisors_response.parsed  # List[Supervisor]
            print(f"Retrieved {len(supervisors_list)} supervisor chains from the API.")
        else:
            print("No supervisors found for this tool and run.")
    except Exception as e:
        print(f"Error retrieving supervisors: {e}")
    
    return supervisors_list


def send_supervision_request(supervisor: Supervisor, supervision_context, execution_id: UUID, tool_id: UUID, tool_args: list[Any], tool_kwargs: dict[str, Any]) -> UUID:
    client = supervision_config.client
    run_id = supervision_config.run_id

    # Prepare the SupervisionRequest with additional information
    if run_id is None:
        raise ValueError("Run ID is required to create a supervision request.")

    # Use the helper function to serialize arguments
    arguments_dict = _serialize_arguments(tool_args, tool_kwargs)

    tool_requests = [ToolRequest(
        tool_id=tool_id,
        arguments=ToolRequestArguments.from_dict(arguments_dict)
    )]

    supervision_request = SupervisionRequest(
        run_id=run_id,
        execution_id=execution_id,
        supervisor_id=supervisor.id,
        task_state=supervision_context.to_task_state(),  # TODO: Make sure this is the correct format
        tool_requests=tool_requests,
        messages=[supervision_context.get_api_messages()[-1]]  # TODO: API otherwise returns Number of tool choices and messages must be the same
    )

    # Send the SupervisionRequest to /api/reviews
    try:
        supervision_status_response = create_supervision_request_sync_detailed(
            client=client,
            body=supervision_request
        )
        if (
            supervision_status_response.status_code == 200 and
            supervision_status_response.parsed is not None and
            supervision_status_response.parsed.id is not None
        ):
            review_id = supervision_status_response.parsed.supervision_request_id
            print(f"Created supervision request with ID: {review_id}")
            return review_id
        else:
            raise Exception(f"Failed to create supervision request. Response: {supervision_status_response}")
    except Exception as e:
        print(f"Error creating supervision request: {e}")
        raise

def send_review_result(
    review_id: UUID,
    execution_id: UUID,
    run_id: UUID,
    tool_id: UUID,
    supervisor_id: UUID,
    decision: SupervisionDecision,
    client: Client,
    tool_args: list[Any],
    tool_kwargs: dict[str, Any]
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
    
    # Combine tool_args and tool_kwargs into arguments_dict
    arguments_dict = _serialize_arguments(tool_args, tool_kwargs)

    tool_request_arguments = ToolRequestArguments.from_dict(arguments_dict)

    tool_request = ToolRequest(
        tool_id=tool_id,
        arguments=tool_request_arguments
    )
    # Create the SupervisionResult object
    supervision_result = SupervisionResult(
        id=uuid4(),  # Generate a unique ID 
        supervision_request_id=review_id, # TODO: Is this review_id?
        created_at=datetime.now(timezone.utc),
        decision=api_decision,
        reasoning=decision.explanation or "",
        toolrequest=tool_request,
    )

    # Create the CreateSupervisionResult object
    create_supervision_result_body = CreateSupervisionResult(
        execution_id=execution_id,
        run_id=run_id,
        tool_id=tool_id,
        supervisor_id=supervisor_id,
        supervision_result=supervision_result,
    )
    # Send the supervision result to the API
    try:
        response = create_supervision_result_sync_detailed(
            review_id=review_id,
            client=client,
            body=create_supervision_result_body
        )
        if response.status_code == 200:
            print(f"Successfully submitted supervision result for review ID: {review_id}")
        else:
            print(f"Failed to submit supervision result. Response: {response}")
    except Exception as e:
        print(f"Error submitting supervision result: {e}")
        raise

def register_supervisor(client: Client, supervisor_info: dict, supervision_context) -> UUID:
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
        client=client,
        body=supervisor_data
    )
    
    if (
        supervisor_response.status_code == 200 and
        supervisor_response.parsed is not None and
        supervisor_response.parsed.id is not None
    ):
        supervisor_id = supervisor_response.parsed.id
        supervision_context.add_supervisor_id(supervisor_info['name'], supervisor_id)
        
        if isinstance(supervisor_id, UUID):
            supervision_config.add_local_supervisor(supervisor_id, supervisor_info['func'])
        else:
            raise ValueError("Invalid supervisor_id: Expected UUID")
            
        print(f"Supervisor '{supervisor_info['name']}' registered with ID: {supervisor_id}")
        return supervisor_id
    else:
        raise Exception(f"Failed to register supervisor '{supervisor_info['name']}'. Response: {supervisor_response}")
    
    
def get_tool_supervisors(run_id: UUID, tool_id: UUID, client: Client) -> List[SupervisorChain]:
    """
    Retrieves the supervisors assigned to a tool, grouped by chain, for a specific run.

    Args:
        run_id (UUID): The ID of the run.
        tool_id (UUID): The ID of the tool.
        client (Client): The API client instance.

    Returns:
        List[SupervisorChain]: A list of supervisor chains.
    """
    try:
        response = get_run_tool_supervisors_sync_detailed(
            run_id=run_id,
            tool_id=tool_id,
            client=client,
        )
        if response.status_code == 200 and response.parsed:
            supervisor_chains = response.parsed
            print(f"Retrieved {len(supervisor_chains)} supervisor chains for run ID {run_id} and tool ID {tool_id}")
            return supervisor_chains
        else:
            print(
                f"Failed to retrieve supervisor chains for run ID {run_id} and tool ID {tool_id}. "
                f"Response: {response}"
            )
            return []
    except Exception as e:
        print(f"Error retrieving supervisor chains: {e}")
        return []

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