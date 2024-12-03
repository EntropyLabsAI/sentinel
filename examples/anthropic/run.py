import json
import inspect
from typing import Callable, Any, List, Dict, get_type_hints, Optional
from enum import Enum
from uuid import UUID, uuid4

import requests
from bs4 import BeautifulSoup
import base64
from litellm import completion
import litellm
from duckduckgo_search import DDGS

from entropy_labs.supervision import supervise
from entropy_labs.api import register_project, create_run, register_task, submit_run_status, submit_run_result, Status
from entropy_labs.supervision.supervisors import human_supervisor, llm_supervisor, Supervisor
from entropy_labs.supervision.config import (
    SupervisionDecision,
    SupervisionDecisionType,
    SupervisionContext,
    get_supervision_context,
)

# Initialize the Anthropic client
# client = anthropic.Anthropic()


ENTROPY_LABS_EMAIL = "devs@entropy-labs.ai"

def check_email_address_supervisor(whitelisted_emails: List[str]) -> Supervisor:
    """
    Factory function that creates a supervisor function to check if the email address
    is in the whitelist.

    Args:
        whitelisted_emails (List[str]): List of whitelisted email addresses.

    Returns:
        Supervisor: A supervisor function checking the email address against the whitelist.
    """
    def supervisor(
        func: Callable,
        supervision_context: SupervisionContext,
        tool_kwargs: Dict[str, Any],
        **kwargs
    ) -> SupervisionDecision:
        """
        Checks if the email address is in the whitelist.

        Args:
            func (Callable): The function being supervised.
            supervision_context (SupervisionContext): Context of the supervision.
            tool_kwargs (Dict[str, Any]): Keyword arguments for the tool function.

        Returns:
            SupervisionDecision: The decision to approve or escalate the tool call.
        """
        to_email = tool_kwargs.get('to')
        if to_email is None:
            return SupervisionDecision(
                decision=SupervisionDecisionType.ESCALATE,
                explanation="No email address provided."
            )
        if to_email in whitelisted_emails:
            return SupervisionDecision(
                decision=SupervisionDecisionType.APPROVE,
                explanation=f"The email address '{to_email}' is allowed."
            )
        else:
            return SupervisionDecision(
                decision=SupervisionDecisionType.ESCALATE,
                explanation=(
                    f"The email address '{to_email}' is not in the whitelist. "
                    f"Whitelisted emails: {whitelisted_emails}"
                )
            )
    supervisor.__name__ = check_email_address_supervisor.__name__
    supervisor.supervisor_attributes = {"whitelisted_emails": whitelisted_emails}
    return supervisor


@supervise()
def internet_search(query: str, max_results: int = 3) -> str:
    """
    Search the internet for information using DuckDuckGo and fetch content from the first few links.

    Parameters:
        query (str): Query to search the internet with.
        max_results (int): Maximum number of results to fetch content from.

    Returns:
        str: Concatenated content from the search results.
    """
    # Define search settings
    region = "wt-wt"  # Worldwide
    safesearch = "moderate"  # Safe search level

    try:
        with DDGS() as ddgs:
            results = []
            # Perform the search
            search_results = ddgs.text(
                query,
                region=region,
                safesearch=safesearch,
                max_results=max_results
            )

            for r in search_results:
                result = {
                    'title': r['title'],
                    'href': r['href'],
                    'snippet': r['body'],
                    'content': ''
                }
                try:
                    # Fetch content from the link
                    response = requests.get(r['href'], timeout=5)
                    soup = BeautifulSoup(response.content, 'html.parser')
                    # Extract text content from the web page
                    paragraphs = soup.find_all('p')
                    text_content = ' '.join([p.get_text() for p in paragraphs])
                    result['content'] = text_content
                except Exception as e:
                    print(f"Error fetching content from {r['href']}: {str(e)}")
                results.append(result)

            # Combine content from all results
            combined_content = '\n'.join([res['content'] for res in results if res['content']])
            return combined_content if combined_content else f"No content found for '{query}'."
    except Exception as e:
        print(f"Error performing search: {str(e)}")
        return f"Error performing search for '{query}'."

# Policies
EMAIL_INVITATION_POLICY = (
    "Ensure that the email invitation is clear, concise, and includes all necessary details about the event. "
    "Verify that the recipient's email address is correct given the previous messages."
)


@supervise(supervision_functions=[
    [check_email_address_supervisor(whitelisted_emails=[ENTROPY_LABS_EMAIL])],
    [llm_supervisor(instructions=EMAIL_INVITATION_POLICY), human_supervisor()]
])
def send_email(to: str, subject: str, body: str):
    """
    Send an email to the specified recipient.

    Parameters:
        to (str): Recipient's email address.
        subject (str): Subject of the email.
        body (str): Body content of the email.

    Returns:
        str: A message indicating the result of the email sending process.
    """
    # Mocking the email sending process
    return f"Email sent to {to} with subject '{subject}'"


# ### `create_calendar_event`
# 
# Creates a calendar event, supervised to ensure correct parameters given the previous messages.

CORRECT_TOOL_PARAMETERS_POLICY = (
    "Make sure that the tool parameters are correct given the previous messages. If incorrect, fix them."
)


@supervise(supervision_functions=[[llm_supervisor(instructions=CORRECT_TOOL_PARAMETERS_POLICY)]])
def create_calendar_event(title: str, start_time: str, end_time: str):
    """
    Create a calendar event.

    Parameters:
        title (str): Title of the event.
        start_time (str): Start time of the event.
        end_time (str): End time of the event.

    Returns:
        str: A message indicating the result of the calendar event creation process.
    """
    # Mocking the calendar event creation process
    return f"Event '{title}' created from {start_time} to {end_time}"


# ### `book_flight`
# 
# Books a flight ticket, requiring human supervision always.

@supervise(supervision_functions=[[human_supervisor()]])
def book_flight(departure_city: str, arrival_city: str, datetime: str, maximum_price: float):
    """
    Book a flight ticket.

    Parameters:
        departure_city (str): Departure city.
        arrival_city (str): Arrival city.
        datetime (str): Departure date and time.
        maximum_price (float): Maximum acceptable price for the flight.

    Returns:
        str: A message indicating the result of the flight booking process.
    """
    # Mocking the flight booking process
    return f"Flight booked from {departure_city} to {arrival_city} on {datetime}."


# ## Assistant Agent
# 
# Let's build the assistant agent that will interact with the user, utilize the tools, and handle OpenAI responses.

# ### Creating OpenAI Tools
# 
# We need to create tool definitions that conform to OpenAI's expected schema.

def create_anthropic_tool(func: Callable) -> Dict[str, Any]:
    """
    Create a tool definition for Anthropic's API from a function.
    """
    signature = inspect.signature(func)
    type_hints = get_type_hints(func)

    def python_type_to_json_type(py_type: Any) -> str:
        """
        Convert a Python type to a JSON schema type.
        """
        if py_type == str:
            return "string"
        elif py_type == int:
            return "integer"
        elif py_type == float:
            return "number"
        elif py_type == bool:
            return "boolean"
        elif py_type == dict:
            return "object"
        elif py_type == list:
            return "array"
        elif isinstance(py_type, type) and issubclass(py_type, Enum):
            return "string"
        else:
            return "string"  # Default to string for unsupported types

    properties = {}
    required_params = []

    for param_name, param in signature.parameters.items():
        param_type = type_hints.get(param_name, str)
        param_schema = {
            "type": python_type_to_json_type(param_type),
            "description": param_name
        }

        # If the parameter has an Enum type, add the enum options
        if isinstance(param_type, type) and issubclass(param_type, Enum):
            param_schema["enum"] = [e.value for e in param_type]

        properties[param_name] = param_schema
        required_params.append(param_name)

    input_schema = {
        "type": "object",
        "properties": properties,
        "required": required_params
    }

    # Build the tool definition
    return {
        "name": func.__name__,
        "description": func.__doc__ or "",
        "input_schema": input_schema
    }


# ### Chat Function with OpenAI
# 
# Handles interaction with the OpenAI GPT model.

def chat_with_anthropic(messages: List[Dict], tools: List[dict]):
    """
    Interact with the model using LiteLLM.
    """
    response = completion(
        model="claude-3-opus-20240229",
        messages=messages,
        tools=tools,
        max_tokens=1024,
    )

    # If there isn't a choice, raise an error
    if len(response.choices) == 0:
        raise ValueError("No choices returned from the model")

    # Need to access the actual message through choices[0].message
    message = response.choices[0].message
    new_message = {
        "role": "assistant",
        "content": message.content,
    }
    messages = update_messages(messages, new_message, run_id)
    return response


# ### Executing Tool Calls
# 
# Executes the tool calls as decided by the assistant.

def execute_tool_call(tool_use, tools):
    """
    Execute a tool call as decided by the assistant.
    """
    function_name = tool_use.name
    # Parse the JSON string from arguments instead of accessing input
    arguments = json.loads(tool_use.arguments)
    print(f"Executing tool call: {function_name} with arguments: {arguments}")

    # Map function names to actual functions
    for tool in tools:
        if tool.__name__ == function_name:
            result = tool(**arguments)
            print(f"Tool call {function_name} result: {result}")
            return result

    print("Function not found.")
    return "Function not found."


# ### Updating Messages
# 
# Updates the conversation messages and the supervision context.

def update_messages(
    messages: List[Dict],
    new_message: Dict[str, Any],
    run_id: UUID
) -> List[Dict]:
    """
    Update the conversation messages and supervision context.

    Parameters:
        messages (List[Dict]): The current list of messages.
        new_message (Dict[str, Any] | ChatCompletionMessage): The new message to add.
        run_id (UUID): The ID of the current run.

    Returns:
        List[Dict]: The updated list of messages.
    """
    supervision_context = get_supervision_context(run_id)
    supervision_context.update_messages(messages + [new_message])
    return messages + [new_message]


# ### Starting the Chatbot
# 
# The main loop that starts the chatbot and handles user interaction.

def in_jupyter_notebook():
    try:
        from IPython import get_ipython
        shell = get_ipython().__class__.__name__
        if shell == 'ZMQInteractiveShell':
            return True   # Jupyter notebook or qtconsole
        else:
            return False  # Other type (likely standard Python interpreter)
    except NameError:
        return False      # Probably standard Python interpreter

def encode_image(image_path):
    with open(image_path, "rb") as image_file:
        return base64.b64encode(image_file.read()).decode('utf-8')

def start_chatbot(
    start_prompt: str,
    tools: List[Callable],
    run_id: UUID,
):
    """
    Start the chatbot interaction, working both in Jupyter Notebook and standard Python script.
    """
    # Detect environment
    is_jupyter = in_jupyter_notebook()

    if is_jupyter:
        # Import Jupyter-specific modules inside the conditional block
        import ipywidgets as widgets
        from IPython.display import display

        # Create the output and input widgets
        output_widget = widgets.Output(layout={'border': '1px solid black'})
        input_widget = widgets.Text(
            value='',
            placeholder='Type your message here...',
            description='You:',
            disabled=False,
            continuous_update=False  # Update value only when Enter is pressed
        )

        # Display the widgets
        display(output_widget)
        display(input_widget)

        # Clear the output widget to avoid duplicate messages
        output_widget.clear_output()
    else:
        output_widget = None
        input_widget = None



    # Create Anthropic tool definitions
    anthropic_tools = [create_anthropic_tool(func) for func in tools]

    waiting_for_user_input = False

    def process_assistant_response(assistant_message):
        """
        Process the assistant's response, executing tool calls as needed.
        """
        nonlocal messages, waiting_for_user_input

        stop_reason = assistant_message.choices[0].finish_reason
        message = assistant_message.choices[0].message

        # Add assistant's message to conversation
        messages.append({
            "role": "assistant",
            "content": message.content,
            "tool_calls": message.tool_calls if hasattr(message, 'tool_calls') else None
        })

        # Print assistant's content
        if is_jupyter and output_widget:
            with output_widget:
                print(f"Assistant: {message.content}\n")
        else:
            print(f"Assistant: {message.content}\n")

        # Check for tool calls
        if stop_reason == "tool_calls" and message.tool_calls:
            tool_call = message.tool_calls[0]  # Handle first tool call
            if tool_call:
                # Execute the tool call with parsed arguments
                result = execute_tool_call(tool_call.function, tools)

                # Add tool result as a tool message
                messages.append({
                    "role": "tool",
                    "content": str(result),
                    "tool_call_id": tool_call.id,
                    "name": tool_call.function.name,
                })

                # Get assistant's response to the tool result
                message = chat_with_anthropic(messages, anthropic_tools)
                process_assistant_response(message)
        else:
            waiting_for_user_input = True

    messages = []

    # Start the conversation
    user_message = {"role": "user", "content": start_prompt}
    messages.append(user_message)
    
    if is_jupyter and output_widget:
        with output_widget:
            print(f"You: {start_prompt}\n")
    else:
        print(f"You: {start_prompt}\n")

    # Get assistant's initial response
    message = chat_with_anthropic(messages, anthropic_tools)
    process_assistant_response(message)

    # Input handling
    if is_jupyter and input_widget:
        # Remove existing event handlers to prevent multiple triggers
        input_widget.unobserve_all()

        def handle_submit(change):
            nonlocal messages, waiting_for_user_input

            if not waiting_for_user_input or not change['new']:
                return  # Ignore input if not waiting or empty

            user_input = change['new']
            input_widget.value = ''  # Clear input box

            if user_input.lower() == 'exit':
                input_widget.disabled = True
                if output_widget:
                    with output_widget:
                        print("Goodbye! 👋")
                else:
                    print("Goodbye! 👋")
                return

            # Add user message to conversation
            user_message = {"role": "user", "content": user_input}
            messages.append(user_message)

            if output_widget:
                with output_widget:
                    print(f"You: {user_input}\n")
            else:
                print(f"You: {user_input}\n")

            waiting_for_user_input = False

            # Get assistant response
            message = chat_with_anthropic(messages, anthropic_tools)
            process_assistant_response(message)

        # Register the event handler
        input_widget.observe(handle_submit, names='value')
    else:
        # Standard input/output loop
        while True:
            if waiting_for_user_input:
                user_input = input("You: ").strip()
                if user_input.lower() == 'exit':
                    print("Goodbye! 👋")
                    break
                if not user_input:
                    continue  # Ignore empty input

                # Add user message to conversation
                user_message = {"role": "user", "content": user_input}
                messages.append(user_message)

                print(f"You: {user_input}\n")
                waiting_for_user_input = False

                # Get assistant response
                message = chat_with_anthropic(messages, anthropic_tools)
                process_assistant_response(message)
            else:
                # Waiting for assistant to process
                pass
    submit_run_status(run_id=run_id, status=Status.COMPLETED)
    submit_run_result(run_id=run_id, result="passed")

# ## Running the Assistant
# 
# Finally, we set up the environment and start the assistant. You can now interact with the assistant by typing messages in the input box and see the supervisors in your web browser at: http://localhost:3000/.

# Define the initial prompt for the chatbot
start_prompt = (
    "Go and find the most interesting events happening in AI next week in San Francisco. "
    "Then create a calendar event for the most interesting one. When done, ask me email addresses "
    "where you should send invitations for that event. After the email is sent, book me a flight ticket "
    "from London to San Francisco to attend that event. Make sure that the flight price is less than 1000 GBP."
)

# List of tools available to the assistant
tools = [
    internet_search,
    send_email,
    create_calendar_event,
    book_flight
]

# Register project, task, and run with Entropy Labs
entropy_labs_backend_url = "http://localhost:8099"

# Entropy Labs backend needs to be running
project_id = register_project(
    project_name="Email Assistant",
    entropy_labs_backend_url=entropy_labs_backend_url
)
task_id = register_task(project_id=project_id, task_name="Email Assistant")
run_id = create_run(project_id=project_id, task_id=task_id, tools=tools)

# Start the chatbot
start_chatbot(start_prompt, tools, run_id)


