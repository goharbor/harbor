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
import buildweb_utils
import govc_utils
import nlogging
logger = nlogging.create_logger(__name__)
import test_executor
from deployer import *

if len(sys.argv)!=9 :
    logger.info("python launch.py <build_type> <image_url> <test suitename> <config_file> <dry_run> <auth_mode> <destory>")
    logger.info("Wrong parameters, quit test")
    quit()

build_type = sys.argv[1]
image_url = sys.argv[2]
test_suite = sys.argv[3]
config_file = sys.argv[4]
deploy_count = int(sys.argv[5])
dry_run = sys.argv[6]
auth_mode = sys.argv[7]
destory = sys.argv[8]
config_file = os.getcwd() + "/harbor_nightly_test/testenv.ini"
config = ConfigParser.ConfigParser()
config.read(config_file)
harbor_endpoints = []
vm_names = []

# ----- deploy harbor build -----
if build_type == "ova" :
    vc_host = config.get("vcenter", "vc_host")
    vc_user = config.get("vcenter", "vc_user")
    vc_password = config.get("vcenter", "vc_password")
    datastore = config.get("vcenter", "datastore")
    cluster = config.get("vcenter", "cluster")
    ova_password = config.get("vcenter", "ova_password")
    ova_name = config.get("vcenter", "ova_name")

    ldap_url = config.get("ldap", "ldap_url")
    ldap_searchdn = config.get("ldap", "ldap_searchdn")
    ldap_search_pwd = config.get("ldap", "ldap_search_pwd")
    ldap_filter = config.get("ldap", "ldap_filter")
    ldap_basedn = config.get("ldap", "ldap_basedn")
    ldap_uid = config.get("ldap", "ldap_uid")
    ldap_scope = config.get("ldap", "ldap_scope")
    ldap_timeout = config.get("ldap", "ldap_timeout")

    if image_url == "latest" :
        image_url = buildweb_utils.get_latest_build_url('master','beta')
    logger.info("Get latest image url:" + image_url)

    ova_deployer = OVADeployer(vc_host, 
                vc_user,
                vc_password, 
                datastore, 
                cluster, 
                image_url, 
                ova_name, 
                ova_password,
                deploy_count,
                dry_run,
                auth_mode,
                ldap_url,
                ldap_searchdn,
                ldap_search_pwd,
                ldap_filter,
                ldap_basedn,
                ldap_uid,
                ldap_scope,
                ldap_timeout)

    logger.info("Going to deploy harbor ova..")
    harbor_endpoints, vm_names = ova_deployer.deploy()

    # ----- wait for harbor ready -----
    for item in harbor_endpoints:
        is_harbor_ready = harbor_util.wait_for_harbor_ready("https://"+item)
        if not is_harbor_ready:
            logger.info("Harbor is not ready after 10 minutes.")
            sys.exit(-1)
        logger.info("%s is ready for test now..." % item)
        # ----- log harbor version -----
        harbor_version = harbor_util.get_harbor_version(item, 'admin', 'Harbor12345')
        logger.info("Harbor version: %s ..." % harbor_version)
        with open(os.getcwd() + '/build.properties', 'w') as the_file:
            the_file.write('harbor_version=%s' % harbor_version)

    # ----- execute test cases -----
    try:
        execute_results = test_executor.execute(harbor_endpoints, vm_names, ova_password, test_suite, auth_mode, vc_host, vc_user, vc_password)
        if not execute_results:
            logger.info("execute test failure.")
            sys.exit(-1)
    except Exception, e:
        logger.info(e)
        sys.exit(-1)
    finally:
        if destory == "true":
            ova_deployer.destory()

elif build_type == "installer" :
    logger.info("Going to download installer image to install")
elif build_type == "all" :
    logger.info("launch ova and installer")
