#!/usr/bin/python2

import os, subprocess

SHELL_SCRIPT_DIR = '/harbor/workspace/harbor_nightly_test/tests/nightly-test/shellscript/'

def getvmip(vc_url, vc_user, vc_password, vm_name) :
    cmd = (SHELL_SCRIPT_DIR+'getvmip.sh %s %s %s %s ' % (vc_url, vc_user, getPasswordInShell(vc_password), vm_name))
    result = subprocess.check_output(cmd,shell=True)
    return result

def destroyvm(vc_url, vc_user, vc_password, vm_name) :
    cmd = (SHELL_SCRIPT_DIR+'destroyvm.sh %s %s %s %s ' % (vc_url, vc_user, getPasswordInShell(vc_password), vm_name))  
    result = subprocess.check_output(cmd, shell=True)
    return result

def getPasswordInShell(password) :
    return password.replace("!", "\!")
