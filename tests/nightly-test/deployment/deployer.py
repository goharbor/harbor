import abc

class Deployer(abc.ABC):
    __metaclass__ = abc.ABCMeta

    @abc.abstractmethod
    def deploy(self):
        return
    
class OVADeployer(Deployer):

    def __init__(self, vc_host, vc_user, vc_password, ds, cluster, ova_path, ova_name, ova_root_password, dry_run, count):
        self.vc_host = '' 
        self.vc_user = '' 
        self.vc_password = '' 
        self.ds = '' 
        self.cluster = ''  
        self.ova_path = '' 
        self.ova_name = '' 
        self.ova_root_password = '' 
        self.dry_run = ''
        self.count = 1 
        
        self.auth_mode="db_auth", 
        self.harbor_password="Harbor12345", 
        self.log_path=None 
        self.ip=None
        self.netmask=None
        self.gateway=None
        self.dns=None
        self.ovf_tool_path=None
        self.DEFAULT_LOCAL_OVF_TOOL_PATH = '/home/harbor-ci/ovftool/ovftool'
        self.ova_endpoints = []
        self.ova_names = []

    def __generate_ova_names(self):
        for i in range(0, self.count):
            ova_name_temp = self.ova_name +"-"+ datetime.now().isoformat().replace(":", "-").replace(".", "-")
            time.sleep(1)
            self.ova_names.append(ova_name_temp)
    
    def __set_ovf_tool(self):
        if not self.ovf_tool_path:
            self.ovf_tool_path = self.DEFAULT_LOCAL_OVF_TOOL_PATH
        if not os.path.isfile(self.ovf_tool_path):
            LOG.error("ovftool not found.")
        return        
   
    def deploy(self):
        self.__generate_ova_names()
        self.__set_ovf_tool()

        for i in range(0, self.count):
            cmd = (
                '"%s" --X:"logFile"="./deploy_oms.log" --overwrite --powerOn --datastore=\'%s\' --noSSLVerify --acceptAllEulas --name=%s \
                --X:injectOvfEnv --X:enableHiddenProperties  --prop:root_pwd=\'%s\' --prop:permit_root_login=true --prop:auth_mode=\'%s\' \
                --prop:harbor_admin_password=\'%s\' --prop:max_job_workers=5   %s  \
                vi://%s:\'%s\'@%s/Datacenter/host/%s'
                % (self.ovf_tool_path, self.ds, self.ova_names.get(i),
                    self.ova_root_password, self.auth_mode,
                    self.harbor_password, self.ova_path,
                    self.vc_user, self.vc_password, self.vc_host, self.cluster
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

            ova_endpoint = ''
            ova_endpoint = govc_utils.getvmip(self.vc_host, self.vc_user, self.vc_password, self.ova_names.get(i))
            if ova_endpoint is not '':
                self.ova_endpoints.append(ova_endpoint) 

        return self.ova_endpoints


class OfflineDeployer(Deployer):

    def __init__(self):
        self.vm_host = '' 
        self.vm_user = '' 
        self.vm_password = '' 

    def deploy(self):
        pass