from __future__ import absolute_import


import unittest

import library.repository
import library.docker_api
from library.base import _assert_status_code
from testutils import ADMIN_CLIENT
from testutils import harbor_server

from testutils import TEARDOWN
from library.project import Project
from library.user import User
from library.repository import Repository
from library.artifact import Artifact
from library.repository import push_image_to_project
from library.repository import pull_harbor_image

class TestProjects(unittest.TestCase):
    @classmethod
    def setUpClass(self):
        self.project= Project()
        self.user= User()
        self.artifact = Artifact()
        self.repo= Repository()
        self.url = ADMIN_CLIENT["endpoint"]
        self.user_password = "Aa123456"
        self.repo_name = "hello-world"

    @classmethod
    def tearDownClass(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete repository(RA,IA) by user(UA);
        self.repo.delete_repoitory(TestProjects.project_name, self.repo_name, **TestProjects.USER_CLIENT)

        #2. Delete project(PA);
        self.project.delete_project(TestProjects.project_id, **TestProjects.USER_CLIENT)

        #3. Delete user(UA).
        self.user.delete_user(TestProjects.user_id, **ADMIN_CLIENT)

    def testCreateDeleteTag(self):
        """
        Test case:
            Create/Delete tag
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Push an image(IA) to Harbor by docker successfully;
            4. Create a tag(1.0) for the image(IA);
            5. Get the image(latest) from Harbor successfully;
            6. Verify the image(IA) contains tag named 1.0;
            7. Delete the tag(1.0) from image(IA);
            8. Get the image(IA) from Harbor successfully;
            9. Verify the image(IA) contains no tag named 1.0;
        Tear down:
            1. Delete repository(RA,IA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA).
        """
        #1. Create a new user(UA);
        TestProjects.user_id, user_name = self.user.create_user(user_password = self.user_password, **ADMIN_CLIENT)

        TestProjects.USER_CLIENT=dict(with_tag = True, endpoint = self.url, username = user_name, password = self.user_password)

        #2. Create a new project(PA) by user(UA);
        TestProjects.project_id, TestProjects.project_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_CLIENT)

        #3. Push an image(IA) to Harbor by docker successfully;
        repo_name, tag = push_image_to_project(TestProjects.project_name, harbor_server, 'admin', 'Harbor12345', self.repo_name, "latest")

        #4. Create a tag(1.0) for the image(IA)
        self.artifact.create_tag(TestProjects.project_name, self.repo_name, tag, "1.0",**TestProjects.USER_CLIENT)

        #5. Get the image(IA) from Harbor successfully;
        artifact = self.artifact.get_reference_info(TestProjects.project_name, self.repo_name, tag, **TestProjects.USER_CLIENT)

        #6. Verify the image(IA) contains tag named 1.0;
        self.assertEqual(artifact[0].tags[0].name, tag)
        self.assertEqual(artifact[0].tags[1].name, "1.0")

        #7. Delete the tag(1.0) from image(IA);
        self.artifact.delete_tag(TestProjects.project_name, self.repo_name, tag, "1.0",**TestProjects.USER_CLIENT)

        #8. Get the image(latest) from Harbor successfully;
        artifact = self.artifact.get_reference_info(TestProjects.project_name, self.repo_name, tag, **TestProjects.USER_CLIENT)

        #9. Verify the image(IA) contains no tag named 1.0;
        self.assertEqual(artifact[0].tags[0].name, tag)



if __name__ == '__main__':
    unittest.main()

