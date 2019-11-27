import os

from g import data_dir, REDIS_UID, REDIS_GID
from utils.misc import prepare_dir

redis_data_path = os.path.join(data_dir, 'redis')

def prepare_redis(config_dict):
    prepare_dir(redis_data_path, uid=REDIS_UID, gid=REDIS_GID)
