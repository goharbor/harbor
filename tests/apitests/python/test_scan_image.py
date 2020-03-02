from __future__ import absolute_import
import unittest

from testutils import harbor_server
from testutils import TEARDOWN
from testutils import ADMIN_CLIENT
from library.project import Project
from library.user import User
from library.repository import Repository
from library.repository import push_image_to_project

class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        self.project= Project()
        self.user= User()
        self.repo= Repository()

    @classmethod
    def tearDown(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == True, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete repository(RA) by user(UA);
        self.repo.delete_repoitory(TestProjects.repo_name, **TestProjects.USER_SCAN_IMAGE_CLIENT)

        #2. Delete project(PA);
        self.project.delete_project(TestProjects.project_scan_image_id, **TestProjects.USER_SCAN_IMAGE_CLIENT)

        #3. Delete user(UA);
        self.user.delete_user(TestProjects.user_scan_image_id, **ADMIN_CLIENT)

    def testScanImage(self):
        """
        Test case:
            Scan A Image
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

        TestProjects.USER_SCAN_IMAGE_CLIENT=dict(endpoint = url, username = user_scan_image_name, password = user_001_password)

        #2. Create a new private project(PA) by user(UA);
        TestProjects.project_scan_image_id, project_scan_image_name = self.project.create_project(metadata = {"public": "false"}, **ADMIN_CLIENT)

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
        TestProjects.repo_name, tag = push_image_to_project(project_scan_image_name, harbor_server, user_scan_image_name, user_001_password, image, src_tag)

        #6. Send scan image command and get tag(TA) information to check scan result, it should be finished;
        self.repo.scan_image(TestProjects.repo_name, tag, **TestProjects.USER_SCAN_IMAGE_CLIENT)
        self.repo.check_image_scan_result(TestProjects.repo_name, tag, **TestProjects.USER_SCAN_IMAGE_CLIENT)

if __name__ == '__main__':
    unittest.main()