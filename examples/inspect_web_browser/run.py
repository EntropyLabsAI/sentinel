from inspect_ai import Task, task
from inspect_ai.dataset import Sample
from inspect_ai.scorer import includes
from inspect_ai.solver import generate, use_tools
from inspect_ai.tool import web_browser
from inspect_ai import Task, eval
from pathlib import Path
from entropy_labs.api import register_project, register_samples_with_entropy_labs

@task
def browser():
    return Task(
        dataset=[
            Sample(
                input="Go to http://entropy-labs.ai/ and sign up your interest with your email address which is 'hello@test.com'."
            )
        ],
        solver=[
            use_tools(web_browser()),
            generate(),
        ],
        scorer=includes(),
        sandbox="docker",
    )



if __name__ == "__main__":
    approval_file_name = "approval.yaml"
    approval = (Path(__file__).parent / approval_file_name).as_posix()
    
    tasks = browser()
    
    # Register the project and create the run with Entropy Labs
    project_id = register_project(project_name="inspect-web-browsing", entropy_labs_backend_url="http://localhost:8080")
    tasks.dataset.samples = register_samples_with_entropy_labs(tasks, project_id, approval)
    
    eval(tasks, trace=True, model="openai/gpt-4o-mini", approval=approval)
    