from __future__ import print_function
import utils
import os
acceptable_versions = ['1.6.0']
keys = [
    'hostname',
    'ui_url_protocol',
    'customize_crt',
    'ssl_cert',
    'ssl_cert_key',
    'secretkey_path',
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
    'clair_db_host',
    'clair_db_password',
    'clair_db_port',
    'clair_db_username',
    'clair_db',
    'uaa_endpoint',
    'uaa_clientid',
    'uaa_clientsecret',
    'uaa_verify_cert',
    'uaa_ca_cert',
    'registry_storage_provider_name',
    'registry_storage_provider_config'
    ]

def migrate(input_cfg, output_cfg):
    d = utils.read_conf(input_cfg)
    val = {}
    for k in keys:
        val[k] = d.get(k,'')
    #append registry to no_proxy
    np_list = d.get('no_proxy','').split(',')
    new_np_list = ['core' if x=='ui' else x for x in np_list]
    val['no_proxy'] = ','.join(new_np_list)
    tpl_path = os.path.join(os.path.dirname(__file__), 'harbor.cfg.tpl')
    utils.render(tpl_path, output_cfg, **val)
