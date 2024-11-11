In this example, we show how to use custom agent using purely openai functions and tools. We wrap the tools with the @supervise() decorator.

This is example of customer support agent that can perform tasks on behalf of a customer.

Tools it has access to:
- respond_to_customer: Respond to a customer
- authenticate_customer: Authenticate a customer

pip install duckduckgo-search