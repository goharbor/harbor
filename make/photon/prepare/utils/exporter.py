import os
from g import config_dir, templates_dir, DEFAULT_GID, DEFAULT_UID
from utils.jinja import render_jinja
from utils.misc import prepare_dir

EXPORTER_CONFIG_DIR = os.path.join(config_dir, "exporter")
EXPORTER_CONF_ENV = os.path.join(config_dir, "exporter", "env")
EXPORTER_ENV_TEMPLATE_PATH = os.path.join(templates_dir, "exporter", "env.jinja")

def prepare_exporter(config_dict):
    prepare_dir(EXPORTER_CONFIG_DIR, uid=DEFAULT_UID, gid=DEFAULT_GID)

    render_jinja(
        EXPORTER_ENV_TEMPLATE_PATH,
        EXPORTER_CONF_ENV,
        **config_dict)
