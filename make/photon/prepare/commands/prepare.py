# pylint: disable=no-value-for-parameter

import sys
import logging

import click

from utils.misc import delfile
from utils.configs import validate, parse_yaml_config
from utils.cert import prepare_registry_ca, SSL_CERT_KEY_PATH, SSL_CERT_PATH, get_secret_key, prepare_trust_ca
from utils.db import prepare_db
from utils.jobservice import prepare_job_service
from utils.registry import prepare_registry
from utils.registry_ctl import prepare_registry_ctl
from utils.core import prepare_core
from utils.log import prepare_log_configs
from utils.docker_compose import prepare_docker_compose
from utils.nginx import prepare_nginx, nginx_confd_dir
from utils.redis import prepare_redis
from utils.internal_tls import prepare_tls
from utils.trivy_adapter import prepare_trivy_adapter
from utils.portal import prepare_portal
from utils.exporter import prepare_exporter
from g import (config_dir, input_config_path, private_key_pem_path, root_crt_path, secret_key_dir,
old_private_key_pem_path, old_crt_path)

@click.command()
@click.option('--conf', default=input_config_path, help="the path of Harbor configuration file")
@click.option('--with-trivy', is_flag=True, help="the Harbor instance is to be deployed with Trivy")
@click.option('--with-podman', is_flag=True, help="the Harbor instance will be deployed using Podman (avoid Docker logging options)")
def prepare(conf, with_trivy, with_podman):

    delfile(config_dir)
    config_dict = parse_yaml_config(conf, with_trivy=with_trivy)
    try:
        validate(config_dict)
    except Exception as e:
        click.echo('Error happened in config validation...')
        logging.error(e)
        sys.exit(-1)

    prepare_portal(config_dict)
    prepare_log_configs(config_dict)
    prepare_nginx(config_dict)
    prepare_core(config_dict, with_trivy=with_trivy)
    prepare_registry(config_dict)
    prepare_registry_ctl(config_dict)
    prepare_db(config_dict)
    prepare_job_service(config_dict)
    prepare_redis(config_dict)
    prepare_tls(config_dict)
    prepare_trust_ca(config_dict)

    get_secret_key(secret_key_dir)

    #  If Customized cert enabled
    prepare_registry_ca(
        private_key_pem_path=private_key_pem_path,
        root_crt_path=root_crt_path,
        old_private_key_pem_path=old_private_key_pem_path,
        old_crt_path=old_crt_path)

    if config_dict['metric'].enabled:
        prepare_exporter(config_dict)

    if with_trivy:
        prepare_trivy_adapter(config_dict)

    prepare_docker_compose(config_dict, with_trivy, with_podman)
