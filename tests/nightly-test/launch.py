#!/usr/bin/python2

import sys
import os
import ConfigParser
from subprocess import call
from datetime import datetime
import time
dir_path = os.path.dirname(os.path.realpath(__file__))
sys.path.append(dir_path + '/utils')
sys.path.append(dir_path + '/deployment')
import harbor_util
import test_executor
import deployer

if len(sys.argv)!=7 :
    print "python launch.py <build_type> <image_url> <test suitename> <config_file> <dry_run>"
    print "Wrong parameters, quit test"
    quit()

build_type = sys.argv[1]
image_url = sys.argv[2]
test_suite = sys.argv[3]
config_file = sys.argv[4]
deploy_count = int(sys.argv[5])
dry_run = sys.argv[6]
config_file = "/harbor/workspace/harbor_nightly_test_yan/harbor_nightly_test/testenv.ini"
#  config_file = "/Users/daojunz/Documents/harbor_nightly_test/testenv.ini"
config = ConfigParser.ConfigParser()
config.read(config_file)
harbor_endpoints = []

# ----- deploy harbor build -----
if build_type == "ova" :
    vc_host = config.get("vcenter", "vc_host")
    vc_user = config.get("vcenter", "vc_user")
    vc_password = config.get("vcenter", "vc_password")
    datastore = config.get("vcenter", "datastore")
    cluster = config.get("vcenter", "cluster")
    ova_password = config.get("vcenter", "ova_password")
    ova_name = config.get("vcenter", "ova_name")
    ova_name = ova_name +"-"+ datetime.now().isoformat().replace(":", "-").replace(".", "-")

    ova_deployer = OVADeployer(vc_host, 
                vc_user,
                vc_password, 
                datastore, 
                cluster, 
                image_url, 
                ova_name, 
                ova_password,
                deploy_count,
                dry_run)

    harbor_endpoints = ova_deployer.deploy()

elif build_type == "installer" :
    print "Going to download installer image to install"
elif build_type == "all" :
    print "launch ova and installer"

# ----- wait for harbor ready -----
for item in harbor_endpoints:
    is_harbor_ready = harbor_util.wait_for_harbor_ready("https://"+item)
    if not is_harbor_ready:
        print "Harbor is not ready after 10 minutes."
        sys.exit(-1)
    print "%s is ready for test now..." % item

# ----- execute test cases -----
execute_results = test_executor.execute(harbor_endpoints, ova_password, test_suite)
if not execute_results:
    print "execute test failure."
    sys.exit(-1)
