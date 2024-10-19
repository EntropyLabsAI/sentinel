import ast
import shlex
from typing import List, Dict, Optional, Tuple, Set

def check_bash_command(
    command: str,
    allowed_commands: List[str],
    allow_sudo: bool = False,
    command_specific_rules: Optional[Dict[str, List[str]]] = None
) -> Tuple[bool, str]:
    """
    Check if a bash command is allowed based on the given rules.

    Args:
        command (str): The bash command to check.
        allowed_commands (List[str]): List of allowed bash commands.
        allow_sudo (bool, optional): Whether to allow sudo commands. Defaults to False.
        command_specific_rules (Optional[Dict[str, List[str]]], optional): Dictionary of command-specific rules.

    Returns:
        Tuple[bool, str]: A tuple containing a boolean (True if approved, False if rejected) and an explanation string.
    """
    dangerous_chars = ["&", "|", ";", ">", "<", "`", "$", "(", ")"]
    command_specific_rules = command_specific_rules or {}

    if not command.strip():
        return False, "Empty command"

    try:
        tokens = shlex.split(command)
    except ValueError as e:
        return False, f"Invalid command syntax: {str(e)}"

    if any(char in command for char in dangerous_chars):
        return False, f"Command contains potentially dangerous characters: {', '.join(char for char in dangerous_chars if char in command)}"

    base_command = tokens[0]

    if base_command == "sudo":
        if not allow_sudo:
            return False, "sudo is not allowed"
        if len(tokens) < 2:
            return False, "Invalid sudo command"
        base_command = tokens[1]
        tokens = tokens[1:]

    if base_command not in allowed_commands:
        return False, f"Command '{base_command}' is not in the allowed list. Allowed commands: {', '.join(allowed_commands)}"

    if base_command in command_specific_rules:
        allowed_subcommands = command_specific_rules[base_command]
        if len(tokens) > 1 and tokens[1] not in allowed_subcommands:
            return False, f"{base_command} subcommand '{tokens[1]}' is not allowed. Allowed subcommands: {', '.join(allowed_subcommands)}"

    return True, f"Command '{command}' is approved."

def check_python_code(
    code: str,
    allowed_modules: List[str],
    allowed_functions: List[str],
    disallowed_builtins: Optional[Set[str]] = None,
    sensitive_modules: Optional[Set[str]] = None,
    allow_system_state_modification: bool = False
) -> Tuple[bool, str]:
    """
    Check if Python code uses only allowed modules and functions, and applies additional safety checks.

    Args:
        code (str): The Python code to check.
        allowed_modules (List[str]): List of allowed Python modules.
        allowed_functions (List[str]): List of allowed functions.
        disallowed_builtins (Optional[Set[str]]): Set of disallowed built-in functions.
        sensitive_modules (Optional[Set[str]]): Set of sensitive modules to be blocked.
        allow_system_state_modification (bool): Whether to allow modification of system state.

    Returns:
        Tuple[bool, str]: A tuple containing a boolean (True if approved, False if rejected) and an explanation string.
    """
    allowed_modules_set = set(allowed_modules)
    allowed_functions_set = set(allowed_functions)
    disallowed_builtins = disallowed_builtins or {"eval", "exec", "compile", "__import__", "open", "input"}
    sensitive_modules = sensitive_modules or {"os", "sys", "subprocess", "socket", "requests"}

    if not code.strip():
        return False, "Empty code"

    try:
        tree = ast.parse(code)
    except SyntaxError as e:
        return False, f"Invalid Python syntax: {str(e)}"

    for node in ast.walk(tree):
        if isinstance(node, ast.Import):
            for alias in node.names:
                if alias.name not in allowed_modules_set:
                    return False, f"Module '{alias.name}' is not in the allowed list. Allowed modules: {', '.join(allowed_modules)}"
                if alias.name in sensitive_modules:
                    return False, f"Module '{alias.name}' is considered sensitive and not allowed."
        elif isinstance(node, ast.ImportFrom):
            if node.module not in allowed_modules_set:
                return False, f"Module '{node.module}' is not in the allowed list. Allowed modules: {', '.join(allowed_modules)}"
            if node.module in sensitive_modules:
                return False, f"Module '{node.module}' is considered sensitive and not allowed."
        elif isinstance(node, ast.Call):
            if isinstance(node.func, ast.Name):
                if node.func.id in disallowed_builtins:
                    return False, f"Built-in function '{node.func.id}' is not allowed for security reasons."
                if node.func.id not in allowed_functions_set:
                    return False, f"Function '{node.func.id}' is not in the allowed list. Allowed functions: {', '.join(allowed_functions)}"

        if not allow_system_state_modification:
            if isinstance(node, ast.Assign):
                for target in node.targets:
                    if isinstance(target, ast.Attribute) and target.attr.startswith("__"):
                        return False, "Modification of system state (dunder attributes) is not allowed."

    return True, "Python code is approved."
