# -*- coding: utf-8 -*-

import time
import base
import swagger_client

class Target(base.Base):
    def create_target(self,
        endpoint_target = None,
            username_target = "target_user", password_target = "Aa123456", name_target=base._random_name("target"),
        target_type=0, insecure_target=True, expect_status_code = 201,
        **kwargs):
        if endpoint_target is None:
            endpoint_target = r"https://{}.{}.{}.{}".format(int(round(time.time() * 1000)) % 100,
              int(round(time.time() * 1000)) % 200,
              int(round(time.time() * 1000)) % 100,
              int(round(time.time() * 1000)) % 254)
        client = self._get_client(**kwargs)
        policy = swagger_client.RepTarget(name=name_target, endpoint=endpoint_target,
            username=username_target, password=password_target, type=target_type,
            insecure=insecure_target)

        _, status_code, header = client.targets_post_with_http_info(policy)
        base._assert_status_code(expect_status_code, status_code)
        return base._get_id_from_header(header), name_target

    def  get_target(self, expect_status_code = 200, params = None, **kwargs):
        client = self._get_client(**kwargs)
        data = []
        if params is None:
            params = {}
        data, status_code, _ = client.targets_get_with_http_info(**params)
        base._assert_status_code(expect_status_code, status_code)
        return data

    def delete_target(self, target_id, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        _, status_code, _  = client.targets_id_delete_with_http_info(target_id)
        base._assert_status_code(expect_status_code, status_code)