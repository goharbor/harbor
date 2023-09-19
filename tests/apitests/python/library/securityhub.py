# -*- coding: utf-8 -*-

import base
from v2_swagger_client.rest import ApiException
from urllib.parse import quote


class SecurityHub(base.Base):


    def __init__(self):
        super(SecurityHub, self).__init__(api_type="securityhub")


    def get_security_summary(self, with_dangerous_cve=True, with_dangerous_artifact=True, expect_status_code=200, expect_response_body=None, **kwargs):
        try:
            return_data, status_code, _ = self._get_client(**kwargs).get_security_summary_with_http_info(with_dangerous_cve=with_dangerous_cve, with_dangerous_artifact=with_dangerous_artifact)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        return return_data


    def list_vulnerabilities(self, q=None, tune_count=True, with_tag=True, page=1, page_size=10, expect_status_code=200, expect_response_body=None, **kwargs):
        try:
            if q is not None:
                q = quote(q)
                return_data, status_code, _ = self._get_client(**kwargs).list_vulnerabilities_with_http_info(q=q, tune_count=tune_count, with_tag=with_tag, page=page, page_size=page_size)
            else:
                return_data, status_code, _ = self._get_client(**kwargs).list_vulnerabilities_with_http_info(tune_count=tune_count, with_tag=with_tag, page=page, page_size=page_size)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        return return_data
