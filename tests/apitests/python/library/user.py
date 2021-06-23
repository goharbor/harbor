# -*- coding: utf-8 -*-

import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException


class User(base.Base, object):

    def __init__(self):
        super(User, self).__init__(api_type = "user")

    def create_user(self, name=None,
                    email=None, user_password=None, realname=None, expect_status_code=201, **kwargs):
        if name is None:
            name = base._random_name("user")
        if realname is None:
            realname = base._random_name("realname")
        if email is None:
            email = '%s@%s.com' % (realname, "harbortest")
        if user_password is None:
            user_password = "Harbor12345678"
        user_req = v2_swagger_client.UserCreationReq(username=name, email=email, password=user_password, realname=realname)
        try:
            _, status_code, header = self._get_client(**kwargs).create_user_with_http_info(user_req)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
        else:
            base._assert_status_code(expect_status_code, status_code)
            return base._get_id_from_header(header), name

    def get_users(self, user_name=None, email=None, page=None, page_size=None, expect_status_code=200, **kwargs):
        query = []
        if user_name is not None:
            query.append("username=" + user_name)
        if email is not None:
            query.append("email=" + email)

        params = {}
        if len(query) > 0:
            params["q"] = ",".join(query)
        if page is not None:
            params["page"] = page
        if page_size is not None:
            params["page_size"] = page_size
        try:
            data, status_code, _ = self._get_client(**kwargs).list_users_with_http_info(**params)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
        else:
            base._assert_status_code(expect_status_code, status_code)
            return data

    def get_user_by_id(self, user_id, **kwargs):
        data, status_code, _ = self._get_client(**kwargs).get_user_with_http_info(user_id)
        base._assert_status_code(200, status_code)
        return data

    def get_user_by_name(self, name, expect_status_code=200, **kwargs):
        users = self.get_users(user_name=name, expect_status_code=expect_status_code, **kwargs)
        for user in users:
            if user.username == name:
                return user
        return None

    def get_user_current(self, **kwargs):
        data, status_code, _ = self._get_client(**kwargs).get_current_user_info_with_http_info()
        base._assert_status_code(200, status_code)
        return data

    def delete_user(self, user_id, expect_status_code=200, **kwargs):
        _, status_code, _ = self._get_client(**kwargs).delete_user_with_http_info(user_id)
        base._assert_status_code(expect_status_code, status_code)
        return user_id

    def update_user_pwd(self, user_id, new_password=None, old_password=None, **kwargs):
        if old_password is None:
            old_password = ""
        password = v2_swagger_client.PasswordReq(old_password=old_password, new_password=new_password)
        _, status_code, _ = self._get_client(**kwargs).update_user_password_with_http_info(user_id, password)
        base._assert_status_code(200, status_code)
        return user_id

    def update_user_profile(self, user_id, email=None, realname=None, comment=None, **kwargs):
        user_profile = v2_swagger_client.UserProfile(email=email, realname=realname, comment=comment)
        _, status_code, _ = self._get_client(**kwargs).update_user_profile_with_http_info(user_id, user_profile)
        base._assert_status_code(200, status_code)
        return user_id

    def update_user_role_as_sysadmin(self, user_id, IsAdmin, **kwargs):
        sysadmin_flag = v2_swagger_client.UserSysAdminFlag(sysadmin_flag=IsAdmin)
        _, status_code, _ = self._get_client(**kwargs).set_user_sys_admin_with_http_info(user_id, sysadmin_flag)
        base._assert_status_code(200, status_code)
        return user_id
