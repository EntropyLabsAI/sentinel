## Running Inspect AI Example

1. Make sure Inspect AI is installed in your python environment:

   ```bash
   pip install inspect-ai
   ```

2. Change to the example directory:

   ```
   git clone https://github.com/UKGovernmentBEIS/inspect_ai.git
   cd inspect_ai
   pip install -e ".[dev]"
   ```

3. Run single agent example:

   ```bash
   inspect eval run.py --approval approval.yaml --model openai/gpt-4o --trace
   ```

4. Or run the agents.sh helper script to run multiple agents in parallel inside a tmux window (requires [tmux](https://github.com/tmux/tmux/wiki)):

   ```bash
   ./agents.sh
   ```
   This script will start a tmux session with 4 panes, each running a different agent.
   You can change the number of agents to run by changing the `PANE_COUNT` variable, and you can change the model by changing the `MODEL` variable. 
