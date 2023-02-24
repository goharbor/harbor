# -*- coding: utf-8 -*-

import os
import base


def helm3_7_registry_login(ip, user, password):
    command = ["helm3.7", "registry", "login", ip, "-u", user, "-p", password]
    base.run_command(command)

def helm3_7_package(file_path):
    command = ["helm3.7", "package", file_path]
    base.run_command(command)

def helm3_7_push(file_path, ip, project_name):
    command = ["helm3.7", "push", file_path, "oci://{}/{}".format(ip, project_name)]
    base.run_command(command)
