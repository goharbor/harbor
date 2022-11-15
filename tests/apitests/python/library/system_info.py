# -*- coding: utf-8 -*-

import base
from v2_swagger_client.rest import ApiException


class System_info(base.Base):

    def __init__(self):
        super(System_info, self).__init__(api_type="system_info")

    def get_system_info(self, expect_status_code=200, expect_response_body=None, **kwargs):
        try:
            return_data, status_code, _ = self._get_client(**kwargs).get_system_info_with_http_info()
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        return return_data