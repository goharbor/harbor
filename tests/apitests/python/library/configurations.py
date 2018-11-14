# -*- coding: utf-8 -*-

import base
import swagger_client

class Configurations(base.Base):
    def get_configurations(self, item_name = None, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        data, status_code, _ = client.configurations_get_with_http_info()
        base._assert_status_code(expect_status_code, status_code)
        if item_name is not None:
            return {
            'project_creation_restriction': data.project_creation_restriction.value,
            }.get(item_name,'error')
        return data

    def set_configurations_of_project_creation_restriction_success(self, project_creation_restriction, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        conf = swagger_client.Configurations(project_creation_restriction = project_creation_restriction)
        _, status_code, _ = client.configurations_put_with_http_info(conf)
        base._assert_status_code(200, status_code)

        item_value = self.get_configurations(item_name = "project_creation_restriction", **kwargs)
        if item_value != project_creation_restriction:
            raise Exception("Failed to set system configuration item {} to value {},\
                actual value is {}".format("project_creation_restriction", project_creation_restriction, item_value))
