import os

from g import config_dir, templates_dir, data_dir, PG_UID, PG_GID
from utils.misc import prepare_dir
from utils.jinja import render_jinja

db_config_dir = os.path.join(config_dir, "db")
db_env_template_path = os.path.join(templates_dir, "db", "env.jinja")
db_conf_env = os.path.join(config_dir, "db", "env")
database_data_path = os.path.join(data_dir, 'database')

def prepare_db(config_dict):
    prepare_dir(database_data_path, uid=PG_UID, gid=PG_GID, mode=0o700)
    prepare_dir(db_config_dir)
    render_jinja(
        db_env_template_path,
        db_conf_env,
        harbor_db_password=config_dict['harbor_db_password'])
