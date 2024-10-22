from el.supervision.config import set_global_mock_policy
from el.supervision.langchain.supervisors import human_supervisor
from el.supervision.decorators import supervise
from el.mocking.policies import MockPolicy
from langchain_openai import ChatOpenAI
from langchain.tools import tool
from langchain_core.messages import HumanMessage
from pydantic import BaseModel
from typing import List

@tool
@supervise(
    mock_policy=MockPolicy.SAMPLE_LIST,
    mock_responses=[10, 20, 30]
)
def add(a: float, b: float) -> float:
    """Add two numbers."""
    return a + b

@tool
@supervise(
    mock_policy=MockPolicy.SAMPLE_RANDOM
)
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
    mock_policy=MockPolicy.SAMPLE_LLM,
    supervision_functions=[human_supervisor()]
)
def upload_api(input_data: str) -> UploadResponse:
    """Upload the input data to the API and receive a response.
    
    Args:
        input_data (str): The data to upload.
    
    Returns:
        UploadResponse: The response from the API.
    """
    # Example API response
    # Actual API response would go here
    return UploadResponse(status="success", data=[{"id": 1, "name": "example"}], message="Upload successful")
    

if __name__ == "__main__":
    # Initialize tools and logging
    tools = [add, divide, upload_api]
    
    # Set global mock policy (can be overridden by individual tools)
    set_global_mock_policy(MockPolicy.NO_MOCK, override_local_policy=False)
    
    # Initialize the LLM
    llm = ChatOpenAI(model="gpt-4", temperature=0)
    
    # Bind tools to the LLM
    llm_with_tools = llm.bind_tools(tools)
    messages = [HumanMessage(content="Divide 12 by 3, then add 5 to the result. Finally call upload_api with the result.")]
    
    while True:
        ai_msg = llm_with_tools.invoke(messages)
        messages.append(ai_msg)
        
        if ai_msg.tool_calls:
            for tool_call in ai_msg.tool_calls:
                selected_tool = {
                    "add": add,
                    "divide": divide,
                    "upload_api": upload_api
                }.get(tool_call['name'].lower())
                tool_msg = selected_tool.invoke(tool_call)
                messages.append(tool_msg)
        else:
            print(ai_msg.content)
            user_input = input("Ask another question or type 'exit' to quit: ")
            if user_input.lower() == "exit":
                break
            messages.append(HumanMessage(content=user_input))
