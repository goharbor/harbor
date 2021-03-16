# -*- coding: utf-8 -*-

import time
import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException

class Preheat(base.Base, object):
    def __init__(self):
        super(Preheat,self).__init__(api_type = "preheat")

    def create_instance(self, name = None, description="It's a dragonfly instance", vendor="dragonfly",
                        endpoint_url="http://20.32.244.16", auth_mode="NONE", enabled=True, insecure=True,
                        expect_status_code = 201, expect_response_body = None, **kwargs):
        if name is None:
            name = base._random_name("instance")
        instance = v2_swagger_client.Instance(name=name, description=description,vendor=vendor,
            endpoint=endpoint_url, auth_mode=auth_mode, enabled=enabled)
        print("instance:",instance)
        try:
            _, status_code, header = self._get_client(**kwargs).create_instance_with_http_info(instance)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(201, status_code)
        return base._get_id_from_header(header), name

    def create_policy(self, project_name, project_id, provider_id, name = None, description="It's a dragonfly policy",
                        filters=r'[{"type":"repository","value":"re*"},{"type":"tag","value":"v1.0*"}]', trigger=r'{"type":"manual","trigger_setting":{"cron":""}}', enabled=True,
                        expect_status_code = 201, expect_response_body = None, **kwargs):
        if name is None:
            name = base._random_name("policy")
        policy = v2_swagger_client.PreheatPolicy(name=name, project_id=project_id, provider_id=provider_id,
                                                   description=description,filters=filters,
                                                   trigger=trigger, enabled=enabled)
        print("policy:",policy)
        try:
            data, status_code, header = self._get_client(**kwargs).create_policy_with_http_info(project_name, policy)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(201, status_code)
        return base._get_id_from_header(header), name

    def get_instance(self, instance_name, **kwargs):
        return self._get_client(**kwargs).get_instance(instance_name)

    def get_policy(self, project_name, preheat_policy_name, **kwargs):
        return self._get_client(**kwargs).get_policy(project_name, preheat_policy_name)

    def update_policy(self, project_name, preheat_policy_name, policy, **kwargs):
        return self._get_client(**kwargs).update_policy(project_name, preheat_policy_name, policy)

    def delete_instance(self, preheat_instance_name, expect_status_code = 200, expect_response_body = None, **kwargs):
        try:
            _, status_code, header = _, status_code, _ = self._get_client(**kwargs).delete_instance_with_http_info(preheat_instance_name)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
        else:
            base._assert_status_code(expect_status_code, status_code)
            base._assert_status_code(200, status_code)
