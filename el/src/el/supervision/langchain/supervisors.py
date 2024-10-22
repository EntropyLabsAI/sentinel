from typing import Any, Callable, Dict, List, Optional, Set
from el.supervision.config import SupervisionDecisionType, SupervisionDecision
from el.utils.utils import prompt_user_input_or_api
import random
from el.supervision.common import check_bash_command, check_python_code

# Supervisor functions

def human_supervisor():
    """
    Creates a human supervisor function that requests approval either via CLI or API.
    
    Returns:
        Callable: A supervisor function that requests human approval.
    """
    def supervisor(func: Callable, *args, **kwargs) -> SupervisionDecision:
        """Human supervisor that requests approval either via CLI or API."""
        function_name = func.__name__
        # Prepare the data to send to the backend or display in CLI
        data = {
            "function": function_name,
            "args": args,
            "kwargs": kwargs
        }
        # Use a utility function to handle CLI or API-based human approval
        decision = prompt_user_input_or_api(data)
        return decision
    supervisor.__name__ = human_supervisor.__name__
    return supervisor

def llm_supervisor():
    """
    Creates an LLM supervisor function that evaluates the proposed action.
    
    Returns:
        Callable: A supervisor function that uses an LLM to make decisions.
    """
    def supervisor(func: Callable, *args, **kwargs) -> SupervisionDecision:
        """LLM supervisor that evaluates the proposed action."""
        # Prepare the prompt for the LLM
        prompt = f"Should the function '{func.__name__}' be executed with arguments {args} and {kwargs}? Respond with 'approve', 'reject', or 'escalate'."
        # Send the prompt to the LLM (this is a placeholder for actual LLM integration)
        decision_str = simulate_llm_response(prompt)
        
        return SupervisionDecision(decision=decision_str)
    
    supervisor.__name__ = llm_supervisor.__name__
    return supervisor

def simulate_llm_response(prompt: str) -> str:
    """Simulate an LLM response (placeholder function)."""
    # In an actual implementation, this function would send the prompt to an LLM model
    # TODO: Implement actual LLM response
    print(f"LLM Prompt: {prompt}")
    # For simulation, we'll randomly choose a decision
    return random.choice(["approve", "reject", "escalate"])

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
        """Supervisor that checks if the bash command is allowed."""
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
        """Supervisor that checks if the Python code is allowed."""
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
