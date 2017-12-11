import logging
import os
import sys
import time

#from subprocess import call
import subprocess

LOG = logging.getLogger(__name__)
#DEFAULT_LOCAL_OVF_TOOL_PATH = "/Users/daojunz/Documents/ovftool/ovftool/ovftool"
DEFAULT_LOCAL_OVF_TOOL_PATH = '/home/harbor-ci/ovftool/ovftool'
OMJS_PATH = '/opt/vmware/vio/etc/omjs.properties'


def get_ovf_tool_path():
    platform = sys.platform
    os.environ['TCROOT'] = "/build/toolchain/"

    if platform.startswith('linux'):
        path = 'lin64'
    elif platform.startswith('win'):
        path = 'win64'
    elif platform.startswith('darwin'):
        path = 'mac32'
    else:
        LOG.debug("unsupported platform %s" % platform)
        return None
    ovf_path = os.environ['OVF_TOOL'] = "%s/%s/ovftool-4.1.0/ovftool" \
                                        % (os.environ['TCROOT'], path)
    if os.path.isfile(ovf_path):
        # check if file exists
        LOG.debug("ovf tool exists at the following location: %s" % ovf_path)
    else:
        LOG.debug("couldn't not find ovftool in toolchain %s " % ovf_path)
        ovf_path = None
    return ovf_path


def wait_for_mgmt_service(oms_ip, vc_user, vc_password):
    LOG.info('Waiting for management service')
    cmd_utils.wait_for(func=OmsController, timeout=500, delay=10, oms=oms_ip,
                        sso_user=vc_user, sso_pwd=vc_password)
    LOG.info('Management service is running.')



def deploy_ova(vc_host, vc_user, vc_password, ds, cluster,  ova_path, ova_name, ova_root_password, dry_run, auth_mode="db_auth", harbor_password="Harbor12345", log_path=None, ip=None, netmask=None, gateway=None, dns=None, ovf_tool_path=None):
    if not ovf_tool_path:
        ovf_tool_path = DEFAULT_LOCAL_OVF_TOOL_PATH
    if not os.path.isfile(ovf_tool_path):
        LOG.error("ovftool not found.")

    cmd = (
         '"%s" --X:"logFile"="./deploy_oms.log" --overwrite --powerOn --datastore=\'%s\' --noSSLVerify --acceptAllEulas --name=%s \
          --X:injectOvfEnv --X:enableHiddenProperties  --prop:root_pwd=\'%s\' --prop:permit_root_login=true --prop:auth_mode=\'%s\' \
          --prop:harbor_admin_password=\'%s\' --prop:max_job_workers=5   %s  \
          vi://%s:\'%s\'@%s/Datacenter/host/%s'
          % (ovf_tool_path, ds, ova_name,
            ova_root_password, auth_mode,
            harbor_password, ova_path,
            vc_user, vc_password, vc_host, cluster
          )
    )

    print cmd
    print 'Start to deploy harbor OVA.'
    print dry_run
    if dry_run == "true" :
        print "Dry run ..."
    else:
        subprocess.check_output(cmd, shell=True)
    print 'Successfully deployed harbor OVA.'   

