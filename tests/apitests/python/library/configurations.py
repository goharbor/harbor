# -*- coding: utf-8 -*-

import base
import swagger_client
from swagger_client.rest import ApiException

def set_configurations(client, expect_status_code = 200, expect_response_body = None, **config):
    conf = swagger_client.Configurations()

    if "project_creation_restriction" in config:
        conf.project_creation_restriction = config.get("project_creation_restriction")
    if "token_expiration" in config:
        conf.token_expiration = config.get("token_expiration")

    try:
        _, status_code, _ = client.configurations_put_with_http_info(conf)
    except ApiException as e:
        base._assert_status_code(expect_status_code, e.status)
        if expect_response_body is not None:
            base._assert_status_body(expect_response_body, e.body)
        return

    base._assert_status_code(expect_status_code, status_code)

class Configurations(base.Base):
    def get_configurations(self, item_name = None, expect_status_code = 200, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)

        try:
            data, status_code, _ = client.configurations_get_with_http_info()
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return

        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(200, status_code)

        if item_name is not None:
            return {
            'project_creation_restriction': data.project_creation_restriction.value,
            'token_expiration': data.token_expiration.value,
            }.get(item_name,'Get Configutation Error: Item name {} is not exist'.format(item_name))

        return data

    def set_configurations_of_project_creation_restriction(self, project_creation_restriction, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)

        config=dict(project_creation_restriction=project_creation_restriction)
        set_configurations(client, expect_status_code = expect_status_code, **config)

    def set_configurations_of_token_expiration(self, token_expiration, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)

        config=dict(token_expiration=token_expiration)
        set_configurations(client, expect_status_code = expect_status_code, **config)
