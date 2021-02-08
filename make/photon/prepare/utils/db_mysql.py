import os

from g import config_dir, templates_dir, data_dir, MYSQL_UID, MYSQL_GID
from utils.misc import prepare_dir
from utils.jinja import render_jinja

db_config_dir = os.path.join(config_dir, "db_mysql")
db_env_template_path = os.path.join(templates_dir, "db_mysql", "env.jinja")
db_conf_env = os.path.join(config_dir, "db_mysql", "env")
database_data_path = os.path.join(data_dir, 'mysql')

def prepare_db_mysql(config_dict):
    prepare_dir(database_data_path, uid=MYSQL_UID, gid=MYSQL_GID, mode=0o700)
    prepare_dir(db_config_dir)
    render_jinja(
        db_env_template_path,
        db_conf_env,
        harbor_db_password=config_dict['harbor_db_password'])
