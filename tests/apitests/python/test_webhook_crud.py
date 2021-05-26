from __future__ import absolute_import


import unittest
import v2_swagger_client
from testutils import ADMIN_CLIENT, suppress_urllib3_warning
from testutils import harbor_server
from testutils import TEARDOWN
from library.base import _assert_status_code
from library.project import Project
from library.user import User
from library.repository import Repository
from library.repository import push_self_build_image_to_project
from library.webhook import Webhook

class TestProjects(unittest.TestCase):
    @suppress_urllib3_warning
    def setUp(self):
        self.url = ADMIN_CLIENT["endpoint"]
        self.user_password = "Aa123456"
        self.project= Project()
        self.user= User()
        self.webhook= Webhook()
        self.user_id, self.project_id = [None] * 2
        self.user_id, self.user_name = self.user.create_user(user_password = self.user_password, **ADMIN_CLIENT)
        self.USER_CLIENT = dict(with_signature = True, with_immutable_status = True, endpoint = self.url, username = self.user_name, password = self.user_password)

    @unittest.skipIf(TEARDOWN == True, "Test data won't be erased.")
    def tearDown(self):
        #1. Delete project(PA);
        self.project.delete_project(self.project_id, **self.USER_CLIENT)

        #2. Delete user(UA).
        self.user.delete_user(self.user_id, **ADMIN_CLIENT)

    def testDelRepo(self):
        """
        Test case:
            Webhook CRUD
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Create a new webhook(WA) in project(PA) by user(UA);
            4. Modify properties of webhook(WA), it should be successful;
            5. Delete webhook(WA) by user(UA), it should be successful.
        Tear down:
            1. Delete project(PA);
            2. Delete user(UA).
        """
        #2. Create a new project;
        self.project_id, project_name = self.project.create_project(metadata = {"public": "false"}, **self.USER_CLIENT)
        print("project_id:",self.project_id)
        print("project_name:",project_name)

        target_1 = v2_swagger_client.WebhookTargetObject(
            address = "https://hooks.slack.com/services",
            skip_cert_verify = False,
            type = "slack",
            auth_header = "aaa"
        )
        target_2 = v2_swagger_client.WebhookTargetObject(
            address = "https://202.10.12.13",
            skip_cert_verify = False,
            type = "http",
            auth_header = "aaa"
        )
        #This need to be removed once issue #13378 fixed.
        policy_id, policy_name = self.webhook.create_webhook(self.project_id, [target_1, target_2], **self.USER_CLIENT)
        target_1 = v2_swagger_client.WebhookTargetObject(
            address = "https://hooks.slack.com/services/new",
            skip_cert_verify = True,
            type = "http",
            auth_header = "bbb"
        )
        self.webhook.get_webhook(self.project_id, policy_id, **self.USER_CLIENT)
        self.webhook.update_webhook(self.project_id, policy_id, name = "new_name",auth_header = "new_header",
            event_types = ["DELETE_ARTIFACT", "TAG_RETENTION"], enabled = False, targets = [target_1], **self.USER_CLIENT)
        self.webhook.get_webhook(self.project_id, policy_id, **self.USER_CLIENT)

        self.webhook.delete_webhook(self.project_id, policy_id, **self.USER_CLIENT)
        self.webhook.get_webhook(self.project_id, policy_id, expect_status_code = 404, **self.USER_CLIENT)
if __name__ == '__main__':
    unittest.main()

