# -*- coding: utf-8 -*-

import time
import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException

class Robot(base.Base, object):
    def __init__(self):
        super(Robot,self).__init__(api_type = "robot")

    def create_project_robot(self, project_name, expires_at, robot_name = None, robot_desc = None, has_pull_right = True,  has_push_right = True, has_chart_read_right = True,  has_chart_create_right = True, expect_status_code = 201, **kwargs):
        if robot_name is None:
            robot_name = base._random_name("robot")
        if robot_desc is None:
            robot_desc = base._random_name("robot_desc")
        if has_pull_right is False and has_push_right is False:
            has_pull_right = True
        access_list = []
        action_pull = "pull"
        action_push = "push"
        action_read = "read"
        action_create = "create"
        if has_pull_right is True:
            robotAccountAccess = v2_swagger_client.Access(resource = "repository", action = action_pull)
            access_list.append(robotAccountAccess)
        if has_push_right is True:
            robotAccountAccess = v2_swagger_client.Access(resource = "repository", action = action_push)
            access_list.append(robotAccountAccess)
        if has_chart_read_right is True:
            robotAccountAccess = v2_swagger_client.Access(resource = "helm-chart", action = action_read)
            access_list.append(robotAccountAccess)
        if has_chart_create_right is True:
            robotAccountAccess = v2_swagger_client.Access(resource = "helm-chart-version", action = action_create)
            access_list.append(robotAccountAccess)

        robotaccountPermissions = v2_swagger_client.Permission(kind = "project", namespace = project_name, access = access_list)
        permission_list = []
        permission_list.append(robotaccountPermissions)
        robotAccountCreate = v2_swagger_client.RobotCreate(name=robot_name, description=robot_desc, expires_at=expires_at, level="project", permissions = permission_list)

        client = self._get_client(**kwargs)
        data = []
        data, status_code, header = client.create_robot_with_http_info(robotAccountCreate)
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(201, status_code)
        return base._get_id_from_header(header), data

    def get_robot_account_by_id(self, robot_id, **kwargs):
        client = self._get_client(**kwargs)
        data, status_code, _ = client.get_robot_by_id_with_http_info(robot_id)
        return data

    def disable_robot_account(self, robot_id, disable, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        data = self.get_robot_account_by_id(robot_id, **kwargs)
        robotAccountUpdate = v2_swagger_client.RobotCreate(name=data.name, description=data.description, expires_at=data.expires_at, level=data.level, permissions = data.permissions, disable = disable)

        _, status_code, _ = client.update_robot_with_http_info(robot_id, robotAccountUpdate)
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(200, status_code)

    def delete_robot_account(self, robot_id, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        _, status_code, _ = client.delete_robot_with_http_info(robot_id)
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(200, status_code)
