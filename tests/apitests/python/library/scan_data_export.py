# -*- coding: utf-8 -*-

import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException


class Scan_data_export(base.Base):

    def __init__(self):
        super(Scan_data_export, self).__init__(api_type="scan_data_export")

    def get_scan_data_export_execution_list(self, expect_status_code=200, expect_response_body=None, **kwargs):
        try:
            return_data, status_code, _ = self._get_client(**kwargs).get_scan_data_export_execution_list_with_http_info()
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        return return_data

    def get_scan_data_export_execution(self, execution_id, expect_status_code=200, expect_response_body=None, **kwargs):
        try:
            return_data, status_code, _ = self._get_client(**kwargs).get_scan_data_export_execution_with_http_info(execution_id)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        return return_data

    def export_scan_data(self, x_scan_data_type, projects, labels=None, repositories=None, cve_ids=None, tags=None, expect_status_code=200, expect_response_body=None, **kwargs):
        criteria = v2_swagger_client.ScanDataExportRequest(projects=projects, labels=labels, repositories=repositories, cve_ids=cve_ids, tags=tags)
        try:
            return_data, status_code, _ = self._get_client(**kwargs).export_scan_data_with_http_info(x_scan_data_type, criteria)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        return return_data

    def download_scan_data(self, execution_id, expect_status_code=200, expect_response_body=None, **kwargs):
        try:
            return_data, status_code, _ = self._get_client(**kwargs).download_scan_data_with_http_info(execution_id)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        return return_data