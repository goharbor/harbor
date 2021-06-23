# -*- coding: utf-8 -*-
import os
import subprocess
import time

import client
import swagger_client
import v2_swagger_client

try:
    from urllib import getproxies
except ImportError:
    from urllib.request import getproxies

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

def _create_client(server, credential, debug, api_type="products"):
    cfg = None
    if api_type in ('projectv2', 'artifact', 'repository', 'scanner', 'scan', 'scanall', 'preheat', 'quota',
                    'replication', 'registry', 'robot', 'gc', 'retention', 'immutable', 'system_cve_allowlist',
                    'configure', 'user', 'member', 'health', 'label', 'webhook'):
        cfg = v2_swagger_client.Configuration()
    else:
        cfg = swagger_client.Configuration()

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
        "chart": client.ChartRepositoryApi(client.ApiClient(cfg)),
        "products": swagger_client.ProductsApi(swagger_client.ApiClient(cfg)),
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
        "webhook": v2_swagger_client.WebhookApi(v2_swagger_client.ApiClient(cfg))
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

def restart_process(process):
    if process == "dockerd":
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
    print("Command: ", command)

    try:
        proc = subprocess.Popen(command, universal_newlines=True, shell=True,
                            stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
        output, errors = proc.communicate()
    except Exception as e:
        print("Run command caught exception:", e)
        output = None
    else:
        print(proc.returncode, errors, output)
    finally:
        proc.stdout.close()
        print("output: ", output)
        return output

def run_command(command, expected_error_message = None):
    print("Command: ", subprocess.list2cmdline(command))
    try:
        output = subprocess.check_output(command,
                                         stderr=subprocess.STDOUT,
                                         universal_newlines=True)
    except subprocess.CalledProcessError as e:
        print("Run command error:", str(e))
        print("expected_error_message:", expected_error_message)
        if expected_error_message is not None:
            if str(e.output).lower().find(expected_error_message.lower()) < 0:
                raise Exception(r"Error message {} is not as expected {}".format(str(e.output), expected_error_message))
        else:
            raise Exception('Error: Exited with error code: %s. Output:%s'% (e.returncode, e.output))
    else:
        print("output:", output)
        return output

class Base(object):
    def __init__(self, server=None, credential=None, debug=True, api_type="products"):
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
