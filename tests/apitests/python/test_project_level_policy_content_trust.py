from __future__ import absolute_import

import unittest

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
    def setUp(self):
        self.project= Project()
        self.user= User()
        self.artifact= Artifact(api_type='artifact')
        self.repo= Repository(api_type='repository')

    @classmethod
    def tearDown(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete repository(RA) by user(UA);
        self.repo.delete_repoitory(TestProjects.project_content_trust_name, TestProjects.repo_name.split('/')[1], **TestProjects.USER_CONTENT_TRUST_CLIENT)

        #2. Delete project(PA);
        self.project.delete_project(TestProjects.project_content_trust_id, **TestProjects.USER_CONTENT_TRUST_CLIENT)

        #3. Delete user(UA);
        self.user.delete_user(TestProjects.user_content_trust_id, **ADMIN_CLIENT)

    def testProjectLevelPolicyContentTrust(self):
        """
        Test case:
            Project Level Policy Content Trust
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Push a new image(IA) in project(PA) by admin;
            4. Image(IA) should exist;
            5. Pull image(IA) successfully;
            6. Enable content trust in project(PA) configuration;
            7. Pull image(IA) failed and the reason is "The image is not signed in Notary".
        Tear down:
            1. Delete repository(RA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA);
        """
        url = ADMIN_CLIENT["endpoint"]
        image = "hello-world"
        admin_name = ADMIN_CLIENT["username"]
        admin_password = ADMIN_CLIENT["password"]
        user_content_trust_password = "Aa123456"

        #1. Create a new user(UA);
        TestProjects.user_content_trust_id, user_content_trust_name = self.user.create_user(user_password = user_content_trust_password, **ADMIN_CLIENT)

        TestProjects.USER_CONTENT_TRUST_CLIENT=dict(endpoint = url, username = user_content_trust_name, password = user_content_trust_password)

        #2. Create a new project(PA) by user(UA);
        TestProjects.project_content_trust_id, TestProjects.project_content_trust_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_CONTENT_TRUST_CLIENT)

        #3. Push a new image(IA) in project(PA) by admin;
        TestProjects.repo_name, tag = push_image_to_project(TestProjects.project_content_trust_name, harbor_server, admin_name, admin_password, image, "latest")

        #4. Image(IA) should exist;
        artifact = self.artifact.get_reference_info(TestProjects.project_content_trust_name, image, tag, **TestProjects.USER_CONTENT_TRUST_CLIENT)
        self.assertEqual(artifact[0].tags[0].name, tag)

        #5. Pull image(IA) successfully;
        pull_harbor_image(harbor_server, admin_name, admin_password, TestProjects.repo_name, tag)

        #6. Enable content trust in project(PA) configuration;
        self.project.update_project(TestProjects.project_content_trust_id, metadata = {"enable_content_trust": "true"}, **TestProjects.USER_CONTENT_TRUST_CLIENT)

        #7. Pull image(IA) failed and the reason is "The image is not signed in Notary".
        pull_harbor_image(harbor_server, admin_name, admin_password, TestProjects.repo_name, tag, expected_error_message = "The image is not signed in Notary")

if __name__ == '__main__':
    unittest.main()

