import os, shutil

from g import templates_dir, config_dir, DEFAULT_UID, DEFAULT_GID
from .jinja import render_jinja
from .misc import prepare_config_dir

clair_template_dir = os.path.join(templates_dir, "clair")

def prepare_clair(config_dict):
    clair_config_dir = prepare_config_dir(config_dir, "clair")

    if os.path.exists(os.path.join(clair_config_dir, "postgresql-init.d")):
        print("Copying offline data file for clair DB")
        shutil.rmtree(os.path.join(clair_config_dir, "postgresql-init.d"))

    shutil.copytree(os.path.join(clair_template_dir, "postgresql-init.d"), os.path.join(clair_config_dir, "postgresql-init.d"))

    postgres_env_path = os.path.join(clair_config_dir, "postgres_env")
    postgres_env_template = os.path.join(clair_template_dir, "postgres_env.jinja")

    clair_config_path = os.path.join(clair_config_dir, "config.yaml")
    clair_config_template = os.path.join(clair_template_dir, "config.yaml.jinja")

    clair_env_path = os.path.join(clair_config_dir, "clair_env")
    clair_env_template = os.path.join(clair_template_dir, "clair_env.jinja")

    render_jinja(
        postgres_env_template,
        postgres_env_path,
        password=config_dict['db_password'])

    render_jinja(
        clair_config_template,
        clair_config_path,
        uid=DEFAULT_UID,
        gid=DEFAULT_GID,
        password= config_dict['db_password'],
        username= config_dict['db_user'],
        host= config_dict['db_host'],
        port= config_dict['db_port'],
        dbname= config_dict['clair_db'],
        interval= config_dict['clair_updaters_interval'])

    # config http proxy for Clair
    render_jinja(
        clair_env_template,
        clair_env_path,
        **config_dict)
