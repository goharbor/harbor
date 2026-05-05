# -*- coding: utf-8 -*-
import os
import subprocess
import time

import importlib

try:
    from urllib import getproxies
except ImportError:
    from urllib.request import getproxies

def swagger_module():
        module = importlib.import_module("v2_swagger_client")
        return module

class Server:
    def __init__(self, endpoint, verify_ssl):
        self.endpoint = endpoint
        self.verify_ssl = verify_ssl

class Credential:
    def __init__(self, type, username, password):
        self.type = type
        self.username = username
        self.password = password

def get_endpoint():
    harbor_server = os.environ.get("HARBOR_HOST", "localhost:8080")
    return os.environ.get("HARBOR_HOST_SCHEMA", "https")+ "://"+harbor_server+"/api/v2.0"

def _create_client(server, credential, debug, api_type):
    v2_swagger_client = swagger_module()
    cfg = v2_swagger_client.Configuration()
    cfg.host = server.endpoint
    cfg.verify_ssl = server.verify_ssl
    # support basic auth only for now
    cfg.username = credential.username
    cfg.password = credential.password
    cfg.debug = debug

    proxies = getproxies()
    proxy = proxies.get('http', proxies.get('all', None))
    if proxy:
        cfg.proxy = proxy

    if cfg.username is None and cfg.password is None:
        # returns {} for auth_settings for anonymous access
        import types
        cfg.auth_settings = types.MethodType(lambda self: {}, cfg)

    return {
        "projectv2":v2_swagger_client.ProjectApi(v2_swagger_client.ApiClient(cfg)),
        "artifact": v2_swagger_client.ArtifactApi(v2_swagger_client.ApiClient(cfg)),
        "preheat": v2_swagger_client.PreheatApi(v2_swagger_client.ApiClient(cfg)),
        "quota": v2_swagger_client.QuotaApi(v2_swagger_client.ApiClient(cfg)),
        "repository": v2_swagger_client.RepositoryApi(v2_swagger_client.ApiClient(cfg)),
        "scan": v2_swagger_client.ScanApi(v2_swagger_client.ApiClient(cfg)),
        "scanall": v2_swagger_client.ScanAllApi(v2_swagger_client.ApiClient(cfg)),
        "scanner": v2_swagger_client.ScannerApi(v2_swagger_client.ApiClient(cfg)),
        "replication": v2_swagger_client.ReplicationApi(v2_swagger_client.ApiClient(cfg)),
        "registry": v2_swagger_client.RegistryApi(v2_swagger_client.ApiClient(cfg)),
        "robot": v2_swagger_client.RobotApi(v2_swagger_client.ApiClient(cfg)),
        "gc": v2_swagger_client.GcApi(v2_swagger_client.ApiClient(cfg)),
        "retention": v2_swagger_client.RetentionApi(v2_swagger_client.ApiClient(cfg)),
        "immutable": v2_swagger_client.ImmutableApi(v2_swagger_client.ApiClient(cfg)),
        "system_cve_allowlist": v2_swagger_client.SystemCVEAllowlistApi(v2_swagger_client.ApiClient(cfg)),
        "configure": v2_swagger_client.ConfigureApi(v2_swagger_client.ApiClient(cfg)),
        "label": v2_swagger_client.LabelApi(v2_swagger_client.ApiClient(cfg)),
        "user": v2_swagger_client.UserApi(v2_swagger_client.ApiClient(cfg)),
        "member": v2_swagger_client.MemberApi(v2_swagger_client.ApiClient(cfg)),
        "health": v2_swagger_client.HealthApi(v2_swagger_client.ApiClient(cfg)),
        "webhook": v2_swagger_client.WebhookApi(v2_swagger_client.ApiClient(cfg)),
        "purge": v2_swagger_client.PurgeApi(v2_swagger_client.ApiClient(cfg)),
        "audit_log": v2_swagger_client.AuditlogApi(v2_swagger_client.ApiClient(cfg)),
        "scan_data_export": v2_swagger_client.ScanDataExportApi(v2_swagger_client.ApiClient(cfg)),
        "statistic": v2_swagger_client.StatisticApi(v2_swagger_client.ApiClient(cfg)),
        "system_info": v2_swagger_client.SysteminfoApi(v2_swagger_client.ApiClient(cfg)),
        "jobservice": v2_swagger_client.JobserviceApi(v2_swagger_client.ApiClient(cfg)),
        "schedule": v2_swagger_client.ScheduleApi(v2_swagger_client.ApiClient(cfg)),
        "securityhub": v2_swagger_client.SecurityhubApi(v2_swagger_client.ApiClient(cfg)),
    }.get(api_type,'Error: Wrong API type')

def _assert_status_code(expect_code, return_code, err_msg = r"HTTPS status code s not as we expected. Expected {}, while actual HTTPS status code is {}."):
    if str(return_code) != str(expect_code):
        raise Exception(err_msg.format(expect_code, return_code))

def _assert_status_body(expect_status_body, returned_status_body):
    if str(returned_status_body.strip()).lower().find(expect_status_body.lower()) < 0:
        raise Exception(r"HTTPS status body s not as we expected. Expected {}, while actual HTTPS status body is {}.".format(expect_status_body, returned_status_body))

def _random_name(prefix):
    return "%s-%d" % (prefix, int(round(time.time() * 1000)))

def _get_id_from_header(header):
    try:
        location = header["Location"]
        return int(location.split("/")[-1])
    except Exception:
        return None

def _get_string_from_unicode(udata):
    result=''
    for u in udata:
        tmp = u.encode('utf8')
        result = result + tmp.strip('\n\r\t')
    return result

def getenv_bool(name: str, default: bool = False) -> bool:
    val = os.getenv(name)
    if val is None:
        return default
    return val.strip().lower() in ( "true", "1")

def restart_process(process):
    if "dockerd" in process:
        full_process_name = process
    elif process == "containerd":
        full_process_name = "/usr/local/bin/containerd"
    else:
        raise Exception("Please input dockerd or containerd for process retarting.")
    run_command_with_popen("ps aux |grep " + full_process_name)
    for i in range(10):
        pid = run_command_with_popen(["pidof " + full_process_name])
        if pid in [None, ""]:
            break
        run_command_with_popen(["kill " + str(pid)])
        time.sleep(3)

    run_command_with_popen("ps aux |grep " + full_process_name)
    run_command_with_popen("rm -rf /var/lib/" + process + "/*")
    run_command_with_popen(full_process_name + " > ./daemon-local.log 2>&1 &")
    time.sleep(3)
    pid = run_command_with_popen(["pidof " + full_process_name])
    if pid in [None, ""]:
        raise Exception("Failed to start process {}.".format(full_process_name))
    run_command_with_popen("ps aux |grep " + full_process_name)

def run_command_with_popen(command):
    try:
        proc = subprocess.Popen(command, universal_newlines=True, shell=True,
                            stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
        output, errors = proc.communicate()
    except Exception as e:
        output = None
    finally:
        proc.stdout.close()
        return output

def run_command(command, expected_error_message = None):
    try:
        output = subprocess.check_output(command,
                                         stderr=subprocess.STDOUT,
                                         universal_newlines=True)
    except subprocess.CalledProcessError as e:
        if expected_error_message is not None:
            if str(e.output).lower().find(expected_error_message.lower()) < 0:
                raise Exception(r"Error message is not as expected {}".format(expected_error_message))
        else:
            raise Exception('Error: Exited with error code: %s, error message: %s' % (e.returncode, e.output))
    else:
        return output

class Base(object):
    def __init__(self, server=None, credential=None, debug=True, api_type=""):
        if server is None:
            server = Server(endpoint=get_endpoint(), verify_ssl=False)
        if not isinstance(server.verify_ssl, bool):
            server.verify_ssl = server.verify_ssl == "True"

        if credential is None:
            credential = Credential(type="basic_auth", username="admin", password="Harbor12345")

        self.server = server
        self.credential = credential
        self.debug = debug
        self.api_type = api_type
        self.client = _create_client(server, credential, debug, api_type=api_type)

    def _get_client(self, **kwargs):
        if len(kwargs) == 0:
            return self.client

        server = self.server
        if "endpoint" in kwargs:
            server.endpoint = kwargs.get("endpoint")
        if "verify_ssl" in kwargs:
            if not isinstance(kwargs.get("verify_ssl"), bool):
                server.verify_ssl = kwargs.get("verify_ssl") == "True"
            else:
                server.verify_ssl = kwargs.get("verify_ssl")

        credential = Credential(
            kwargs.get("type", self.credential.type),
            kwargs.get("username", self.credential.username),
            kwargs.get("password", self.credential.password),
        )

        return _create_client(server, credential, self.debug, kwargs.get('api_type', self.api_type))
