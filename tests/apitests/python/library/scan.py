# -*- coding: utf-8 -*-

import time
import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException

class Scan(base.Base):
    def scan_artifact(self, project_name, repo_name, reference, expect_status_code = 202, **kwargs):
        client = self._get_client(**kwargs)
        data, status_code, _ = client.scan_artifact_with_http_info(project_name, repo_name, reference)
        base._assert_status_code(expect_status_code, status_code)
        return data
