# -*- coding: utf-8 -*-

import time
import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException
from library.base import _assert_status_code

class ProjectV2(base.Base, object):
    def __init__(self):
        super(ProjectV2,self).__init__(api_type = "projectv2")

    def get_project_log(self, project_name, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        body, status_code, _ = client.get_logs_with_http_info(project_name)
        base._assert_status_code(expect_status_code, status_code)
        return body

    def filter_project_logs(self, project_name, operator, resource, resource_type, operation, **kwargs):
        access_logs = self.get_project_log(project_name, **kwargs)
        count = 0
        for each_access_log in list(access_logs):
            if each_access_log.username == operator and \
                    each_access_log.resource_type == resource_type and \
                    each_access_log.resource == resource and \
                    each_access_log.operation == operation:
                count = count + 1
        return count

    def query_user_logs(self, project_name, status_code=200, **kwargs):
        try:
            logs = self.get_project_log(project_name, expect_status_code=status_code, **kwargs)
            count = 0
            for log in list(logs):
                count = count + 1
            return count
        except ApiException as e:
            _assert_status_code(status_code, e.status)
            return 0

