# -*- coding: utf-8 -*-

import sys
import base
import swagger_client

class Registry(base.Base):
    def create_registry(self, endpoint, name=None, username="", 
        password="", insecure=True, **kwargs):
        if name is None:
            name = base._random_name("registry")
        client = self._get_client(**kwargs)
        registry = swagger_client.RepTargetPost(name=name, endpoint=endpoint, 
            username=username, password=password, insecure=insecure)
        _, _, header = client.targets_post_with_http_info(registry)
        return base._get_id_from_header(header), name

    def get_registry_id_by_endpoint(self, endpoint, **kwargs):
        client = self._get_client(**kwargs)
        registries = client.targets_get()
        for registry in registries or []:
            if registry.endpoint == endpoint:
                return registry.id
        raise Exception("registry %s not found" % endpoint)