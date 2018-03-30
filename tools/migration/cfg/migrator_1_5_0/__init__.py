from __future__ import print_function
import utils
import os
acceptable_versions = ['1.2.0', '1.3.0', '1.4.0']

#In 1.5 template the placeholder has the same value as the attribute name.
default = {
    'log_rotate_count':'50',
    'log_rotate_size':'200M',
    'db_host':'mysql',
    'db_port':'3306',
    'db_user':'root',
    'redis_url':'',
    'clair_db_host':'postgres',
    'clair_db_port':'5432',
    'clair_db_username':'postgres',
    'clair_db':'postgres',
    'uaa_endpoint':'uaa.mydomain.org',
    'uaa_clientid':'id',
    'uaa_clientsecret':'secret',
    'uaa_verify_cert':'true',
    'uaa_ca_cert':'/path/to/ca.pem',
    'registry_storage_provider_name':'filesystem',
    'registry_storage_provider_config':''
}

def migrate(input_cfg, output_cfg):
    d = utils.read_conf(input_cfg)
    keys = list(default.keys())
    keys.extend(['hostname', 'ui_url_protocol', 'max_job_workers', 'customize_crt', 
            'ssl_cert', 'ssl_cert_key', 'secretkey_path', 'admiral_url', 'db_password', 'clair_db_password'])
    val = {}
    for k in keys:
        if k in d:
            val[k] = d[k]
        else:
            val[k] = default[k]
    tpl_path = os.path.join(os.path.dirname(__file__), 'harbor.cfg.tpl')
    utils.render(tpl_path, output_cfg, **val)


