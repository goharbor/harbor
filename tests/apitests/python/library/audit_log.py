# -*- coding: utf-8 -*-

import base
from v2_swagger_client.rest import ApiException


class Audit_Log(base.Base):

    def __init__(self):
        super(Audit_Log, self).__init__(api_type="audit_log")

    def get_latest_audit_log(self):
        return self.list_audit_logs(sort="-creation_time", page_size=1, page=1)[0]

    def list_audit_logs(self, sort, page_size, page, expect_status_code=200, expect_response_body=None, **kwargs):
        try:
            return_data, status_code, _ = self._get_client(**kwargs).list_audit_logs_with_http_info(sort=sort, page_size=page_size, page=page)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        return return_data