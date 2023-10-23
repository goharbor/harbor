from __future__ import absolute_import

import unittest
import time

from testutils import ADMIN_CLIENT, suppress_urllib3_warning, files_directory
from testutils import TEARDOWN
from testutils import harbor_server
from library.user import User
from library.project import Project
from library.repository import Repository
from library.base import _assert_status_code
from library.repository import push_special_image_to_project, push_self_build_image_to_project
from library.artifact import Artifact
from library.gc import GC
from library import docker_api, cosign


class TestProjects(unittest.TestCase):
    @suppress_urllib3_warning
    def setUp(self):
        self.gc = GC()
        self.project = Project()
        self.user = User()
        self.repo = Repository()
        self.artifact = Artifact()
        self.repo_name = "test_repo"
        self.repo_name_untag = "test_untag"
        self.tag = "v1.0"
        self.gc_success_status = "Success"


    def testGarbageCollection(self):
        """
        Test case:
            Garbage Collection
        Test step and expected result:
            1. Create a new user(UA);
            2. Create project(PA) and project(PB) by user(UA);
            3. Push a image in project(PA) and then delete repository by admin;
            4. Get repository by user(UA), it should get nothing;
            5. Tigger garbage collection operation;
            6. Check garbage collection job was finished;
            7. Get garbage collection log, check there is a number of files was deleted;
            8. Push a image in project(PB) by admin and delete the only tag;
            9. Tigger garbage collection operation;
            10. Check garbage collection job was finished;
            11. Repository with untag image should be still there;
            12. But no any artifact in repository anymore.
        """
        url = ADMIN_CLIENT["endpoint"]
        admin_name = ADMIN_CLIENT["username"]
        admin_password = ADMIN_CLIENT["password"]
        user_gc_password = "Aa123456"

        #1. Create a new user(UA);
        TestProjects.user_gc_id, user_gc_name = self.user.create_user(user_password = user_gc_password, **ADMIN_CLIENT)

        TestProjects.USER_GC_CLIENT=dict(endpoint = url, username = user_gc_name, password = user_gc_password)

        #2. Create project(PA) and project(PB) by user(UA);
        TestProjects.project_gc_id, TestProjects.project_gc_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_GC_CLIENT)
        TestProjects.project_gc_untag_id, TestProjects.project_gc_untag_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_GC_CLIENT)

        #3. Push a image in project(PA) and then delete repository by admin;
        push_special_image_to_project(TestProjects.project_gc_name, harbor_server, admin_name, admin_password, self.repo_name, ["latest", "v1.2.3"])
        self.repo.delete_repository(TestProjects.project_gc_name, self.repo_name, **TestProjects.USER_GC_CLIENT)

        #4. Get repository by user(UA), it should get nothing;
        repo_data = self.repo.list_repositories(TestProjects.project_gc_name, **TestProjects.USER_GC_CLIENT)
        _assert_status_code(len(repo_data), 0)

        #8. Push a image in project(PB) by admin and delete the only tag;
        push_special_image_to_project(TestProjects.project_gc_untag_name, harbor_server, admin_name, admin_password, self.repo_name_untag, [self.tag])
        self.artifact.delete_tag(TestProjects.project_gc_untag_name, self.repo_name_untag, self.tag, self.tag, **ADMIN_CLIENT)

        #5. Tigger garbage collection operation;
        gc_id = self.gc.gc_now(**ADMIN_CLIENT)

        #6. Check garbage collection job was finished;
        self.gc.validate_gc_job_status(gc_id, self.gc_success_status, **ADMIN_CLIENT)

        #7. Get garbage collection log, check there is a number of files was deleted;
        self.gc.validate_deletion_success(gc_id, **ADMIN_CLIENT)

        artifacts = self.artifact.list_artifacts(TestProjects.project_gc_untag_name, self.repo_name_untag, **TestProjects.USER_GC_CLIENT)
        _assert_status_code(len(artifacts), 1)

        time.sleep(5)

        #9. Tigger garbage collection operation;
        gc_id = self.gc.gc_now(is_delete_untagged=True, **ADMIN_CLIENT)

        #10. Check garbage collection job was finished;
        self.gc.validate_gc_job_status(gc_id, self.gc_success_status, **ADMIN_CLIENT)

        #7. Get garbage collection log, check there is a number of files was deleted;
        self.gc.validate_deletion_success(gc_id, **ADMIN_CLIENT)

        #11. Repository with untag image should be still there;
        repo_data_untag = self.repo.list_repositories(TestProjects.project_gc_untag_name, **TestProjects.USER_GC_CLIENT)
        _assert_status_code(len(repo_data_untag), 1)
        self.assertEqual(TestProjects.project_gc_untag_name + "/" + self.repo_name_untag , repo_data_untag[0].name)

        #12. But no any artifact in repository anymore.
        artifacts = self.artifact.list_artifacts(TestProjects.project_gc_untag_name, self.repo_name_untag, **TestProjects.USER_GC_CLIENT)
        self.assertEqual(artifacts,[])


    def testGarbageCollectionAccessory(self):
        """
        Test case:
            Garbage Collection Accessory
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Push a new image(IA) in project(PA) by user(UA);
            4. Push image(IA) SBOM to project(PA) by user(UA);
            5. Sign image(IA) with cosign;
            6. Sign image(IA) SBOM with cosign;
            7. Sign image(IA) Signature with cosign;
            8. Delete image(IA) Signature of signature by user(UA);
            9. Trigger GC and wait for GC to succeed;
            10. Get the GC log and check that the image(IA) Signature of signature is deleted;
            11. Delete image(IA) Signature by user(UA);
            12. Trigger GC and wait for GC to succeed;
            13. Get the GC log and check that the image(IA) Signature is deleted;
            14. Delete image(IA) SBOM by user(UA);
            15. Trigger GC and wait for GC to succeed;
            16. Get the GC log and check that the image(IA) SBOM and Signature of SBOM is deleted;
            17. Push image(IA) SBOM to project(PA) by user(UA);
            18. Sign image(IA) with cosign;
            19. Sign image(IA) SBOM with cosign;
            20. Sign image(IA) Signature with cosign;
            21. Trigger GC and wait for GC to succeed;
            22. Get the GC log and check that it is not deleted;
            23. Delete tag of image(IA) by user(UA);
            24. Trigger GC and wait for GC to succeed;
            25. Get the GC log and check that the image(IA) and all aeecssory is deleted;
        """
        url = ADMIN_CLIENT["endpoint"]
        user_password = "Aa123456"
        deleted_prefix =  "delete blob from storage: "
        # 1. Create user(UA)
        _, user_name = self.user.create_user(user_password = user_password, **ADMIN_CLIENT)
        user_client = dict(endpoint = url, username = user_name, password = user_password, with_accessory = True)

        # 2. Create private project(PA) by user(UA)
        self.image = "test_image"
        _, project_name = self.project.create_project(metadata = {"public": "false"}, **user_client)

        # 3. Push a new image(IA) in project(PA) by user(UA)
        push_self_build_image_to_project(project_name, harbor_server, user_name, user_password, self.image, self.tag)

        # 4. Push image(IA) SBOM to project(PA) by user(UA)
        self.sbom_path = files_directory + "sbom_test.json"
        docker_api.docker_login_cmd(harbor_server, user_name, user_password, enable_manifest = False)
        cosign.push_artifact_sbom("{}/{}/{}:{}".format(harbor_server, project_name, self.image, self.tag), self.sbom_path)
        artifact_info = self.artifact.get_reference_info(project_name, self.image, self.tag, **user_client)
        sbom_digest = artifact_info.accessories[0].digest

        # 5. Sign image(IA) with cosign
        cosign.generate_key_pair()
        cosign.sign_artifact("{}/{}/{}:{}".format(harbor_server, project_name, self.image, self.tag))
        artifact_info = self.artifact.get_reference_info(project_name, self.image, self.tag, **user_client)
        self.assertEqual(len(artifact_info.accessories), 2)
        image_signature_digest = None
        for accessory in artifact_info.accessories:
            if accessory.digest != sbom_digest:
                image_signature_digest = accessory.digest
                break

        # 6. Sign image(IA) SBOM cosign
        cosign.sign_artifact("{}/{}/{}@{}".format(harbor_server, project_name, self.image, sbom_digest))
        sbom_info = self.artifact.get_reference_info(project_name, self.image, sbom_digest, **user_client)
        sbom_signature_digest = sbom_info.accessories[0].digest

        # 7. Sign image(IA) Signature with cosign
        cosign.sign_artifact("{}/{}/{}@{}".format(harbor_server, project_name, self.image, image_signature_digest))
        signature_info = self.artifact.get_reference_info(project_name, self.image, image_signature_digest, **user_client)
        signature_signature_digest = signature_info.accessories[0].digest

        # 8. Delete image(IA) Signature of signature by user(UA)
        self.artifact.delete_artifact(project_name, self.image, signature_signature_digest, **user_client)

        # 9. Trigger GC and wait for GC to succeed
        gc_id = self.gc.gc_now(**ADMIN_CLIENT)
        self.gc.validate_gc_job_status(gc_id, self.gc_success_status, **ADMIN_CLIENT)

        # 10. Get the GC log and check that the image(IA) Signature of signature is deleted
        gc_log = self.gc.get_gc_log_by_id(gc_id, **ADMIN_CLIENT)
        self.assertIn(deleted_prefix + signature_signature_digest, gc_log)

        # 11. Delete image(IA) Signature by user(UA)
        self.artifact.delete_artifact(project_name, self.image, image_signature_digest, **user_client)

        # 12. Trigger GC and wait for GC to succeed
        gc_id = self.gc.gc_now(**ADMIN_CLIENT)
        self.gc.validate_gc_job_status(gc_id, self.gc_success_status, **ADMIN_CLIENT)

        # 13. Get the GC log and check that the image(IA) Signature is deleted
        gc_log = self.gc.get_gc_log_by_id(gc_id, **ADMIN_CLIENT)
        self.assertIn(deleted_prefix + image_signature_digest, gc_log)

        # 14. Delete image(IA) SBOM by user(UA)
        self.artifact.delete_artifact(project_name, self.image, sbom_digest, **user_client)

        # 15. Trigger GC and wait for GC to succeed
        gc_id = self.gc.gc_now(**ADMIN_CLIENT)
        self.gc.validate_gc_job_status(gc_id, self.gc_success_status, **ADMIN_CLIENT)

        # 16. Get the GC log and check that the image(IA) SBOM and Signature of SBOM is deleted
        gc_log = self.gc.get_gc_log_by_id(gc_id, **ADMIN_CLIENT)
        self.assertIn(deleted_prefix + sbom_digest, gc_log)
        self.assertIn(deleted_prefix + sbom_signature_digest, gc_log)

        # 17. Push image(IA) SBOM to project(PA) by user(UA)
        self.sbom_path = files_directory + "sbom_test.json"
        docker_api.docker_login_cmd(harbor_server, user_name, user_password, enable_manifest = False)
        cosign.push_artifact_sbom("{}/{}/{}:{}".format(harbor_server, project_name, self.image, self.tag), self.sbom_path)
        artifact_info = self.artifact.get_reference_info(project_name, self.image, self.tag, **user_client)
        sbom_digest = artifact_info.accessories[0].digest

        # 18. Sign image(IA) with cosign
        cosign.sign_artifact("{}/{}/{}:{}".format(harbor_server, project_name, self.image, self.tag))
        artifact_info = self.artifact.get_reference_info(project_name, self.image, self.tag, **user_client)
        self.assertEqual(len(artifact_info.accessories), 2)
        image_signature_digest = None
        for accessory in artifact_info.accessories:
            if accessory.digest != sbom_digest:
                image_signature_digest = accessory.digest
                break

        # 19. Sign image(IA) SBOM cosign
        cosign.sign_artifact("{}/{}/{}@{}".format(harbor_server, project_name, self.image, sbom_digest))
        sbom_info = self.artifact.get_reference_info(project_name, self.image, sbom_digest, **user_client)
        sbom_signature_digest = sbom_info.accessories[0].digest

        # 20. Sign image(IA) Signature with cosign
        cosign.sign_artifact("{}/{}/{}@{}".format(harbor_server, project_name, self.image, image_signature_digest))
        signature_info = self.artifact.get_reference_info(project_name, self.image, image_signature_digest, **user_client)
        signature_signature_digest = signature_info.accessories[0].digest

        # 21. Trigger GC and wait for GC to succeed
        gc_id = self.gc.gc_now(**ADMIN_CLIENT)
        self.gc.validate_gc_job_status(gc_id, self.gc_success_status, **ADMIN_CLIENT)

        # 22. Get the GC log and check that it is not deleted
        gc_log = self.gc.get_gc_log_by_id(gc_id, **ADMIN_CLIENT)
        self.assertNotIn(deleted_prefix + sbom_digest, gc_log)
        self.assertNotIn(deleted_prefix + sbom_signature_digest, gc_log)
        self.assertNotIn(deleted_prefix + image_signature_digest, gc_log)
        self.assertNotIn(deleted_prefix + signature_signature_digest, gc_log)

        # 23. Delete tag of image(IA) by user(UA)
        self.artifact.delete_tag(project_name, self.image, self.tag, self.tag, **user_client)

        # 24. Trigger GC and wait for GC to succeed
        gc_id = self.gc.gc_now(is_delete_untagged=True, **ADMIN_CLIENT)
        self.gc.validate_gc_job_status(gc_id, self.gc_success_status, **ADMIN_CLIENT)

        # 25. Get the GC log and check that the image(IA) and all aeecssory is deleted
        gc_log = self.gc.get_gc_log_by_id(gc_id, **ADMIN_CLIENT)
        self.assertIn(self.image, gc_log)
        self.assertIn(deleted_prefix + sbom_digest, gc_log)
        self.assertIn(deleted_prefix + sbom_signature_digest, gc_log)
        self.assertIn(deleted_prefix + image_signature_digest, gc_log)
        self.assertIn(deleted_prefix + signature_signature_digest, gc_log)


if __name__ == '__main__':
    unittest.main()
