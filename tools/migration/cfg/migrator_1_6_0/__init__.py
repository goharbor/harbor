from __future__ import print_function
import utils
import os
acceptable_versions = ['1.5.0']
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
    if not 'registry' in np_list:
        np_list.append('registry')
        val['no_proxy'] = ','.join(np_list)
    #handle harbor db information, if it previously pointed to internal mariadb, point it to the new default db instance of pgsql,
    #update user to default pgsql user.
    if 'mysql' == d['db_host']:
        val['db_host'] = 'postgresql'
        val['db_port'] = '5432'
        val['db_user'] = 'postgres'
    #handle clair db information, if it pointed to internal pgsql in previous deployment, point it to the new default db instance of pgsql,
    #the user should be the same user as harbor db
    if 'postgres' == d['clair_db_host']:
        val['clair_db_host'] = 'postgresql'
        val['cliar_db_user'] = val['db_user']
        val['clair_db_password'] = val['db_password']
    tpl_path = os.path.join(os.path.dirname(__file__), 'harbor.cfg.tpl')
    utils.render(tpl_path, output_cfg, **val)
