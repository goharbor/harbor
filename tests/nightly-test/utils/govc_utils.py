#!/usr/bin/python2

import os, subprocess
import time

SHELL_SCRIPT_DIR = os.getcwd() + '/tests/nightly-test/shellscript/'

def getvmip(vc_url, vc_user, vc_password, vm_name, timeout=600) :
    cmd = (SHELL_SCRIPT_DIR+'getvmip.sh %s %s %s %s ' % (vc_url, vc_user, getPasswordInShell(vc_password), vm_name))
    interval = 10
    while True:
        try:
            if timeout <= 0:
                return ''
            result = subprocess.check_output(cmd,shell=True).strip()
            if result is not '':
                if result != 'photon-machine':
                    return result
        except Exception, e:
            timeout -= interval
            time.sleep(interval)
            continue
        timeout -= interval
        time.sleep(interval)

def destoryvm(vc_url, vc_user, vc_password, vm_name) :
    cmd = (SHELL_SCRIPT_DIR+'destoryvm.sh %s %s %s %s ' % (vc_url, vc_user, getPasswordInShell(vc_password), vm_name))  
    result = subprocess.check_output(cmd, shell=True)
    return result

def getPasswordInShell(password) :
    return password.replace("!", "\!")