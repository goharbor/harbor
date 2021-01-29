# -*- coding: utf-8 -*-
import subprocess
from testutils import notary_url, BASE_IMAGE_ABS_PATH_NAME
from docker_api import docker_load_image, docker_image_clean_all

def sign_image(registry_ip, project_name, image, tag):
    docker_load_image(BASE_IMAGE_ABS_PATH_NAME)
    try:
        ret = subprocess.check_output(["./tests/apitests/python/sign_image.sh", registry_ip, project_name, image, tag, notary_url], shell=False)
        print("sign_image return: ", ret)
    except subprocess.CalledProcessError as e:
        raise Exception("Failed to sign image error is {} {}.".format(e.returncode, e.output))
    finally:
        docker_image_clean_all()

