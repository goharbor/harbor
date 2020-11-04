from __future__ import absolute_import
import unittest
import urllib

from testutils import ADMIN_CLIENT, suppress_urllib3_warning
from testutils import harbor_server
from testutils import TEARDOWN
import library.oras
from library.sign import sign_image
from library.user import User
from library.project import Project
from library.repository import Repository
from library.artifact import Artifact


class TestProjects(unittest.TestCase):
    @suppress_urllib3_warning
    def setUp(self):
        self.project = Project()
        self.user = User()
        self.artifact = Artifact()
        self.repo = Repository()
        self.repo_name = "hello-artifact"
        self.tag = "test_v2"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def tearDown(self):
        #1. Delete user(UA);
        self.user.delete_user(TestProjects.user_sign_image_id, **ADMIN_CLIENT)

    def testOrasCli(self):
        """
        Test case:
            Push Artifact With ORAS CLI
        Test step and expected result:
            1. Create user-001
            2. Create a new private project(PA) by user(UA);
            3. ORAS CLI push artifacts;
            4. Get repository from Harbor successfully, and verfiy repository name is repo pushed by ORAS CLI;
            5. Get and verify artifacts by tag;
            6. ORAS CLI pull artifacts index by tag;
            7. Verfiy MD5 between artifacts pushed by ORAS and artifacts pulled by ORAS;
        Tear down:
            NA
        """
        url = ADMIN_CLIENT["endpoint"]
        user_001_password = "Aa123456"

        #1. Create user-001
        TestProjects.user_sign_image_id, user_name = self.user.create_user(user_password = user_001_password, **ADMIN_CLIENT)

        TestProjects.USER_CLIENT=dict(with_signature = True, endpoint = url, username = user_name, password = user_001_password)

        #2. Create a new private project(PA) by user(UA);
        TestProjects.project_id, TestProjects.project_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_CLIENT)

        #3. ORAS CLI push artifacts;
        md5_list_push = library.oras.oras_push(harbor_server, user_name, user_001_password, TestProjects.project_name, self.repo_name, self.tag)

        #4. Get repository from Harbor successfully, and verfiy repository name is repo pushed by ORAS CLI;
        repo_data = self.repo.get_repository(TestProjects.project_name, self.repo_name, **TestProjects.USER_CLIENT)
        self.assertEqual(repo_data.name, TestProjects.project_name + "/" + self.repo_name)

        #5. Get and verify artifacts by tag;
        artifact = self.artifact.get_reference_info(TestProjects.project_name, self.repo_name, self.tag, **TestProjects.USER_CLIENT)
        self.assertEqual(artifact.tags[0].name, self.tag)

        #6. ORAS CLI pull artifacts index by tag;
        md5_list_pull = library.oras.oras_pull(harbor_server, user_name, user_001_password, TestProjects.project_name, self.repo_name, self.tag)

        #7. Verfiy MD5 between artifacts pushed by ORAS and artifacts pulled by ORAS;
        if set(md5_list_push) != set(md5_list_pull):
            raise Exception(r"MD5 check failed with {} and {}.".format(str(md5_list_push), str(md5_list_pull)))

if __name__ == '__main__':
    unittest.main()
