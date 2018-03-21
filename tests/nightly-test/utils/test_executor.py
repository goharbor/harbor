#!/usr/bin/python2

import os, subprocess
import time
import sys

from subprocess import call
import json

import nlogging
logger = nlogging.create_logger(__name__)

# Needs have docker installed.
def execute(harbor_endpoints, vm_names, harbor_root_pwd, test_suite, auth_mode ,vc_host, vc_user, vc_password, harbor_pwd='Harbor12345') :
    cmd = ''
    exe_result = -1
    cmd_base = "docker run -i --privileged -v %s:/drone -w /drone vmware/harbor-e2e-engine:1.38 " % os.getcwd()

    if len(harbor_endpoints) == 1:
        cmd_pybot = "pybot -v ip:%s -v vm_name:%s -v ip1: -v HARBOR_PASSWORD:%s -v SSH_PWD:%s -v vc_host:%s -v vc_user:%s -v vc_password:%s " % (harbor_endpoints[0], vm_names[0], harbor_pwd, harbor_root_pwd, vc_host, vc_user, vc_password)
    
    if len(harbor_endpoints) == 2:
        cmd_pybot = "pybot -v ip:%s -v vm_name:%s -v ip1:%s -v vm_name1:%s -v HARBOR_PASSWORD:%s -v SSH_PWD:%s -v vc_host:%s -v vc_user:%s -v vc_password:%s " % (harbor_endpoints[0], vm_names[0], harbor_endpoints[1], vm_names[1], harbor_pwd, harbor_root_pwd, vc_host, vc_user, vc_password)

    cmd = cmd_base + cmd_pybot
    if test_suite == 'Nightly':
        if auth_mode == 'ldap_auth':
            cmd = cmd + "/drone/tests/robot-cases/Group11-Nightly/LDAP.robot"           
        else:           
            cmd = cmd + "/drone/tests/robot-cases/Group11-Nightly/Nightly.robot"

        logger.info(cmd)
        p = subprocess.Popen(cmd, shell=True, stderr=subprocess.PIPE)
        while True:
            out = p.stderr.read(1)
            if out == '' and p.poll() != None:
                break
            if out != '':
                sys.stdout.write(out)
                sys.stdout.flush()
        exe_result = p.returncode

    if test_suite == 'Replication':
        cmd = cmd + "/drone/tests/robot-cases/Group11-Nightly/Replication.robot"

        logger.info(cmd)
        p = subprocess.Popen(cmd, shell=True, stderr=subprocess.PIPE)
        while True:
            out = p.stderr.read(1)
            if out == '' and p.poll() != None:
                break
            if out != '':
                sys.stdout.write(out)
                sys.stdout.flush()
        exe_result = p.returncode

    if test_suite == 'Longevity':
        cmd = cmd + "/drone/tests/robot-cases/Group12-Longevity/Longevity.robot > /dev/null 2>&1"
        logger.info(cmd)
        exe_result = subprocess.call(cmd, shell=True, stderr=subprocess.PIPE, stdout=subprocess.PIPE)
    
    collect_log()
    return exe_result == 0

# Needs to move log.html to another path it will be overwrite by any pybot run.
def collect_log():
    pass