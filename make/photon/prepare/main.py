import click

from utils.admin_server import prepare_adminserver
from utils.misc import delfile
from utils.configs import validate, parse_yaml_config
from utils.cert import prepare_ca, SSL_CERT_KEY_PATH, SSL_CERT_PATH, get_secret_key
from utils.db import prepare_db
from utils.jobservice import prepare_job_service
from utils.registry import prepare_registry
from utils.registry_ctl import prepare_registry_ctl
from utils.core import prepare_core
from utils.uaa import prepare_uaa_cert_file
from utils.notary import prepare_notary
from utils.log import prepare_log_configs
from utils.clair import prepare_clair
from utils.chart import prepare_chartmuseum
from utils.docker_compose import prepare_docker_compose
from utils.nginx import prepare_nginx, nginx_confd_dir
from g import (config_dir, private_key_pem_template, config_file_path, core_cert_dir, private_key_pem, 
root_crt, root_cert_path_template, registry_custom_ca_bundle_config)

# Main function
@click.command()
@click.option('--conf', default=config_file_path, help="the path of Harbor configuration file")
@click.option('--with-notary', is_flag=True, help="the Harbor instance is to be deployed with notary")
@click.option('--with-clair', is_flag=True, help="the Harbor instance is to be deployed with clair")
@click.option('--with-chartmuseum', is_flag=True, help="the Harbor instance is to be deployed with chart repository supporting")
def main(conf, with_notary, with_clair, with_chartmuseum):

    delfile(config_dir)
    config_dict = parse_yaml_config(conf)
    validate(config_dict, notary_mode=with_notary)

    prepare_log_configs(config_dict)
    prepare_nginx(config_dict)
    prepare_adminserver(config_dict, with_notary=with_notary, with_clair=with_clair, with_chartmuseum=with_chartmuseum)
    prepare_core(config_dict)
    prepare_registry(config_dict)
    prepare_registry_ctl(config_dict)
    prepare_db(config_dict)
    prepare_job_service(config_dict)

    get_secret_key(config_dict['secretkey_path'])
    if config_dict['auth_mode'] == "uaa_auth":
        prepare_uaa_cert_file(config_dict['uaa_ca_cert'], core_cert_dir)

    #  If Customized cert enabled
    prepare_ca(
        customize_crt=config_dict['customize_crt'],
        private_key_pem_path=private_key_pem,
        private_key_pem_template=private_key_pem_template,
        root_crt_path=root_crt,
        root_cert_template_path=root_cert_path_template,
        registry_custom_ca_bundle_path=config_dict['registry_custom_ca_bundle_path'],
        registry_custom_ca_bundle_config=registry_custom_ca_bundle_config)

    if with_notary:
        prepare_notary(config_dict, nginx_confd_dir, SSL_CERT_PATH, SSL_CERT_KEY_PATH)

    if with_clair:
        prepare_clair(config_dict)

    if with_chartmuseum:
        prepare_chartmuseum(config_dict)

    prepare_docker_compose(config_dict, with_clair, with_notary, with_chartmuseum)

if __name__ == '__main__':
    main()