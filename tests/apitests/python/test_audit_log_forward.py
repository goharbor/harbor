from __future__ import absolute_import

import unittest

from testutils import ADMIN_CLIENT, LOG_PATH, harbor_server, suppress_urllib3_warning
from library.audit_log import Audit_Log
from library.user import User
from library.project import Project
from library.artifact import Artifact
from library.configurations import Configurations
from library.repository import Repository, push_self_build_image_to_project


class TestAuditLogForword(unittest.TestCase, object):

    @suppress_urllib3_warning
    def setUp(self):
        self.project= Project()
        self.user= User()
        self.artifact = Artifact()
        self.repo = Repository()
        self.config = Configurations()
        self.audit_log = Audit_Log()
        self.image = "hello-world"
        self.tag = "latest"
        self.audit_log_path = LOG_PATH + "audit.log"
        self.audit_log_forward_endpoint = "harbor-log:10514"
        # 1. Reset audit log forword
        self.config.set_configurations_of_audit_log_forword("", False)

    def tearDown(self):
        # 1. Reset audit log forword
        self.config.set_configurations_of_audit_log_forword("", False)
        # 2. Close audit log file
        TestAuditLogForword.audit_log_file.close

    def testAuditLogForword(self):
        """
        Test case:
            Audit Log Forword
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Verify that Skip Audit Log Database cannot be enabled without Audit Log Forward;
            4. Enable Audit Log Forward;
            5. Push a new image(IA) in project(PA) by user(UA);
            6. Verify that the Audit Log should be in the log database;
            7. Verify that the Audit Log should be in the audit.log;
            8. Enable Skip Audit Log Database;
            9. Delete image(IA);
            10. Verify that the Audit Log should not be in log database;
            11. Verify that the Audit Log should be in the audit.log;
            12. Verify that Skip Audit Log Database cannot be enabled without Audit Log Forward;
        Tear down:
            1 Reset audit log forword.
        """
        url = ADMIN_CLIENT["endpoint"]
        user_password = "Aa123456"

        # 1. Create user(UA)
        user_id, user_name = self.user.create_user(user_password = user_password, **ADMIN_CLIENT)
        user_client = dict(endpoint = url, username = user_name, password = user_password, with_accessory = True)

        # 2.1. Create private project(PA) by user(UA)
        project_id, project_name = self.project.create_project(metadata = {"public": "false"}, **user_client)
        # 2.2. Get private project of user(UA), user(UA) can see only one private project which is project(PA)
        self.project.projects_should_exist(dict(public=False), expected_count = 1, expected_project_id = project_id, **user_client)
        
        # 3 Verify that Skip Audit Log Database cannot be enabled without Audit Log Forward
        self.config.set_configurations_of_audit_log_forword(skip_audit_log_database=True, expect_status_code=400)
        
        # 4 Enable Audit Log Forward
        self.config.set_configurations_of_audit_log_forword(audit_log_forward_endpoint=self.audit_log_forward_endpoint, expect_status_code=200)
        # 4.1 Verify configuration
        configurations = self.config.get_configurations()
        self.assertEqual(configurations.audit_log_forward_endpoint.value, self.audit_log_forward_endpoint)
        self.assertFalse(configurations.skip_audit_log_database.value)
        
        # 5 Push a new image(IA) in project(PA) by user(UA)
        repo_name, tag = push_self_build_image_to_project(project_name, harbor_server, user_name, user_password, self.image, self.tag)

        # 6. Verify that the Audit Log should be in the log database
        first_audit_log = self.audit_log.get_latest_audit_log()
        self.assertEqual(first_audit_log.operation, "create")
        self.assertEqual(first_audit_log.resource, "{}:{}".format(repo_name, tag))
        self.assertEqual(first_audit_log.resource_type, "artifact")
        self.assertEqual(first_audit_log.username, user_name)
        self.assertIsNotNone(first_audit_log.op_time)
        
        # 7. Verify that the Audit Log should be in the audit.log
        TestAuditLogForword.audit_log_file = open(self.audit_log_path, "r")
        latest_line = TestAuditLogForword.audit_log_file.readlines()[-1]
        self.assertIn('operator="{}"'.format(user_name), latest_line)
        self.assertIn('resourceType="artifact"', latest_line)
        self.assertIn('action:create', latest_line)
        self.assertIn('resource:{}:{}'.format(repo_name, tag), latest_line)
        self.assertIn('time="20', latest_line)
        
        # 8.1 Enable Skip Audit Log Database
        self.config.set_configurations_of_audit_log_forword(skip_audit_log_database=True)
        # 8.1 Verify configuration
        configurations = self.config.get_configurations()
        self.assertEqual(configurations.audit_log_forward_endpoint.value, self.audit_log_forward_endpoint)
        self.assertTrue(configurations.skip_audit_log_database.value)

        # 9. Delete image(IA)
        self.artifact.delete_artifact(project_name, self.image, self.tag, **user_client)

        # 10. Verify that the Audit Log should not be in log database
        second_audit_log = self.audit_log.get_latest_audit_log()
        self.assertEqual(first_audit_log.operation, second_audit_log.operation)
        self.assertEqual(first_audit_log.resource, second_audit_log.resource)
        self.assertEqual(first_audit_log.resource_type,second_audit_log.resource_type)
        self.assertEqual(first_audit_log.username, second_audit_log.username)
        self.assertEqual(first_audit_log.op_time, second_audit_log.op_time)

        # 11. Verify that the Audit Log should be in the audit.log
        latest_line = TestAuditLogForword.audit_log_file.readlines()[-1]
        self.assertIn('operator="{}"'.format(user_name), latest_line)
        self.assertIn('resourceType="artifact"', latest_line)
        self.assertIn('action:delete', latest_line)
        self.assertIn('resource:{}'.format(repo_name), latest_line)
        self.assertIn('time="20', latest_line)
        
        # 12 Verify that Skip Audit Log Database cannot be enabled without Audit Log Forward
        self.config.set_configurations_of_audit_log_forword(audit_log_forward_endpoint="", expect_status_code=400)

if __name__ == '__main__':
    unittest.main()