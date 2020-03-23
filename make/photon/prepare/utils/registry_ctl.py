import os, shutil

from g import config_dir, templates_dir, DEFAULT_GID, DEFAULT_UID
from utils.misc import prepare_dir
from utils.jinja import render_jinja

registryctl_config_dir = os.path.join(config_dir, "registryctl")
registryctl_config_template_path = os.path.join(templates_dir, "registryctl", "config.yml.jinja")
registryctl_conf = os.path.join(config_dir, "registryctl", "config.yml")
registryctl_env_template_path = os.path.join(templates_dir, "registryctl", "env.jinja")
registryctl_conf_env = os.path.join(config_dir, "registryctl", "env")

def prepare_registry_ctl(config_dict):
    # prepare dir
    prepare_dir(registryctl_config_dir)

    # Render Registryctl env
    render_jinja(
        registryctl_env_template_path,
        registryctl_conf_env,
        **config_dict)

    # Render Registryctl config
    render_jinja(
        registryctl_config_template_path,
        registryctl_conf,
        uid=DEFAULT_UID,
        gid=DEFAULT_GID,
        **config_dict)
