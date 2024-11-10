import logging
from typing import List, Dict, Optional, Set, Any, Callable
from inspect_ai.approval import Approval, Approver, approver
from inspect_ai.solver import TaskState
from inspect_ai.tool import ToolCall, ToolCallView
from inspect_ai._util.registry import registry_lookup
from entropy_labs.sentinel_api_client.sentinel_api_client.client import Client
from uuid import UUID
from entropy_labs.supervision.config import supervision_config
from entropy_labs.supervision.config import SupervisionDecision, SupervisionDecisionType
from functools import wraps

# Helper functions
def get_or_create_execution_id(run_id: UUID, tool_id: UUID, client: Client) -> Optional[UUID]:
    from entropy_labs.api.sentinel_api_client_helper import create_execution, get_executions
    from entropy_labs.sentinel_api_client.sentinel_api_client.models.status import Status
    executions = get_executions(run_id, client)
    if executions:
        for execution in executions:
            if execution.tool_id == tool_id and execution.run_id == run_id and execution.status == Status.PENDING:
                return execution.id
    return create_execution(tool_id, run_id, client)

def get_supervisor_by_name(run_id: UUID, tool_id: UUID, supervisor_name: str, client: Client):
    from entropy_labs.api.sentinel_api_client_helper import get_tool_supervisors
    supervisors_chains = get_tool_supervisors(run_id, tool_id, client)
    # Only one chain is supported with Inspect AI as of now
    supervisors = supervisors_chains[0].supervisors
    for supervisor in supervisors:
        if supervisor.name == supervisor_name:
            return supervisor
    return None

def send_supervision_request_wrapper(supervisor, supervision_context, execution_id, tool_id, call_arguments):
    from entropy_labs.api.sentinel_api_client_helper import send_supervision_request
    return send_supervision_request(
        supervisor=supervisor,
        supervision_context=supervision_context,
        execution_id=execution_id,
        tool_id=tool_id,
        tool_args=[],
        tool_kwargs=call_arguments
    )

def send_review_result_wrapper(review_id, execution_id, run_id, tool_id, supervisor_id, decision, client, call_arguments):
    from entropy_labs.api.sentinel_api_client_helper import send_review_result
    send_review_result(
        review_id=review_id,
        execution_id=execution_id,
        run_id=run_id,
        tool_id=tool_id,
        supervisor_id=supervisor_id,
        decision=decision,
        client=client,
        tool_args=[],  # TODO: If modified, send modified args and kwargs
        tool_kwargs=call_arguments  # TODO: update for modified
    )

def prepare_approval(decision: SupervisionDecision) -> Approval:
    from entropy_labs.supervision.common import _transform_entropy_labs_approval_to_inspect_ai_approval
    return _transform_entropy_labs_approval_to_inspect_ai_approval(decision)

def get_tool_info(call: ToolCall, supervision_context):
    entry = supervision_context.get_supervised_function_entry_by_name(call.function)
    if entry:
        tool_id = entry['tool_id']
        ignored_attributes = entry['ignored_attributes']
        return tool_id, ignored_attributes
    else:
        raise Exception(f"Tool ID for function {call.function} not found in the registry.")

# Decorator for common Entropy Sentinel API interactions
def with_entropy_supervision(supervisor_name_param: Optional[str] = None):
    def decorator(approve_func: Callable):
        @wraps(approve_func)
        async def wrapper(
            message: str,
            call: ToolCall,
            view: ToolCallView,
            state: Optional[TaskState] = None,
            **kwargs
        ) -> Approval:
            from entropy_labs.api.sentinel_api_client_helper import SupervisorType

            supervision_context = supervision_config.context
            run_id = supervision_config.run_id
            client = supervision_config.client  # Get the Sentinel API client
            if run_id is None:
                raise Exception("Run ID is required to create a supervision request.")

            # Retrieve the tool_id from the registry
            tool_id, _ = get_tool_info(call, supervision_context)

            # Get or create request_group
            execution_id = get_or_create_execution_id(run_id, tool_id, client)
            if not execution_id:
                raise Exception(f"Failed to get or create execution ID")

            # Supervisor name can be specified or use the name of the approve function
            supervisor_name = supervisor_name_param or approve_func.__name__
            # Get the relevant supervisor for this function
            supervisor = get_supervisor_by_name(run_id, tool_id, supervisor_name, client)
            if supervisor is None or supervisor.id is None:
                raise Exception(f"No supervisor found for function {call.function}")

            supervision_context.inspect_ai_state = state
            review_id = send_supervision_request_wrapper(
                supervisor, supervision_context, execution_id, tool_id, call.arguments
            )

            # Call the specific approval logic
            decision = await approve_func(
                message, call, view, state,
                supervision_context=supervision_context,
                review_id=review_id,
                client=client,
                **kwargs
            )
            print(f"Decision: {decision} for review ID: {review_id}")

            if supervisor.type != SupervisorType.HUMAN_SUPERVISOR:
                # Send the decision to the API
                send_review_result_wrapper(
                    review_id,
                    execution_id,
                    run_id,
                    tool_id,
                    supervisor.id,
                    decision,
                    client,
                    call.arguments
                )

            return prepare_approval(decision)

        # Set the __name__ of the wrapper function to the supervisor name or the outer function's name
        wrapper.__name__ = supervisor_name_param or approve_func.__name__
        return wrapper
    return decorator

# Updated Approvers
@approver
def bash_approver(
    allowed_commands: List[str],
    allow_sudo: bool = False,
    command_specific_rules: Optional[Dict[str, List[str]]] = None,
) -> Approver:
    @with_entropy_supervision(supervisor_name_param="bash_approver")
    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: Optional[TaskState] = None,
        supervision_context=None,
        review_id: Optional[UUID] = None,
        client: Optional[Client] = None,
    ) -> SupervisionDecision:
        """
        Bash approver for Inspect AI.
        """
        from entropy_labs.supervision.common import check_bash_command
        command = str(next(iter(call.arguments.values()))).strip()
        is_approved, explanation = check_bash_command(
            command, allowed_commands, allow_sudo, command_specific_rules
        )

        if is_approved:
            decision = SupervisionDecision(
                decision=SupervisionDecisionType.APPROVE, explanation=explanation
            )
        else:
            decision = SupervisionDecision(
                decision=SupervisionDecisionType.ESCALATE, explanation=explanation
            )
        return decision
    return approve

@approver
def python_approver(
    allowed_modules: List[str],
    allowed_functions: List[str],
    disallowed_builtins: Optional[Set[str]] = None,
    sensitive_modules: Optional[Set[str]] = None,
    allow_system_state_modification: bool = False,
) -> Approver:
    @with_entropy_supervision(supervisor_name_param="python_approver")
    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: Optional[TaskState] = None,
        supervision_context=None,
        review_id: Optional[UUID] = None,
        client: Optional[Client] = None,
    ) -> SupervisionDecision:
        """
        Python approver for Inspect AI.
        """
        from entropy_labs.supervision.common import check_python_code
        code = str(next(iter(call.arguments.values()))).strip()
        is_approved, explanation = check_python_code(
            code,
            allowed_modules,
            allowed_functions,
            disallowed_builtins,
            sensitive_modules,
            allow_system_state_modification
        )

        if is_approved:
            decision = SupervisionDecision(
                decision=SupervisionDecisionType.APPROVE, explanation=explanation
            )
        else:
            decision = SupervisionDecision(
                decision=SupervisionDecisionType.ESCALATE, explanation=explanation
            )
        return decision
    return approve

@approver
def human_approver(
    agent_id: str,
    approval_api_endpoint: Optional[str] = None,
    n: int = 3,
    timeout: int = 300
) -> Approver:
    @with_entropy_supervision(supervisor_name_param="human_approver")
    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: Optional[TaskState] = None,
        supervision_context=None,
        review_id: Optional[UUID] = None,
        client: Optional[Client] = None,
    ) -> SupervisionDecision:
        """
        Human approver for Inspect AI.
        """
        from entropy_labs.supervision.common import human_supervisor_wrapper
        if state is None:
            return SupervisionDecision(
                decision=SupervisionDecisionType.ESCALATE,
                explanation="TaskState is required for this approver."
            )

        logging.info(f"Generating {n} tool call suggestions for user review")

        approval_decision = await human_supervisor_wrapper(
            task_state=state,
            call=call,
            backend_api_endpoint=approval_api_endpoint,
            timeout=timeout,
            use_inspect_ai=True,
            n=n,
            review_id=review_id,
            client=client
        )
        return approval_decision
    return approve

@approver
def llm_approver(
    instructions: str,
    openai_model: str,
    system_prompt: Optional[str] = None,
    include_context: bool = False,
    agent_id: Optional[str] = None
) -> Approver:
    @with_entropy_supervision(supervisor_name_param="llm_approver")
    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: Optional[TaskState] = None,
        supervision_context=None,
        review_id: Optional[UUID] = None,
        client: Optional[Client] = None,
    ) -> SupervisionDecision:
        """
        LLM approver for Inspect AI.
        """
        from entropy_labs.supervision.supervisors import llm_supervisor
        if state is None:
            return SupervisionDecision(
                decision=SupervisionDecisionType.ESCALATE,
                explanation="TaskState is required for this approver."
            )

        llm_supervisor_func = llm_supervisor(
            instructions=instructions,
            openai_model=openai_model,
            system_prompt=system_prompt,
            include_context=include_context
        )

        function_name = call.function
        func = registry_lookup('tool', function_name)
        approval_decision = llm_supervisor_func(
            func=func,
            supervision_context=supervision_context,
            **call.arguments
        )
        return approval_decision
    return approve
