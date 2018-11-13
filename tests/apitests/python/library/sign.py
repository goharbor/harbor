# -*- coding: utf-8 -*-
import os
import subprocess

def sign_image_old(registry_ip, project_name, image, tag):
    #result_code = subprocess.call(["./tests/apitests/python/sign_image.sh", registry_ip, project_name, image, tag], shell=False)
    result_code = subprocess.call(["./sign_image.sh", registry_ip, project_name, image, tag], shell=False)
    if result_code != 0:
        raise Exception("Failed to sign image error code is {}.".format(result_code))

def sign_image(registry_ip, project_name, image, tag):
    #result_code = os.system("./tests/apitests/python/sign_image.sh {} {} {} {}".format(registry_ip, project_name, image, tag))
    result_code = os.system("export DOCKER_CONTENT_TRUST=1 && export DOCKER_CONTENT_TRUST_SERVER=https://10.192.127.8:4443")
    #result_code = os.system("./sign_image.sh {} {} {} {}".format(registry_ip, project_name, image, tag))