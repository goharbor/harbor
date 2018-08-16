# -*- coding: utf-8 -*-

import sys
import base
import swagger_client

class Registry(base.Base):
    def create_registry(self, endpoint, name = base._random_name("registry"), username="", 
        password="", insecure=True, **kwargs):
        client = self._get_client(**kwargs)
        registry = swagger_client.RepTargetPost(name=name, endpoint=endpoint, 
            username=username, password=password, insecure=insecure)
        _, _, header = client.targets_post_with_http_info(registry)
        return base._get_id_from_header(header), name