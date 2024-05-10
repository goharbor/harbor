# -*- coding: utf-8 -*-

from __future__ import absolute_import
import time

import unittest

from testutils import harbor_server, suppress_urllib3_warning
from testutils import ADMIN_CLIENT
from library.scan_data_export import Scan_data_export
from library.project import Project
from library.user import User
from library.artifact import Artifact
from library.scan import Scan
from library.repository import push_self_build_image_to_project


class TestScanDataExport(unittest.TestCase):

    @suppress_urllib3_warning
    def setUp(self):
        self.scan_data_export = Scan_data_export()
        self.project = Project()
        self.user = User()
        self.scan = Scan()
        self.artifact = Artifact()
        self.image = "alpine"
        self.tag = "latest"
        self.x_scan_data_type = "application/vnd.security.vulnerability.report; version=1.1"

    def testScanDataExportArtifact(self):
        """
        Test case:
            Scan Data Export API
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Push a new image(IA) in project(PA) by user(UA);
            4. Send scan image command and get tag(TA) information to check scan result, it should be finished;
            5. Verify trigger export scan data execution but does not specify Scan-Data-Type status code should be 422;
            6. Verify trigger export scan data execution but specifying multiple project status code should be 400;
            7. Trigger export scan data execution correctly;
            8. Verify that the export scan data execution triggered by the user(UA) cannot be queried by other users;
            9. User (UA) should be able to query the triggered export scan data execution;
            10. Wait for the export scan data execution to succeed;
            11. Verify that the export scan data execution triggered by the user (UA) cannot be download by other users;
            12. User (UA) should be able to download the triggered export scan data execution
            13. Verify that the downloaded export scan data execution cannot be downloaded again
            14. Verify the status message if no cve found or matched
        """
        url = ADMIN_CLIENT["endpoint"]
        user_password = "Aa123456"

        # 1. Create user(UA)
        user_id, user_name = self.user.create_user(user_password=user_password, **ADMIN_CLIENT)
        user_client = dict(endpoint=url, username=user_name, password=user_password)

        # 2.1. Create private project(PA) by user(UA)
        project_id, project_name = self.project.create_project(metadata={"public": "false"}, **user_client)
        # 2.2. Get private project of uesr-001, uesr-001 can see only one private project which is project-001
        self.project.projects_should_exist(dict(public=False), expected_count=1, expected_project_id=project_id, **user_client)

        # 3. Push a new image(IA) in project(PA) by user(UA)
        push_self_build_image_to_project(project_name, harbor_server, user_name, user_password, self.image, self.tag)

        # 4. Send scan image command and get tag(TA) information to check scan result, it should be finished
        self.scan.scan_artifact(project_name, self.image, self.tag, **user_client)
        self.artifact.check_image_scan_result(project_name, self.image, self.tag, with_scan_overview=True, **user_client)

        # 5. Verify trigger export scan data execution but does not specify Scan-Data-Type status code should be 422
        self.scan_data_export.export_scan_data("", projects=[project_id], expect_status_code=422, expect_response_body="X-Scan-Data-Type in header is required")

        # 6. Verify trigger export scan data execution but specifying multiple project status code should be 400
        self.scan_data_export.export_scan_data(self.x_scan_data_type, projects=[1, project_id], expect_status_code=400, expect_response_body="bad request: only support export single project")

        # 7. Trigger export scan data execution correctly
        execution_id = self.scan_data_export.export_scan_data(self.x_scan_data_type, projects=[project_id], **user_client).id
        print("execution_id:", execution_id)

        # 8.1. Verify that the export scan data execution triggered by the user(UA) cannot be queried by other users by get scan data export execution list API
        execution_list = self.scan_data_export.get_scan_data_export_execution_list()
        if not execution_list:
            self.assertNotEqual(execution_id, execution_list.items[0].id)
            self.assertEqual(ADMIN_CLIENT["username"], execution_list.items[0].user_name)

        # 8.2. Verify that the export scan data execution triggered by the user(UA) cannot be queried by other users by get scan_data export execution API
        self.scan_data_export.get_scan_data_export_execution(execution_id, expect_status_code=403, expect_response_body="FORBIDDEN")

        # 9. User (UA) should be able to query the triggered export scan data execution
        execution_list = self.scan_data_export.get_scan_data_export_execution_list(**user_client)
        self.assertEqual(execution_id, execution_list.items[0].id)
        self.assertEqual(user_name, execution_list.items[0].user_name)

        # 10. Wait for the export scan data execution to succeed
        execution = None
        for i in range(15):
            print("wait for the job to finish:", i)
            execution = self.scan_data_export.get_scan_data_export_execution(execution_id, **user_client)
            if execution.status == "Success":
                self.assertEqual(user_name, execution.user_name)
                self.assertEqual(user_id, execution.user_id)
                break
            time.sleep(2)
        self.assertEqual(execution.status, "Success")

        # 11. Verify that the export scan data execution triggered by the user (UA) cannot be download by other users
        self.scan_data_export.download_scan_data(execution_id, expect_status_code=403)

        # The csv file will not be able to downloaded if it is empty, so only check download if the file is present, otherwise check the status message
        if execution.file_present:
            # 12. User (UA) should be able to download the triggered export scan data execution
            self.scan_data_export.download_scan_data(execution_id, **user_client)

            # 13. Verify that the downloaded export scan data execution cannot be downloaded again
            self.scan_data_export.download_scan_data(execution_id, expect_status_code=404, **user_client)
        else:
            # 14. Verify the status message if no cve found or matched
            self.assertEqual("No vulnerabilities found or matched", execution.status_text)


if __name__ == '__main__':
    unittest.main()
