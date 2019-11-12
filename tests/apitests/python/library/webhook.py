# -*- coding: utf-8 -*-

import base
import swagger_client


class Webhook(base.Base):
    def create_webhook_policy(self, name=None, description="", project_id=1, creator="", targets=None,
                              event_types=None, enabled=True,
                              expect_status_code=201, **kwargs):
        if event_types is None:
            event_types = ['pushImage', 'pullImage', 'deleteImage', 'uploadChart', 'deleteChart',
                           'downloadChart', 'scanningFailed', 'scanningCompleted']
        if name is None:
            name = base._random_name("wbpolicy")
        if targets is None:
            targets = [swagger_client.WebhookTargetObject(type="http", skip_cert_verify=True,
                                                          address="http://127.0.0.1:9009")]
        client = self._get_client(**kwargs)
        policy = swagger_client.WebhookPolicy(name=name, description=description, targets=targets,
                                              event_types=event_types, enabled=enabled)
        _, status_code, header = client.projects_project_id_webhook_policies_test_post_with_http_info(project_id,
                                                                                                      policy)
        base._assert_status_code(expect_status_code, status_code)
        return base._get_id_from_header(header), name

    def get_webhook_policy(self, project_id=1, policy_id=None, expect_status_code=200, **kwargs):
        client = self._get_client(**kwargs)
        data, status_code, _ = client.projects_project_id_webhook_policies_policy_id_get_with_http_info(project_id,
                                                                                                        policy_id)
        base._assert_status_code(expect_status_code, status_code)
        return data

    def check_webook_policy_should_exist(self, project_id, check_policy_id, **kwargs):
        policy_data = self.get_webhook_policy(project_id, check_policy_id, **kwargs)
        if policy_data.id != check_policy_id:
            raise Exception(
                r"Check webhook policy failed, expect <{}> actual <{}>.".format(check_policy_id, policy_data.id))
        else:
            print r"Check Webhook policy passed, policy ID <{}>.".format(check_policy_id)

    def get_webhook_last_trigger_info(self, project_id, **kwargs):
        client = self._get_client(**kwargs)
        return client.projects_project_id_webhook_lasttrigger_get_with_http_info(project_id)

    def delete_webhook_policy(self, project_id, policy_id, expect_status_code=200, **kwargs):
        client = self._get_client(**kwargs)
        _, status_code, _ = client.projects_project_id_webhook_policies_policy_id_delete_with_http_info(project_id,
                                                                                                        policy_id)
        base._assert_status_code(expect_status_code, status_code)
