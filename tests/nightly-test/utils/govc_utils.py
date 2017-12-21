#!/usr/bin/python2

import os, subprocess
import time

SHELL_SCRIPT_DIR = '/harbor/workspace/harbor_nightly_test/tests/nightly-test/shellscript/'

def getvmip(vc_url, vc_user, vc_password, vm_name, timeout=600) :
    cmd = (SHELL_SCRIPT_DIR+'getvmip.sh %s %s %s %s ' % (vc_url, vc_user, getPasswordInShell(vc_password), vm_name))
    print cmd
    interval = 10
    results = []
    while True:
        try:
            if timeout <= 0:
                print "timeout to get ova ip"
                return -1

            result = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
            for line in result.stdout.readlines():
                print line
                results.append(line)

            print "######"
            print results[0]
            print results[0] is not ''
            print results[0] is not None                        
            print "######"            
            if results[0] is not '' and results[0] is not "photon-machine":
                print results[0]
                return 0
        except Exception, e:
            timeout -= interval
            time.sleep(interval)
            continue
        timeout -= interval
        time.sleep(interval)
    return result

def destroyvm(vc_url, vc_user, vc_password, vm_name) :
    cmd = (SHELL_SCRIPT_DIR+'destroyvm.sh %s %s %s %s ' % (vc_url, vc_user, getPasswordInShell(vc_password), vm_name))  
    result = subprocess.check_output(cmd, shell=True)
    return result

def getPasswordInShell(password) :
    return password.replace("!", "\!")
