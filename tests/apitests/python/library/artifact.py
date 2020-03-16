# -*- coding: utf-8 -*-

import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException

class Artifact(base.Base):
    def get_reference_info(self, project_name, repo_name, reference, **kwargs):
        client = self._get_client(**kwargs)
        params = {}
        if "with_signature" in kwargs:
            params["with_signature"] = kwargs["with_signature"]
        if "with_tag" in kwargs:
            params["with_tag"] = kwargs["with_tag"]
        return client.get_artifact_with_http_info(project_name, repo_name, reference, **params )

    def add_label_to_reference(self, project_name, repo_name, reference, label_id, **kwargs):
        client = self._get_client(**kwargs)
        label = v2_swagger_client.Label(id = label_id)
        return client.add_label_with_http_info(project_name, repo_name, reference, label)

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

    def create_tag(self, project_name, repo_name, reference, tag_name, expect_status_code = 201, **kwargs):
        client = self._get_client(**kwargs)
        tag = v2_swagger_client.Tag(name = tag_name)
        _, status_code, _ = client.create_tag_with_http_info(project_name, repo_name, reference, tag)
        base._assert_status_code(expect_status_code, status_code)

    def delete_tag(self, project_name, repo_name, reference, tag_name, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        _, status_code, _ = client.delete_tag_with_http_info(project_name, repo_name, reference, tag_name)
        base._assert_status_code(expect_status_code, status_code)
