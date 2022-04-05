# -*- coding: utf-8 -*-

import time
import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException

class StopScan(base.Base, object):
    def __init__(self):
        super(StopScan,self).__init__(api_type = "scan")

    def stop_scan_artifact(self, project_name, repo_name, reference, expect_status_code = 202, expect_response_body = None, **kwargs):
        try:
            data, status_code, _ = self._get_client(**kwargs).stop_scan_artifact_with_http_info(project_name, repo_name, reference)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return

        base._assert_status_code(expect_status_code, status_code)

        return data