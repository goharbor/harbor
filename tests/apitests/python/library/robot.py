# -*- coding: utf-8 -*-

import time
import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException
from base import _assert_status_code

class Robot(base.Base, object):
    def __init__(self):
        super(Robot,self).__init__(api_type = "robot")

    def list_robot(self, expect_status_code = 200, **kwargs):
        try:
            body, status_code, _ = self._get_client(**kwargs).list_robot_with_http_info()
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            return []
        else:
            base._assert_status_code(expect_status_code, status_code)
            base._assert_status_code(200, status_code)
            return body

    def create_access_list(self, right_map = [True] * 10):
        _assert_status_code(10, len(right_map), r"Please input full access list for system robot account. Expected {}, while actual input count is {}.")
        action_pull = "pull"
        action_push = "push"
        action_read = "read"
        action_create = "create"
        action_del = "delete"

        access_def_list = [
            ("repository", action_pull),
            ("repository", action_push),
            ("artifact", action_del),
            ("helm-chart", action_read),
            ("helm-chart-version", action_create),
            ("helm-chart-version", action_del),
            ("tag", action_create),
            ("tag", action_del),
            ("artifact-label", action_create),
            ("scan", action_create)
        ]

        access_list = []
        for i in range(len(access_def_list)):
            if right_map[i] is True:
                robotAccountAccess = v2_swagger_client.Access(resource = access_def_list[i][0], action = access_def_list[i][1])
                access_list.append(robotAccountAccess)
        return access_list

    def create_project_robot(self, project_name, duration, robot_name = None, robot_desc = None,
            has_pull_right = True,  has_push_right = True, has_chart_read_right = True,
            has_chart_create_right = True, expect_status_code = 201, expect_response_body = None,
            **kwargs):
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

        robotaccountPermissions = v2_swagger_client.RobotPermission(kind = "project", namespace = project_name, access = access_list)
        permission_list = []
        permission_list.append(robotaccountPermissions)
        robotAccountCreate = v2_swagger_client.RobotCreate(name=robot_name, description=robot_desc, duration=duration, level="project", permissions = permission_list)

        data = []
        try:
            data, status_code, header = self._get_client(**kwargs).create_robot_with_http_info(robotAccountCreate)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
        else:
            base._assert_status_code(expect_status_code, status_code)
            base._assert_status_code(201, status_code)
            return base._get_id_from_header(header), data

    def get_robot_account_by_id(self, robot_id, **kwargs):
        data, status_code, _ = self._get_client(**kwargs).get_robot_by_id_with_http_info(robot_id)
        return data

    def disable_robot_account(self, robot_id, disable, expect_status_code = 200, **kwargs):
        data = self.get_robot_account_by_id(robot_id, **kwargs)
        robotAccountUpdate = v2_swagger_client.RobotCreate(name=data.name, description=data.description, duration=data.duration, level=data.level, permissions = data.permissions, disable = disable)

        _, status_code, _ = self._get_client(**kwargs).update_robot_with_http_info(robot_id, robotAccountUpdate)
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(200, status_code)

    def delete_robot_account(self, robot_id, expect_status_code = 200, **kwargs):
        _, status_code, _ = self._get_client(**kwargs).delete_robot_with_http_info(robot_id)
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(200, status_code)

    def create_system_robot(self, permission_list, duration, robot_name = None, robot_desc = None, expect_status_code = 201, **kwargs):
        if robot_name is None:
            robot_name = base._random_name("robot")
        if robot_desc is None:
            robot_desc = base._random_name("robot_desc")

        robotAccountCreate = v2_swagger_client.RobotCreate(name=robot_name, description=robot_desc, duration=duration, level="system", disable = False, permissions = permission_list)
        data = []
        data, status_code, header = self._get_client(**kwargs).create_robot_with_http_info(robotAccountCreate)
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(201, status_code)
        return base._get_id_from_header(header), data

    def update_robot_account(self, robot_id, robot, expect_status_code = 200, **kwargs):
        _, status_code, _ = self._get_client(**kwargs).update_robot_with_http_info(robot_id, robot)
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(200, status_code)

    def update_system_robot_account(self, robot_id, robot_name, robot_account_Permissions_list, disable = None, expect_status_code = 200, **kwargs):
        robot = v2_swagger_client.Robot(id = robot_id, name = robot_name, level = "system", permissions = robot_account_Permissions_list)
        if disable in (True, False):
            robot.disable = disable
        self.update_robot_account(robot_id, robot, expect_status_code = expect_status_code, **kwargs)

    def refresh_robot_account_secret(self, robot_id, robot_new_sec, expect_status_code = 200, **kwargs):
        robot_sec = v2_swagger_client.RobotSec(secret = robot_new_sec)
        data, status_code, _ = self._get_client(**kwargs).refresh_sec_with_http_info(robot_id, robot_sec)
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(200, status_code)
        print("Refresh new secret:", data)
        return data
