from __future__ import absolute_import
import unittest

from testutils import harbor_server, suppress_urllib3_warning
from testutils import TEARDOWN
from testutils import ADMIN_CLIENT
from library.project import Project
from library.user import User
from library.repository import Repository
from library.repository import push_self_build_image_to_project
from library.artifact import Artifact
from library.scan_all import ScanAll
from library.scan_all_stop import StopScanAll

class TestStopScanAll(unittest.TestCase):
    @suppress_urllib3_warning
    def setUp(self):
        self.project= Project()
        self.user= User()
        self.artifact = Artifact()
        self.repo = Repository()
        self.scan_all = ScanAll()
        self.stop_scan_all = StopScanAll()

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def tearDown(self):
        #1. Delete Alice's repository and Luca's repository;
        self.repo.delete_repository(TestStopScanAll.project_Alice_name, TestStopScanAll.repo_Alice_name.split('/')[1], **ADMIN_CLIENT)
        self.repo.delete_repository(TestStopScanAll.project_Luca_name, TestStopScanAll.repo_Luca_name.split('/')[1], **ADMIN_CLIENT)

        #2. Delete Alice's project and Luca's project;
        self.project.delete_project(TestStopScanAll.project_Alice_id, **ADMIN_CLIENT)
        self.project.delete_project(TestStopScanAll.project_Luca_id, **ADMIN_CLIENT)

        #3. Delete user Alice and Luca.
        self.user.delete_user(TestStopScanAll.user_Alice_id, **ADMIN_CLIENT)
        self.user.delete_user(TestStopScanAll.user_Luca_id, **ADMIN_CLIENT)
        print("Case completed")

    def testSystemLevelScanALL(self):
        """
        Test case:
            System level Stop Scan All
        Test step and expected result:
            1. Create user Alice and Luca;
            2. Create 2 new private projects project_Alice and project_Luca;
            3. Push a image to project_Alice and push another image to project_Luca;
            4. Trigger scan all event;
            5. Send stop scan all request.
        Tear down:
            1. Delete Alice's repository and Luca's repository;
            2. Delete Alice's project and Luca's project;
            3. Delete user Alice and Luca.
        """
        url = ADMIN_CLIENT["endpoint"]
        user_common_password = "Aa123456"

        #1. Create user Alice and Luca;
        TestStopScanAll.user_Alice_id, user_Alice_name = self.user.create_user(user_password = user_common_password, **ADMIN_CLIENT)
        TestStopScanAll.user_Luca_id, user_Luca_name = self.user.create_user(user_password = user_common_password, **ADMIN_CLIENT)

        USER_ALICE_CLIENT=dict(endpoint = url, username = user_Alice_name, password = user_common_password, with_scan_overview = True)
        USER_LUCA_CLIENT=dict(endpoint = url, username = user_Luca_name, password = user_common_password, with_scan_overview = True)

        #2. Create 2 new private projects project_Alice and project_Luca;
        TestStopScanAll.project_Alice_id, TestStopScanAll.project_Alice_name = self.project.create_project(metadata = {"public": "false"}, **USER_ALICE_CLIENT)
        TestStopScanAll.project_Luca_id, TestStopScanAll.project_Luca_name = self.project.create_project(metadata = {"public": "false"}, **USER_LUCA_CLIENT)

        #3. Push a image to project_Alice and push another image to project_Luca;

        #Note: Please make sure that this Image has never been pulled before by any other cases,
        #          so it is a not-scanned image rigth after repository creation.
        #image = "tomcat"
        image_a = "mariadb"
        src_tag = "latest"
        #3.1 Push a image to project_Alice;
        TestStopScanAll.repo_Alice_name, tag_Alice = push_self_build_image_to_project(TestStopScanAll.project_Alice_name, harbor_server, user_Alice_name, user_common_password, image_a, src_tag)

        #Note: Please make sure that this Image has never been pulled before by any other cases,
        #          so it is a not-scanned image rigth after repository creation.
        image_b = "httpd"
        src_tag = "latest"
        #3.2 push another image to project_Luca;
        TestStopScanAll.repo_Luca_name, tag_Luca = push_self_build_image_to_project(TestStopScanAll.project_Luca_name, harbor_server, user_Luca_name, user_common_password, image_b, src_tag)

        #4. Trigger scan all event;
        self.scan_all.scan_all_now(**ADMIN_CLIENT)
        # self.scan_all.wait_until_scans_all_finish(**ADMIN_CLIENT)

        #5. Send stop scan all request.
        self.stop_scan_all.stop_scan_all()

if __name__ == '__main__':
    unittest.main()