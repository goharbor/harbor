# -*- coding: utf-8 -*-

from swagger_client.rest import ApiException
from v2_swagger_client.rest import ApiException

import base


def set_configurations(client, expect_status_code = 200, expect_response_body = None, **config):
    conf = {}

    if "project_creation_restriction" in config and config.get("project_creation_restriction") is not None:
        conf["project_creation_restriction"] = config.get("project_creation_restriction")
    if "token_expiration" in config and config.get("token_expiration") is not None:
        conf["token_expiration"] = config.get("token_expiration")
    if "ldap_filter" in config and config.get("ldap_filter") is not None:
        conf["ldap_filter"] = config.get("ldap_filter")
    if "ldap_group_attribute_name" in config and config.get("ldap_group_attribute_name") is not None:
        conf["ldap_group_attribute_name"] = config.get("ldap_group_attribute_name")
    if "ldap_group_base_dn" in config:
        conf["ldap_group_base_dn"] = config.get("ldap_group_base_dn")
    if "ldap_group_search_filter" in config and config.get("ldap_group_search_filter") is not None:
        conf["ldap_group_search_filter"] = config.get("ldap_group_search_filter")
    if "ldap_group_search_scope" in config and config.get("ldap_group_search_scope") is not None:
        conf["ldap_group_search_scope"] = config.get("ldap_group_search_scope")
    if "ldap_group_admin_dn" in config and config.get("ldap_group_admin_dn") is not None:
        conf["ldap_group_admin_dn"] = config.get("ldap_group_admin_dn")

    try:
        _, status_code, _ = client.update_configurations_with_http_info(conf)
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

