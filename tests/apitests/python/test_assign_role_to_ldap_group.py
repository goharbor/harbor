from __future__ import absolute_import
import unittest

from testutils import harbor_server
from testutils import TEARDOWN
from testutils import ADMIN_CLIENT
from testutils import created_user, created_project
from library.project import Project
from library.user import User
from library.repository import Repository
from library.repository import push_image_to_project
from library.configurations import Configurations


class TestAssignRoleToLdapGroup(unittest.TestCase):
    @classmethod
    def setUp(self):
        self.conf= Configurations()
        self.project = Project()
        self.repo = Repository()

    @classmethod
    def tearDown(self):
        print("Case completed")

    def testAssignRoleToLdapGroup(self):
        """
        Test case:
            Assign Role To Ldap Group
        Test step and expected result:
            1. Set LDAP Auth configurations;
            2. Create a new public project(PA) by Admin;
            3. Add 3 member groups to project(PA);
            4. Push image by each member role;
            5. Verfify that admin_user and dev_user can push image, guest_user can not push image;
            6. Verfify that admin_user, dev_user and guest_user can view logs, test user can not view logs.
            7. Delete repository(RA) by user(UA);
            8. Delete project(PA);
        """
        url = ADMIN_CLIENT["endpoint"]
        USER_ADMIN=dict(endpoint = url, username = "admin_user", password = "zhu88jie", repo = "hello-world")
        USER_DEV=dict(endpoint = url, username = "dev_user", password = "zhu88jie", repo = "alpine")
        USER_GUEST=dict(endpoint = url, username = "guest_user", password = "zhu88jie", repo = "busybox")
        USER_TEST=dict(endpoint = url, username = "test", password = "123456")

        self.conf.set_configurations_of_ldap(ldap_filter="", ldap_group_attribute_name="cn", ldap_group_base_dn="ou=groups,dc=example,dc=com",
                                             ldap_group_search_filter="objectclass=groupOfNames", ldap_group_search_scope=2, **ADMIN_CLIENT)

        with created_project(metadata={"public": "false"}) as (project_id, project_name):
            self.project.add_project_members(project_id, member_role_id = 1, _ldap_group_dn = "cn=harbor_admin,ou=groups,dc=example,dc=com", **ADMIN_CLIENT)
            self.project.add_project_members(project_id, member_role_id = 2, _ldap_group_dn = "cn=harbor_dev,ou=groups,dc=example,dc=com", **ADMIN_CLIENT)
            self.project.add_project_members(project_id, member_role_id = 3, _ldap_group_dn = "cn=harbor_guest,ou=groups,dc=example,dc=com", **ADMIN_CLIENT)
            projects = self.project.get_projects(dict(name=project_name), **USER_ADMIN)
            self.assertTrue(len(projects) == 1)
            self.assertEqual(1, projects[0].current_user_role_id)

            repo_name_admin, tag_name_admin  = push_image_to_project(project_name, harbor_server, USER_ADMIN["username"], USER_ADMIN["password"], USER_ADMIN["repo"], "latest")
            self.repo.image_should_exist(repo_name_admin, tag_name_admin, **USER_ADMIN)
            repo_name_dev, tag_name_dev = push_image_to_project(project_name, harbor_server, USER_DEV["username"], USER_DEV["password"], USER_DEV["repo"], "latest")
            self.repo.image_should_exist(repo_name_dev, tag_name_dev, **USER_DEV)
            repo_name_guest, tag_name_guest = push_image_to_project(project_name, harbor_server, USER_GUEST["username"], USER_GUEST["password"], USER_GUEST["repo"], "latest")
            self.repo.image_should_not_exist(repo_name_guest, tag_name_guest, **USER_GUEST)


            self.assertTrue(self.project.query_user_logs(project_id, **USER_ADMIN)>0, "admin user can see logs")
            self.assertTrue(self.project.query_user_logs(project_id, **USER_DEV)>0, "dev user can see logs")
            self.assertTrue(self.project.query_user_logs(project_id, **USER_GUEST)>0, "guest user can see logs")
            self.assertTrue(self.project.query_user_logs(project_id, status_code=403, **USER_TEST)==0, "test user can not see any logs")

            self.repo.delete_repoitory(repo_name_admin, **USER_ADMIN)
            self.repo.delete_repoitory(repo_name_dev, **USER_ADMIN)

if __name__ == '__main__':
    unittest.main()