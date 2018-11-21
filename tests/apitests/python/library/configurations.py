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
    caught_api_exception = False
    try:
        _, status_code, _ = client.configurations_put_with_http_info(conf)
    except ApiException as e:
        caught_api_exception = True
        if e.status == expect_status_code:
            if expect_response_body is not None and e.body.strip() != expect_response_body.strip():
                raise Exception(r"Set configuration response body is not as we expected. Expected {}, while actual response body is {}.".format(expect_response_body.strip(), e.body.strip()))
            else:
                return e.reason, e.body
        else:
            raise Exception(r"Set configuration status code is not as we expected {}, while actual status code is {}.".format(expect_status_code, e.status))
    if expect_status_code != 200 and caught_api_exception == False:
        raise Exception(r"Failed to catch error {} when set configurations.".format(expect_status_code))

    base._assert_status_code(expect_status_code, status_code)

class Configurations(base.Base):
    def get_configurations(self, item_name = None, expect_status_code = 200, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)
        HasError = False
        try:
            data, status_code, _ = client.configurations_get_with_http_info()
        except ApiException as e:
            HasError = True
            if e.status == expect_status_code:
                if expect_response_body is not None and e.body.strip() != expect_response_body.strip():
                    raise Exception(r"Set configuration response body is not as we expected. Expected {}, while actual response body is {}..".format(expect_response_body.strip(), e.body.strip()))
                else:
                    return e.reason, e.body
            else:
                raise Exception(r"Set configuration status code is not as we expected {}, while actual status code is {}.".format(expect_status_code, e.status))
        if expect_status_code != 200 and HasError == False:
            raise Exception(r"Failed to catch error {} when get configurations.".format(expect_status_code))

        base._assert_status_code(expect_status_code, status_code)

        if item_name is not None:
            return {
            'project_creation_restriction': data.project_creation_restriction.value,
            'token_expiration': data.token_expiration.value,
            }.get(item_name,'error')
        return data

    def set_configurations_of_project_creation_restriction_success(self, project_creation_restriction, **kwargs):
        client = self._get_client(**kwargs)
        config=dict(project_creation_restriction=project_creation_restriction)
        set_configurations(client, **config)
        item_value = self.get_configurations(item_name = "project_creation_restriction", **kwargs)
        if item_value != project_creation_restriction:
            raise Exception("Failed to set system configuration item {} to value {},\
                actual value is {}".format("project_creation_restriction", project_creation_restriction, item_value))

    def set_configurations_of_token_expiration(self, token_expiration, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        config=dict(token_expiration=token_expiration)
        set_configurations(client, expect_status_code = expect_status_code, **config)
