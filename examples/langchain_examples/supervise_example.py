from entropy_labs.supervision.config import set_global_supervision_functions
from entropy_labs.supervision.supervisors import Supervisor, llm_supervisor
from entropy_labs.supervision.langchain.supervisors import human_supervisor
from entropy_labs.supervision.decorators import supervise
from entropy_labs.supervision.config import (
    set_global_supervision_functions,
    SupervisionDecision,
    SupervisionDecisionType,
    SupervisionContext,
)
from entropy_labs.supervision.langchain.logging import EntropyLabsCallbackHandler
from langchain_openai import ChatOpenAI
from langchain.tools import tool
from langchain_core.messages import HumanMessage
from typing import Callable

BACKEND_API_ENDPOINT = "http://localhost:8080"


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


@tool
@supervise(supervision_functions=[llm_supervisor(instructions="If the numbers are not positive, reject them.")])
def add(a: float, b: float) -> float:
    """Add two numbers."""
    return a + b

# Set multiply to be supervised by multiply_supervisor
@tool
@supervise(supervision_functions=[divide_supervisor(min_b=0)])
def divide(a: float, b: float) -> float:
    """Divide two numbers."""
    print(f"Dividing {a} and {b}")
    return a / b



if __name__ == "__main__":
    # Initialize tools
    tools = [add, divide]
    
    # Create the callback handler with single_log_file=True
    callbacks = [EntropyLabsCallbackHandler(
        tools=tools,
        single_log_file=True
    )]
    
    
    # You can also set global supervision functions
    set_global_supervision_functions([human_supervisor(backend_api_endpoint=BACKEND_API_ENDPOINT)])
    # for these tools:
    # add: human_supervisor will be used
    # divide: if divide_supervisor rejects it will be escalated to human_supervisor
    
    # Unchanged LangChain simple agent logic
    llm = ChatOpenAI(model="gpt-4o", temperature=0)
    
    llm_with_tools = llm.bind_tools(tools)
    messages = [HumanMessage(content="Divide 12 and -2, then add 5 to the result.")]
    
    while True:
        ai_msg = llm_with_tools.invoke(messages, config={"callbacks": callbacks})
        messages.append(ai_msg)
        
        if ai_msg.tool_calls:
            for tool_call in ai_msg.tool_calls:
                selected_tool = {"add": add, "divide": divide}[tool_call["name"].lower()]
                tool_msg = selected_tool.invoke(tool_call, messages=messages, config={"callbacks": callbacks})
                messages.append(tool_msg)
        else:
            print(ai_msg.content)
            user_input = input("Ask another question or type 'exit' to quit.")
            if user_input == "exit":
                break
        