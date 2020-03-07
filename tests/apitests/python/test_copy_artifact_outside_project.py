from __future__ import absolute_import


import unittest

from library.base import _assert_status_code
from testutils import ADMIN_CLIENT
from testutils import harbor_server

from testutils import TEARDOWN
from library.artifact import Artifact
from library.project import Project
from library.user import User
from library.repository import Repository
from library.repository import push_image_to_project
from library.repository import pull_harbor_image

class TestProjects(unittest.TestCase):
    @classmethod
    def setUpClass(self):
        self.project = Project()
        self.user = User()
        self.artifact = Artifact(api_type='artifact')
        self.repo = Repository(api_type='repository')

    @classmethod
    def tearDownClass(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete repository(RA);
        self.repo.delete_repoitory(TestProjects.project_src_repo_name, (TestProjects.src_repo_name).split('/')[1], **TestProjects.USER_RETAG_CLIENT)

        #2. Delete repository by retag;
        self.repo.delete_repoitory(TestProjects.project_dst_repo_name, (TestProjects.dst_repo_name).split('/')[1], **TestProjects.USER_RETAG_CLIENT)

        #3. Delete project(PA);
        self.project.delete_project(TestProjects.project_src_repo_id, **TestProjects.USER_RETAG_CLIENT)
        self.project.delete_project(TestProjects.project_dst_repo_id, **TestProjects.USER_RETAG_CLIENT)

        #4. Delete user(UA).
        self.user.delete_user(TestProjects.user_retag_id, **ADMIN_CLIENT)

    def testRetag(self):
        """
        Test case:
            Retag Image
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Create a new project(PB) by user(UA);
            4. Update role of user-retag as guest member of project(PB);
            5. Create a new repository(RA) in project(PA) by user(UA);
            6. Get repository in project(PA), there should be one repository which was created by user(UA);
            7. Get repository(RA)'s image tag detail information;
            8. Retag image in project(PA) to project(PB), it should be forbidden;
            9. Update role of user-retag as admin member of project(PB);
            10. Retag image in project(PA) to project(PB), it should be successful;
            11. Get repository(RB)'s image tag detail information;
            12. Read digest of retaged image, it must be the same with the image in repository(RA);
            13. Pull image from project(PB) by user_retag, it must be successful;
        Tear down:
            1. Delete repository(RA);
            2. Delete repository by retag;
            3. Delete project(PA);
            4. Delete user(UA).
        """
        url = ADMIN_CLIENT["endpoint"]
        user_retag_password = "Aa123456"
        pull_tag_name = "latest"
        dst_repo_sub_name = "repo"

        #1. Create a new user(UA);
        TestProjects.user_retag_id, user_retag_name = self.user.create_user(user_password = user_retag_password, **ADMIN_CLIENT)

        TestProjects.USER_RETAG_CLIENT=dict(endpoint = url, username = user_retag_name, password = user_retag_password)

        #2. Create a new project(PA) by user(UA);
        TestProjects.project_src_repo_id, TestProjects.project_src_repo_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_RETAG_CLIENT)

        #3. Create a new project(PB) by user(UA);
        TestProjects.project_dst_repo_id, TestProjects.project_dst_repo_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_RETAG_CLIENT)

        retag_member_id = self.project.get_project_member_id(TestProjects.project_dst_repo_id, user_retag_name, **TestProjects.USER_RETAG_CLIENT)

        #4. Update role of user-retag as guest member of project(PB);
        self.project.update_project_member_role(TestProjects.project_dst_repo_id, retag_member_id, 3, **ADMIN_CLIENT)

        #5. Create a new repository(RA) in project(PA) by user(UA);
        TestProjects.src_repo_name, tag_name = push_image_to_project(TestProjects.project_src_repo_name, harbor_server, 'admin', 'Harbor12345', "hello-world", pull_tag_name)

        #6. Get repository in project(PA), there should be one repository which was created by user(UA);
        src_repo_data = self.repo.list_repositories(TestProjects.project_src_repo_name, **TestProjects.USER_RETAG_CLIENT)
        _assert_status_code(TestProjects.src_repo_name, src_repo_data[0].name)

        #7. Get repository(RA)'s image tag detail information;
        src_tag_data = self.artifact.get_reference_info(TestProjects.project_src_repo_name, TestProjects.src_repo_name.split('/')[1], tag_name, **TestProjects.USER_RETAG_CLIENT)
        TestProjects.dst_repo_name = TestProjects.project_dst_repo_name+"/"+ dst_repo_sub_name
        #8. Retag image in project(PA) to project(PB), it should be forbidden;
        self.artifact.copy_artifact(TestProjects.project_dst_repo_name, dst_repo_sub_name, TestProjects.src_repo_name+"@"+src_tag_data[0].digest, expect_status_code=403, **TestProjects.USER_RETAG_CLIENT)

        #9. Update role of user-retag as admin member of project(PB);
        self.project.update_project_member_role(TestProjects.project_dst_repo_id, retag_member_id, 1, **ADMIN_CLIENT)

        #10. Retag image in project(PA) to project(PB), it should be successful;
        self.artifact.copy_artifact(TestProjects.project_dst_repo_name, dst_repo_sub_name, TestProjects.src_repo_name+"@"+src_tag_data[0].digest, **TestProjects.USER_RETAG_CLIENT)

        #11. Get repository(RB)'s image tag detail information;
        dst_tag_data = self.artifact.get_reference_info(TestProjects.project_dst_repo_name, dst_repo_sub_name, tag_name, **TestProjects.USER_RETAG_CLIENT)

        #12. Read digest of retaged image, it must be the same with the image in repository(RA);
        self.assertEqual(src_tag_data[0].digest, dst_tag_data[0].digest)

        #13. Pull image from project(PB) by user_retag, it must be successful;"
        pull_harbor_image(harbor_server, user_retag_name, user_retag_password, TestProjects.dst_repo_name, tag_name)


if __name__ == '__main__':
    unittest.main()

