# -*- coding: utf-8 -*-

import os
import subprocess

def set_sign_env(registry_ip, project_name, image, tag):
    result_code = subprocess.call(["/home/travis/gopath/src/github.com/goharbor/harbor/tests/robot-cases/Group3-Upgrade/sign_image.sh", registry_ip, project_name, image, tag])
    if result_code != 0:
        raise Exception("Failed to sign image error code is {} error info is {}.".format(result_code, info))

