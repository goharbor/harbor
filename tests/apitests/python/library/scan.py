# -*- coding: utf-8 -*-

import time
import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException

class Scan(base.Base, object):
    def __init__(self):
        super(Scan,self).__init__(api_type = "scan")

    def scan_artifact(self, project_name, repo_name, reference, expect_status_code = 202, expect_response_body = None, **kwargs):
        try:
            data, status_code, _ = self._get_client(**kwargs).scan_artifact_with_http_info(project_name, repo_name, reference)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return

        base._assert_status_code(expect_status_code, status_code)

        return data

    def sbom_generation_of_artifact(self, project_name, repo_name, reference, expect_status_code = 202, expect_response_body = None, **kwargs):
        try:
            req_param = dict(scan_type = {"scan_type":"sbom"})
            data, status_code, _ = self._get_client(**kwargs).scan_artifact_with_http_info(project_name, repo_name, reference, **req_param)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return

        base._assert_status_code(expect_status_code, status_code)

        return data

