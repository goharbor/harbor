# -*- coding: utf-8 -*-

import time
import re
import base
import swagger_client
from swagger_client.rest import ApiException

class System(base.Base):
    def get_gc_history(self, expect_status_code = 200, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)

        try:
            data, status_code, _ = client.system_gc_get_with_http_info()
        except ApiException as e:
            if e.status == expect_status_code:
                if expect_response_body is not None and e.body.strip() != expect_response_body.strip():
                    raise Exception(r"Get configuration response body is not as expected {} actual status is {}.".format(expect_response_body.strip(), e.body.strip()))
                else:
                    return e.reason, e.body
            else:
                raise Exception(r"Get configuration result is not as expected {} actual status is {}.".format(expect_status_code, e.status))
        base._assert_status_code(expect_status_code, status_code)
        return data

    def get_gc_status_by_id(self, job_id, expect_status_code = 200, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)

        try:
            data, status_code, _ = client.system_gc_id_get_with_http_info(job_id)
        except ApiException as e:
            if e.status == expect_status_code:
                if expect_response_body is not None and e.body.strip() != expect_response_body.strip():
                    raise Exception(r"Get configuration response body is not as expected {} actual status is {}.".format(expect_response_body.strip(), e.body.strip()))
                else:
                    return e.reason, e.body
            else:
                raise Exception(r"Get configuration result is not as expected {} actual status is {}.".format(expect_status_code, e.status))
        base._assert_status_code(expect_status_code, status_code)
        return data

    def get_gc_log_by_id(self, job_id, expect_status_code = 200, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)

        try:
            data, status_code, _ = client.system_gc_id_log_get_with_http_info(job_id)
        except ApiException as e:
            if e.status == expect_status_code:
                if expect_response_body is not None and e.body.strip() != expect_response_body.strip():
                    raise Exception(r"Get configuration response body is not as expected {} actual status is {}.".format(expect_response_body.strip(), e.body.strip()))
                else:
                    return e.reason, e.body
            else:
                raise Exception(r"Get configuration result is not as expected {} actual status is {}.".format(expect_status_code, e.status))
        base._assert_status_code(expect_status_code, status_code)
        return data

    def get_gc_schedule(self, expect_status_code = 200, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)

        try:
            data, status_code, _ = client.system_gc_schedule_get_with_http_info()
        except ApiException as e:
            if e.status == expect_status_code:
                if expect_response_body is not None and e.body.strip() != expect_response_body.strip():
                    raise Exception(r"Get configuration response body is not as expected {} actual status is {}.".format(expect_response_body.strip(), e.body.strip()))
                else:
                    return e.reason, e.body
            else:
                raise Exception(r"Get configuration result is not as expected {} actual status is {}.".format(expect_status_code, e.status))
        base._assert_status_code(expect_status_code, status_code)
        return data

    def set_cve_allowlist(self, expires_at=None, expected_status_code=200, *cve_ids, **kwargs):
        client = self._get_client(**kwargs)
        cve_list = [swagger_client.CVEAllowlistItem(cve_id=c) for c in cve_ids]
        allowlist = swagger_client.CVEAllowlist(expires_at=expires_at, items=cve_list)
        try:
            r = client.system_cve_allowlist_put_with_http_info(allowlist=allowlist, _preload_content=False)
        except Exception as e:
            base._assert_status_code(expected_status_code, e.status)
        else:
            base._assert_status_code(expected_status_code, r.status)

    def get_cve_allowlist(self, **kwargs):
        client = self._get_client(**kwargs)
        return client.system_cve_allowlist_get()

    def get_project_quota(self, reference, reference_id, **kwargs):
        params={}
        params['reference'] = reference
        params['reference_id'] = reference_id

        client = self._get_client(api_type='quota', **kwargs)
        data, status_code, _ = client.list_quotas_with_http_info(**params)
        base._assert_status_code(200, status_code)
        return data
