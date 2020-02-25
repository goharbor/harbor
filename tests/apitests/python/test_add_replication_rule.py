from __future__ import absolute_import
import unittest

from testutils import ADMIN_CLIENT
from testutils import harbor_server
from testutils import TEARDOWN
from library.project import Project
from library.user import User
from library.replication import Replication
from library.registry import Registry
import swagger_client

class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        self.project = Project()
        self.user = User()
        self.replication = Replication()
        self.registry = Registry()

    @classmethod
    def tearDown(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete rule(RA);
        self.replication.delete_replication_rule(TestProjects.rule_id, **ADMIN_CLIENT)

        #2. Delete registry(TA);
        self.registry.delete_registry(TestProjects.registry_id, **ADMIN_CLIENT)

        #3. Delete project(PA);
        self.project.delete_project(TestProjects.project_add_rule_id, **TestProjects.USER_add_rule_CLIENT)

        #4. Delete user(UA);
        self.user.delete_user(TestProjects.user_add_rule_id, **ADMIN_CLIENT)

    def testAddReplicationRule(self):
        """
        Test case:
            Add Replication Rule
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new private project(PA) by user(UA);
            3. Create a new registry;
            4. Create a new rule for this registry;
            5. Check rule should be exist.
        Tear down:
            1. Delete rule(RA);
            2. Delete registry(TA);
            3. Delete project(PA);
            4. Delete user(UA).
        """
        url = ADMIN_CLIENT["endpoint"]
        user_add_rule_password = "Aa123456"

        #1. Create user(UA)
        TestProjects.user_add_rule_id, user_add_rule_name = self.user.create_user(user_password = user_add_rule_password, **ADMIN_CLIENT)

        TestProjects.USER_add_rule_CLIENT=dict(endpoint = url, username = user_add_rule_name, password = user_add_rule_password)

        #2.1. Create private project(PA) by user(UA)
        TestProjects.project_add_rule_id, _ = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_add_rule_CLIENT)

        #2.2. Get private project of uesr-001, uesr-001 can see only one private project which is project-001
        self.project.projects_should_exist(dict(public=False), expected_count = 1,
            expected_project_id = TestProjects.project_add_rule_id, **TestProjects.USER_add_rule_CLIENT)

        #3. Create a new registry
        TestProjects.registry_id, _ = self.registry.create_registry("https://" + harbor_server,**ADMIN_CLIENT)
        print "TestProjects.registry_id:", TestProjects.registry_id

        #4. Create a new rule for this registry;
        TestProjects.rule_id, rule_name = self.replication.create_replication_policy(dest_registry=swagger_client.Registry(id=int(TestProjects.registry_id)), **ADMIN_CLIENT)

        #5. Check rule should be exist
        self.replication.check_replication_rule_should_exist(TestProjects.rule_id, rule_name, **ADMIN_CLIENT)


if __name__ == '__main__':
    unittest.main()
