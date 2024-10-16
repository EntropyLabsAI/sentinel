## Running Inspect AI Example

1. Make sure Inspect AI and Entropy Labs are installed in your python environment:

   ```bash
   pip install inspect-ai entropy_labs --upgrade
   ```

2. Change to the example directory:

   ```bash
   cd examples/inspect_example
   ```

3. Run single agent example when using human approval for all tools:

   ```bash
   inspect eval run.py --approval approval_human.yaml --model openai/gpt-4o --trace
   ```
   You can change the approval configuration by changing the `approval_human.yaml` file. There are also allowlist examples in the `approval_escalation.yaml` file. You can run this by changing the `--approval` argument:

   ```bash
   inspect eval run.py --approval approval_escalation.yaml --model openai/gpt-4o --trace
   ```

4. Or run the agents.sh helper script to run multiple agents in parallel inside a tmux window (requires [tmux](https://github.com/tmux/tmux/wiki)):

   ```bash
   ./agents.sh
   ```
      This script will start a tmux session with 4 panes, each running a different agent.
   You can change the number of agents to run by changing the `PANE_COUNT` variable, and you can change the model by changing the `MODEL` variable. 



## Approval configuation
Examples of approval configuration are in `approval_human.yaml`, `approval_llm.yaml` and `approval_escalation.yaml` files. We follow the Inspect AI format for the approval configuration. More information here: [https://inspect.ai-safety-institute.org.uk/approval.html](https://inspect.ai-safety-institute.org.uk/approval.html). Our approvers are in the `el.approvers` module. You can install the `entropy-labs` package to get these approvers:

```bash
pip install entropy-labs
```
To use the approvers with inspect you need to make sure to import the approvers from the `el` package:

```python
from el.approvers import _registry
```
Then you can reference the approvers in the approval configuration like `el/human_approver`. You can find more information about the `entropy-labs` package at [https://pypi.org/project/entropy-labs/](https://pypi.org/project/entropy-labs/).


In the `approval_escalation.yaml` example, bash approval is done by checking if the command is in the allowed list. Python approval is done by checking if the modules and functions used are in the allowed list. If bash or python approval is not passed, the human approval is triggered. You can remove bash or python approval if you want escalate directly to human approval. Make sure the human api `approval_api_endpoint` is set to the local approval api endpoint.

## Approval Configuration Details

The `approval_escalation.yaml` file contains the following configuration:

### Python Allowlist
- Allowed modules: requests, json, csv, datetime, re, math, random, time
- Allowed functions: print, len, range, str, int, float, list, dict, set, tuple, sum, max, min
- Disallowed builtins: eval, exec, compile, __import__, open, input
- Sensitive modules: os, sys, subprocess, socket
- System state modification: Not allowed

### Bash Allowlist
- Allowed commands: ls, cd, pwd, echo, cat, grep, mkdir, cp, wget, curl, pip
- Sudo: Not allowed
- Command-specific rules for pip: install, list, show

### Human API
- Applies to all tools
- Approval API endpoint: http://localhost:8080
- Agent ID: sample_3
- Number of approvals required: 5

This configuration ensures that only approved Python modules and functions, as well as specific bash commands, can be used without triggering human approval. Any attempts to use disallowed or sensitive operations will require human intervention through the specified API endpoint.







