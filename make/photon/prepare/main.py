from commands.prepare import prepare
from commands.gencerts import gencert
from commands.migrate import migrate
import click

@click.group()
def cli():
    pass

cli.add_command(prepare)
cli.add_command(gencert)
cli.add_command(migrate)

if __name__ == '__main__':
    cli()
