#!/usr/bin/python2

import os, subprocess
import time

# Needs have docker installed.
def execute_test_ova(harbor_endpoint, harbor_root_pwd, test_suite, harbor_pwd='Harbor12345') :
    cmd = "docker run -it --privileged -v /harbor/workspace/harbor_nightly_test_yan:/drone -w /drone vmware/harbor-e2e-engine:1.38 pybot -v ip:%s -v HARBOR_PASSWORD:%s -v SSH_PWD:%s " % (harbor_endpoint, harbor_pwd, harbor_root_pwd)
    if test_suite == 'Nightly':
        cmd = cmd + "/drone/tests/robot-cases/Group11-Nightly/Nightly.robot"

    if test_suite == 'Longevity':
        cmd = cmd + "/drone/tests/robot-cases/Group10-Longevity/Longevity.robot"
    
    print cmd
    rc = subprocess.check_output(cmd,shell=True).strip()
    collect_log()
    return rc

# Needs to move log.html to another path it will be overwrite by any pybot run.
def collect_log():
    pass