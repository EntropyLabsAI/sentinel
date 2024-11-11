import json
from typing import Callable, get_type_hints, Any, Optional
from pydantic import BaseModel, create_model, Field
import inspect
from openai.types.chat import ChatCompletionToolParam
from openai import OpenAI
import base64
import hashlib
from entropy_labs.supervision import supervise
from entropy_labs.api.project_registration import register_project, create_run, register_task
from entropy_labs.supervision.supervisors import human_supervisor, llm_supervisor
from entropy_labs.supervision.config import SupervisionDecision, SupervisionDecisionType, SupervisionContext, get_supervision_context
from entropy_labs.supervision.supervisors import Supervisor
from enum import Enum

client = OpenAI()

PASSWORD = "ENTROPYLABS"
YOUR_EMAIL = "" #Fill it out
ENTROPY_LABS_EMAIL = "devs@entropy-labs.ai"

EMAIL_INVITATION_POLICY = "If the email is about inviting to an event, check that the email is inviting to the correct event given the previous messages, that the dates are correct, the email is brief and that the sender is authenticated."


def check_email_address_supervisor(whitelisted_emails: list[str]) -> Supervisor:
    """
    Factory function that creates a supervisor function to check if the email address
    is in the whitelist.

    Args:
        whitelisted_emails (list[str]): List of whitelisted email addresses.

    Returns:
        Supervisor: A supervisor function checking the email address against the whitelist.
    """
    def supervisor(
        func: Callable,
        supervision_context: SupervisionContext,
        tool_kwargs: dict[str, Any],
        **kwargs
    ) -> SupervisionDecision:
        """Checks if the email address is in the whitelist."""
        to_email = tool_kwargs.get('to')
        if to_email is None:
            return SupervisionDecision(
                decision=SupervisionDecisionType.ESCALATE,
                explanation="No email address provided."
            )
        if to_email in whitelisted_emails:
            return SupervisionDecision(
                decision=SupervisionDecisionType.APPROVE,
                explanation=f"The email address '{to_email}' is allow."
            )
        else:
            return SupervisionDecision(
                decision=SupervisionDecisionType.ESCALATE,
                explanation=f"The email address '{to_email}' is not in the whitelist."
            )
    supervisor.__name__ = check_email_address_supervisor.__name__
    supervisor.supervisor_attributes = {"whitelisted_emails": whitelisted_emails}
    return supervisor


# Tools
@supervise()
def internet_search(query: str, max_results: int = 3) -> str:
    """
    Search the internet for information using DuckDuckGo and fetch content from the first few links.

    Parameters:
    - query (str): Query to search the internet with.
    - max_results (int): Maximum number of results to fetch content from.

    Returns:
    - str: Concatenated content from the search results.
    """
    from duckduckgo_search import DDGS
    import requests
    from bs4 import BeautifulSoup

    region = "wt-wt"  # Worldwide
    safesearch = "moderate"

    try:
        with DDGS() as ddgs:
            results = []
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
    
    
@supervise(supervision_functions=[[llm_supervisor(instructions=EMAIL_INVITATION_POLICY), human_supervisor()], [check_email_address_supervisor(whitelisted_emails=[ENTROPY_LABS_EMAIL]), human_supervisor()]])
def send_email(to: str, subject: str, body: str):
    """
    Send an email to the specified recipient.

    Parameters:
    - to (str): Recipient's email address.
    - subject (str): Subject of the email.
    - body (str): Body content of the email.

    Returns:
    - str: A message indicating the result of the email sending process.
    """
    return f"Email sent to {to} with subject '{subject}'"
    # TODO: Implement
    # import smtplib
    # from email.mime.text import MIMEText

    # # Email account credentials
    # smtp_server = 'smtp.gmail.com'
    # smtp_port = 587
    # sender_email = 'sentinel.demo.el@gmail.com'
    # s_pd = "ScWBH3hS%JSWPE5C8jg"  #base64.urlsafe_b64encode(hashlib.sha256(sender_email.encode()).digest()).decode()[:20]

    # # Create the email message
    # message = MIMEText(body)
    # message['From'] = sender_email
    # message['To'] = to
    # message['Subject'] = subject

    # try:
    #     # Connect to the SMTP server and send the email
    #     with smtplib.SMTP(smtp_server, smtp_port) as server:
    #         server.starttls()
    #         server.login(sender_email, s_pd)
    #         server.send_message(message)
    #     return f"Email sent to {to} with subject '{subject}'"
    # except Exception as e:
    #     return f"Failed to send email: {str(e)}"
@supervise()
def create_calendar_event(title: str, start_time: str, end_time: str):
    return f"Event '{title}' created from {start_time} to {end_time}"

@supervise()
def authenticate_user(user_name: str, password: str) -> bool:
    """
    Authenticate user by password.
    """
    return password == PASSWORD


def create_openai_tool(func: Callable) -> dict:
    """
    Create an OpenAI tool definition from a function, conforming to OpenAI's expected schema.
    """
    import inspect
    from typing import get_type_hints, Any
    from enum import Enum

    signature = inspect.signature(func)
    type_hints = get_type_hints(func)

    def python_type_to_json_type(py_type: Any) -> str:
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

    parameters: dict[str, Any] = {
        "type": "object",
        "properties": {},
        "additionalProperties": False  # Ensure no additional properties are allowed
    }

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

        parameters["properties"][param_name] = param_schema

        # Include all parameters in required when strict is True
        required_params.append(param_name)

    parameters["required"] = required_params

    # Build the function definition
    return {
        "type": "function",
        "function": {
            "name": func.__name__,
            "description": func.__doc__ or "",
            "parameters": parameters,
            "strict": True
        }
    }

# Function to interact with OpenAI
def chat_with_openai(messages, tools):
    completion = client.chat.completions.create(
        model="gpt-4o",
        messages=messages,
        tools=tools,
        parallel_tool_calls=False)
    return completion

# Function to execute tool calls
def execute_tool_call(tool_call, tools):
    function_name = tool_call.function.name
    arguments = json.loads(tool_call.function.arguments)
    print(f"Executing tool call: {function_name} with arguments: {arguments}")
    
    # Map function names to actual functions
    for tool in tools:
        if tool.__name__ == function_name:
            result = tool(**arguments)
            print(f"Tool call {function_name} successful.")
            return result
    print("Function not found.")
    return "Function not found."

def update_messages(messages: list[dict], new_message: dict | Any) -> list[dict]:
    supervision_context = get_supervision_context()
    if not isinstance(new_message, dict):
        new_message = new_message.to_dict()
    supervision_context.update_openai_messages(messages + [new_message])
    return messages + [new_message]

# Function to start the chatbot
def start_chatbot(start_prompt: str, tools: list[Callable]):
    print("👋 Welcome! I'm your chatbot. Type 'exit' to end the chat.\n")
    openai_tools = [create_openai_tool(func) for func in tools]
    
    messages = [{"role": "system", "content": "You are a helpful assistant."}]
    
    
    if not start_prompt:
        start_prompt = input("Input a message: ")
    messages = update_messages(messages, {"role": "user", "content": start_prompt})    
    
    while True:
        response = chat_with_openai(messages, openai_tools)
        messages = update_messages(messages, response.choices[0].message)
        
        # Check if the response includes a tool call
        tool_calls = response.choices[0].message.tool_calls
        if tool_calls:
            tool_arguments = ", ".join([f"{arg_name}: {arg_value}" for arg_name, arg_value in json.loads(tool_calls[0].function.arguments).items()])
            print(f"Bot decided to call a tool: {tool_calls[0].function.name} with arguments: {tool_arguments}")
            tool_call = tool_calls[0]
            result = execute_tool_call(tool_call, tools)
            
            # Create a message containing the result of the function call
            function_call_result_message = {
                "role": "tool",
                "content": json.dumps(result),
                "tool_call_id": tool_call.id
            }
            
            # Send the tool call result back to the model
            messages = update_messages(messages, function_call_result_message)
        else:
            #ask user for new input
            print(f"Bot: {response.choices[0].message.content}\n")
            user_input = input("Input a message: ")
            if user_input.lower() == 'exit':
                print("Goodbye! 👋")
                break
            messages = update_messages(messages, {"role": "user", "content": user_input})
            
        
        
        

# Start the chatbot
if __name__ == "__main__":
    start_prompt = "Go and find most interesting events happening in AI next week in San Francisco. Then create a calendar event for the most interesting one. After done ask me email addresses where you should send invitations for that event."
        
    tools = [
        internet_search, 
        send_email, 
        create_calendar_event, 
        authenticate_user
    ]
    
    # Register project, task and run with Entropy Labs. Entropy Labs docker container needs to be running.
    entropy_labs_backend_url = "http://localhost:8080"
    project_id = register_project(project_name="Email Assistant", entropy_labs_backend_url=entropy_labs_backend_url)
    task_id = register_task(project_id=project_id, task_name="Email Assistant")
    run_id = create_run(project_id=project_id, task_id=task_id, tools=tools)
    
    start_chatbot(start_prompt, tools)