import os

from g import templates_dir, config_dir
from .jinja import render_jinja
from .misc import prepare_dir

clair_adapter_template_dir = os.path.join(templates_dir, "clair-adapter")

def prepare_clair_adapter(config_dict):
    clair_adapter_config_dir = prepare_dir(config_dir, "clair-adapter")

    clair_adapter_env_path = os.path.join(clair_adapter_config_dir, "env")
    clair_adapter_env_template = os.path.join(clair_adapter_template_dir, "env.jinja")

    render_jinja(
        clair_adapter_env_template,
        clair_adapter_env_path,
        **config_dict)
