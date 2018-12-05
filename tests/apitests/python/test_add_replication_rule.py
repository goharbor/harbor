from __future__ import absolute_import
import unittest

from testutils import ADMIN_CLIENT
from testutils import TEARDOWN
from library.project import Project
from library.user import User
from library.replication import Replication
from library.target import Target
import swagger_client

class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        project = Project()
        self.project= project

        user = User()
        self.user= user

        replication = Replication()
        self.replication= replication

        target = Target()
        self.target= target

    @classmethod
    def tearDown(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete rule(RA);
        for rule_id in TestProjects.rule_id_list:
            self.replication.delete_replication_rule(rule_id, **ADMIN_CLIENT)

        #2. Delete target(TA);
        self.target.delete_target(TestProjects.target_id, **ADMIN_CLIENT)

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
            3. Create a new target(TA)/registry;
            4. Create a new rule for project(PA) and target(TA);
            5. Check if rule is exist.
        Tear down:
            1. Delete rule(RA);
            2. Delete targe(TA);
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

        #3. Create a new target(TA)/registry
        TestProjects.target_id, _ = self.target.create_target(**ADMIN_CLIENT)
        print "TestProjects.target_id:", TestProjects.target_id

        TestProjects.rule_id_list = []

        trigger_values_to_set = ["Manual", "Immediate"]
        for value in trigger_values_to_set:
            #4. Create a new rule for project(PA) and target(TA)
            rule_id, rule_name = self.replication.create_replication_rule([TestProjects.project_add_rule_id],
                [TestProjects.target_id], trigger=swagger_client.RepTrigger(kind=value), **ADMIN_CLIENT)
            TestProjects.rule_id_list.append(rule_id)

            #5. Check rule should be exist
            self.replication.check_replication_rule_should_exist(rule_id, rule_name, expect_trigger = value, **ADMIN_CLIENT)


if __name__ == '__main__':
    unittest.main()
