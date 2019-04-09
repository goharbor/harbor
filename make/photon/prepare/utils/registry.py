import os, copy

from g import config_dir, templates_dir, DEFAULT_GID, DEFAULT_UID
from utils.misc import prepare_config_dir
from utils.jinja import render_jinja


registry_config_dir = os.path.join(config_dir, "registry")
registry_config_template_path = os.path.join(templates_dir, "registry", "config.yml.jinja")
registry_conf = os.path.join(config_dir, "registry", "config.yml")


def prepare_registry(config_dict):
    prepare_config_dir(registry_config_dir)

    storage_provider_info = get_storage_provider_info(
    config_dict['storage_provider_name'],
    config_dict['storage_provider_config'])

    render_jinja(
        registry_config_template_path,
        registry_conf,
        uid=DEFAULT_UID,
        gid=DEFAULT_GID,
        storage_provider_info=storage_provider_info,
        **config_dict)


def get_storage_provider_info(provider_name, provider_config):
    provider_config_copy = copy.deepcopy(provider_config)
    if provider_name == "filesystem":
        if not (provider_config_copy and provider_config_copy.has_key('rootdirectory')):
            provider_config_copy['rootdirectory'] = '/storage'
    if provider_name == 'gcs' and provider_config_copy.get('keyfile'):
        provider_config_copy['keyfile'] = '/etc/registry/gcs.key'
    # generate storage configuration section in yaml format
    storage_provider_conf_list = [provider_name + ':']
    for config in provider_config_copy.items():
        storage_provider_conf_list.append('{}: {}'.format(*config))
    storage_provider_info = ('\n' + ' ' * 4).join(storage_provider_conf_list)
    return storage_provider_info
