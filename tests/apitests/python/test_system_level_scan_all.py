from __future__ import absolute_import
import unittest

from testutils import harbor_server
from testutils import TEARDOWN
from testutils import ADMIN_CLIENT
from library.system import System
from library.project import Project
from library.user import User
from library.repository import Repository
from library.repository import push_image_to_project
from library.artifact import Artifact

class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        self.system = System()
        self.project= Project()
        self.user= User()
        self.artifact = Artifact(api_type='artifact')
        self.repo = Repository(api_type='repository')

    @classmethod
    def tearDown(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete Alice's repository and Luca's repository;
        self.repo.delete_repoitory(TestProjects.project_Alice_name, TestProjects.repo_Alice_name.split('/')[1], **ADMIN_CLIENT)
        self.repo.delete_repoitory(TestProjects.project_Luca_name, TestProjects.repo_Luca_name.split('/')[1], **ADMIN_CLIENT)

        #2. Delete Alice's project and Luca's project;
        self.project.delete_project(TestProjects.project_Alice_id, **ADMIN_CLIENT)
        self.project.delete_project(TestProjects.project_Luca_id, **ADMIN_CLIENT)

        #3. Delete user Alice and Luca.
        self.user.delete_user(TestProjects.user_Alice_id, **ADMIN_CLIENT)
        self.user.delete_user(TestProjects.user_Luca_id, **ADMIN_CLIENT)

    def testSystemLevelScanALL(self):
        """
        Test case:
            System level Scan All
        Test step and expected result:
            1. Create user Alice and Luca;
            2. Create 2 new private projects project_Alice and project_Luca;
            3. Push a image to project_Alice and push another image to project_Luca;
            4. Trigger scan all event;
            5. Check if image in project_Alice and another image in project_Luca were both scanned.
        Tear down:
            1. Delete Alice's repository and Luca's repository;
            2. Delete Alice's project and Luca's project;
            3. Delete user Alice and Luca.
        """
        url = ADMIN_CLIENT["endpoint"]
        user_common_password = "Aa123456"

        #1. Create user Alice and Luca;
        TestProjects.user_Alice_id, user_Alice_name = self.user.create_user(user_password = user_common_password, **ADMIN_CLIENT)
        TestProjects.user_Luca_id, user_Luca_name = self.user.create_user(user_password = user_common_password, **ADMIN_CLIENT)

        USER_ALICE_CLIENT=dict(endpoint = url, username = user_Alice_name, password = user_common_password, with_scan_overview = True)
        USER_LUCA_CLIENT=dict(endpoint = url, username = user_Luca_name, password = user_common_password, with_scan_overview = True)

        #2. Create 2 new private projects project_Alice and project_Luca;
        TestProjects.project_Alice_id, TestProjects.project_Alice_name = self.project.create_project(metadata = {"public": "false"}, **USER_ALICE_CLIENT)
        TestProjects.project_Luca_id, TestProjects.project_Luca_name = self.project.create_project(metadata = {"public": "false"}, **USER_LUCA_CLIENT)

        #3. Push a image to project_Alice and push another image to project_Luca;

        #Note: Please make sure that this Image has never been pulled before by any other cases,
        #          so it is a not-scanned image rigth after repository creation.
        #image = "tomcat"
        image_a = "mariadb"
        src_tag = "latest"
        #3.1 Push a image to project_Alice;
        TestProjects.repo_Alice_name, tag_Alice = push_image_to_project(TestProjects.project_Alice_name, harbor_server, user_Alice_name, user_common_password, image_a, src_tag)

        #Note: Please make sure that this Image has never been pulled before by any other cases,
        #          so it is a not-scanned image rigth after repository creation.
        image_b = "httpd"
        src_tag = "latest"
        #3.2 push another image to project_Luca;
        TestProjects.repo_Luca_name, tag_Luca = push_image_to_project(TestProjects.project_Luca_name, harbor_server, user_Luca_name, user_common_password, image_b, src_tag)

        #4. Trigger scan all event;
        self.system.scan_now(**ADMIN_CLIENT)

        #5. Check if image in project_Alice and another image in project_Luca were both scanned.
        #self.repo.check_image_scan_result(TestProjects.repo_Alice_name, tag_Alice, **USER_ALICE_CLIENT)
        #self.repo.check_image_scan_result(TestProjects.repo_Luca_name, tag_Luca, **USER_LUCA_CLIENT)
        self.artifact.check_image_scan_result(TestProjects.project_Alice_name, image_a, tag_Alice, **USER_ALICE_CLIENT)
        self.artifact.check_image_scan_result(TestProjects.project_Luca_name, image_b, tag_Luca, **USER_LUCA_CLIENT)

if __name__ == '__main__':
    unittest.main()