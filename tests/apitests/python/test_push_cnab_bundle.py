from __future__ import absolute_import

import sys
import unittest

from testutils import ADMIN_CLIENT, suppress_urllib3_warning
from testutils import harbor_server
from testutils import TEARDOWN
import library.repository
import library.cnab
from library.project import Project
from library.user import User
from library.repository import Repository
from library.artifact import Artifact
from library.scan import Scan

class TestCNAB(unittest.TestCase):
    @suppress_urllib3_warning
    def setUp(self):
        print("Setup")

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def do_tearDown(self):
        """
        Tear down:
            1. Delete repository(RA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA);
        """
        #1. Delete repository(RA) by user(UA);
        TestCNAB.repo.delete_repository(TestCNAB.project_name, TestCNAB.cnab_repo_name, **TestCNAB.USER_CLIENT)

        #2. Delete project(PA);
        TestCNAB.project.delete_project(TestCNAB.project_id, **TestCNAB.USER_CLIENT)

        #3. Delete user(UA).
        TestCNAB.user.delete_user(TestCNAB.user_id, **ADMIN_CLIENT)

    def test_01_PushBundleByCnab(self):
        """
        Test case:
            Push Bundle By Cnab
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Push bundle to harbor as repository(RA);
            4. Get repository from Harbor successfully;
            5. Verfiy bundle name;
            6. Get artifact by sha256;
            7. Verify artifact information.
        """
        TestCNAB.project= Project()
        TestCNAB.user= User()
        TestCNAB.artifact = Artifact()
        TestCNAB.repo= Repository()
        TestCNAB.scan = Scan()
        TestCNAB.url = ADMIN_CLIENT["endpoint"]
        TestCNAB.user_push_cnab_password = "Aa123456"
        TestCNAB.cnab_repo_name = "test_cnab"
        TestCNAB.cnab_tag = "test_cnab_tag"
        TestCNAB.project_name = None
        TestCNAB.artifacts_config_ref_child_list = None
        TestCNAB.artifacts_ref_child_list = None

        #1. Create a new user(UA);
        TestCNAB.user_id, TestCNAB.user_name = TestCNAB.user.create_user(user_password = TestCNAB.user_push_cnab_password, **ADMIN_CLIENT)
        TestCNAB.USER_CLIENT=dict(endpoint = TestCNAB.url, username = TestCNAB.user_name, password = TestCNAB.user_push_cnab_password, with_scan_overview = True)

        #2. Create a new project(PA) by user(UA);
        TestCNAB.project_id, TestCNAB.project_name = TestCNAB.project.create_project(metadata = {"public": "false"}, **TestCNAB.USER_CLIENT)

        #3. Push bundle to harbor as repository(RA);
        target = harbor_server + "/" + TestCNAB.project_name  + "/" + TestCNAB.cnab_repo_name  + ":" + TestCNAB.cnab_tag
        TestCNAB.reference_sha256 = library.cnab.push_cnab_bundle(harbor_server, TestCNAB.user_name, TestCNAB.user_push_cnab_password, "goharbor/harbor-log:v1.10.0", "kong:latest", target)

        #4. Get repository from Harbor successfully;
        TestCNAB.cnab_bundle_data = TestCNAB.repo.get_repository(TestCNAB.project_name, TestCNAB.cnab_repo_name, **TestCNAB.USER_CLIENT)
        print(TestCNAB.cnab_bundle_data)

        #4.1 Get refs of CNAB bundle;
        TestCNAB.artifacts = TestCNAB.artifact.list_artifacts(TestCNAB.project_name, TestCNAB.cnab_repo_name, **TestCNAB.USER_CLIENT)
        print("artifacts:", TestCNAB.artifacts)
        TestCNAB.artifacts_ref_child_list = []
        TestCNAB.artifacts_config_ref_child_list = []
        for ref in TestCNAB.artifacts[0].references:
            if ref.annotations["io.cnab.manifest.type"] != 'config':
                TestCNAB.artifacts_ref_child_list.append(ref.child_digest)
            else:
                TestCNAB.artifacts_config_ref_child_list.append(ref.child_digest)
        self.assertEqual(len(TestCNAB.artifacts_ref_child_list), 2, msg="Image artifact count should be 2.")
        self.assertEqual(len(TestCNAB.artifacts_config_ref_child_list), 1, msg="Bundle count should be 1.")
        print(TestCNAB.artifacts_ref_child_list)

        #4.2 Cnab bundle can be pulled by ctr successfully;
        # This step might not successful since ctr does't support cnab fully, it might be uncomment sometime in future.
        # Please keep them in comment!
        #library.containerd.ctr_images_pull(TestCNAB.user_name, TestCNAB.user_push_cnab_password, target)
        #library.containerd.ctr_images_list(oci_ref = target)

        #5. Verfiy bundle name;
        self.assertEqual(TestCNAB.cnab_bundle_data.name, TestCNAB.project_name + "/" + TestCNAB.cnab_repo_name)

        #6. Get artifact by sha256;
        artifact = TestCNAB.artifact.get_reference_info(TestCNAB.project_name, TestCNAB.cnab_repo_name, TestCNAB.reference_sha256, **TestCNAB.USER_CLIENT)

        #7. Verify artifact information;
        self.assertEqual(artifact.type, 'CNAB')
        self.assertEqual(artifact.digest, TestCNAB.reference_sha256)

    def test_02_ScanCNAB(self):
        """
        Test case:
            Scan CNAB
        Test step and expected result:
            1. Scan config artifact, it should be failed with 400 status code;
            2. Scan 1st child artifact, it should be scanned, the other should be not scanned, repository should not be scanned;
            3. Scan 2cn child artifact, it should be scanned, repository should not be scanned;
            4. Scan repository, it should be scanned;
        Tear down:
        """
        #1. Scan config artifact, it should be failed with 400 status code;
        TestCNAB.scan.scan_artifact(TestCNAB.project_name, TestCNAB.cnab_repo_name, TestCNAB.artifacts_config_ref_child_list[0], expect_status_code = 400, **TestCNAB.USER_CLIENT)

        #2. Scan 1st child artifact, it should be scanned, the other should be not scanned, repository should not be scanned;
        TestCNAB.scan.scan_artifact(TestCNAB.project_name, TestCNAB.cnab_repo_name, TestCNAB.artifacts_ref_child_list[0], **TestCNAB.USER_CLIENT)
        TestCNAB.artifact.check_image_scan_result(TestCNAB.project_name, TestCNAB.cnab_repo_name, TestCNAB.artifacts_ref_child_list[0], **TestCNAB.USER_CLIENT)
        TestCNAB.artifact.check_image_scan_result(TestCNAB.project_name, TestCNAB.cnab_repo_name, TestCNAB.artifacts_ref_child_list[1], expected_scan_status = "Not Scanned", **TestCNAB.USER_CLIENT)
        TestCNAB.artifact.check_image_scan_result(TestCNAB.project_name, TestCNAB.cnab_repo_name, TestCNAB.artifacts_config_ref_child_list[0], expected_scan_status = "No Scan Overview", **TestCNAB.USER_CLIENT)
        TestCNAB.artifact.check_image_scan_result(TestCNAB.project_name, TestCNAB.cnab_repo_name, TestCNAB.artifacts[0].digest, expected_scan_status = "Not Scanned", **TestCNAB.USER_CLIENT)

        #3. Scan 2cn child artifact, it should be scanned, repository should not be scanned;
        TestCNAB.scan.scan_artifact(TestCNAB.project_name, TestCNAB.cnab_repo_name, TestCNAB.artifacts_ref_child_list[1], **TestCNAB.USER_CLIENT)
        TestCNAB.artifact.check_image_scan_result(TestCNAB.project_name, TestCNAB.cnab_repo_name, TestCNAB.artifacts_ref_child_list[1], **TestCNAB.USER_CLIENT)
        TestCNAB.artifact.check_image_scan_result(TestCNAB.project_name, TestCNAB.cnab_repo_name, TestCNAB.artifacts_config_ref_child_list[0], expected_scan_status = "No Scan Overview", **TestCNAB.USER_CLIENT)
        TestCNAB.artifact.check_image_scan_result(TestCNAB.project_name, TestCNAB.cnab_repo_name, TestCNAB.artifacts[0].digest, expected_scan_status = "Not Scanned", **TestCNAB.USER_CLIENT)

        #4. Scan repository, it should be scanned;
        TestCNAB.scan.scan_artifact(TestCNAB.project_name, TestCNAB.cnab_repo_name, TestCNAB.artifacts[0].digest, **TestCNAB.USER_CLIENT)
        TestCNAB.artifact.check_image_scan_result(TestCNAB.project_name, TestCNAB.cnab_repo_name, TestCNAB.artifacts[0].digest, **TestCNAB.USER_CLIENT)

        self.do_tearDown()

if __name__ == '__main__':
    suite = unittest.TestSuite(unittest.makeSuite(TestCNAB))
    result = unittest.TextTestRunner(sys.stdout, verbosity=2, failfast=True).run(suite)
    if not result.wasSuccessful():
        raise Exception(r"CNAB test failed: {}".format(result))

