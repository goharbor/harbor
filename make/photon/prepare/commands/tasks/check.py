import click
import json
from pathlib import Path
from rich.console import Console
from rich.table import Table

from .quality_checker.quality_checker import QualityChecker

console = Console()


@click.command()
@click.argument("task_path", type=click.Path(exists=True, path_type=Path))
@click.option(
    "-u",
    "--unit-test-path",
    default="tests/test_outputs.py",
    help="Relative path to the unit test file.",
)
@click.option(
    "-d",
    "--dockerfile-path",
    default="Dockerfile",
    help="Relative path to the Dockerfile.",
)
@click.option(
    "-m",
    "--model",
    default="anthropic/claude-3-5-sonnet-20241022",
    help="Name of the model to use for quality checking.",
)
@click.option(
    "-o",
    "--output",
    type=click.Path(path_type=Path),
    help="Path to write JSON quality check results.",
)
@click.option(
    "--task-name",
    help="Task name (used when task_path is a tasks directory).",
)
@click.option(
    "--dataset",
    help="Dataset name (used with task-name).",
)
def check(task_path, unit_test_path, dockerfile_path, model, output, task_name, dataset):
    """Check the quality of a task using an LLM.

    The task_path should point to a directory containing:
    - task.yaml: Task configuration
    - solution.sh or solution.yaml: Solution script
    - Dockerfile: Docker configuration
    - tests/test_outputs.py: Unit tests

    Examples:
        harbor tasks check /path/to/task
        harbor tasks check tasks/ --task-name my_task --dataset my_dataset
    """
    # Handle different path formats
    if task_name:
        # If task_name is provided, construct the path
        if dataset:
            task_dir = Path(task_path) / dataset / task_name
        else:
            task_dir = Path(task_path) / task_name
    else:
        # Use the provided path directly
        task_dir = Path(task_path)

    if not task_dir.exists():
        click.echo(f"Error: Task directory '{task_dir}' not found.", err=True)
        raise SystemExit(1)

    # Create the quality checker
    try:
        checker = QualityChecker(
            task_dir,
            model,
            unit_test_relative_path=unit_test_path,
            dockerfile_relative_path=dockerfile_path,
        )
    except FileNotFoundError as e:
        click.echo(f"Error: {e}", err=True)
        raise SystemExit(1)

    # Run the quality check
    click.echo(f"Running quality check on task: {task_dir}")
    click.echo(f"Using model: {model}")

    try:
        result = checker.check()
    except Exception as e:
        click.echo(f"Error during quality check: {e}", err=True)
        raise SystemExit(1)

    # Save to file if requested
    if output:
        output.write_text(result.model_dump_json(indent=4))
        click.echo(f"Results saved to: {output}")

    # Display results in a table
    table = Table(title="Quality Check Results", show_lines=True)
    table.add_column("Check", style="cyan")
    table.add_column("Outcome", style="bold")
    table.add_column("Explanation")

    for field_name, field_value in result.model_dump().items():
        outcome = field_value["outcome"]
        outcome_style = {
            "pass": "green",
            "fail": "red",
            "not_applicable": "yellow",
        }.get(outcome.lower(), "white")

        table.add_row(
            field_name.replace("_", " ").title(),
            f"[{outcome_style}]{outcome}[/{outcome_style}]",
            field_value["explanation"],
        )

    console.print(table)

    # Check if there are any failures
    failures = [
        field_name
        for field_name, field_value in result.model_dump().items()
        if field_value["outcome"] == "fail"
    ]

    if failures:
        click.echo(f"\n⚠️  Found {len(failures)} quality check failure(s).", err=True)
        raise SystemExit(1)
    else:
        click.echo("\n✅ All quality checks passed!")