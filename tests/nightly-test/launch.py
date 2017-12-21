#!/usr/bin/python2

import sys
import os
import ConfigParser
from subprocess import call
from datetime import datetime
import time

dir_path = os.path.dirname(os.path.realpath(__file__))
sys.path.append(dir_path + '/utils')

import ova_utils
import govc_utils
import harbor_util
import buildweb_utils

if len(sys.argv)!=6 :
    print "python launch.py <build_type> <image_url> <test suitename> <config_file> <dry_run>"
    print "Wrong parameters, quit test"
    quit()

build_type = sys.argv[1]
image_url = sys.argv[2]
test_suite = sys.argv[3]
config_file = sys.argv[4]
dry_run = sys.argv[5]
config_file = "/harbor/workspace/harbor_nightly_test/harbor_nightly_test/testenv.ini"
#  config_file = "/Users/daojunz/Documents/harbor_nightly_test/testenv.ini"

harbor_ova_endpoint = ''

config = ConfigParser.ConfigParser()
config.read(config_file)

if build_type == "ova" :
    print "Going to install ova on target machine!"
    vc_host = config.get("vcenter", "vc_host")
    print "vc_host:", vc_host
    vc_user = config.get("vcenter", "vc_user")
    print "vc_user:", vc_user
    vc_password = config.get("vcenter", "vc_password")
    print "vc_password:", vc_password
    datastore = config.get("vcenter", "datastore")
    cluster = config.get("vcenter", "cluster")
    ova_password = config.get("vcenter", "ova_password")
    ova_name = config.get("vcenter", "ova_name")
    
    ova_name = ova_name +"-"+ datetime.now().isoformat().replace(":", "-").replace(".", "-")
    print "ova_name:", ova_name

    print "image url:", image_url

    if image_url == "latest" :
        buildweb = buildweb_utils.BuildWebUtil()
        build_id=buildweb.get_latest_recommend_build('harbor_build', 'master')
        image_url = buildweb.get_deliverable_by_build_id(build_id, '.*.ovf')
        print "Get latest image url:" + image_url

    ova_utils.deploy_ova(vc_host, 
                vc_user,
                vc_password, 
                datastore, 
                cluster, 
                image_url, 
                ova_name, 
                ova_password,
                dry_run)

    time.sleep(20)

    harbor_ova_endpoint = govc_utils.getvmip(vc_host, vc_user, vc_password, ova_name)
    print "OVA install complete, start to test now, fqdn=" + harbor_ova_endpoint    
    print "run test now"
    print "test done"
    print "Destorying vm after test"

elif build_type == "installer" :
    print "Going to download installer image to install"
    vm_host = config.get("vm", "vm_host")
    vm_user = config.get("vm", "vm_user")
    vm_password = config.get("vm", "vm_password")
elif build_type == "all" :
    print "launch ova and installer"

print "All test done!"

if harbor_ova_endpoint is not None:
    result = harbor_util.wait_for_harbor_ready("https://"+harbor_ova_endpoint)
    if result != 0:
        print "Harbor is not ready after 10 minutes."