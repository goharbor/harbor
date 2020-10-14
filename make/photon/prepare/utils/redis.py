import os

from g import config_dir, templates_dir, data_dir, REDIS_UID, REDIS_GID
from utils.misc import prepare_dir
from utils.jinja import render_jinja

redis_config_dir = os.path.join(config_dir, "redis")
redis_env_template_path = os.path.join(templates_dir, "redis", "env.jinja")
redis_conf_env = os.path.join(config_dir, "redis", "env")
redis_data_path = os.path.join(data_dir, 'redis')

def prepare_redis(config_dict):
    prepare_dir(redis_data_path, uid=REDIS_UID, gid=REDIS_GID)
    prepare_dir(redis_config_dir)
    render_jinja(
        redis_env_template_path,
        redis_conf_env,
        harbor_redis_password=config_dict['harbor_redis_password'])
