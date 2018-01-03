#!/usr/bin/python2

import os, subprocess
import time
import sys

from subprocess import call
import json

import nlogging
logger = nlogging.create_logger()

# Needs have docker installed.
def execute(harbor_endpoints, harbor_root_pwd, test_suite, harbor_pwd='Harbor12345') :
    cmd = ''
    cmd_base = "docker run -i --privileged -v /harbor/workspace/harbor_nightly_test_yan:/drone -w /drone vmware/harbor-e2e-engine:1.38 "

    if len(harbor_endpoints) == 1:
        cmd_pybot = "pybot -v ip:%s -v HARBOR_PASSWORD:%s -v SSH_PWD:%s " % (harbor_endpoints[0], harbor_pwd, harbor_root_pwd)
    
    if len(harbor_endpoints) == 2:
        cmd_pybot = "pybot -v ip:%s ip1:%s -v HARBOR_PASSWORD:%s -v SSH_PWD:%s " % (harbor_endpoints[1], harbor_pwd, harbor_root_pwd)

    cmd = cmd_base + cmd_pybot
    if test_suite == 'Nightly':
        cmd = cmd + "/drone/tests/robot-cases/Group11-Nightly/Nightly.robot"

    if test_suite == 'Longevity':
        cmd = cmd + "/drone/tests/robot-cases/Group10-Longevity/Longevity.robot"
    
    logger.info(cmd)
    result = os.system(cmd)

    collect_log()
    return results_flag()

# Needs to move log.html to another path it will be overwrite by any pybot run.
def collect_log():
    pass

# True means test pass
def results_flag():
    return true