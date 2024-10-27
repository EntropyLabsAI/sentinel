import logging
from typing import List, Dict, Optional, Set
from inspect_ai.approval import Approval, Approver, approver
from inspect_ai.solver import TaskState
from inspect_ai.tool import ToolCall, ToolCallView
from inspect_ai._util.registry import registry_lookup

@approver
def bash_approver(
    allowed_commands: List[str],
    allow_sudo: bool = False,
    command_specific_rules: Optional[Dict[str, List[str]]] = None,
) -> Approver:
    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: Optional[TaskState] = None,
    ) -> Approval:
        from entropy_labs.supervision.common import check_bash_command
        command = str(next(iter(call.arguments.values()))).strip()
        is_approved, explanation = check_bash_command(command, allowed_commands, allow_sudo, command_specific_rules)
        
        if is_approved:
            return Approval(decision="approve", explanation=explanation)
        else:
            return Approval(decision="escalate", explanation=explanation)
    return approve

@approver
def python_approver(
    allowed_modules: List[str],
    allowed_functions: List[str],
    disallowed_builtins: Optional[Set[str]] = None,
    sensitive_modules: Optional[Set[str]] = None,
    allow_system_state_modification: bool = False,
) -> Approver:
    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: Optional[TaskState] = None,
    ) -> Approval:
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
            return Approval(decision="approve", explanation=explanation)
        else:
            return Approval(decision="escalate", explanation=explanation)
    return approve

@approver
def human_approver(agent_id: str, approval_api_endpoint: Optional[str] = None, n: int = 3, timeout: int = 300) -> Approver:
    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: Optional[TaskState] = None,
    ) -> Approval:
        from entropy_labs.supervision.common import human_supervisor_wrapper        
        from entropy_labs.supervision.common import _transform_entropy_labs_approval_to_inspect_ai_approval
        if state is None:
            return Approval(decision="escalate", explanation="TaskState is required for this approver.")

        logging.info(f"Generating {n} tool call suggestions for user review")

        approval_decision = await human_supervisor_wrapper(task_state=state, call=call, backend_api_endpoint=approval_api_endpoint, agent_id=agent_id, timeout=timeout, use_inspect_ai=True, n=n)
        inspect_approval = _transform_entropy_labs_approval_to_inspect_ai_approval(approval_decision)
        return inspect_approval
    return approve

@approver
def llm_approver(instructions: str, openai_model: str, system_prompt: Optional[str] = None, include_context: bool = False, agent_id: Optional[str] = None) -> Approver:
    async def approve(
        message: str,
        call: ToolCall,
        view: ToolCallView,
        state: Optional[TaskState] = None,
    ) -> Approval:
        if state is None:
            return Approval(decision="escalate", explanation="TaskState is required for this approver.")
        from entropy_labs.supervision.supervisors import llm_supervisor
        from entropy_labs.supervision.config import supervision_config
        from entropy_labs.supervision.common import _transform_entropy_labs_approval_to_inspect_ai_approval

        llm_supervisor_func = llm_supervisor(instructions=instructions, openai_model=openai_model, system_prompt=system_prompt, include_context=include_context)
        
        function_name = call.function 
        func = registry_lookup('tool', function_name)
        supervision_config.context.inspect_ai_state = state
        approval_decision = llm_supervisor_func(func=func, supervision_context=supervision_config.context,
                                               **call.arguments)
        inspect_approval = _transform_entropy_labs_approval_to_inspect_ai_approval(approval_decision)
        return inspect_approval      
    return approve
