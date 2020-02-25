# -*- coding: utf-8 -*-

import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException

class Artifact(base.Base):
    def get_reference_info(self, project_name, repo_name, reference, **kwargs):
        client = self._get_client(**kwargs)
        return client.get_artifact_with_http_info(project_name, repo_name, reference)

    def copy_artifact(self, project_name, repo_name, _from, expect_status_code = 201, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)

        try:
            data, status_code, _ = client.copy_artifact_with_http_info(project_name, repo_name, _from)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return

        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(201, status_code)
        return data
