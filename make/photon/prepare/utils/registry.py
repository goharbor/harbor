import os, shutil

from g import config_dir, templates_dir, DEFAULT_GID, DEFAULT_UID
from utils.misc import prepare_config_dir
from utils.jinja import render_jinja


registry_config_dir = os.path.join(config_dir, "registry")
registry_config_template_path = os.path.join(templates_dir, "registry", "config.yml.jinja")
registry_conf = os.path.join(config_dir, "registry", "config.yml")


def prepare_registry(config_dict):
    prepare_registry_config_dir()

    storage_provider_info = get_storage_provider_info(
    config_dict['storage_provider_name'],
    config_dict['storage_provider_config'],
    registry_config_dir)

    render_jinja(
        registry_config_template_path,
        registry_conf,
        uid=DEFAULT_UID,
        gid=DEFAULT_GID,
        storage_provider_info=storage_provider_info,
        **config_dict)

def prepare_registry_config_dir():
    prepare_config_dir(registry_config_dir)

def get_storage_provider_info(provider_name, provider_config, registry_config_dir_path):
    if provider_name == "filesystem":
        if not provider_config:
            storage_provider_config = "rootdirectory: /storage"
        elif "rootdirectory:" not in storage_provider_config:
            storage_provider_config = "rootdirectory: /storage" + "," + storage_provider_config
    # generate storage configuration section in yaml format
    storage_provider_conf_list = [provider_name + ':']
    for c in storage_provider_config.split(","):
        kvs = c.split(": ")
        if len(kvs) == 2:
            if kvs[0].strip() == "keyfile":
                srcKeyFile = kvs[1].strip()
                if os.path.isfile(srcKeyFile):
                    shutil.copyfile(srcKeyFile, os.path.join(registry_config_dir_path, "gcs.key"))
                    storage_provider_conf_list.append("keyfile: %s" % "/etc/registry/gcs.key")
                    continue
        storage_provider_conf_list.append(c.strip())
    storage_provider_info = ('\n' + ' ' * 4).join(storage_provider_conf_list)
    return storage_provider_info
