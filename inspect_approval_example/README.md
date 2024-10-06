# How to run this before the inspect branch is merged

Currently, the functionality is in a separate branch. Follow these steps:

1. If inspect is already installed, uninstall it:

   ```
   pip uninstall inspect_ai
   ```

2. Clone the feature/approvals branch:

   ```
   cd inspect_approval_example
   git clone -b feature/approvals https://github.com/UKGovernmentBEIS/inspect_ai.git inspect_ai_feature_approvals
   ```

3. Change to the cloned directory:

   ```
   cd inspect_ai_feature_approvals
   ```

4. Install the package in editable mode with development dependencies:

   ```
   pip install -e ".[dev]"
   ```

5. Run the example:

   ```
   cd inspect_approval_example
   inspect eval approval.py --approval approval.yaml --trace --model openai/gpt-4o
   ```
