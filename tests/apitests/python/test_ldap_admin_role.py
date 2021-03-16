from __future__ import absolute_import
import unittest

from testutils import harbor_server, suppress_urllib3_warning
from testutils import TEARDOWN
from testutils import ADMIN_CLIENT
from library.user import User
from library.project import Project
from library.configurations import Configurations


class TestLdapAdminRole(unittest.TestCase):
    @suppress_urllib3_warning
    def setUp(self):
        url = ADMIN_CLIENT["endpoint"]
        self.conf= Configurations()
        self.user = User()
        self.project = Project()
        self.USER_MIKE=dict(endpoint = url, username = "mike", password = "zhu88jie")
        self.project_id = None


    def testLdapAdminRole(self):
        """
        Test case:
            LDAP Admin Role
        Test step and expected result:
            1. Set LDAP Auth configurations;
            2. Create a new public project(PA) by LDAP user mike;
            3. Check project is created successfully;
            4. Check mike's SysAdminFlag is false, but AdminRoleInAuth should be true
            5. Delete project(PA);
        """


        self.conf.set_configurations_of_ldap(ldap_group_admin_dn="cn=harbor_users,ou=groups,dc=example,dc=com", **ADMIN_CLIENT)

        self.project_id, project_name = self.project.create_project(metadata = {"public": "false"}, **self.USER_MIKE)
        print("self.project_id:", self.project_id)
        self.project.check_project_name_exist(name=project_name, **self.USER_MIKE)

        _user = self.user.get_user_current(**self.USER_MIKE)
        print( _user)
        self.assertFalse(_user.sysadmin_flag)
        self.assertTrue(_user.admin_role_in_auth)


    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def tearDown(self):
        print("self.project_id:", self.project_id)
        self.project.delete_project(self.project_id, **self.USER_MIKE)
        print("Case completed")

if __name__ == '__main__':
    unittest.main()