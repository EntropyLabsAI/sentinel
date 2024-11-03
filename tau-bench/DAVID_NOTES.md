python run.py --agent-strategy tool-calling --env retail --model gpt-4o --model-provider openai --user-model gpt-4o --user-model-provider openai --user-strategy llm --max-concurrency 10 --task-ids 2


conda activate tau-bench
cd /Users/david/PycharmProjects/agent-control-plane/entropy_labs
pip install -e ".[dev]" 


# RESULTS

## Retail

No supervision:
Model: gpt-4o-mini
🏆 Average reward: 0.4260869565217391
📈 Pass^k
  k=1: 0.4260869565217391
📄 Results saved to results/tool-calling-gpt-4o-mini-0.0_range_0--1_user-gpt-4o-mini-llm_1102182420.json


Model: gpt-4o
🏆 Average reward: 0.6260869565217392
📈 Pass^k
  k=1: 0.6260869565217392
📄 Results saved to results/tool-calling-gpt-4o-0.0_range_0--1_user-gpt-4o-llm_1102193240.json


Superivision

Model: gpt-4o
🏆 Average reward: 0.4956521739130435
📈 Pass^k
  k=1: 0.4956521739130435
📄 Results saved to results/tool-calling-gpt-4o-0.0_range_0--1_user-gpt-4o-llm_1102235139.json

Example failed tasks: task_id=112, 113


## Airline