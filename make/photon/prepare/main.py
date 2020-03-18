from commands.prepare import prepare
from commands.gencerts import gencert
import click

@click.group()
def cli():
    pass

cli.add_command(prepare)
cli.add_command(gencert)

if __name__ == '__main__':
    cli()
