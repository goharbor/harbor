from __future__ import absolute_import

import sys
import unittest

from testutils import ADMIN_CLIENT, suppress_urllib3_warning
from testutils import harbor_server
from testutils import TEARDOWN
import library.repository
import library.docker_api
import library.containerd
from library.base import _assert_status_code
from library.project import Project
from library.user import User
from library.repository import Repository
from library.artifact import Artifact
from library.repository import push_self_build_image_to_project
from library.repository import pull_harbor_image
from library.scan import Scan

class TestManifest(unittest.TestCase):
    @suppress_urllib3_warning
    def setUp(self):
        print("Setup")

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def do_tearDown(self):
        """
        Tear down:
            1. Delete repository(RA,RB,IA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA).
        """
        #1. Delete repository(RA,RB,IA) by user(UA);
        TestManifest.repo.delete_repository(TestManifest.project_name, TestManifest.index_name, **TestManifest.USER_CLIENT)
        TestManifest.repo.delete_repository(TestManifest.project_name, TestManifest.image_a, **TestManifest.USER_CLIENT)
        TestManifest.repo.delete_repository(TestManifest.project_name, TestManifest.image_b, **TestManifest.USER_CLIENT)

        #2. Delete project(PA);
        TestManifest.project.delete_project(TestManifest.project_id, **TestManifest.USER_CLIENT)

        #3. Delete user(UA).
        TestManifest.user.delete_user(TestManifest.user_id, **ADMIN_CLIENT)

    def test_01_AddIndexByDockerManifest(self):
        """
        Test case:
            Push Index By Docker Manifest
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Create 2 new repositorys(RA,RB) in project(PA) by user(UA);
            4. Push an index(IA) to Harbor by docker manifest CLI successfully;
            5. Get Artifacts successfully;
            6. Get index(IA) by reference successfully;
            7. Verify harbor index is index(IA) pushed by docker manifest CLI;
            8.1 Verify harbor index(IA) can be pulled by docker CLI successfully;
            8.2 Verify harbor index(IA) can be pulled by docker CLI successfully;
            9. Get addition successfully;
            10. Unable to Delete artifact in manifest list;
            11. Delete index successfully.
        Tear down:
            1. Delete repository(RA,RB,IA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA).
        """
        TestManifest.project= Project()
        TestManifest.user= User()
        TestManifest.artifact = Artifact()
        TestManifest.repo= Repository()
        TestManifest.scan = Scan()
        TestManifest.url = ADMIN_CLIENT["endpoint"]
        TestManifest.user_push_index_password = "Aa123456"
        TestManifest.index_name = "ci_test_index"
        TestManifest.index_tag = "test_tag"
        TestManifest.image_a = "alpine"
        TestManifest.image_b = "busybox"

        #1. Create a new user(UA);
        TestManifest.user_id, TestManifest.user_name = TestManifest.user.create_user(user_password = TestManifest.user_push_index_password, **ADMIN_CLIENT)

        TestManifest.USER_CLIENT=dict(endpoint = TestManifest.url, username = TestManifest.user_name, password = TestManifest.user_push_index_password, with_scan_overview = True)

        #2. Create a new project(PA) by user(UA);
        TestManifest.project_id, TestManifest.project_name = TestManifest.project.create_project(metadata = {"public": "false"}, **TestManifest.USER_CLIENT)

        #3. Create 2 new repositorys(RA,RB) in project(PA) by user(UA);
        repo_name_a, tag_a = push_self_build_image_to_project(TestManifest.project_name, harbor_server, 'admin', 'Harbor12345', TestManifest.image_a, "latest")
        repo_name_b, tag_b = push_self_build_image_to_project(TestManifest.project_name, harbor_server, 'admin', 'Harbor12345', TestManifest.image_b, "latest")

        #4. Push an index(IA) to Harbor by docker manifest CLI successfully;
        manifests = [harbor_server+"/"+repo_name_a+":"+tag_a, harbor_server+"/"+repo_name_b+":"+tag_b]
        index = harbor_server+"/"+TestManifest.project_name+"/"+TestManifest.index_name+":"+TestManifest.index_tag
        TestManifest.index_sha256_cli_ret, TestManifest.manifests_sha256_cli_ret = library.docker_api.docker_manifest_push_to_harbor(index, manifests, harbor_server, TestManifest.user_name, TestManifest.user_push_index_password)

        #5. Get Artifacts successfully;
        artifacts = TestManifest.artifact.list_artifacts(TestManifest.project_name, TestManifest.index_name, **TestManifest.USER_CLIENT)
        TestManifest.artifacts_ref_child_list = [artifacts[0].references[1].child_digest, artifacts[0].references[0].child_digest]
        self.assertEqual(TestManifest.artifacts_ref_child_list.count(TestManifest.manifests_sha256_cli_ret[0]), 1)
        self.assertEqual(TestManifest.artifacts_ref_child_list.count(TestManifest.manifests_sha256_cli_ret[1]), 1)

        #6. Get index(IA) by reference successfully;
        index_data = TestManifest.artifact.get_reference_info(TestManifest.project_name, TestManifest.index_name, TestManifest.index_tag, **TestManifest.USER_CLIENT)
        print("index_data:",index_data)
        manifests_sha256_harbor_ret = [index_data.references[1].child_digest, index_data.references[0].child_digest]

        #7. Verify harbor index is index(IA) pushed by docker manifest CLI;
        self.assertEqual(index_data.digest, TestManifest.index_sha256_cli_ret)
        self.assertEqual(manifests_sha256_harbor_ret.count(TestManifest.manifests_sha256_cli_ret[0]), 1)
        self.assertEqual(manifests_sha256_harbor_ret.count(TestManifest.manifests_sha256_cli_ret[1]), 1)

        #8.1 Verify harbor index(IA) can be pulled by docker CLI successfully;
        pull_harbor_image(harbor_server, TestManifest.user_name, TestManifest.user_push_index_password, TestManifest.project_name+"/"+TestManifest.index_name, TestManifest.index_tag)

        #8.2 Verify harbor index(IA) can be pulled by ctr successfully;
        oci_ref = harbor_server+"/"+TestManifest.project_name+"/"+TestManifest.index_name+":"+TestManifest.index_tag
        library.containerd.ctr_images_pull(TestManifest.user_name, TestManifest.user_push_index_password, oci_ref)
        library.containerd.ctr_images_list(oci_ref = oci_ref)

        #9. Get addition successfully;
        addition_v = TestManifest.artifact.get_addition(TestManifest.project_name, TestManifest.index_name, TestManifest.index_tag, "vulnerabilities", **TestManifest.USER_CLIENT)
        self.assertEqual(addition_v[0], '{}')
        #This artifact has no build history

        addition_v = TestManifest.artifact.get_addition(TestManifest.project_name, TestManifest.index_name, TestManifest.manifests_sha256_cli_ret[0], "vulnerabilities", **TestManifest.USER_CLIENT)
        self.assertEqual(addition_v[0], '{}')
        addition_b = TestManifest.artifact.get_addition(TestManifest.project_name, TestManifest.index_name, TestManifest.manifests_sha256_cli_ret[0], "build_history", **TestManifest.USER_CLIENT)
        self.assertIn("ADD file:", addition_b[0])
        image_data = TestManifest.artifact.get_reference_info(TestManifest.project_name, TestManifest.index_name, TestManifest.manifests_sha256_cli_ret[0], **TestManifest.USER_CLIENT)

        addition_v = TestManifest.artifact.get_addition(TestManifest.project_name, TestManifest.index_name, TestManifest.manifests_sha256_cli_ret[1], "vulnerabilities", **TestManifest.USER_CLIENT)
        self.assertEqual(addition_v[0], '{}')
        addition_b = TestManifest.artifact.get_addition(TestManifest.project_name, TestManifest.index_name, TestManifest.manifests_sha256_cli_ret[1], "build_history", **TestManifest.USER_CLIENT)
        self.assertIn("ADD file:", addition_b[0])
        image_data = TestManifest.artifact.get_reference_info(TestManifest.project_name, TestManifest.index_name, TestManifest.manifests_sha256_cli_ret[0], **TestManifest.USER_CLIENT)

    @unittest.skipIf(TEARDOWN == True, "Test data won't be erased.")
    def test_99_ManifestDeletion(self):
        #10. Unable to Delete artifact in manifest list;
        TestManifest.artifact.delete_artifact(TestManifest.project_name, TestManifest.index_name, TestManifest.manifests_sha256_cli_ret[0], expect_status_code = 412, **TestManifest.USER_CLIENT)

        #11. Delete index successfully.
        TestManifest.artifact.delete_artifact(TestManifest.project_name, TestManifest.index_name, TestManifest.index_tag, **TestManifest.USER_CLIENT)

        TestManifest.do_tearDown()

    def test_02_ScanManifestList(self):
        """
        Test case:
            Scan Manifest List
        Test step and expected result:
            1. Scan 1st child artifact, it should be scanned, 2nd child should be Not Scanned, index should be Not Scanned;
            2. Scan 2nd child artifact, it should be scanned, index should be scanned;
        """
        #1. Scan 1st child artifact, it should be scanned, 2nd child should be Not Scanned, index should be Not Scanned;
        TestManifest.scan.scan_artifact(TestManifest.project_name, TestManifest.index_name, TestManifest.artifacts_ref_child_list[0], **TestManifest.USER_CLIENT)
        TestManifest.artifact.check_image_scan_result(TestManifest.project_name, TestManifest.index_name, TestManifest.artifacts_ref_child_list[0], **TestManifest.USER_CLIENT)

        TestManifest.artifact.check_image_scan_result(TestManifest.project_name, TestManifest.index_name, TestManifest.artifacts_ref_child_list[1], expected_scan_status = "Not Scanned", **TestManifest.USER_CLIENT)
        TestManifest.artifact.check_image_scan_result(TestManifest.project_name, TestManifest.index_name, TestManifest.index_sha256_cli_ret, expected_scan_status = "Not Scanned", **TestManifest.USER_CLIENT)

        #2. Scan 2nd child artifact, it should be scanned, index should be scanned;
        TestManifest.scan.scan_artifact(TestManifest.project_name, TestManifest.index_name, TestManifest.artifacts_ref_child_list[1], **TestManifest.USER_CLIENT)
        TestManifest.artifact.check_image_scan_result(TestManifest.project_name, TestManifest.index_name, TestManifest.artifacts_ref_child_list[1], **TestManifest.USER_CLIENT)
        TestManifest.artifact.check_image_scan_result(TestManifest.project_name, TestManifest.index_name, TestManifest.index_sha256_cli_ret, **TestManifest.USER_CLIENT)


if __name__ == '__main__':
    suite = unittest.TestSuite(unittest.makeSuite(TestManifest))
    result = unittest.TextTestRunner(sys.stdout, verbosity=2, failfast=True).run(suite)
    if not result.wasSuccessful():
        raise Exception(r"Manifest test failed: {}".format(result))

