# How to run this before the inspect branch is merged

Currently, the functionality is in a separate branch. Follow these steps:

1. If inspect is already installed, uninstall it:

   ```
   pip uninstall inspect_ai
   ```

2. Install the package in editable mode with development dependencies:

   ```
    git clone https://github.com/UKGovernmentBEIS/inspect_ai.git
 cd inspect_ai
 pip install -e ".[dev]"
   ```

3. Run the example, ensuring that the `--approval` flag points to a yaml that in turn points to your Approvals server

   ```
   cd inspect_example
   inspect eval run.py --approval approval.yaml --model openai/gpt-4o-mini
   ```

4. Run the agents.sh script to run multiple agents in parallel:

   ```
   ./agents.sh
   ```
This script will start a tmux session with 4 panes, each running a different agent.
You can change the number of agents to run by changing the `PANE_COUNT` variable, and you can change the model by changing the `MODEL` variable. 
