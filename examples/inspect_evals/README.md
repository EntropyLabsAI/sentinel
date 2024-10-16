This folder contains examples of how to use the `inspect eval` with approvals with evaluations and benchmarks examples already implemented in the Inspect repository. The list of the examples is here: [https://inspect.ai-safety-institute.org.uk/examples/](https://inspect.ai-safety-institute.org.uk/examples/)

To run an evaluation with approval, use the following command:

```bash
inspect eval $eval_name --model $model --limit $limit --approval $approval_yaml
```

Replace `$eval_name` with the name of the evaluation you want to run, `$model` with the model you want to use, `$limit` with the number of samples you want to run, and `$approval_yaml` with the path to the approval yaml file.

For example, to run the SWE-Bench evaluation with the GPT-4o model, use the following command:

```bash
inspect eval inspect_evals/swe_bench --model openai/gpt-4o --limit 10 --approval approval.yaml
```

We have examples of approval yaml files for some of the evaluations. But you can run any evaluation already implemented in Inspect: [https://github.com/UKGovernmentBEIS/inspect_evals/tree/main/src/inspect_evals](https://github.com/UKGovernmentBEIS/inspect_evals/tree/main/src/inspect_evals).
