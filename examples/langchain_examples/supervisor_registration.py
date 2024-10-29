from entropy_labs.supervision.config import supervision_config
from entropy_labs.supervision.langchain.supervisors import human_supervisor
from entropy_labs.supervision.langchain.logging import EntropyLabsCallbackHandler
from entropy_labs.supervision.decorators import supervise
from entropy_labs.api.sentinel_api_client_helper import register_project, create_run, register_tools_and_supervisors
from entropy_labs.sentinel_api_client.sentinel_api_client import Client
from entropy_labs.supervision.supervisors import Supervisor
from langchain_openai import ChatOpenAI
from langchain.tools import tool
from langchain_core.messages import HumanMessage
from pydantic import BaseModel
from typing import List, Callable
from entropy_labs.supervision.config import (
    SupervisionDecision,
    SupervisionDecisionType,
    SupervisionContext,
)


def divide_supervisor(min_a: int = 0, min_b: int = 0) -> Supervisor:
    """
    Factory function that creates a supervisor function to check if both
    numbers are above specified minimums.

    Args:
        min_a (int): Minimum allowed value for the first argument. Default is 0.
        min_b (int): Minimum allowed value for the second argument. Default is 0.

    Returns:
        Supervisor: A supervisor function checking the arguments against minimums.
    """
    def supervisor(
        func: Callable,
        supervision_context: SupervisionContext,
        a: float,
        b: float,
        **kwargs
    ) -> SupervisionDecision:
        """Checks if both numbers are above the specified minimums."""
        if a > min_a and b > min_b:
            return SupervisionDecision(
                decision=SupervisionDecisionType.APPROVE,
                explanation=f"Both numbers are above their minimums (a > {min_a}, b > {min_b})."
            )
        else:
            return SupervisionDecision(
                decision=SupervisionDecisionType.ESCALATE,
                explanation=f"One or both numbers are below their minimums (a ≤ {min_a} or b ≤ {min_b})."
            )
    return supervisor
# Define tool functions with supervision decorators
@tool
@supervise()
def add(a: float, b: float) -> float:
    """Add two numbers."""
    return a + b

@tool
@supervise(supervision_functions=[divide_supervisor(min_b=0)])
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
    supervision_functions=[
        human_supervisor(backend_api_endpoint="http://localhost:8080"),
    ]
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
    client = Client(base_url="http://localhost:8080")

    supervision_config.client = client #TODO: Move this somewhere else
    # Register the project and create a run
    project_id = register_project(client, "langchain-example")
    run_id = create_run(client, project_id)

    # Automatically register tools and supervisors
    register_tools_and_supervisors(client, run_id)

    # Initialize tools and logging
    tools = [add, divide, upload_api]
    log_directory = ".logs/supervisor_registration_example"

    # First run: Create logs
    callbacks = [EntropyLabsCallbackHandler(
        tools=tools,
        log_directory=log_directory,
        single_log_file=True,
        run_id=run_id  # Pass the run_id to the callback handler
    )]
    # Initialize the LLM
    llm = ChatOpenAI(model="gpt-4", temperature=0)
    # Bind tools to the LLM
    llm_with_tools = llm.bind_tools(tools)
    # Initial message from the user
    messages = [HumanMessage(content="Divide 12 by 3, then add 5 to the result. Finally, call upload_api with the result.")]
    # Run the agent
    run_agent(llm_with_tools, messages, callbacks)
