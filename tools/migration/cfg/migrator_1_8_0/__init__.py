from __future__ import print_function
import utils
import os
from jinja2 import Environment, FileSystemLoader
acceptable_versions = ['1.7.0']
keys = [
    'hostname',
    'ui_url_protocol',
    'ssl_cert',
    'ssl_cert_key',
    'admiral_url',
    'log_rotate_count',
    'log_rotate_size',
    'http_proxy',
    'https_proxy',
    'no_proxy',
    'db_host',
    'db_password',
    'db_port',
    'db_user',
    'redis_host',
    'redis_port',
    'redis_password',
    'redis_db_index',
    'clair_updaters_interval',
    'max_job_workers',
    'registry_storage_provider_name',
    'registry_storage_provider_config',
    'registry_custom_ca_bundle'
    ]

def migrate(input_cfg, output_cfg):
    d = utils.read_conf(input_cfg)
    val = {}
    for k in keys:
        val[k] = d.get(k,'')
    if val['db_host'] == 'postgresql' and val['db_port'] == '5432':
        val['external_db'] = False
    else:
        val['external_db'] = True
    # If using default filesystem, didn't need registry_storage_provider_config config
    if val['registry_storage_provider_name'] == 'filesystem' and not val.get('registry_storage_provider_config'):
        val['storage_provider_info'] = ''
    else:
        val['storage_provider_info'] = utils.get_storage_provider_info(
            val['registry_storage_provider_name'],
            val['registry_storage_provider_config']
            )
    if val['redis_host'] == 'redis' and val['redis_port'] == '6379':
        val['external_redis'] = False
    else:
        val['registry_db_index'], val['jobservice_db_index'], val['chartmuseum_db_index'] = map(int, val['redis_db_index'].split(','))
        val['external_redis'] = True

    this_dir = os.path.dirname(__file__)
    tpl = Environment(loader=FileSystemLoader(this_dir)).get_template('harbor.yml.jinja')

    with open(output_cfg, 'w') as f:
        f.write(tpl.render(**val))