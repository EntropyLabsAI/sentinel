from inspect_ai import Task, task
from inspect_ai.dataset import Sample
from inspect_ai.solver import generate, use_tools
from inspect_ai.tool import web_browser
from inspect_ai import Task, eval
from pathlib import Path
from entropy_labs.api import register_project, register_samples_with_entropy_labs, update_run_status_by_sample_id, get_sample_result
from inspect_ai.solver import TaskState
from inspect_ai.scorer import (
    Score,
    Scorer,
    Target,
    accuracy,
    scorer,
    stderr,
)
from typing import Optional



@scorer(metrics=[accuracy(), stderr()])
def entropy_labs_web_ui_scorer(timeout: Optional[int] = 300, wait_for_result: bool = True) -> Scorer:
    async def score(state: TaskState, target: Target) -> Score:
        
        # Update the run status to completed
        update_run_status_by_sample_id(str((state.sample_id)), status="completed")
        
        # Get the result of the run - It's necessary to submit the result in the Entropy Labs UI
        if wait_for_result:
            result = get_sample_result(str(state.sample_id))
        else:
            result = ''
        
        return Score(
            value=1 if result == 'passed' else 0,
            answer="",
            explanation="",
        )

    return score

@task
def browser():
    return Task(
        dataset=[
            Sample(
                input="Go to http://entropy-labs.ai/ and finish and sign up your interest with your email address which is 'hello@test.com'."
            )
        ],
        solver=[
            use_tools(web_browser()),
            generate(),
        ],
        scorer=entropy_labs_web_ui_scorer(),
        sandbox="docker",
    )



if __name__ == "__main__":
    approval_file_name = "approval.yaml"
    approval = (Path(__file__).parent / approval_file_name).as_posix()
    
    tasks = browser()
    
    # Register the project and create the run with Entropy Labs
    project_id = register_project(project_name="inspect-web-browsing", entropy_labs_backend_url="http://localhost:8099")
    tasks.dataset.samples = register_samples_with_entropy_labs(tasks, project_id, approval)
    
    eval(tasks, trace=True, model="openai/gpt-4o-mini", approval=approval)
    
