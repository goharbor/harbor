# -*- coding: utf-8 -*-

import commands
import os

def set_sign_env(registry_ip, project_name, image, tag):
    cmd = r"../../robot-cases/Group3-Upgrade/sign_image.sh {} {} {} {} ".format(registry_ip, project_name, image, tag)
    result_code, info = commands.getstatusoutput(cmd)
    if result_code != 0:
        raise Exception("image %s:%s exists" % (result_code, info))

