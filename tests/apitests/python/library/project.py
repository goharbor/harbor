# -*- coding: utf-8 -*-

import v2_swagger_client
from library.base import _assert_status_code
from v2_swagger_client.models.role_request import RoleRequest
from v2_swagger_client.rest import ApiException

import base


def is_member_exist_in_project(members, member_user_name, expected_member_role_id = None):
    result = False
    for member in members:
        if member.entity_name == member_user_name:
            if expected_member_role_id != None:
                if member.role_id == expected_member_role_id:
                    return True
            else:
                return True
    return result

def get_member_id_by_name(members, member_user_name):
    for member in members:
        if member.entity_name == member_user_name:
            return member.id
    return None

class Project(base.Base):
    def __init__(self, username=None, password=None):
        kwargs = dict(api_type="projectv2")
        if username and password:
            kwargs["credential"] = base.Credential('basic_auth', username, password)
        super(Project, self).__init__(**kwargs)

    def create_project(self, name=None, registry_id=None, metadata=None, expect_status_code = 201, expect_response_body = None, **kwargs):
        if name is None:
            name = base._random_name("project")
        if metadata is None:
            metadata = {}
        if registry_id is None:
            registry_id = registry_id
        project = v2_swagger_client.ProjectReq(project_name=name, registry_id = registry_id, metadata=metadata)
        try:
            _, status_code, header = self._get_client(**kwargs).create_project_with_http_info(project)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(201, status_code)
        return base._get_id_from_header(header), name

    def get_projects(self, params, **kwargs):
        data = []
        data, status_code, _ = self._get_client(**kwargs).list_projects_with_http_info(**params)
        base._assert_status_code(200, status_code)
        return data

    def get_project_id(self, project_name, **kwargs):
        project_data = self.get_projects(dict(), **kwargs)
        actual_count = len(project_data)
        if actual_count == 1 and str(project_data[0].project_name) != str(project_name):
            return project_data[0].project_id
        else:
            return None

    def projects_should_exist(self, params, expected_count = None, expected_project_id = None, **kwargs):
        project_data = self.get_projects(params, **kwargs)
        actual_count = len(project_data)
        if expected_count is not None and actual_count!= expected_count:
            raise Exception(r"Private project count should be {}.".format(expected_count))
        if expected_project_id is not None and actual_count == 1 and str(project_data[0].project_id) != str(expected_project_id):
            raise Exception(r"Project-id check failed, expect {} but got {}, please check this test case.".format(str(expected_project_id), str(project_data[0].project_id)))

    def check_project_name_exist(self, name=None, **kwargs):
        try:
            _, status_code, _ = self._get_client(**kwargs).head_project_with_http_info(name)
        except ApiException as e:
            status_code = -1
        return {
            200: True,
            404: False,
        }.get(status_code,False)

    def get_project(self, project_id, expect_status_code = 200, expect_response_body = None, **kwargs):
        try:
            data, status_code, _ = self._get_client(**kwargs).get_project_with_http_info(project_id)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return

        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(200, status_code)
        print("Project {} info: {}".format(project_id, data))
        return data

    def update_project(self, project_id, expect_status_code=200, metadata=None, cve_allowlist=None, **kwargs):
        project = v2_swagger_client.ProjectReq(metadata=metadata, cve_allowlist=cve_allowlist)
        try:
            _, sc, _ = self._get_client(**kwargs).update_project_with_http_info(project_id, project)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
        else:
            base._assert_status_code(expect_status_code, sc)

    def delete_project(self, project_id, expect_status_code = 200, **kwargs):
        _, status_code, _ = self._get_client(**kwargs).delete_project_with_http_info(project_id)
        base._assert_status_code(expect_status_code, status_code)

    def get_project_log(self, project_name, expect_status_code = 200, **kwargs):
        body, status_code, _ = self._get_client(**kwargs).get_logs_with_http_info(project_name)
        base._assert_status_code(expect_status_code, status_code)
        return body

    def filter_project_logs(self, project_name, operator, resource, resource_type, operation, **kwargs):
        access_logs = self.get_project_log(project_name, **kwargs)
        count = 0
        for each_access_log in list(access_logs):
            if each_access_log.username == operator and \
                    each_access_log.resource_type == resource_type and \
                    each_access_log.resource == resource and \
                    each_access_log.operation == operation:
                count = count + 1
        return count

    def get_project_members(self, project_id, **kwargs):
        kwargs['api_type'] = 'member'
        return self._get_client(**kwargs).list_project_members(project_id)

    def get_project_member(self, project_id, member_id, expect_status_code = 200, expect_response_body = None, **kwargs):
        from swagger_client.rest import ApiException
        kwargs['api_type'] = 'member'
        data = []
        try:
            data, status_code, _ = self._get_client(**kwargs).get_project_member_with_http_info(project_id, member_id,)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return

        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(200, status_code)
        return data

    def get_project_member_id(self, project_id, member_user_name, **kwargs):
        kwargs['api_type'] = 'member'
        members = self.get_project_members(project_id, **kwargs)
        result = get_member_id_by_name(list(members), member_user_name)
        if result == None:
            raise Exception(r"Failed to get member id of member {} in project {}.".format(member_user_name, project_id))
        else:
            return result

    def check_project_member_not_exist(self, project_id, member_user_name, **kwargs):
        kwargs['api_type'] = 'member'
        members = self.get_project_members(project_id, **kwargs)
        result = is_member_exist_in_project(list(members), member_user_name)
        if result == True:
            raise Exception(r"User {} should not be a member of project with ID {}.".format(member_user_name, project_id))

    def check_project_members_exist(self, project_id, member_user_name, expected_member_role_id = None, **kwargs):
        kwargs['api_type'] = 'member'
        members = self.get_project_members(project_id, **kwargs)
        result = is_member_exist_in_project(members, member_user_name, expected_member_role_id = expected_member_role_id)
        if result == False:
            raise Exception(r"User {} should be a member of project with ID {}.".format(member_user_name, project_id))

    def update_project_member_role(self, project_id, member_id, member_role_id, expect_status_code = 200, **kwargs):
        kwargs['api_type'] = 'member'
        role = RoleRequest(role_id = member_role_id)
        data, status_code, _ = self._get_client(**kwargs).update_project_member_with_http_info(project_id, member_id, role = role)
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(200, status_code)
        return data

    def delete_project_member(self, project_id, member_id, expect_status_code = 200, **kwargs):
        kwargs['api_type'] = 'member'
        _, status_code, _ = self._get_client(**kwargs).delete_project_member_with_http_info(project_id, member_id)
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(200, status_code)

    def add_project_members(self, project_id, user_id = None, member_role_id = None, _ldap_group_dn=None, expect_status_code = 201, **kwargs):
        kwargs['api_type'] = 'member'
        projectMember = v2_swagger_client.ProjectMember()
        if user_id is not None:
           projectMember.member_user = {"user_id": int(user_id)}
        if member_role_id is None:
            projectMember.role_id = 1
        else:
            projectMember.role_id = member_role_id
        if _ldap_group_dn is not None:
            projectMember.member_group = v2_swagger_client.UserGroup(ldap_group_dn=_ldap_group_dn)

        data = []
        try:
            data, status_code, header = self._get_client(**kwargs).create_project_member_with_http_info(project_id, project_member = projectMember)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
        else:
            base._assert_status_code(expect_status_code, status_code)
            return base._get_id_from_header(header)

    def query_user_logs(self, project_name, status_code=200, **kwargs):
        try:
            logs = self.get_project_log(project_name, expect_status_code=status_code, **kwargs)
            count = 0
            for log in list(logs):
                count = count + 1
            return count
        except ApiException as e:
            _assert_status_code(status_code, e.status)
            return 0