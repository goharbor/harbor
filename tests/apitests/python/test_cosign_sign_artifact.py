# -*- coding: utf-8 -*-

from __future__ import absolute_import
import unittest

from testutils import harbor_server, suppress_urllib3_warning
from library import cosign
from testutils import ADMIN_CLIENT
from testutils import TEARDOWN
from library.project import Project
from library.user import User
from library.repository import Repository
from library.repository import push_self_build_image_to_project
from library import docker_api
from library.artifact import Artifact

class TestCosign(unittest.TestCase):

    @suppress_urllib3_warning
    def setUp(self):
        self.project= Project()
        self.user= User()
        self.artifact = Artifact()
        self.repo = Repository()
        self.image = "alpine"
        self.tag = "latest"
        self.expect_accessory_type = "signature.cosign"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def tearDown(self):
        #1. Delete repository by user(UA);
        self.repo.delete_repository(TestCosign.project_name, self.image, **TestCosign.user_client)
        #2. Delete project(PA);
        self.project.delete_project(TestCosign.project_id, **TestCosign.user_client)
        #3. Delete user(UA).
        self.user.delete_user(TestCosign.user_id, **ADMIN_CLIENT)

    def testCosignArtifact(self):
        """
        Test case:
            Cosign Artifact API
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Push a new image(IA) in project(PA) by user(UA);
            4. Verify that the image (IA) is not signed by cosign;
            5. Sign image(IA) with cosign;
            6. Verify that the image (IA) is signed by cosign;
        Tear down:
            1. Delete project(PA);
            2. Delete user(UA).
        """
        url = ADMIN_CLIENT["endpoint"]
        user_password = "Aa123456"

        # 1. Create user(UA)
        TestCosign.user_id, user_name = self.user.create_user(user_password = user_password, **ADMIN_CLIENT)
        TestCosign.user_client = dict(endpoint = url, username = user_name, password = user_password, with_accessory = True)

        # 2.1. Create private project(PA) by user(UA)
        TestCosign.project_id, TestCosign.project_name = self.project.create_project(metadata = {"public": "false"}, **TestCosign.user_client)
        # 2.2. Get private project of uesr-001, uesr-001 can see only one private project which is project-001
        self.project.projects_should_exist(dict(public=False), expected_count = 1, expected_project_id = TestCosign.project_id, **TestCosign.user_client)

        # 3. Push a new image(IA) in project(PA) by user(UA)
        TestCosign.repo_name, tag = push_self_build_image_to_project(TestCosign.project_name, harbor_server, user_name, user_password, self.image, self.tag)

        # 4.1. Verify list_artifacts API;
        artifact_list = self.artifact.list_artifacts(TestCosign.project_name, self.image, **TestCosign.user_client)
        first_artifact = artifact_list[0]
        artifact_reference = first_artifact.digest
        self.assertTrue(len(artifact_list) == 1)
        self.assertIsNone(artifact_list[0].accessories)
        # 4.2. Verify get_reference_info API;
        artifact_info = self.artifact.get_reference_info(TestCosign.project_name, self.image, artifact_reference, **TestCosign.user_client)
        self.assertIsNone(artifact_info.accessories)
        # 4.3. Verify list_accessories API;
        accessory_list = self.artifact.list_accessories(TestCosign.project_name, self.image, artifact_reference, **TestCosign.user_client)
        self.assertTrue(len(accessory_list) == 0)

        # 5.1. Generate cosign key pair;
        cosign.generate_key_pair()
        # 5.2. Generate cosign key pair;
        docker_api.docker_login_cmd(harbor_server, user_name, user_password, enable_manifest = False)
        cosign.sign_artifact("{}/{}/{}:{}".format(harbor_server, TestCosign.project_name, self.image, self.tag))

        # 6.1. Verify list_artifacts API;
        artifact_list = self.artifact.list_artifacts(TestCosign.project_name, self.image, **TestCosign.user_client)
        self.assertTrue(len(artifact_list) == 1)
        first_artifact = artifact_list[0]
        self.assertTrue(len(first_artifact.accessories) == 1)
        first_accessory = first_artifact.accessories[0]
        self.assertEqual(first_accessory.type, self.expect_accessory_type)
        accessory_reference = first_accessory.digest
        # 6.2. Verify get_reference_info API;
        artifact_info = self.artifact.get_reference_info(TestCosign.project_name, self.image, artifact_reference, **TestCosign.user_client)
        self.assertEqual(artifact_info.accessories[0].type, self.expect_accessory_type)
        # 6.3. Verify list_accessories API;
        accessory_list = self.artifact.list_accessories(TestCosign.project_name, self.image, artifact_reference, **TestCosign.user_client)
        self.assertTrue(len(accessory_list) == 1)
        self.assertEqual(accessory_list[0].type, self.expect_accessory_type)
        # 6.4. Verify list_accessories API;
        accessory_info = self.artifact.get_reference_info(TestCosign.project_name, self.image, accessory_reference, **TestCosign.user_client)
        self.assertEqual(accessory_info.digest, accessory_reference)

if __name__ == '__main__':
    unittest.main()
