# -*- coding: utf-8 -*-

import sys
import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException

class Label(base.Base, object):
    def __init__(self):
        super(Label,self).__init__(api_type = "label")

    def create_label(self, name=None, desc="", color="", scope="g",
            project_id=0, expect_status_code = 201, **kwargs):
        if name is None:
            name = base._random_name("label")
        label = v2_swagger_client.Label(name=name,
            description=desc, color=color,
            scope=scope, project_id=project_id)
        client = self._get_client(**kwargs)

        try:
            _, status_code, header = client.create_label_with_http_info(label)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
        else:
            base._assert_status_code(expect_status_code, status_code)
            base._assert_status_code(201, status_code)
            return base._get_id_from_header(header), name

    def delete_label(self, label_id, **kwargs):
        client = self._get_client(**kwargs)
        return client.delete_label_with_http_info(int(label_id))