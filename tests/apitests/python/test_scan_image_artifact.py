from __future__ import absolute_import
import unittest

from testutils import harbor_server
from testutils import TEARDOWN
from testutils import ADMIN_CLIENT
from library.project import Project
from library.user import User
from library.repository import Repository
from library.repository import push_image_to_project
from library.artifact import Artifact
from library.scan import Scan
class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        self.project= Project()
        self.user= User()
        self.artifact = Artifact(api_type='artifact')
        self.repo = Repository(api_type='repository')
        self.scan = Scan(api_type='scan')

    @classmethod
    def tearDown(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete repository(RA) by user(UA);
        self.repo.delete_repoitory(TestProjects.project_scan_image_name, TestProjects.repo_name.split('/')[1], **TestProjects.USER_SCAN_IMAGE_CLIENT)

        #2. Delete project(PA);
        self.project.delete_project(TestProjects.project_scan_image_id, **TestProjects.USER_SCAN_IMAGE_CLIENT)

        #3. Delete user(UA);
        self.user.delete_user(TestProjects.user_scan_image_id, **ADMIN_CLIENT)

    def testScanImageArtifact(self):
        """
        Test case:
            Scan An Image Artifact
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new private project(PA) by user(UA);
            3. Add user(UA) as a member of project(PA) with project-admin role;
            4. Get private project of user(UA), user(UA) can see only one private project which is project(PA);
            5. Create a new repository(RA) and tag(TA) in project(PA) by user(UA);
            6. Send scan image command and get tag(TA) information to check scan result, it should be finished;
        Tear down:
            1. Delete repository(RA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA);
        """
        url = ADMIN_CLIENT["endpoint"]
        user_001_password = "Aa123456"

        #1. Create user-001
        TestProjects.user_scan_image_id, user_scan_image_name = self.user.create_user(user_password = user_001_password, **ADMIN_CLIENT)

        TestProjects.USER_SCAN_IMAGE_CLIENT=dict(endpoint = url, username = user_scan_image_name, password = user_001_password, with_scan_overview = True)

        #2. Create a new private project(PA) by user(UA);
        TestProjects.project_scan_image_id, TestProjects.project_scan_image_name = self.project.create_project(metadata = {"public": "false"}, **ADMIN_CLIENT)

        #3. Add user(UA) as a member of project(PA) with project-admin role;
        self.project.add_project_members(TestProjects.project_scan_image_id, TestProjects.user_scan_image_id, **ADMIN_CLIENT)

        #4. Get private project of user(UA), user(UA) can see only one private project which is project(PA);
        self.project.projects_should_exist(dict(public=False), expected_count = 1,
            expected_project_id = TestProjects.project_scan_image_id, **TestProjects.USER_SCAN_IMAGE_CLIENT)

        #Note: Please make sure that this Image has never been pulled before by any other cases,
        #          so it is a not-scanned image right after repository creation.
        #image = "tomcat"
        image = "docker"
        src_tag = "1.13"
        #5. Create a new repository(RA) and tag(TA) in project(PA) by user(UA);
        TestProjects.repo_name, tag = push_image_to_project(TestProjects.project_scan_image_name, harbor_server, user_scan_image_name, user_001_password, image, src_tag)

        #6. Send scan image command and get tag(TA) information to check scan result, it should be finished;
        self.scan.scan_artifact(TestProjects.project_scan_image_name, TestProjects.repo_name.split('/')[1], tag, **TestProjects.USER_SCAN_IMAGE_CLIENT)

        #6. Send scan image command and get tag(TA) information to check scan result, it should be finished;
        self.artifact.check_image_scan_result(TestProjects.project_scan_image_name, image, tag, **TestProjects.USER_SCAN_IMAGE_CLIENT)

if __name__ == '__main__':
    unittest.main()