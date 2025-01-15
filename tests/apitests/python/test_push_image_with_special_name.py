from __future__ import absolute_import
import unittest
import urllib

from testutils import ADMIN_CLIENT, suppress_urllib3_warning
from testutils import harbor_server
from testutils import TEARDOWN
from library.artifact import Artifact
from library.project import Project
from library.user import User
from library.repository import Repository
from library.repository import push_self_build_image_to_project

class TestProjects(unittest.TestCase):
    @suppress_urllib3_warning
    def setUp(self):
        self.project = Project()
        self.user = User()
        self.artifact = Artifact()
        self.repo = Repository()

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def tearDown(self):
        # remove the deletion as the signed image cannot be deleted.
        #1. Delete repository(RA) by user(UA);
        #self.repo.delete_repository(TestProjects.project_sign_image_name, TestProjects.repo_name.split('/')[1], **TestProjects.USER_sign_image_CLIENT)

        #2. Delete project(PA);
        #self.project.delete_project(TestProjects.project_sign_image_id, **TestProjects.USER_sign_image_CLIENT)

        #3. Delete user(UA);
        self.user.delete_user(TestProjects.user_sign_image_id, **ADMIN_CLIENT)

    def testPushImageWithSpecialName(self):
        """
        Test case:
            Push Image With Special Name
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new private project(PA) by user(UA);
            3. Add user(UA) as a member of project(PA) with project-admin role;
            4. Get private project of user(UA), user(UA) can see only one private project which is project(PA);
            5. Create a new repository(RA) and tag(TA) in project(PA) by user(UA);
            6. Sign image with tag(TA) which was tagged by step #5;
            7. Get signature of image with tag(TA), it should be exist.
        Tear down:
            NA
        """
        url = ADMIN_CLIENT["endpoint"]
        user_001_password = "Aa123456"

        #1. Create user-001
        TestProjects.user_sign_image_id, user_sign_image_name = self.user.create_user(user_password = user_001_password, **ADMIN_CLIENT)

        TestProjects.USER_sign_image_CLIENT=dict(endpoint = url, username = user_sign_image_name, password = user_001_password)

        #2. Create a new private project(PA) by user(UA);
        TestProjects.project_sign_image_id, TestProjects.project_sign_image_name = self.project.create_project(metadata = {"public": "false"}, **ADMIN_CLIENT)

        #3. Add user(UA) as a member of project(PA) with project-admin role;
        self.project.add_project_members(TestProjects.project_sign_image_id, user_id=TestProjects.user_sign_image_id, **ADMIN_CLIENT)

        #4. Get private project of user(UA), user(UA) can see only one private project which is project(PA);
        self.project.projects_should_exist(dict(public=False), expected_count = 1,
            expected_project_id = TestProjects.project_sign_image_id, **TestProjects.USER_sign_image_CLIENT)

        image = "redis"
        src_tag = "latest"
        profix = "aaa/bbb"

        #5. Create a new repository(RA) and tag(TA) in project(PA) by user(UA);
        TestProjects.repo_name, tag = push_self_build_image_to_project(TestProjects.project_sign_image_name, harbor_server, user_sign_image_name, user_001_password, profix+"/"+image, src_tag)
        #7. Get signature of image with tag(TA), it should be exist.
        full_name = urllib.parse.quote(profix+"/"+image,'utf-8')

        artifact = self.artifact.get_reference_info(TestProjects.project_sign_image_name, full_name, tag, **TestProjects.USER_sign_image_CLIENT)
        self.assertEqual(artifact.type, 'IMAGE')

if __name__ == '__main__':
    unittest.main()
