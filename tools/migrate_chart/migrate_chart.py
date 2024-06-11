#!/usr/local/bin/python3

import subprocess
import signal
import sys
import os
from pathlib import Path
import tarfile
import yaml

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
        click.echo("There are {} errors exist".format(len(errs)), err=True)
        for e in errs:
            click.echo(e, err=True)
            # Write the error to file
            with open("/chart_storage/migration_errors.txt", "a") as f:
                f.write(e + "\n")

def graceful_exit(signum, frame):
    print_exist_errs()
    sys.exit()

signal.signal(signal.SIGINT, graceful_exit)
signal.signal(signal.SIGTERM, graceful_exit)

def find_chart_yaml(tar, path=''):
    # Iterate through the members of the tarfile
    for member in tar.getmembers():
        # If the member is a directory, recursively search within it
        if member.isdir():
            find_chart_yaml(tar, os.path.join(path, member.name))
        # If the member is a file and its name is 'chart.yaml', return its path
        if "Chart.yaml" in member.name:
            return os.path.join(path, member.name)

def read_chart_version(chart_tgz_path):
    # Open the chart tgz file
    with tarfile.open(chart_tgz_path, 'r:gz') as tar:
        # Find the path to chart.yaml within the tarball
        chart_yaml_path = find_chart_yaml(tar)
        if chart_yaml_path:
            # Extract the chart.yaml file
            chart_yaml_file = tar.extractfile(chart_yaml_path)
            if chart_yaml_file is not None:
                # Load the YAML content from chart.yaml
                chart_data = yaml.safe_load(chart_yaml_file)
                # Read the version from chart.yaml
                version = chart_data.get('version')
                name = chart_data.get('name')
                return name, version
            else:
                raise Exception("Failed to read chart.yaml from the chart tgz file. filename {}".format(chart_tgz_path))
        else:
            raise Exception("chart.yaml not found in the chart tgz file. filename {}".format(chart_tgz_path))

class ChartV2:

    def __init__(self, filepath:Path):
        self.filepath = filepath
        self.project = self.filepath.parts[-2]
        self.name = ""
        self.version = ""
        try:
            self.name, self.version = read_chart_version(filepath)
            if self.name == "" or self.version == "" or self.name is None or self.version is None :
                raise Exception('chart name: {} is illegal'.format('-'.join(parts)))
        except Exception as e:
            click.echo("Skipped chart: {} due to illegal chart name. Error: {}".format(filepath, e), err=True)
        return

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

        oci_ref = "oci://{host}/{project}".format(
            host=hostname,
            project=self.project)

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
    charts = [ChartV2(c) for p in CHART_SOURCE_DIR.iterdir() if p.is_dir() for c in p.iterdir() if c.is_file() and c.name.endswith(".tgz")]
    with click.progressbar(charts, label="Migrating chart ...", length=len(charts),
    item_show_func=lambda x: "{}/{}:{} total errors: {}".format(x.project, x.name, x.version, len(errs)) if x else '') as bar:
        for chart in bar:
            try:
                if chart.name == "" or chart.version == "" :
                    print("skip the chart {} has no name or version info".format(chart.filepath))
                    continue
                result = chart.migrate(hostname, username, password)
                if result.stderr:
                    errs.append("chart: {name}:{version} in {project} has err: {err}".format(
                        name=chart.name,
                        version=chart.version,
                        project=chart.project,
                        err=result.stderr
                    ))
            except Exception as e:
                errs.append("chart: {name}:{version} in {project}, path {path} has err: {err}".format(
                    name=chart.name,
                    version=chart.version,
                    project=chart.project,
                    path = chart.filepath,
                    err=e))
    click.echo("Migration is Done.")
    print_exist_errs()

if __name__ == '__main__':
    migrate()
