## Running Inspect AI Example

1. Make sure Inspect AI is installed in your python environment:

   ```bash
   pip install git+https://github.com/UKGovernmentBEIS/inspect_ai
   ```

2. Change to the example directory:

   ```bash
   cd examples/inspect_example
   ```

3. Run single agent example:

   ```bash
   inspect eval run.py --approval approval.yaml --model openai/gpt-4o --trace
   ```

This will start a single agent and run it against the approval server.

4. Or run multiple agents in parallel:

   ```bash
   ./agents.sh
   ```
This script will start a tmux session with 4 panes, each running a different agent.
You can change the number of agents to run by changing the `PANE_COUNT` variable, and you can change the model by changing the `MODEL` variable. 
