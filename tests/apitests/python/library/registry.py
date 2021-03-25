# -*- coding: utf-8 -*-

import base
import v2_swagger_client

class Registry(base.Base, object):
    def __init__(self):
        super(Registry,self).__init__(api_type = "registry")

    def create_registry(self, url, registry_type= "harbor", description="", credentialType = "basic",
            access_key = "admin", access_secret = "Harbor12345", name=base._random_name("registry"),
            insecure=True, expect_status_code = 201, **kwargs):

        client = self._get_client(**kwargs)
        registryCredential = v2_swagger_client.RegistryCredential(type=credentialType, access_key=access_key, access_secret=access_secret)
        registry = v2_swagger_client.Registry(name=name, url=url,
                                           description= description, type=registry_type,
                                           insecure=insecure, credential=registryCredential)
        print("registry:", registry)
        _, status_code, header = client.create_registry_with_http_info(registry)
        base._assert_status_code(expect_status_code, status_code)
        return base._get_id_from_header(header), _

    def get_registry_id_by_endpoint(self, endpoint, **kwargs):
        client = self._get_client(**kwargs)
        registries = client.list_registries()
        for registry in registries or []:
            if registry.endpoint == endpoint:
                return registry.id
        raise Exception("registry %s not found" % endpoint)

    def delete_registry(self, registry_id, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        _, status_code, _  = client.delete_registry_with_http_info(registry_id)
        base._assert_status_code(expect_status_code, status_code)