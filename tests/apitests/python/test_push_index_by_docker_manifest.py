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
        self.artifact = Artifact(api_type='artifact')
        self.repo= Repository(api_type='repository')
        self.url = ADMIN_CLIENT["endpoint"]
        self.user_push_index_password = "Aa123456"
        self.index_name = "ci_test_index"
        self.index_tag = "test_tag"
        self.image_a = "alpine"
        self.image_b = "busybox"

    @classmethod
    def tearDownClass(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete repository(RA,RB,IA) by user(UA);
        self.repo.delete_repoitory(TestProjects.project_push_index_name, self.index_name, **TestProjects.USER_CLIENT)
        self.repo.delete_repoitory(TestProjects.project_push_index_name, self.image_a, **TestProjects.USER_CLIENT)
        self.repo.delete_repoitory(TestProjects.project_push_index_name, self.image_b, **TestProjects.USER_CLIENT)

        #2. Delete project(PA);
        self.project.delete_project(TestProjects.project_push_index_id, **TestProjects.USER_CLIENT)

        #3. Delete user(UA).
        self.user.delete_user(TestProjects.user_id, **ADMIN_CLIENT)

    def testAddIndexByDockerManifest(self):
        """
        Test case:
            Push Index By Docker Manifest
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Create 2 new repositorys(RA,RB) in project(PA) by user(UA);
            4. Push an index(IA) to Harbor by docker manifest CLI successfully;
            5. Get index(IA) from Harbor successfully;
            6. Verify harbor index is index(IA) pushed by docker manifest CLI;
            7. Verify harbor index(IA) can be pulled by docker manifest CLI successfully.
        Tear down:
            1. Delete repository(RA,RB,IA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA).
        """
        #1. Create a new user(UA);
        TestProjects.user_id, user_name = self.user.create_user(user_password = self.user_push_index_password, **ADMIN_CLIENT)

        TestProjects.USER_CLIENT=dict(endpoint = self.url, username = user_name, password = self.user_push_index_password)

        #2. Create a new project(PA) by user(UA);
        TestProjects.project_push_index_id, TestProjects.project_push_index_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_CLIENT)

        #3. Create 2 new repositorys(RA,RB) in project(PA) by user(UA);
        repo_name_a, tag_a = push_image_to_project(TestProjects.project_push_index_name, harbor_server, 'admin', 'Harbor12345', self.image_a, "latest")
        repo_name_b, tag_b = push_image_to_project(TestProjects.project_push_index_name, harbor_server, 'admin', 'Harbor12345', self.image_b, "latest")

        #4. Push an index(IA) to Harbor by docker manifest CLI successfully;
        manifests = [harbor_server+"/"+repo_name_a+":"+tag_a, harbor_server+"/"+repo_name_b+":"+tag_b]
        index = harbor_server+"/"+TestProjects.project_push_index_name+"/"+self.index_name+":"+self.index_tag
        index_sha256_cli_ret, manifests_sha256_cli_ret = library.docker_api.docker_manifest_push_to_harbor(index, manifests, harbor_server, user_name, self.user_push_index_password)

        #5. Get index(IA) from Harbor successfully;
        index_data = self.artifact.get_reference_info(TestProjects.project_push_index_name, self.index_name, self.index_tag, **TestProjects.USER_CLIENT)
        manifests_sha256_harbor_ret = [index_data[0].references[1].child_digest, index_data[0].references[0].child_digest]

        #6. Verify harbor index is index(IA) pushed by docker manifest CLI;
        self.assertEqual(index_data[0].digest, index_sha256_cli_ret)
        self.assertEqual(manifests_sha256_harbor_ret.count(manifests_sha256_cli_ret[0]), 1)
        self.assertEqual(manifests_sha256_harbor_ret.count(manifests_sha256_cli_ret[1]), 1)

        #7. Verify harbor index(IA) can be pulled by docker manifest CLI successfully;
        pull_harbor_image(harbor_server, user_name, self.user_push_index_password, TestProjects.project_push_index_name+"/"+self.index_name, self.index_tag)

if __name__ == '__main__':
    unittest.main()

