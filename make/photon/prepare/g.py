import os
from pathlib import Path

## Const
DEFAULT_UID = 10000
DEFAULT_GID = 10000

PG_UID = 999
PG_GID = 999

REDIS_UID = 999
REDIS_GID = 999

## Global variable
host_root_dir = '/hostfs'

base_dir = '/harbor_make'
templates_dir = "/usr/src/app/templates"
config_dir = '/config'
data_dir = '/data'
secret_dir = '/secret'
secret_key_dir='/secret/keys'

old_private_key_pem_path = Path('/config/core/private_key.pem')
old_crt_path = Path('/config/registry/root.crt')

private_key_pem_path = Path('/secret/core/private_key.pem')
root_crt_path = Path('/secret/registry/root.crt')

config_file_path = '/compose_location/harbor.yml'
input_config_path = '/input/harbor.yml'
versions_file_path = Path('/usr/src/app/versions')

cert_dir = os.path.join(config_dir, "nginx", "cert")
core_cert_dir = os.path.join(config_dir, "core", "certificates")