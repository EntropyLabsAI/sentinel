# LangChain Examples

This directory contains three examples demonstrating how to use LangChain with supervision and mocking tools to build intelligent agents that can interact with tools, handle supervision, and reuse previous runs.

## Requirements

- **Python 3.11+**
- **Libraries:**
  - `langchain_openai`
  - `langchain_core`
  - `entropy-labs`
- **API Keys:**
  - Ensure that your OpenAI API key is set in your environment.

## Setup Instructions

1. **Install Dependencies:**

    ```bash
    pip install langchain_openai langchain_core entropy-labs
    ```

2. **Set API Keys:**

    Ensure that your OpenAI API key is set in your environment:

    ```bash
    export OPENAI_API_KEY='your-api-key-here'
    ```

---

## Examples Overview

1. [supervise_example.py](#supervise_examplepy): Demonstrates how to use supervision functions with tools to enforce constraints during tool execution.
2. [mock_functions_example.py](#mock_functionsexamplepy): Shows how to use mocking policies to simulate tool responses for testing and development.
3. [reuse_runs_example.py](#reuse_runsexamplepy): Illustrates how to reuse previous execution logs to mock tool responses in subsequent runs.

---

## supervise_example.py

### Description

In `supervise_example.py`, we define two tools: `add` and `divide`. We apply supervision functions to enforce constraints during the execution of the `divide` tool.

- **`add(a, b)`:** Adds two numbers.
- **`divide(a, b)`:** Divides two numbers but is supervised to prevent division by zero or negative numbers.

We use the `divide_supervisor` to supervise the `divide` function and set a global supervision function `human_supervisor` that can escalate if necessary.

### Key Functionalities

- **Supervision Decorators:** Applying `@supervise` to tools to enforce constraints.
- **Global Supervision Functions:** Using `set_global_supervision_functions` to apply supervision to all tools.
- **Tool Interaction:** Demonstrating how the agent interacts with tools and handles supervision.

### How to Run

Navigate to the `examples/langchain_examples` directory and run:

```bash
python supervise_example.py
```

---

## mock_functions_example.py

### Description

In `mock_functions_example.py`, we demonstrate how to use mocking policies to simulate tool responses. This is useful for testing and development without executing the actual tool logic.

- **`add(a, b)`:** Uses a predefined list of mock responses.
- **`divide(a, b)`:** Returns random mock responses.
- **`upload_api(input_data)`:** Uses an LLM to generate mock responses, supervised by `human_supervisor`.

### Key Functionalities

- **Mocking Policies:** Using `MockPolicy` to define how functions are mocked.
- **Supervision with Mocking:** Combining supervision functions with mocking.
- **Global Mock Policy:** Setting a global policy for mocking.

### How to Run

Navigate to the `examples/langchain_examples` directory and run:

```bash
python mock_functions_example.py
```

Follow the prompts in the console to interact with the agent.

---

## reuse_runs_example.py

### Description

In `reuse_runs_example.py`, we show how to reuse previous execution logs to mock tool responses in new runs. This can save time and resources by avoiding redundant computations.

- **First Run:** Executes the agent normally and logs the execution.
- **Second Run:** Reuses the logs from the first run to mock the tool responses.

### Key Functionalities

- **Execution Logging:** Using `EntropyLabsCallbackHandler` to log executions.
- **Mocking from Previous Calls:** Setting up mock policies to sample from previous runs.
- **Agent Loop Refactoring:** Encapsulating the agent loop in a reusable function.

### How to Run

Make sure you have a `.logs` directory or adjust the `log_directory` path as needed.

Navigate to the `examples/langchain_examples` directory and run:

```bash
python reuse_runs_example.py
```

---

## Conclusion

These examples demonstrate advanced usage of LangChain agents with supervision and mocking capabilities. By incorporating supervision functions and mocking policies, you can build more robust and controlled AI applications.

---

If you have any questions or need further assistance, feel free to reach out.
