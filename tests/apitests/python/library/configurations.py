# -*- coding: utf-8 -*-

from v2_swagger_client.rest import ApiException

import base
import json


def set_configurations(client, expect_status_code = 200, expect_response_body = None, **config):
    try:
        _, status_code, _ = client.update_configurations_with_http_info(config)
    except ApiException as e:
        base._assert_status_code(expect_status_code, e.status)
        if expect_response_body is not None:
            base._assert_status_body(expect_response_body, e.body)
        return
    base._assert_status_code(expect_status_code, status_code)

class Configurations(base.Base, object):
    def __init__(self):
        super(Configurations,self).__init__(api_type = "configure")

    def get_configurations(self, item_name = None, expect_status_code = 200, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)

        try:
            data, status_code, _ = client.get_configurations_with_http_info()
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

    def set_configurations_of_ldap(self, ldap_filter=None, ldap_group_attribute_name=None,
            ldap_group_base_dn=None, ldap_group_search_filter=None, ldap_group_search_scope=None, ldap_group_admin_dn=None, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        config=dict(ldap_filter=ldap_filter, ldap_group_attribute_name=ldap_group_attribute_name,
                           ldap_group_base_dn=ldap_group_base_dn, ldap_group_search_filter=ldap_group_search_filter, ldap_group_admin_dn=ldap_group_admin_dn, ldap_group_search_scope=ldap_group_search_scope)
        set_configurations(client, expect_status_code = expect_status_code, **config)

    def set_configurations_of_audit_log_forword(self, audit_log_forward_endpoint=None, skip_audit_log_database=None, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        config=dict(audit_log_forward_endpoint=audit_log_forward_endpoint, skip_audit_log_database=skip_audit_log_database)
        set_configurations(client, expect_status_code = expect_status_code, **config)

    def set_configurations_of_retain_image_last_pull_time(self, is_skip, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        config=dict(scanner_skip_update_pulltime=is_skip)
        set_configurations(client, expect_status_code = expect_status_code, **config)

    def set_configurations_of_banner_message(self, message, message_type=None, closable=None, from_date=None, to_date=None, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        banner_message = None
        if message == "":
            banner_message = ""
        else:
            banner_message = {
                "message": message,
                "type": message_type,
                "closable": closable,
                "fromDate": from_date,
                "toDate": to_date
            }
            banner_message = json.dumps(banner_message)
        config=dict(banner_message=banner_message)
        set_configurations(client, expect_status_code = expect_status_code, **config)
