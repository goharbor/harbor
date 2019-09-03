import os

from g import config_dir, templates_dir, DEFAULT_GID, DEFAULT_UID
from utils.misc import prepare_dir
from utils.jinja import render_jinja

log_config_dir = os.path.join(config_dir, "log")

# logrotate config file
logrotate_template_path = os.path.join(templates_dir, "log", "logrotate.conf.jinja")
log_rotate_config = os.path.join(config_dir, "log", "logrotate.conf")

# syslog docker config file
log_syslog_docker_template_path = os.path.join(templates_dir, 'log', 'rsyslog_docker.conf.jinja')
log_syslog_docker_config = os.path.join(config_dir, 'log', 'rsyslog_docker.conf')

def prepare_log_configs(config_dict):
    prepare_dir(log_config_dir)

    # Render Log config
    render_jinja(
        logrotate_template_path,
        log_rotate_config,
        uid=DEFAULT_UID,
        gid=DEFAULT_GID,
        **config_dict)

   # Render syslog docker config
    render_jinja(
        log_syslog_docker_template_path,
        log_syslog_docker_config,
        uid=DEFAULT_UID,
        gid=DEFAULT_GID,
        **config_dict
   )