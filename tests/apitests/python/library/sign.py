# -*- coding: utf-8 -*-

import commands
import os

def set_sign_env(registry_ip, project_name, image, tag):
    print "current path:", os.getcwd()
    cmd = r"/home/travis/gopath/src/github.com/goharbor/harbor/test/robot-cases/Group3-Upgrade/sign_image.sh {} {} {} {} ".format(registry_ip, project_name, image, tag)
    result_code, info = commands.getstatusoutput(cmd)
    if result_code != 0:
        raise Exception("image %s:%s exists" % (result_code, info))

