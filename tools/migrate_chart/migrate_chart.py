#!/usr/local/bin/python3

import subprocess
import signal
import sys
from pathlib import Path

import click
import requests

MIGRATE_CHART_SCRIPT = '/migrate_chart.sh'
HELM_CMD = '/linux-amd64/helm'
CA_UPDATE_CMD = 'update-ca-certificates'
CHART_URL_PATTERN = "https://{host}/api/v2.0/projects/{project}/repositories/{name}/artifacts/{version}"
CHART_SOURCE_DIR = Path('/chart_storage')

errs = []

def print_exist_errs():
    if errs:
        click.echo("Following errors exist", err=True)
        for e in errs:
            click.echo(e, err=True)

def graceful_exit(signum, frame):
    print_exist_errs()
    sys.exit()

signal.signal(signal.SIGINT, graceful_exit)
signal.signal(signal.SIGTERM, graceful_exit)

class ChartV2:

    def __init__(self, filepath:Path):
        self.filepath = filepath
        self.project = self.filepath.parts[-2]
        parts = self.filepath.stem.split('-')
        flag = False
        for i in range(len(parts)-1, -1, -1):
            if parts[i][0].isnumeric():
                self.name, self.version = '-'.join(parts[:i]), '-'.join(parts[i:])
                flag = True
                break
        if not flag:
            raise Exception('chart name: {} is illegal'.format('-'.join(parts)))

    def __check_exist(self, hostname, username, password):
        return requests.get(CHART_URL_PATTERN.format(
                host=hostname,
                project=self.project,
                name=self.name,
                version=self.version),
                auth=requests.auth.HTTPBasicAuth(username, password))

    def migrate(self, hostname, username, password):
        res = self.__check_exist(hostname, username, password)
        if res.status_code == 200:
            raise Exception("Artifact already exist in harbor")
        if res.status_code == 401:
            raise Exception(res.reason)

        oci_ref = "{host}/{project}/{name}:{version}".format(
            host=hostname,
            project=self.project,
            name=self.name,
            version=self.version)

        return subprocess.run([MIGRATE_CHART_SCRIPT, HELM_CMD, self.filepath, oci_ref],
        text=True, stdout=subprocess.DEVNULL, stderr=subprocess.PIPE)


@click.command()
@click.option('--hostname', default='127.0.0.1', help='the password to login harbor')
@click.option('--username', default='admin', help='The username to login harbor')
@click.option('--password', default='Harbor12345', help='the password to login harbor')
def migrate(hostname, username, password):
    """
    Migrate chart v2 to harbor oci registry
    """
    if username != 'admin':
        raise Exception('This operation only allowed for admin')
    subprocess.run([CA_UPDATE_CMD])
    subprocess.run([HELM_CMD, 'registry', 'login', hostname, '--username', username, '--password', password])
    charts = [ChartV2(c) for p in CHART_SOURCE_DIR.iterdir() if p.is_dir() for c in p.iterdir() if c.is_file() and c.name != "index-cache.yaml"]
    with click.progressbar(charts, label="Migrating chart ...", length=len(charts),
    item_show_func=lambda x: "{}/{}:{} total errors: {}".format(x.project, x.name, x.version, len(errs)) if x else '') as bar:
        for chart in bar:
            try:
                result = chart.migrate(hostname, username, password)
                if result.stderr:
                    errs.append("chart: {name}:{version} in {project} has err: {err}".format(
                        name=chart.name,
                        version=chart.version,
                        project=chart.project,
                        err=result.stderr
                    ))
            except Exception as e:
                errs.append("chart: {name}:{version} in {project} has err: {err}".format(
                    name=chart.name,
                    version=chart.version,
                    project=chart.project,
                    err=e))
    click.echo("Migration is Done.")
    print_exist_errs()

if __name__ == '__main__':
    migrate()
