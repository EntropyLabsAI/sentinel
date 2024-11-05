from typing import Any, Callable, Dict, List, Optional, Literal
from entropy_labs.sentinel_api_client.sentinel_api_client.types import UNSET
from enum import Enum
import random
import json
from pydantic import BaseModel, Field
from entropy_labs.mocking.policies import MockPolicy
from threading import Lock
from inspect_ai.solver import TaskState
from inspect_ai.tool import ToolCall, Tool
from inspect_ai.model import ChatMessage, ChatMessageAssistant, ChatMessageTool, ModelOutput, ChatMessageUser, ChatMessageSystem
from uuid import UUID, uuid4
from inspect_ai._util.content import Content, ContentText
from entropy_labs.supervision.langchain.utils import extract_messages_from_events
from entropy_labs.sentinel_api_client.sentinel_api_client.models.task_state import TaskState as APITaskState
from entropy_labs.sentinel_api_client.sentinel_api_client.models.message import Message
from entropy_labs.sentinel_api_client.sentinel_api_client.models.tool_call import ToolCall as ApiToolCall
from entropy_labs.sentinel_api_client.sentinel_api_client.models.tool_call_arguments import ToolCallArguments
from entropy_labs.sentinel_api_client.sentinel_api_client.models.tool import Tool as ApiTool
from entropy_labs.sentinel_api_client.sentinel_api_client.models.output import Output as ApiOutput
from entropy_labs.sentinel_api_client.sentinel_api_client.models.message import Message
import json


PREFERRED_LLM_MODEL = "gpt-4o"

class SupervisionDecisionType(str, Enum):
    APPROVE = "approve"
    REJECT = "reject"
    ESCALATE = "escalate"
    TERMINATE = "terminate"
    MODIFY = "modify"

class ModifiedData(BaseModel):
    tool_args: Optional[List[Any]] = None
    """Modified positional arguments for the tool/function."""

    tool_kwargs: Optional[Dict[str, Any]] = None
    """Modified keyword arguments for the tool/function."""

class SupervisionDecision(BaseModel):
    decision: SupervisionDecisionType
    """Supervision decision."""

    modified: Optional[ModifiedData] = Field(default=None)
    """Modified data for decision 'modify'."""

    explanation: Optional[str] = Field(default=None)
    """Explanation for decision."""

class SupervisionContext:
    def __init__(self):
        self.langchain_events: List[dict] = []  # List to store all logged events
        self.lock = Lock()  # Ensure thread safety
        self.metadata: Dict[str, Any] = {}
        self.inspect_ai_state: Optional[TaskState] = None
        self.openai_messages: List[Dict[str, Any]] = []

        # Registries
        self.supervised_functions_registry: Dict[Callable, Dict[str, Any]] = {}
        self.registered_supervisors: Dict[str, UUID] = {}
        self.tool_ids: Dict[str, UUID] = {}
        self.supervisor_ids: Dict[str, UUID] = {}

    def add_event(self, event: dict):
        with self.lock:
            self.langchain_events.append(event)

    def add_metadata(self, key: str, value: Any):
        self.metadata[key] = value

    def to_text(self) -> str:
        """Converts the supervision context into a textual description."""
        texts = []
        with self.lock:
            # Process LangChain events if any
            if self.langchain_events:
                texts.append("## LangChain Events:")
                for event in self.langchain_events:
                    event_description = self._describe_event(event)
                    texts.append(event_description)

            # Process inspect_ai_state if it exists
            if self.inspect_ai_state:
                inspect_ai_text = self._describe_inspect_ai_state()
                texts.append(inspect_ai_text)
                
            # Process OpenAI messages if any
            if self.openai_messages:
                openai_text = self._describe_openai_messages()
                texts.append(openai_text)

        return "\n\n".join(texts)

    def _describe_event(self, event: dict) -> str:
        """Converts a single event into a textual description."""
        event_type = event.get("event", "Unknown Event")
        data = event.get("data", {})
        description = f"### Event: {event_type}\nData:\n```json\n{json.dumps(data, indent=2)}\n```"
        return description

    def _describe_inspect_ai_state(self) -> str:
        """Converts the inspect_ai_state into a textual description."""
        state = self.inspect_ai_state
        texts = []
        
        if state is None:
            return ""

        # Include meta information
        meta_info = f"## Inspect AI State:\n**Model:** {state.model}\n**Sample ID:** {state.sample_id}\n**Epoch:** {state.epoch}"
        texts.append(meta_info)

        # Include messages
        texts.append("### Messages:")
        for message in state.messages:
            message_text = self._describe_chat_message(message)
            texts.append(message_text)

        # Include output if available
        if state.output:
            texts.append("### Output:")
            texts.append(f"```json\n{json.dumps(state.output.dict(), indent=2)}\n```")

        return "\n\n".join(texts)

    def _describe_chat_message(self, message: ChatMessage) -> str:
        """Converts a chat message into a textual description."""
        role = message.role.capitalize()
        text_content = message.text.strip()
        text = f"**{role}:**\n{text_content}"

        if isinstance(message, ChatMessageAssistant) and message.tool_calls:
            text += "\n\n**Tool Calls:**"
            for tool_call in message.tool_calls:
                tool_call_description = self._describe_tool_call(tool_call)
                text += f"\n{tool_call_description}"

        return text

    def _describe_tool_call(self, tool_call: ToolCall) -> str:
        """Converts a ToolCall into a textual description."""
        description = (
            f"- **Tool Call ID:** {tool_call.id}\n"
            f"  - **Function:** {tool_call.function}\n"
            f"  - **Arguments:** `{json.dumps(tool_call.arguments, indent=2)}`\n"
            f"  - **Type:** {tool_call.type}"
        )
        return description

    def _describe_openai_messages(self) -> str:
        """Converts the OpenAI messages into a textual description."""
        messages = [convert_openai_message_to_chat_message(msg) for msg in self.openai_messages]
        texts = []
        texts.append("## OpenAI Messages:")
        for message in messages:
            message_text = self._describe_chat_message(message)
            texts.append(message_text)
        return "\n\n".join(texts)

    # Methods to manage the registries
    def add_supervised_function(self, func: Callable, supervision_functions: List[List[Callable]], ignored_attributes: List[str]):
        with self.lock:
            if func in self.supervised_functions_registry:
                print(f"Function '{func.__qualname__}' is already registered. Skipping.")
                return  # Skip adding the duplicate

            self.supervised_functions_registry[func] = {
                'supervision_functions': supervision_functions or [],
                'ignored_attributes': ignored_attributes or []
            }
            print(f"Registered function '{func.__qualname__}'")

    def get_supervised_function_entry(self, func: Callable) -> Optional[Dict[str, Any]]:
        with self.lock:
            return self.supervised_functions_registry.get(func)

    def update_tool_id(self, func: Callable, tool_id: UUID):
        with self.lock:
            if func in self.supervised_functions_registry:
                self.supervised_functions_registry[func]['tool_id'] = tool_id
                print(f"Updated tool ID for '{func.__qualname__}' to {tool_id}")

    def add_supervisor_id(self, supervisor_name: str, supervisor_id: UUID):
        with self.lock:
            self.registered_supervisors[supervisor_name] = supervisor_id
            print(f"Locally registered supervisor '{supervisor_name}' with ID: {supervisor_id}")

    def get_supervisor_id(self, supervisor_name: str) -> Optional[UUID]:
        with self.lock:
            return self.registered_supervisors.get(supervisor_name)

    def to_task_state(self) -> APITaskState:
        """Converts the supervision context into the API client's TaskState model."""
        if self.inspect_ai_state:
            return convert_task_state(self.inspect_ai_state)
        elif self.openai_messages: #TODO: this is when user uses directly openai messages, update
            # Convert OpenAI messages to TaskState
            messages = [convert_openai_message_to_chat_message(msg) for msg in self.openai_messages]
            task_state = TaskState(
                model='model_name',  # TODO: Update with actual model name
                sample_id=str(uuid4()),
                epoch=1,
                input=messages[0].text if messages else "",
                messages=messages,
                completed=False,
                metadata={}
            )
            return convert_task_state(task_state)
        elif self.langchain_events:
            messages = extract_messages_from_events(self.langchain_events)

            # Create your custom TaskState object
            # TODO: Initialize direct API TaskState ?
            task_state = TaskState(
                model='model_name',  # TODO: Update with actual model name
                sample_id=str(uuid4()),
                epoch=1,
                input=messages[0].content if messages else "",
                messages=messages,
                completed=False,
                metadata={}
            )
            return convert_task_state(task_state)
        else:
            # Handle case when no messages are available
            return APITaskState(
                messages=[],
                tools=[],
                output=None,
                completed=False,
                tool_choice=None,
            )
        
    def get_messages(self) -> List[ChatMessage]:
        """Get the messages from the supervision context as ChatMessage objects."""
        if self.openai_messages:
            # Convert OpenAI messages to ChatMessage objects
            return [convert_openai_message_to_chat_message(msg) for msg in self.openai_messages]
        elif self.langchain_events:
            # Extract messages from LangChain events
            return extract_messages_from_events(self.langchain_events)
        elif self.inspect_ai_state:
            # Use messages from inspect_ai_state
            return self.inspect_ai_state.messages
        else:
            # No messages available
            return []
        
    def get_api_messages(self) -> List[Message]:
        """Get the messages from the supervision context in the API client's Message model."""
        return [convert_message(msg) for msg in self.get_messages()]
        
    def update_openai_messages(self, messages: List[Dict[str, Any]]):
        """Updates the context with a list of OpenAI messages."""
        with self.lock:
            self.openai_messages = messages.copy()

class SupervisionConfig:
    def __init__(self):
        self.global_supervision_functions: List[Callable] = []
        self.global_mock_policy = MockPolicy.NO_MOCK
        self.override_local_policy = False
        self.mock_responses: Dict[str, List[Any]] = {}
        self.previous_calls: Dict[str, List[Any]] = {}
        self.function_supervisors: Dict[str, List[Callable]] = {}  # Function-specific supervision chains
        self.llm = None
        self.context = SupervisionContext()
        self.run_id: Optional[UUID] = None
        self.client = None # Sentinel API client
        self.local_supervisors_by_id: Dict[UUID, Callable] = {}

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

    def set_run_id(self, run_id: UUID):
        self.run_id = run_id

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

    def add_local_supervisor(self, supervisor_id: UUID, supervisor_func: Callable):
        """Add a supervisor function to the config."""
        self.local_supervisors_by_id[supervisor_id] = supervisor_func

    def get_supervisor_by_id(self, supervisor_id: UUID) -> Optional[Callable]:
        """Retrieve a supervisor function by its ID."""
        return self.local_supervisors_by_id.get(supervisor_id)

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

def convert_task_state(task_state: TaskState) -> APITaskState:
    """
    Converts Inspect AI TaskState object into the Sentinel API client's TaskState model.
    They should be identical when serialized.
    """
    from entropy_labs.sentinel_api_client.sentinel_api_client.types import UNSET
    from entropy_labs.sentinel_api_client.sentinel_api_client.models.task_state import TaskState as APITaskState

    # Convert messages
    messages = [convert_message(msg) for msg in task_state.messages]

    # Convert tools - we can't do this because Inspect AI doesn't have tool names    
    tools: list[ApiTool] = []

    # Convert output
    output = convert_output(task_state.output)

    # Convert tool_choice
    tool_choice = UNSET

    return APITaskState(
        messages=messages,
        tools=tools,
        output=output,
        completed=False,
        tool_choice=tool_choice,
    )

def convert_message(msg: ChatMessage) -> Message:
    """
    Converts a ChatMessage to a Message in the API model.
    """

    # Convert content to a string
    if isinstance(msg.content, str):
        content_str = msg.content
    elif isinstance(msg.content, list):
        content_str = "\n".join([convert_content(content) for content in msg.content])
    else:
        content_str = ""
        
    if isinstance(msg, ChatMessageTool):
        tool_call_id = msg.tool_call_id
        function = msg.function
    else:
        tool_call_id = UNSET
        function = UNSET

    if isinstance(msg, ChatMessageAssistant) and msg.tool_calls:
        tool_calls = [convert_tool_call(tool_call) for tool_call in msg.tool_calls]
    else:
        tool_calls = UNSET

    return Message(
        source=msg.source,
        role=msg.role if msg.role != 'tool' else 'assistant',
        content=content_str,
        tool_calls=tool_calls,
        tool_call_id=tool_call_id,
        function=function
    )


def convert_content(content: Content) -> str:
    # Convert Content to a string representation
    if isinstance(content, ContentText):
        return content.text
    elif hasattr(content, 'data_url'):
        return content.data_url  # For images or other content types
    else:
        return str(content)

def convert_tool_call(tool_call: ToolCall) -> ApiToolCall:
    return ApiToolCall(
        id=tool_call.id,
        function=tool_call.function,
        arguments=ToolCallArguments.from_dict(tool_call.arguments),
        type=tool_call.type
    )


def convert_output(output: ModelOutput) -> ApiOutput:
    from entropy_labs.sentinel_api_client.sentinel_api_client.models.output import Output as ApiOutput

    return ApiOutput(
        model=output.model if output.model else UNSET,
        choices=UNSET, #TODO: Implement if needed
        usage=UNSET
    )

def convert_openai_message_to_chat_message(openai_msg: Dict[str, Any]) -> ChatMessage:
    """
    Converts an OpenAI message dict to a ChatMessage object.
    """
    role = openai_msg.get('role')
    content = openai_msg.get('content', '')
    tool_calls = openai_msg.get('tool_calls', None)

    # Ensure content is not None
    if content is None:
        content = ''

    if role == 'assistant':
        # Handle assistant message with potential tool calls
        if tool_calls:
            # Convert function call to ToolCall
            tool_calls = [convert_openai_tool_call(tool_call) for tool_call in tool_calls]
            return ChatMessageAssistant(content=content, tool_calls=tool_calls)
        else:
            return ChatMessageAssistant(content=content)
    elif role == 'user':
        return ChatMessageUser(content=content)
    elif role == 'system':
        return ChatMessageSystem(content=content)
    elif role == 'tool':
        # Handle tool messages
        name = openai_msg.get('name')
        tool_call_id = openai_msg.get('tool_call_id')
        return ChatMessageTool(content=content, function=name, tool_call_id=tool_call_id)
    else:
        # Default handling as user message
        return ChatMessageUser(content=content)

def convert_openai_tool_call(tool_call_dict: Dict[str, Any]) -> ToolCall:
    """
    Converts an OpenAI tool call dict to a ToolCall object.
    """
    id_ = tool_call_dict.get('id')
    function = tool_call_dict.get('function', {})
    function_name = function.get('name')
    arguments = function.get('arguments', {})
    
    if isinstance(arguments, str):
        try:
            arguments = json.loads(arguments)
        except json.JSONDecodeError:
            arguments = {}
    
    type_ = tool_call_dict.get('type', '')

    return ToolCall(id=id_, function=function_name, arguments=arguments, type=type_)
