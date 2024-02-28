from __future__ import absolute_import


import unittest
import time
from testutils import ADMIN_CLIENT, suppress_urllib3_warning
from library.user import User


class TestUser(unittest.TestCase):


    @suppress_urllib3_warning
    def setUp(self):
        self.user = User()


    def testUser(self):
        """
        Test case:
            User CRUD
        Test step and expected result:
            1. Create a new user(UA);
            2. List all users, there should be one user(UA);
            3. Get current user by user(UA), it should be user(UA;
            4. Search user(UA) by name, it should be user(UA);
            5. Get user profile by user(UA), it should be user(UA);
            6. Update user profile by user(UA);
            7. Update user to admin;
            8. Update user password by user(UA);
            9. Update user password by admin;
            10. Get current user permissions by user(UA);
            11. Delete user(UA);
        """
        url = ADMIN_CLIENT["endpoint"]
        user_password = "Aa123456"

        # 1. Create a new user(UA);
        user_id, user_name = self.user.create_user(user_password=user_password, **ADMIN_CLIENT)
        timestamp = user_name.split("-")[1]
        USER_CLIENT=dict(endpoint=url, username=user_name, password=user_password)

        # 2. List all users, there should be one user(UA);
        users = self.user.get_users(**ADMIN_CLIENT)
        self.assertIsNotNone(users)

        # 3. Get current user by user(UA), it should be user(UA;
        current_user = self.user.get_user_current(**USER_CLIENT)
        self.check_user(current_user, user_name, user_id, timestamp)

        # 4. Search user(UA) by name, it should be user(UA);
        users = self.user.search_user_by_username(user_name, **USER_CLIENT)
        user = users[0]
        self.assertEqual(len(users), 1)
        self.assertEqual(user.username, user_name)
        self.assertEqual(user.user_id, user_id)

        # 5. Get user profile by user(UA), it should be user(UA);
        user = self.user.get_user_by_id(user_id, **USER_CLIENT)
        self.check_user(user, user_name, user_id, timestamp)

        # 6. Update user profile by user(UA);
        timestamp = int(round(time.time() * 1000))
        comment = "For testing"
        self.user.update_user_profile(user_id, email="realname-{}@harbortest.com".format(timestamp), realname="realname-{}".format(timestamp), comment=comment, **USER_CLIENT)
        user = self.user.get_user_by_id(user_id, **USER_CLIENT)
        self.check_user(user, user_name, user_id, timestamp, comment)

        # 7. Update user to admin;
        self.user.update_user_role_as_sysadmin(user_id, True, **ADMIN_CLIENT)
        user = self.user.get_user_by_id(user_id, **USER_CLIENT)
        self.check_user(user, user_name, user_id, timestamp, comment, True)

        # 8. Update user password by user(UA);
        new_password = "Aa1234567-New"
        self.user.update_user_pwd(user_id, new_password=new_password, old_password=user_password, **USER_CLIENT)
        self.user.search_user_by_username(user_name, expect_status_code=401, expect_response_body="unauthorized", **USER_CLIENT)
        USER_CLIENT["password"] = new_password

        # 9. Update user password by admin;
        new_password = "Aa1234567-New-Edit"
        self.user.update_user_pwd(user_id, new_password=new_password, old_password=USER_CLIENT["password"], **ADMIN_CLIENT)
        self.user.search_user_by_username(user_name, expect_status_code=401, expect_response_body="unauthorized", **USER_CLIENT)
        USER_CLIENT["password"] = new_password

        # 10. Get current user permissions by user(UA);
        permissions = self.user.get_current_user_permissions(scope="/project/1/repository", relative=True, **USER_CLIENT)
        self.assertTrue(len(permissions) > 0)

        # 11. Delete user(UA);
        self.user.delete_user(user_id, **ADMIN_CLIENT)
        self.user.get_user_by_id(user_id, expect_status_code=404, expect_response_body="user {} not found".format(user_id), **ADMIN_CLIENT)


    def check_user(self, user, user_name, user_id, timestamp, comment=None, sysadmin_flag=False):
        self.assertEqual(user.username, user_name)
        self.assertEqual(user.user_id, user_id)
        self.assertEqual(user.email, "realname-{}@harbortest.com".format(timestamp))
        self.assertEqual(user.comment, comment)
        self.assertEqual(user.realname, "realname-{}".format(timestamp))
        self.assertEqual(user.sysadmin_flag, sysadmin_flag)


if __name__ == '__main__':
    unittest.main()

