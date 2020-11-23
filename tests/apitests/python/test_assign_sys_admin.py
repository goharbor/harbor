from __future__ import absolute_import

import unittest

from testutils import ADMIN_CLIENT, suppress_urllib3_warning
from testutils import TEARDOWN
from library.user import User
from library.configurations import Configurations

class TestProjects(unittest.TestCase):
    @suppress_urllib3_warning
    def setUp(self):
        self.conf= Configurations()
        self.user = User()

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def tearDown(self):
        #1. Delete user(UA);
        self.user.delete_user(TestProjects.user_assign_sys_admin_id, **ADMIN_CLIENT)

    def testAssignSysAdmin(self):
        """
        Test case:
            Assign Sys Admin
        Test step and expected result:
            1. Create a new user(UA);
            2. Set user(UA) has sysadmin role by admin, check user(UA) can modify system configuration;
            3. 3. Set user(UA) has no sysadmin role by admin, check user(UA) can not modify system configuration;
            4. Set user(UA) has sysadmin role by admin, check user(UA) can modify system configuration.
        Tear down:
            1. Delete user(UA).
        """
        url = ADMIN_CLIENT["endpoint"]
        user_assign_sys_admin_password = "Aa123456"

        #1. Create a new user(UA);
        TestProjects.user_assign_sys_admin_id, user_assign_sys_admin_name = self.user.create_user(user_password = user_assign_sys_admin_password, **ADMIN_CLIENT)
        USER_ASSIGN_SYS_ADMIN_CLIENT=dict(endpoint = url, username = user_assign_sys_admin_name, password = user_assign_sys_admin_password)

        #2. Set user(UA) has sysadmin role by admin, check user(UA) can modify system configuration;
        self.user.update_user_role_as_sysadmin(TestProjects.user_assign_sys_admin_id, True, **ADMIN_CLIENT)
        self.conf.set_configurations_of_token_expiration(60, **USER_ASSIGN_SYS_ADMIN_CLIENT)

        #3. Set user(UA) has no sysadmin role by admin, check user(UA) can not modify system configuration;
        self.user.update_user_role_as_sysadmin(TestProjects.user_assign_sys_admin_id, False, **ADMIN_CLIENT)
        self.conf.set_configurations_of_token_expiration(70, expect_status_code = 403, **USER_ASSIGN_SYS_ADMIN_CLIENT)

        #4. Set user(UA) has sysadmin role by admin, check user(UA) can modify system configuration.
        self.user.update_user_role_as_sysadmin(TestProjects.user_assign_sys_admin_id, True, **ADMIN_CLIENT)
        self.conf.set_configurations_of_token_expiration(80, **USER_ASSIGN_SYS_ADMIN_CLIENT)

if __name__ == '__main__':
    unittest.main()

