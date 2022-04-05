# -*- coding: utf-8 -*-

import base
import json
import docker_api

def ctr_images_pull(username, password, oci):
    command = ["ctr", "images", "pull","--snapshotter", "native", "-u", username+":"+password, oci]
    ret = base.run_command(command)
    print("Command return: ", ret)

def ctr_images_list(oci_ref = None):
    command = ["ctr", "images", "list", "--q"]
    print("Command: ", command)
    ret = base.run_command(command)
    print("Command return: ", ret)
    if oci_ref is not None and oci_ref not in ret.split("\n"):
        raise Exception(r" Get OCI ref failed, expected ref is [{}], but return ref list is [{}]".format (ret))


