from typing import Callable, Dict, List, Optional, Set
from entropy_labs.supervision.config import SupervisionDecisionType, SupervisionDecision, supervision_config, SupervisionContext
from entropy_labs.supervision.langchain.utils import create_task_state_from_context
from inspect_ai.tool import ToolCall

def human_supervisor(backend_api_endpoint: Optional[str] = None, agent_id: str = "default_agent", timeout: int = 300, n: int = 1):
    async def supervisor(func: Callable, supervision_context: SupervisionContext, **kwargs) -> SupervisionDecision:
        """
        Human supervisor that requests approval via backend API or CLI.
        """
        from entropy_labs.supervision.common import human_supervisor_wrapper
        
        context = supervision_config.context
        
        # TODO: Reuse Supervise Protocol

        # Create TaskState from context - Right now we are using Inspect AI TaskState format to send to the backend, we need to transform langchain format to TaskState format
        task_state = create_task_state_from_context(context)
        id = 'tool_id'
        tool_call = ToolCall(id=id,function=func.__name__, arguments=kwargs, type='function')
        
        supervisor_decision = await human_supervisor_wrapper(task_state=task_state, call=tool_call, backend_api_endpoint=backend_api_endpoint, agent_id=agent_id, timeout=timeout, use_inspect_ai=False, n=n)

        return supervisor_decision

    supervisor.__name__ = human_supervisor.__name__
    return supervisor

def bash_supervisor(
    allowed_commands: List[str],
    allow_sudo: bool = False,
    command_specific_rules: Optional[Dict[str, List[str]]] = None
):
    """
    Decorator that creates a supervisor function to check bash commands.
    
    Args:
        allowed_commands (List[str]): List of allowed bash commands.
        allow_sudo (bool): Whether to allow sudo commands. Defaults to False.
        command_specific_rules (Optional[Dict[str, List[str]]]): Dictionary of command-specific rules.
    
    Returns:
        Callable: A supervisor function that checks bash commands.
    """
    def supervisor(func: Callable, *args, **kwargs) -> SupervisionDecision:
        from entropy_labs.supervision.common import check_bash_command
        
        command = args[0] if args else kwargs.get('command', '')
        is_approved, explanation = check_bash_command(
            command, allowed_commands, allow_sudo, command_specific_rules
        )
        if is_approved:
            return SupervisionDecision(
                decision=SupervisionDecisionType.APPROVE,
                explanation=explanation
            )
        else:
            return SupervisionDecision(
                decision=SupervisionDecisionType.REJECT,
                explanation=explanation
            )
    
    supervisor.__name__ = bash_supervisor.__name__
    return supervisor

def python_supervisor(
    allowed_modules: List[str],
    allowed_functions: List[str],
    disallowed_builtins: Optional[Set[str]] = None,
    sensitive_modules: Optional[Set[str]] = None,
    allow_system_state_modification: bool = False
):
    """
    Decorator that creates a supervisor function to check Python code.
    
    Args:
        allowed_modules (List[str]): List of allowed Python modules.
        allowed_functions (List[str]): List of allowed functions.
        disallowed_builtins (Optional[Set[str]]): Set of disallowed built-in functions.
        sensitive_modules (Optional[Set[str]]): Set of sensitive modules to be blocked.
        allow_system_state_modification (bool): Whether to allow modification of system state.
    
    Returns:
        Callable: A supervisor function that checks Python code.
    """
    def supervisor(func: Callable, *args, **kwargs) -> SupervisionDecision:
        from entropy_labs.supervision.common import check_python_code
        
        code = args[0] if args else kwargs.get('code', '')
        is_approved, explanation = check_python_code(
            code, allowed_modules, allowed_functions, disallowed_builtins,
            sensitive_modules, allow_system_state_modification
        )
        if is_approved:
            return SupervisionDecision(
                decision=SupervisionDecisionType.APPROVE,
                explanation=explanation
            )
        else:
            return SupervisionDecision(
                decision=SupervisionDecisionType.REJECT,
                explanation=explanation
            )
    
    supervisor.__name__ = python_supervisor.__name__
    return supervisor
