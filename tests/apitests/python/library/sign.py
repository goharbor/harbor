# -*- coding: utf-8 -*-

import commands
import os,sys

try:
    import docker
except ImportError:
    import pip
    pip.main(['install', 'docker'])
    import docker

def set_sign_env(registry_ip, project_name, image, tag):
    result_code, info = commands.getstatusoutput(r"chmod 777 sign_image.sh")
    if result_code != 0:
        raise Exception("image %s:%s exists" % (result_code, info))

    cmd = r"./sign_image.sh {} {} {} {} ".format(registry_ip, project_name, image, tag)
    result_code, info = commands.getstatusoutput(cmd)
    if result_code != 0:
        raise Exception("image %s:%s exists" % (result_code, info))

