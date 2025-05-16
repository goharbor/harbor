import os

from g import config_dir, DEFAULT_GID, DEFAULT_UID, templates_dir
from utils.misc import prepare_dir
from utils.jinja import render_jinja

job_config_dir = os.path.join(config_dir, "jobservice")
job_service_env_template_path = os.path.join(templates_dir, "jobservice", "env.jinja")
job_service_conf_env = os.path.join(config_dir, "jobservice", "env")
job_service_conf_template_path = os.path.join(templates_dir, "jobservice", "config.yml.jinja")
jobservice_conf = os.path.join(config_dir, "jobservice", "config.yml")

def prepare_job_service(config_dict):
    prepare_dir(job_config_dir, uid=DEFAULT_UID, gid=DEFAULT_GID)

    log_level = config_dict['log_level'].upper()

    # Job log and exported reports are stored in data dir
    job_log_dir = os.path.join('/data', "job_logs")
    prepare_dir(job_log_dir, uid=DEFAULT_UID, gid=DEFAULT_GID)

    # Render Jobservice env
    render_jinja(
        job_service_env_template_path,
        job_service_conf_env,
        **config_dict)

    # Render Jobservice config
    render_jinja(
        job_service_conf_template_path,
        jobservice_conf,
        uid=DEFAULT_UID,
        gid=DEFAULT_GID,
        internal_tls=config_dict['internal_tls'],
        max_job_workers=config_dict['max_job_workers'],
        max_job_duration_hours=config_dict['max_job_duration_hours'],
        max_job_duration_seconds=config_dict['max_job_duration_seconds'],
        job_loggers=config_dict['job_loggers'],
        logger_sweeper_duration=config_dict['logger_sweeper_duration'],
        redis_url=config_dict['redis_url_js'],
        level=log_level,
        metric=config_dict['metric'])
