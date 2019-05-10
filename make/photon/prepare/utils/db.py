import os

from g import config_dir, templates_dir
from utils.misc import prepare_config_dir
from utils.jinja import render_jinja

db_config_dir = os.path.join(config_dir, "db")
db_env_template_path = os.path.join(templates_dir, "db", "env.jinja")
db_conf_env = os.path.join(config_dir, "db", "env")

def prepare_db(config_dict):
    prepare_db_config_dir()

    render_jinja(
        db_env_template_path,
        db_conf_env,
        harbor_db_password=config_dict['harbor_db_password'])

def prepare_db_config_dir():
    prepare_config_dir(db_config_dir)