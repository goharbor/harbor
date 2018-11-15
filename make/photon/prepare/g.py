import os

## Const
DEFAULT_UID = 10000
DEFAULT_GID = 10000

## Global variable
base_dir = '/harbor_make'
templates_dir = "/usr/src/app/templates"
config_dir = os.path.join(base_dir, "common/config")
config_file_path = os.path.join(base_dir, 'harbor.yml')

private_key_pem_template = os.path.join(templates_dir, "core", "private_key.pem")
root_cert_path_template = os.path.join(templates_dir, "registry", "root.crt")

cert_dir = os.path.join(config_dir, "nginx", "cert")
core_cert_dir = os.path.join(config_dir, "core", "certificates")
private_key_pem = os.path.join(config_dir, "core", "private_key.pem")
root_crt = os.path.join(config_dir, "registry", "root.crt")
registry_custom_ca_bundle_config = os.path.join(config_dir, "custom-ca-bundle.crt")