import os

from g import data_dir, REDIS_UID, REDIS_GID
from utils.misc import prepare_config_dir

redis_data_path = os.path.join(data_dir, 'redis')

def prepare_redis(config_dict):
    prepare_config_dir(redis_data_path)

    stat_info = os.stat(redis_data_path)
    uid, gid = stat_info.st_uid, stat_info.st_gid
    if not (uid == REDIS_UID and gid == REDIS_GID):
            os.chown(redis_data_path, REDIS_UID, REDIS_GID)