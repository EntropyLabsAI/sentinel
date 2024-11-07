In this example, we show how to use custom agent using purely openai functions and tools. We wrap the tools with the @supervise() decorator.

This is example of customer support agent that can perform tasks on behalf of a customer.

Tools it has access to:
- `get_customer_info`: Get information about a customer
- `update_customer_info`: Update information about a customer
- `change_shipping_address`: Change the shipping address for a customer