import os

from g import base_dir, templates_dir
from .jinja import render_jinja


# render docker-compose
VERSION_TAG = 'dev'
REGISTRY_VERSION = 'v2.7.1'
NOTARY_VERSION = 'v0.6.1-v1.7.1'
CLAIR_VERSION = 'v2.0.7-dev'
CHARTMUSEUM_VERSION = 'v0.7.1-dev'
CLAIR_DB_VERSION = VERSION_TAG
MIGRATOR_VERSION = VERSION_TAG
REDIS_VERSION = VERSION_TAG
NGINX_VERSION = VERSION_TAG
# version of chartmuseum

docker_compose_template_path = os.path.join(templates_dir, 'docker_compose', 'docker-compose.yml.jinja')
docker_compose_yml_path = os.path.join(base_dir, 'docker-compose.yml')

def check_configs(configs):
    pass

def prepare_docker_compose(configs, with_clair, with_notary, with_chartmuseum):
    check_configs(configs)

    rendering_variables = {
        'version': VERSION_TAG,
        'reg_version': "{}-{}".format(REGISTRY_VERSION, VERSION_TAG),
        'redis_version': REDIS_VERSION,
        'notary_version': NOTARY_VERSION,
        'clair_version': CLAIR_VERSION,
        'chartmuseum_version': CHARTMUSEUM_VERSION,
        'data_volume': configs['data_volume'],
        'log_location': configs['log_location'],
        'cert_key_path': configs['cert_key_path'],
        'cert_path': configs['cert_path'],
        'with_notary': with_notary,
        'with_clair': with_clair,
        'with_chartmuseum': with_chartmuseum
    }
    rendering_variables['secretkey_path'] = configs['secretkey_path']

    render_jinja(docker_compose_template_path, docker_compose_yml_path, **rendering_variables)