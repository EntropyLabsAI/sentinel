from el.supervision.config import (
    set_global_mock_policy,
    setup_sample_from_previous_calls
)
from el.supervision.supervisors import divide_supervisor
from el.supervision.langchain.supervisors import human_supervisor
from el.supervision.langchain.logging import EntropyLabsCallbackHandler
from el.supervision.decorators import supervise
from el.mocking.policies import MockPolicy
from langchain_openai import ChatOpenAI
from langchain.tools import tool
from langchain_core.messages import HumanMessage
from pydantic import BaseModel
from typing import List
import os
import glob

# Define tool functions with supervision decorators
@tool
@supervise()
def add(a: float, b: float) -> float:
    """Add two numbers."""
    return a + b

@tool
@supervise()
def divide(a: float, b: float) -> float:
    """Divide two numbers."""
    print(f"Dividing {a} and {b}")
    return a / b

class UploadResponse(BaseModel):
    status: str
    data: List[dict]
    message: str

@tool
@supervise(
    supervision_functions=[human_supervisor()]
)
def upload_api(input_data: str) -> UploadResponse:
    """Upload the input data to the API and receive a response."""
    # Example API response
    # Actual API call code would go here
    return UploadResponse(
        status="success",
        data=[{"id": 1, "name": "example"}],
        message="Upload successful"
    )

def run_agent(llm_with_tools, messages, callbacks):
    """Runs the agent loop, invoking tools as needed."""
    while True:
        # LLM generates a response based on messages
        ai_msg = llm_with_tools.invoke(messages, config={"callbacks": callbacks})
        messages.append(ai_msg)
        
        if ai_msg.tool_calls:
            # If the LLM calls any tools, invoke them
            for tool_call in ai_msg.tool_calls:
                selected_tool = {
                    "add": add,
                    "divide": divide,
                    "upload_api": upload_api
                }.get(tool_call['name'].lower())
                tool_msg = selected_tool.invoke(tool_call, config={"callbacks": callbacks})
                messages.append(tool_msg)
        else:
            # If no tools are called, print the final response and exit
            print(ai_msg.content)
            break

if __name__ == "__main__":
    # Initialize tools and logging
    tools = [add, divide, upload_api]
    log_directory = ".logs/reuse_runs_example"

    # First run: Create logs
    # Create the callback handler with single_log_file=True
    callbacks = [EntropyLabsCallbackHandler(
        tools=tools,
        log_directory=log_directory,
        single_log_file=True
    )]
    # Initialize the LLM
    llm = ChatOpenAI(model="gpt-4", temperature=0)
    # Bind tools to the LLM
    llm_with_tools = llm.bind_tools(tools)
    # Initial message from the user
    messages = [HumanMessage(content="Divide 12 by 3, then add 5 to the result. Finally call upload_api with the result.")]
    # Run the agent
    run_agent(llm_with_tools, messages, callbacks)

    # Second run: Reuse the last execution log
    # Initialize new callbacks without single_log_file
    new_callbacks = [EntropyLabsCallbackHandler(
        tools=tools,
        log_directory=log_directory
    )]
    # Set the global mock policy to sample from previous calls
    set_global_mock_policy(MockPolicy.SAMPLE_PREVIOUS_CALLS, override_local_policy=True)
    # Get the latest log file
    log_files = glob.glob(os.path.join(log_directory, '*'))
    log_file_path = max(log_files, key=os.path.getmtime) if log_files else None
    if log_file_path:
        # Setup sampling from previous calls using the log file
        setup_sample_from_previous_calls(log_file_path)
        # Reinitialize messages
        messages = [HumanMessage(content="Divide 12 by 3, then add 5 to the result. Finally call upload_api with the result.")]
        # Reinitialize the LLM and bind tools
        llm = ChatOpenAI(model="gpt-4", temperature=0)
        llm_with_tools = llm.bind_tools(tools)
        # Run the agent again
        run_agent(llm_with_tools, messages, new_callbacks)
