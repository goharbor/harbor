import os

from g import templates_dir, config_dir, data_dir, DEFAULT_UID, DEFAULT_GID
from .jinja import render_jinja
from .misc import prepare_dir

trivy_adapter_template_dir = os.path.join(templates_dir, "trivy-adapter")


def prepare_trivy_adapter(config_dict):
    trivy_adapter_config_dir = prepare_dir(config_dir, "trivy-adapter")
    prepare_dir(data_dir, "trivy-adapter", "trivy", uid=DEFAULT_UID, gid=DEFAULT_GID)
    prepare_dir(data_dir, "trivy-adapter", "reports", uid=DEFAULT_UID, gid=DEFAULT_GID)

    trivy_adapter_env_path = os.path.join(trivy_adapter_config_dir, "env")
    trivy_adapter_env_template = os.path.join(trivy_adapter_template_dir, "env.jinja")

    render_jinja(
        trivy_adapter_env_template,
        trivy_adapter_env_path,
        **config_dict)
