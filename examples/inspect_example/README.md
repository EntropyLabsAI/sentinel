<<<<<<< HEAD:inspect_example/README.md
=======
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
   inspect eval approval.py --approval approval_1.yaml --model openai/gpt-4o-mini
   ```
>>>>>>> 2707e48 (move examples to /examples):examples/inspect_example/README.md
