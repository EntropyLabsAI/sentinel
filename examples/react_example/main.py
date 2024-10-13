# react_ai_agent.py

from openai import OpenAI
import json

client = OpenAI()

tools = [
    {
        "type": "function",
        "function": {
            "name": "get_weather",
            "strict": True,
            "parameters": {
                "type": "object",
                "properties": {
                    "location": {"type": "string"},
                    "unit": {"type": "string", "enum": ["c", "f"]},
                },
                "required": ["location", "unit"],
                "additionalProperties": False,
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "get_stock_price",
            "strict": True,
            "parameters": {
                "type": "object",
                "properties": {
                    "symbol": {"type": "string"},
                },
                "required": ["symbol"],
                "additionalProperties": False,
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "execute_bash",
            "strict": True,
            "parameters": {
                "type": "object",
                "properties": {"command": {"type": "string"}},
                "required": ["command"],
                "additionalProperties": False,
            },
        }
    }
]

def execute_bash(command):
    print("🔍 Executing bash command: ", command)

    # Hit localhost:8080/api/review with the command
    response = requests.post("http://localhost:8080/api/review", json={"command": command})

    return f"The command '{command}' returned with output: 'output'"

def get_weather(location, unit):
    print("🔍 Fetching weather for ", location, " in ", unit)
    # Implement your weather fetching logic here
    # For simplicity, we'll return a dummy response
    return f"The weather in {location} is 20 degrees {unit.upper()}."

def get_stock_price(symbol):
    # Implement your stock price fetching logic here
    # For simplicity, we'll return a dummy response
    return f"The current price of {symbol} is $100."

# Mapping function names to actual implementations
function_map = {
    "get_weather": get_weather,
    "get_stock_price": get_stock_price,
    "execute_bash": execute_bash,
}

def main_loop():
    messages = []
    print("Welcome to a simple agent! Type 'exit' to quit.")
    
    while True:
        user_input = input("You: ")
        if user_input.lower() in ['exit', 'quit']:
            print("Goodbye!")
            break
        
        messages.append({"role": "user", "content": user_input})
        
        completion = client.chat.completions.create(
            model="gpt-4o",
            messages=messages,
            tools=tools,
            tool_choice="required",
            stream=False,
        )
        
        response_message = completion.choices[0].message
        print("💬 AI response: ", response_message)

        if response_message.tool_calls:
            for tool_call in response_message.tool_calls:
                function_name = tool_call.function.name
                arguments = json.loads(tool_call.function.arguments)
                print("🔍 Calling tool: ", function_name, " with arguments: ", arguments)
                function_map[function_name](**arguments)
        else:
            print(f"AI: {response_message.content}")

if __name__ == "__main__":
    main_loop()
