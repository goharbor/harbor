import os

from g import config_dir, templates_dir, DEFAULT_GID, DEFAULT_UID
from utils.misc import prepare_config_dir
from utils.jinja import render_jinja

log_config_dir = os.path.join(config_dir, "log")
logrotate_template_path = os.path.join(templates_dir, "log", "logrotate.conf.jinja")
log_rotate_config = os.path.join(config_dir, "log", "logrotate.conf")

def prepare_log_configs(config_dict):
    prepare_config_dir(log_config_dir)

    # Render Log config
    render_jinja(
        logrotate_template_path,
        log_rotate_config,
        uid=DEFAULT_UID,
        gid=DEFAULT_GID,
        **config_dict)