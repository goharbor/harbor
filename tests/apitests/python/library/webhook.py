# -*- coding: utf-8 -*-

import time
import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException

class Webhook(base.Base):
    def __init__(self):
        super(Webhook,self).__init__(api_type="webhook")

    def create_webhook(self, project_id, targets, event_types = ["DELETE_ARTIFACT",
            "PULL_ARTIFACT",
            "PUSH_ARTIFACT",
            "DELETE_CHART",
            "DOWNLOAD_CHART",
            "UPLOAD_CHART","QUOTA_EXCEED",
            "QUOTA_WARNING","SCANNING_FAILED",
            "TAG_RETENTION"],
            name = None, desc = None, enabled = True,
            expect_status_code = 201, expect_response_body = None, **kwargs):

        if name is None:
            name = base._random_name("webhook") + (str(project_id))
        if desc is None:
            desc = base._random_name("webhook desc") + (str(project_id))
        policy = v2_swagger_client.WebhookPolicy(
            name = name,
            description = desc,
            project_id = project_id,
            targets = targets,
            event_types = event_types,
            enabled = enabled
        )

        try:
            _, status_code, header = self._get_client(**kwargs).create_webhook_policy_of_project_with_http_info(project_id, policy)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(201, status_code)
        return base._get_id_from_header(header), name

    def update_webhook(self, project_id, webhook_policy_id, event_types = None,
            name = None, desc = None, enabled = None, targets = None,
            expect_status_code=200, expect_response_body=None, **kwargs):

        policy = v2_swagger_client.WebhookPolicy()
        if name is not None:
            policy.name = name
        if desc is not None:
            policy.desc = desc
        if enabled is not None:
            policy.enabled = enabled
        if targets is not None:
            policy.targets = targets
        if event_types is not None:
            policy.event_types = event_types

        try:
            _, status_code, header = self._get_client(**kwargs).update_webhook_policy_of_project_with_http_info(project_id, webhook_policy_id, policy)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(200, status_code)

    def get_webhook(self, project_id, webhook_policy_id, expect_status_code=200, expect_response_body=None, **kwargs):

        try:
            data , status_code, header = self._get_client(**kwargs).get_webhook_policy_of_project_with_http_info(project_id, webhook_policy_id)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(200, status_code)
        print("Webhooks:", data)
        return data

    def delete_webhook(self, project_id, webhook_policy_id, expect_status_code=200, expect_response_body=None, **kwargs):

        try:
            _ , status_code, _ = self._get_client(**kwargs).delete_webhook_policy_of_project_with_http_info(project_id, webhook_policy_id)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(200, status_code)



