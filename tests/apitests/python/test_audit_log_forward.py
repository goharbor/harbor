from __future__ import absolute_import

import unittest
import requests
import json
import time

from testutils import ADMIN_CLIENT, harbor_server, SYSLOG_ENDPOINT, ES_ENDPOINT, suppress_urllib3_warning
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
        self.tag2 = "test"
        # 1. Reset audit log forword
        self.config.set_configurations_of_audit_log_forword("", False)

    def tearDown(self):
        # 1. Reset audit log forword
        self.config.set_configurations_of_audit_log_forword("", False)

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
            9. Create a tag;
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
        self.config.set_configurations_of_audit_log_forword(audit_log_forward_endpoint=SYSLOG_ENDPOINT, expect_status_code=200)
        # 4.1 Verify configuration
        configurations = self.config.get_configurations()
        self.assertEqual(configurations.audit_log_forward_endpoint.value, SYSLOG_ENDPOINT)
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
        self.assertTrue(self.verifyLogInSyslogService(user_name, "{}:{}".format(repo_name, tag), "artifact", "create"))
        
        # 8.1 Enable Skip Audit Log Database
        self.config.set_configurations_of_audit_log_forword(skip_audit_log_database=True)
        # 8.1 Verify configuration
        configurations = self.config.get_configurations()
        self.assertEqual(configurations.audit_log_forward_endpoint.value, SYSLOG_ENDPOINT)
        self.assertTrue(configurations.skip_audit_log_database.value)

        # 9. Create a tag
        self.artifact.create_tag(project_name, self.image, self.tag, self.tag2, **user_client)

        # 10. Verify that the Audit Log should not be in log database
        second_audit_log = self.audit_log.get_latest_audit_log()
        self.assertEqual(first_audit_log.operation, second_audit_log.operation)
        self.assertEqual(first_audit_log.resource, second_audit_log.resource)
        self.assertEqual(first_audit_log.resource_type,second_audit_log.resource_type)
        self.assertEqual(first_audit_log.username, second_audit_log.username)
        self.assertEqual(first_audit_log.op_time, second_audit_log.op_time)

        # 11. Verify that the Audit Log should be in the audit.log
        self.assertTrue(self.verifyLogInSyslogService(user_name, "{}:{}".format(repo_name, self.tag2), "tag", "create"))
        
        # 12 Verify that Skip Audit Log Database cannot be enabled without Audit Log Forward
        self.config.set_configurations_of_audit_log_forword(audit_log_forward_endpoint="", expect_status_code=400)
    
    def verifyLogInSyslogService(self, username, resource, resource_type, operation, expected_count=1):
        url = ES_ENDPOINT + "/_count"
        payload = json.dumps({
            "query": {
                "match": {
                    "message": {
                        "query": "operator=\"{}\" resource:{} resourceType=\"{}\" action:{}".format(username, resource, resource_type, operation),
                        "operator": "and"
                    }
                }
            }
        })
        headers = { 'Content-Type': 'application/json' }
        for _ in range(5):
            response = requests.request("GET", url, headers=headers, data=payload)
            self.assertEqual(response.status_code, 200)
            response_json = response.json()
            if response_json["count"] == expected_count:
                return True
            time.sleep(5)
        return False

if __name__ == '__main__':
    unittest.main()