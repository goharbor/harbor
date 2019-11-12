from __future__ import absolute_import
import unittest

from testutils import ADMIN_CLIENT
from testutils import TEARDOWN
from library.project import Project
from library.user import User
from library.webhook import Webhook


class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        project = Project()
        self.project = project

        user = User()
        self.user = user

        webhook = Webhook()
        self.webhook = webhook

    @classmethod
    def tearDown(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN is False, "Test data won't be erased.")
    def test_ClearData(self):
        # 1. Delete webhook policy;
        self.webhook.delete_webhook_policy(TestProjects.user_add_policy_id, TestProjects.policy_id, **ADMIN_CLIENT)

        # 2. Delete project;
        self.project.delete_project(TestProjects.user_add_policy_id, **TestProjects.USER_add_policy_CLIENT)

        # 3. Delete user;
        self.user.delete_user(TestProjects.user_add_policy_id, **ADMIN_CLIENT)

    def testAddWebhookPolicy(self):
        """
        Test case:
            Add Webhook Policy
        Test step and expected result:
            1. Create a new user;
            2. Create a new private project(PA) by user(UA);
            3. Create a new webhook policy for this project;
            4. Check webhook policy should be exist.
        Tear down:
            1. Delete webhook policy;
            2. Delete project;
            3. Delete user.
        """
        url = ADMIN_CLIENT["endpoint"]
        user_add_policy_password = "Aa123456"

        # 1. Create user
        TestProjects.user_add_policy_id, user_add_policy_name = self.user.create_user(
            user_password=user_add_policy_password,
            **ADMIN_CLIENT)

        TestProjects.USER_add_policy_CLIENT = dict(endpoint=url, username=user_add_policy_name,
                                                   password=user_add_policy_password)

        # 2.1. Create private project(PA) by user(UA)
        TestProjects.project_add_policy_id, _ = self.project.create_project(metadata={"public": "false"},
                                                                            **TestProjects.USER_add_policy_CLIENT)

        # 2.2. Get private project of uesr-001, uesr-001 can see only one private project which is project-001
        self.project.projects_should_exist(dict(public=False), expected_count=1,
                                           expected_project_id=TestProjects.project_add_policy_id,
                                           **TestProjects.USER_add_policy_CLIENT)

        # 3. Create a new policy for this project
        TestProjects.policy_id = self.webhook.create_webhook_policy(project_id=TestProjects.project_add_policy_id,
                                                                    **ADMIN_CLIENT)
        # 4. Check policy should be exist
        self.webhook.check_webook_policy_should_exist(TestProjects.project_add_policy_id, TestProjects.policy_id)


if __name__ == '__main__':
    unittest.main()
