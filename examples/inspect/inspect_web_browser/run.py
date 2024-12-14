from inspect_ai import Task, task
from inspect_ai.dataset import Sample
from inspect_ai.solver import generate, use_tools
from inspect_ai.tool import web_browser
from inspect_ai import Task, eval
from pathlib import Path
from entropy_labs.api import entropy_labs_web_ui_scorer, register_inspect_samples_with_entropy_labs_solver

@task
def browser():
    return Task(
        dataset=[
            Sample(
                input="Go to http://entropy-labs.ai/ and finish and sign up your interest with your email address which is 'hello@test.com'."
            )
        ],
        solver=[
            register_inspect_samples_with_entropy_labs_solver(project_name="inspect-web-browsing", entropy_labs_backend_url="http://localhost:8080"),
            use_tools(web_browser()),
            generate(),
        ],
        scorer=entropy_labs_web_ui_scorer(),
        sandbox="docker",
    )



if __name__ == "__main__":
    approval_file_name = "approval.yaml"
    approval = (Path(__file__).parent / approval_file_name).as_posix()
    
    eval(browser(), trace=True, model="openai/gpt-4o-mini", approval=approval)