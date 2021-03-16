# -*- coding: utf-8 -*-

import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException

class SystemCVEAllowlist(base.Base, object):
    def __init__(self):
        super(SystemCVEAllowlist, self).__init__(api_type = "system_cve_allowlist")

    def set_cve_allowlist(self, expires_at=None, expected_status_code=200, *cve_ids, **kwargs):
        cve_list = [v2_swagger_client.CVEAllowlistItem(cve_id=c) for c in cve_ids]
        allowlist = v2_swagger_client.CVEAllowlist(expires_at=expires_at, items=cve_list)
        try:
            r = self._get_client(**kwargs).put_system_cve_allowlist_with_http_info(allowlist=allowlist, _preload_content=False)
        except ApiException as e:
            base._assert_status_code(expected_status_code, e.status)
        else:
            base._assert_status_code(expected_status_code, r.status)

    def get_cve_allowlist(self, **kwargs):
        return self._get_client(**kwargs).get_system_cve_allowlist()