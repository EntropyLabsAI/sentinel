

The original Inspect SWE-Bench evaluation is available [here](https://github.com/UKGovernmentBEIS/inspect_evals/tree/main/src/inspect_evals/gdm_capabilities/gdm_in_house_ctf).

To run the evaluation with approval, use the following command:

```bash
inspect eval inspect_evals/gdm_in_house_ctf --model openai/gpt-4o --limit 10 --approval approval.yaml
```

Edit the `approval.yaml` file to change the approvers.
