import os

from g import config_dir, templates_dir, data_dir, PG_UID, PG_GID
from utils.misc import prepare_config_dir
from utils.jinja import render_jinja

db_config_dir = os.path.join(config_dir, "db")
db_env_template_path = os.path.join(templates_dir, "db", "env.jinja")
db_conf_env = os.path.join(config_dir, "db", "env")
database_data_path = os.path.join(data_dir, 'database')

def prepare_db(config_dict):
    prepare_config_dir(database_data_path)
    stat_info = os.stat(database_data_path)
    uid, gid = stat_info.st_uid, stat_info.st_gid
    if not (uid == PG_UID and gid == PG_GID):
            os.chown(database_data_path, PG_UID, PG_GID)
    prepare_config_dir(db_config_dir)
    render_jinja(
        db_env_template_path,
        db_conf_env,
        harbor_db_password=config_dict['harbor_db_password'])
