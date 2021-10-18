# -*- coding: utf-8 -*-

import time
import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException

class StopScanAll(base.Base):
    def __init__(self):
        super(StopScanAll,self).__init__(api_type="scanall")

    def stop_scan_all(self, expect_status_code=202, expect_response_body=None, **kwargs):
        try:
            _, status_code, _ = self._get_client(**kwargs).stop_scan_all_with_http_info()
        except ApiException as e:
            if e.status == expect_status_code:
                if expect_response_body is not None and e.body.strip() != expect_response_body.strip():
                    raise Exception(r"Stop scan all response body is not as expected {} actual status is {}.".format(expect_response_body.strip(), e.body.strip()))
                else:
                    return e.reason, e.body
            else:
                raise Exception(r"Stop scan all result is not as expected {} actual status is {}.".format(expect_status_code, e.status))
        base._assert_status_code(expect_status_code, status_code)
