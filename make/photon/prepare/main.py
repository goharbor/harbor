from commands.prepare import prepare
from commands.gencerts import gencert
from commands.migrate import migrate
from commands.tasks.check import check as tasks_check
import click

@click.group()
def cli():
    pass

# Create tasks group
@click.group()
def tasks():
    """Task validation and quality checking commands."""
    pass

# Add check command to tasks group
tasks.add_command(tasks_check, name="check")

# Add existing commands
cli.add_command(prepare)
cli.add_command(gencert)
cli.add_command(migrate)

# Add tasks group to main CLI
cli.add_command(tasks)

if __name__ == '__main__':
    cli()
