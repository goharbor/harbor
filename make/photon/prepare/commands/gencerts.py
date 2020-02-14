import os
import click
import pathlib
from subprocess import check_call, PIPE, STDOUT

from utils.cert import openssl_installed
from utils.misc import get_realpath

gen_tls_script = pathlib.Path(__file__).parent.parent.joinpath('scripts/gencert.sh').absolute()

@click.command()
@click.argument('path')
def gencert(path):
    path = get_realpath(path)
    click.echo('Check openssl ...')
    if not openssl_installed():
        raise(Exception('openssl not installed'))

    click.echo("start generate internal tls certs")
    if not os.path.exists(path):
        click.echo('path {} not exist, create it...'.format(path))
        os.makedirs(path, exist_ok=True)

    shell_stat = check_call([gen_tls_script], stdout=PIPE, stderr=STDOUT, cwd=path)
    if shell_stat != 0:
        click.echo('Can not generate internal tls certs')
