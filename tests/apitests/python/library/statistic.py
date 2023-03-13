# -*- coding: utf-8 -*-

import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException


class Statistic(base.Base):

    def __init__(self):
        super(Statistic, self).__init__(api_type="statistic")

    def get_statistic(self, expect_status_code=200, expect_response_body=None, **kwargs):
        try:
            return_data, status_code, _ = self._get_client(**kwargs).get_statistic_with_http_info()
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        return return_data