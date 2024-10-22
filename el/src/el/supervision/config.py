from typing import Any, Callable, Dict, List, Optional, Literal
from enum import Enum
import random
import json
from pydantic import BaseModel, Field
from el.mocking.policies import MockPolicy

PREFERRED_LLM_MODEL = "gpt-4o"

class SupervisionDecisionType(str, Enum):
    APPROVE = "approve"
    REJECT = "reject"
    ESCALATE = "escalate"
    TERMINATE = "terminate"
    MODIFY = "modify"

class SupervisionDecision(BaseModel):
    decision: SupervisionDecisionType
    """Supervision decision."""

    modified: Any = Field(default=None)
    """Modified data for decision 'modify'."""

    explanation: str | None = Field(default=None)
    """Explanation for decision."""

class SupervisionConfig:
    def __init__(self):
        self.global_supervision_functions: List[Callable] = []
        self.global_mock_policy = MockPolicy.NO_MOCK
        self.override_local_policy = False
        self.mock_responses: Dict[str, List[Any]] = {}
        self.previous_calls: Dict[str, List[Any]] = {}
        self.function_supervisors: Dict[str, List[Callable]] = {}  # Function-specific supervision chains
        self.llm = None

    def set_global_supervision_functions(self, functions: List[Callable]):
        self.global_supervision_functions = functions

    def set_function_supervision_functions(self, function_name: str, functions: List[Callable]):
        self.function_supervisors[function_name] = functions

    def get_supervision_functions(self, function_name: str, local_supervision: Optional[List[Callable]] = None) -> List[Callable]:
        supervisors = []
        # Add local supervision functions specified in the decorator
        if local_supervision:
            supervisors.extend(local_supervision)
        
        # Add function-specific supervision functions
        local_supervisors = self.function_supervisors.get(function_name, [])
        for supervisor in local_supervisors:
            if supervisor not in supervisors:
                supervisors.append(supervisor)
        
        # Add global supervision functions
        for supervisor in self.global_supervision_functions:
            if supervisor not in supervisors:
                supervisors.append(supervisor)
        
        return supervisors

    def add_function_supervisors(self, function_name: str, supervisors: List[Callable]):
        """Add supervisor functions to the supervision chain of a specific function."""
        if function_name not in self.function_supervisors:
            self.function_supervisors[function_name] = []
        self.function_supervisors[function_name].extend(supervisors)

    def set_llm(self, llm):
        self.llm = llm

    def set_mock_policy(self, mock_policy: MockPolicy):
        self.global_mock_policy = mock_policy

    def load_previous_execution_log(self, log_file_path: str, log_format='langchain'):
        """Load and process a previous execution log."""
        with open(log_file_path, 'r') as f:
            log_data = f.readlines()
        
        if log_format == 'langchain':
            for line in log_data:
                try:
                    entry = json.loads(line)
                    if entry['event'] == 'on_tool_end':
                        function_name = entry['data']['kwargs'].get('name')
                        output = entry['data']['output']
                        if function_name:
                            if function_name not in self.previous_calls:
                                self.previous_calls[function_name] = []
                            self.previous_calls[function_name].append(output)
                except json.JSONDecodeError:
                    continue  # Skip lines that are not valid JSON
        else:
            raise ValueError(f"Unsupported log format: {log_format}")
        
        # Update mock_responses with examples from previous calls
        self.mock_responses = self.previous_calls.copy()

    def get_mock_response(self, function_name: str) -> Any:
        """Get a mock response for a specific function."""
        if function_name in self.mock_responses:
            return random.choice(self.mock_responses[function_name])
        else:
            raise ValueError(f"No mock responses available for function: {function_name}")

# Global instance of SupervisionConfig
# TODO: Update this
supervision_config = SupervisionConfig()

def set_global_supervision_functions(functions: List[Callable]):
    supervision_config.set_global_supervision_functions(functions)

def set_function_supervision_functions(function_name: str, functions: List[Callable]):
    supervision_config.set_function_supervision_functions(function_name, functions)

def set_global_mock_policy(mock_policy: MockPolicy, override_local_policy: bool = False):
    supervision_config.set_mock_policy(mock_policy)
    supervision_config.override_local_policy = override_local_policy

def setup_sample_from_previous_calls(log_file_path: str):
    """Set up the SAMPLE_FROM_PREVIOUS_CALLS mock policy."""
    supervision_config.load_previous_execution_log(log_file_path)
    set_global_mock_policy(MockPolicy.SAMPLE_PREVIOUS_CALLS, override_local_policy=True)
