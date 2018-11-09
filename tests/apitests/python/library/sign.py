# -*- coding: utf-8 -*-

from subprocess import call
import shlex

def set_sign_env(registry_ip, project_name, image, tag):
    cmd = r"/home/travis/gopath/src/github.com/goharbor/harbor/tests/robot-cases/Group3-Upgrade/sign_image.sh", registry_ip, project_name, image, tag
    cmd = shlex(cmd)
    result_code = subprocess.call(cmd, shell=False)
    if result_code != 0:
        raise Exception("Failed to sign image error code is {}.".format(result_code))

