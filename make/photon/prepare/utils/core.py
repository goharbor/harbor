import os
import shutil
from g import config_dir, templates_dir, data_dir, DEFAULT_GID, DEFAULT_UID
from utils.jinja import render_jinja
from utils.misc import prepare_dir, generate_random_string

core_config_dir = os.path.join(config_dir, "core", "certificates")
core_env_template_path = os.path.join(templates_dir, "core", "env.jinja")
core_conf_env = os.path.join(config_dir, "core", "env")
core_conf_template_path = os.path.join(templates_dir, "core", "app.conf.jinja")
core_conf = os.path.join(config_dir, "core", "app.conf")

ca_download_dir = os.path.join(data_dir, 'ca_download')


def prepare_core(config_dict, with_trivy):
    prepare_dir(ca_download_dir, uid=DEFAULT_UID, gid=DEFAULT_GID)
    prepare_dir(core_config_dir)
    # Render Core

    render_jinja(
        core_env_template_path,
        core_conf_env,
        with_trivy=with_trivy,
        csrf_key=generate_random_string(32),
        scan_robot_prefix=generate_random_string(8),
        **config_dict)

    render_jinja(
        core_conf_template_path,
        core_conf,
        uid=DEFAULT_UID,
        gid=DEFAULT_GID,
        **config_dict)


def copy_core_config(core_templates_path, core_config_path):
    shutil.copyfile(core_templates_path, core_config_path)
    print("Generated configuration file: %s" % core_config_path)
