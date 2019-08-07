# -*- coding: utf-8 -*-
import os
import subprocess

def sign_image(registry_ip, project_name, image, tag):
    result_code = subprocess.call(["/home/travis/gopath/src/github.com/goharbor/harbor/tests/apitests/python/sign_image.sh", registry_ip, project_name, image, tag], shell=False)
    #result_code = subprocess.call(["./sign_image.sh", registry_ip, project_name, image, tag], shell=False)
    if result_code != 0:
        raise Exception("Failed to sign image error code is {}.".format(result_code))
