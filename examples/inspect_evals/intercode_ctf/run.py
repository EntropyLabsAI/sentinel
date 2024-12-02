from pathlib import Path
from asteroid_sdk.api import register_project, register_samples_with_asteroid    
from inspect_ai import eval
from inspect_evals.gdm_capabilities.intercode_ctf.task import gdm_intercode_ctf

if __name__ == "__main__":
    approval_file_name = "approval.yaml"
    approval = (Path(__file__).parent / approval_file_name).as_posix()
    
    # Register the project with Entropy Labs
    project_id = register_project(
        project_name="intercode-ctf", entropy_labs_backend_url="http://localhost:8080"
    )
    
    # Register samples and create runs
    tasks = gdm_intercode_ctf()
    tasks.dataset.samples = register_samples_with_asteroid(tasks, project_id, approval)
    
    eval(
        tasks,
        approval=approval,
        trace=True,
        model="openai/gpt-4o-mini",
    )
