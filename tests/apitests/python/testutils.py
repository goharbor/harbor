import os
import sys
import warnings
from functools import wraps

sys.path.insert(0, os.environ.get("SWAGGER_CLIENT_PATH", ''))
path=os.getcwd() + "/library"
sys.path.insert(0, path)

path=os.getcwd() + "/tests/apitests/python/library"
sys.path.insert(0, path)
path=os.getcwd() + "/tests/apitests/python/"
sys.path.insert(0, path)
print(sys.path)

files_directory = os.getcwd() + "/tests/files/"

import v2_swagger_client
import swagger_client.models

admin_user = "admin"
admin_pwd = "Harbor12345"

harbor_server = os.environ.get("HARBOR_HOST", '')
harbor_url = "https://{}".format(harbor_server)
#CLIENT=dict(endpoint="https://"+harbor_server+"/api")
ADMIN_CLIENT=dict(endpoint = os.environ.get("HARBOR_HOST_SCHEMA", "https")+ "://"+harbor_server+"/api/v2.0", username = admin_user, password =  admin_pwd)
CHART_API_CLIENT=dict(endpoint = os.environ.get("HARBOR_HOST_SCHEMA", "https")+ "://"+harbor_server+"/api", username = admin_user, password =  admin_pwd)
USER_ROLE=dict(admin=0,normal=1)
TEARDOWN = os.environ.get('TEARDOWN', 'true').lower() in ('true', 'yes')
notary_url = os.environ.get('NOTARY_URL', 'https://'+harbor_server+':4443')
DOCKER_USER = os.environ.get('DOCKER_USER', '')
DOCKER_PWD = os.environ.get('DOCKER_PWD', '')
METRIC_URL = os.environ.get('METRIC_URL', 'http://'+harbor_server+':9090')
BASE_IMAGE = dict(name='busybox', tag='latest')
BASE_IMAGE_ABS_PATH_NAME = '/' + BASE_IMAGE['name'] + '.tar'

def GetProductApi(username, password, harbor_server= os.environ.get("HARBOR_HOST", '')):

    cfg = swagger_client.Configuration()
    cfg.host = "https://"+harbor_server+"/api/v2.0"
    cfg.username = username
    cfg.password = password
    cfg.verify_ssl = False
    cfg.debug = True
    api_client = swagger_client.ApiClient(cfg)
    api_instance = swagger_client.ProductsApi(api_client)
    return api_instance

def GetRepositoryApi(username, password, harbor_server= os.environ.get("HARBOR_HOST", '')):

    cfg = v2_swagger_client.Configuration()
    cfg.host = "https://"+harbor_server+"/api/v2.0"
    cfg.username = username
    cfg.password = password
    cfg.verify_ssl = False
    cfg.debug = True
    api_client = v2_swagger_client.ApiClient(cfg)
    api_instance = v2_swagger_client.RepositoryApi(api_client)
    return api_instance

def GetUserGroupApi(username, password, harbor_server= os.environ.get("HARBOR_HOST", '')):
    cfg = v2_swagger_client.Configuration()
    cfg.host = "https://"+harbor_server+"/api/v2.0"
    cfg.username = username
    cfg.password = password
    cfg.verify_ssl = False
    cfg.debug = True
    api_client = v2_swagger_client.ApiClient(cfg)
    api_instance = v2_swagger_client.UsergroupApi(api_client)
    return api_instance

class TestResult(object):
    def __init__(self):
        self.num_errors = 0
        self.error_message = []
    def add_test_result(self, error_message):
        self.num_errors = self.num_errors + 1
        self.error_message.append(error_message)
    def get_final_result(self):
        if self.num_errors > 0:
            for each_err_msg in self.error_message:
                print("Error message:", each_err_msg)
            raise Exception(r"Test case failed with {} errors.".format(self.num_errors))

def suppress_urllib3_warning(func):
    @wraps(func)
    def inner_func(*args):
        warnings.filterwarnings(action="ignore",
                                message="unclosed",
                                category=ResourceWarning)
        warnings.filterwarnings(action='ignore', message='Unverified HTTPS request')
        func(*args)
    return inner_func

from contextlib import contextmanager

@contextmanager
def created_user(password, _teardown=None):
    from library.user import User

    api = User()
    is_teardown = TEARDOWN
    if _teardown in (True, False):
        is_teardown = _teardown

    user_id, user_name = api.create_user(user_password=password, **ADMIN_CLIENT)
    try:
        yield (user_id, user_name)
    finally:
        if is_teardown:
            api.delete_user(user_id, **ADMIN_CLIENT)

@contextmanager
def created_project(name=None, metadata=None, user_id=None, member_role_id=None, _teardown=None):
    from library.project import Project

    api = Project()

    is_teardown = TEARDOWN
    if _teardown in (True, False):
        is_teardown = _teardown

    project_id, project_name = api.create_project(name=name, metadata=metadata, **ADMIN_CLIENT)
    if user_id:
        api.add_project_members(project_id, user_id=user_id, member_role_id=member_role_id, **ADMIN_CLIENT)

    try:
        yield (project_id, project_name)
    finally:
        if is_teardown:
            api.delete_project(project_id, **ADMIN_CLIENT)
