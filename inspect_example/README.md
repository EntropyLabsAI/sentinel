# How to run this before the inspect branch is merged

Currently, the functionality is in a separate branch. Follow these steps:

1. If inspect is already installed, uninstall it:

   ```
   pip install inspect_ai
   ```

2. Install the package in editable mode with development dependencies:

   ```
   pip install -e ".[dev]"
   ```

3. Run the example:

   ```
   cd inspect_example
   inspect eval approval.py --approval approval.yaml --model openai/gpt-4o
   ```
