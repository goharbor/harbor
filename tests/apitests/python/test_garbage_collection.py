from __future__ import absolute_import

import unittest
import json

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
            5. Push a image in project(PB) by admin and delete the only tag;
            6. Tigger garbage collection operation;
            7. Check garbage collection job was finished;
            8. Get garbage collection log, check there is a number of files was deleted;
            9. Check garbage collection details;
            10. Check the log for garbage collection workers;
            11. Tigger garbage collection operation;
            12. Check garbage collection job was finished;
            13. Get garbage collection log, check there is a number of files was deleted;
            14. Repository with untag image should be still there;
            15. But no any artifact in repository anymore;
            16. Check garbage collection details;
            17. Check the log for garbage collection workers;
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

        #5. Push a image in project(PB) by admin and delete the only tag;
        push_special_image_to_project(TestProjects.project_gc_untag_name, harbor_server, admin_name, admin_password, self.repo_name_untag, [self.tag])
        self.artifact.delete_tag(TestProjects.project_gc_untag_name, self.repo_name_untag, self.tag, self.tag, **ADMIN_CLIENT)

        #6. Tigger garbage collection operation;
        gc_id = self.gc.gc_now(**ADMIN_CLIENT)

        #7. Check garbage collection job was finished;
        self.gc.validate_gc_job_status(gc_id, self.gc_success_status, **ADMIN_CLIENT)

        #8. Get garbage collection log, check there is a number of files was deleted;
        self.gc.validate_deletion_success(gc_id, **ADMIN_CLIENT)
        artifacts = self.artifact.list_artifacts(TestProjects.project_gc_untag_name, self.repo_name_untag, **TestProjects.USER_GC_CLIENT)
        _assert_status_code(len(artifacts), 1)

        #9. Check garbage collection details;
        gc_history = self.gc.get_gc_history(**ADMIN_CLIENT)
        self.checkGarbageCollectionDetails(gc_history[0], False, False, 1000000, 1100000, 2, 1, 1)

        #10. Check the log for garbage collection workers;
        gc_log = self.gc.get_gc_log_by_id(gc_id, **ADMIN_CLIENT)
        self.assertIn("workers: 1", gc_log)

        #11. Tigger garbage collection operation;
        workers = 2
        gc_id = self.gc.gc_now(is_delete_untagged=True, workers=2, **ADMIN_CLIENT)

        #12. Check garbage collection job was finished;
        self.gc.validate_gc_job_status(gc_id, self.gc_success_status, **ADMIN_CLIENT)

        #13. Get garbage collection log, check there is a number of files was deleted;
        self.gc.validate_deletion_success(gc_id, **ADMIN_CLIENT)

        #14. Repository with untag image should be still there;
        repo_data_untag = self.repo.list_repositories(TestProjects.project_gc_untag_name, **TestProjects.USER_GC_CLIENT)
        _assert_status_code(len(repo_data_untag), 1)
        self.assertEqual(TestProjects.project_gc_untag_name + "/" + self.repo_name_untag , repo_data_untag[0].name)

        #15. But no any artifact in repository anymore.
        artifacts = self.artifact.list_artifacts(TestProjects.project_gc_untag_name, self.repo_name_untag, **TestProjects.USER_GC_CLIENT)
        self.assertEqual(artifacts,[])

        #16. Check garbage collection details;
        gc_history = self.gc.get_gc_history(**ADMIN_CLIENT)
        self.checkGarbageCollectionDetails(gc_history[0], True, False, 1500000, 2000000, 3, 1, workers)

        #17. Check the log for garbage collection workers;
        gc_log = self.gc.get_gc_log_by_id(gc_id, **ADMIN_CLIENT)
        self.assertIn("workers: {}".format(workers), gc_log)


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
            10. Check garbage collection details;
            11. Check the log for garbage collection workers;
            12. Get the GC log and check that the image(IA) Signature of signature is deleted;
            13. Delete image(IA) Signature by user(UA);
            14. Trigger GC and wait for GC to succeed;
            15. Check garbage collection details;
            16. Check the log for garbage collection workers;
            17. Get the GC log and check that the image(IA) Signature is deleted;
            18. Delete image(IA) SBOM by user(UA);
            19. Trigger GC and wait for GC to succeed;
            20. Check garbage collection details;
            21. Check the log for garbage collection workers;
            22. Get the GC log and check that the image(IA) SBOM and Signature of SBOM is deleted;
            23. Push image(IA) SBOM to project(PA) by user(UA);
            24. Sign image(IA) with cosign;
            25. Sign image(IA) SBOM with cosign;
            26. Sign image(IA) Signature with cosign;
            27. Trigger GC and wait for GC to succeed;
            28. Check garbage collection details;
            29. Check the log for garbage collection workers;
            30. Get the GC log and check that it is not deleted;
            31. Delete tag of image(IA) by user(UA);
            32. Trigger GC and wait for GC to succeed;
            33. Check garbage collection details;
            34. Check the log for garbage collection workers;
            35. Get the GC log and check that the image(IA) and all aeecssory is deleted;
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
        workers = 3
        gc_id = self.gc.gc_now(workers=3,**ADMIN_CLIENT)
        self.gc.validate_gc_job_status(gc_id, self.gc_success_status, **ADMIN_CLIENT)

        # 10. Check garbage collection details
        gc_history = self.gc.get_gc_history(**ADMIN_CLIENT)
        self.checkGarbageCollectionDetails(gc_history[0], False, False, 1500, 3000, 2, 1, workers)

        # 11. Check the log for garbage collection workers
        gc_log = self.gc.get_gc_log_by_id(gc_id, **ADMIN_CLIENT)
        self.assertIn("workers: {}".format(workers), gc_log)

        # 12. Get the GC log and check that the image(IA) Signature of signature is deleted
        self.assertIn(deleted_prefix + signature_signature_digest, gc_log)

        # 13. Delete image(IA) Signature by user(UA)
        self.artifact.delete_artifact(project_name, self.image, image_signature_digest, **user_client)

        # 14. Trigger GC and wait for GC to succeed
        workers = 4
        gc_id = self.gc.gc_now(workers=workers, **ADMIN_CLIENT)
        self.gc.validate_gc_job_status(gc_id, self.gc_success_status, **ADMIN_CLIENT)

        # 15. Check garbage collection details
        gc_history = self.gc.get_gc_history(**ADMIN_CLIENT)
        self.checkGarbageCollectionDetails(gc_history[0], False, False, 1500, 3000, 2, 1, workers)

        # 16. Check the log for garbage collection workers
        gc_log = self.gc.get_gc_log_by_id(gc_id, **ADMIN_CLIENT)
        self.assertIn("workers: {}".format(workers), gc_log)

        # 17. Get the GC log and check that the image(IA) Signature is deleted
        self.assertIn(deleted_prefix + image_signature_digest, gc_log)

        # 18. Delete image(IA) SBOM by user(UA)
        self.artifact.delete_artifact(project_name, self.image, sbom_digest, **user_client)

        # 19. Trigger GC and wait for GC to succeed
        workers = 5
        gc_id = self.gc.gc_now(workers=workers, **ADMIN_CLIENT)
        self.gc.validate_gc_job_status(gc_id, self.gc_success_status, **ADMIN_CLIENT)

        # 20. Check garbage collection details
        gc_history = self.gc.get_gc_history(**ADMIN_CLIENT)
        self.checkGarbageCollectionDetails(gc_history[0], False, False, 4000, 5000, 4, 2, workers)

        # 21. Check the log for garbage collection workers
        gc_log = self.gc.get_gc_log_by_id(gc_id, **ADMIN_CLIENT)
        self.assertIn("workers: {}".format(workers), gc_log)

        # 22. Get the GC log and check that the image(IA) SBOM and Signature of SBOM is deleted
        self.assertIn(deleted_prefix + sbom_digest, gc_log)
        self.assertIn(deleted_prefix + sbom_signature_digest, gc_log)

        # 23. Push image(IA) SBOM to project(PA) by user(UA)
        self.sbom_path = files_directory + "sbom_test.json"
        docker_api.docker_login_cmd(harbor_server, user_name, user_password, enable_manifest = False)
        cosign.push_artifact_sbom("{}/{}/{}:{}".format(harbor_server, project_name, self.image, self.tag), self.sbom_path)
        artifact_info = self.artifact.get_reference_info(project_name, self.image, self.tag, **user_client)
        sbom_digest = artifact_info.accessories[0].digest

        # 24. Sign image(IA) with cosign
        cosign.sign_artifact("{}/{}/{}:{}".format(harbor_server, project_name, self.image, self.tag))
        artifact_info = self.artifact.get_reference_info(project_name, self.image, self.tag, **user_client)
        self.assertEqual(len(artifact_info.accessories), 2)
        image_signature_digest = None
        for accessory in artifact_info.accessories:
            if accessory.digest != sbom_digest:
                image_signature_digest = accessory.digest
                break

        # 25. Sign image(IA) SBOM cosign
        cosign.sign_artifact("{}/{}/{}@{}".format(harbor_server, project_name, self.image, sbom_digest))
        sbom_info = self.artifact.get_reference_info(project_name, self.image, sbom_digest, **user_client)
        sbom_signature_digest = sbom_info.accessories[0].digest

        # 26. Sign image(IA) Signature with cosign
        cosign.sign_artifact("{}/{}/{}@{}".format(harbor_server, project_name, self.image, image_signature_digest))
        signature_info = self.artifact.get_reference_info(project_name, self.image, image_signature_digest, **user_client)
        signature_signature_digest = signature_info.accessories[0].digest

        # 27. Trigger GC and wait for GC to succeed
        workers = 1
        gc_id = self.gc.gc_now(**ADMIN_CLIENT)
        self.gc.validate_gc_job_status(gc_id, self.gc_success_status, **ADMIN_CLIENT)

        # 28. Check garbage collection details
        gc_history = self.gc.get_gc_history(**ADMIN_CLIENT)
        self.checkGarbageCollectionDetails(gc_history[0], False, False, 0, 0, 0, 0, workers)

        # 29. Check the log for garbage collection workers
        gc_log = self.gc.get_gc_log_by_id(gc_id, **ADMIN_CLIENT)
        self.assertIn("workers: {}".format(workers), gc_log)

        # 30. Get the GC log and check that it is not deleted
        self.assertNotIn(deleted_prefix + sbom_digest, gc_log)
        self.assertNotIn(deleted_prefix + sbom_signature_digest, gc_log)
        self.assertNotIn(deleted_prefix + image_signature_digest, gc_log)
        self.assertNotIn(deleted_prefix + signature_signature_digest, gc_log)

        # 31. Delete tag of image(IA) by user(UA)
        self.artifact.delete_tag(project_name, self.image, self.tag, self.tag, **user_client)

        # 32. Trigger GC and wait for GC to succeed
        workers = 2
        gc_id = self.gc.gc_now(workers=workers, is_delete_untagged=True, **ADMIN_CLIENT)
        self.gc.validate_gc_job_status(gc_id, self.gc_success_status, **ADMIN_CLIENT)

        # 33. Check garbage collection details
        gc_history = self.gc.get_gc_history(**ADMIN_CLIENT)
        self.checkGarbageCollectionDetails(gc_history[0], True, False, 2500000, 3000000, 11, 5, workers)

        # 34. Check the log for garbage collection workers
        gc_log = self.gc.get_gc_log_by_id(gc_id, **ADMIN_CLIENT)
        self.assertIn("workers: {}".format(workers), gc_log)

        # 35. Get the GC log and check that the image(IA) and all aeecssory is deleted
        self.assertIn(self.image, gc_log)
        self.assertIn(deleted_prefix + sbom_digest, gc_log)
        self.assertIn(deleted_prefix + sbom_signature_digest, gc_log)
        self.assertIn(deleted_prefix + image_signature_digest, gc_log)
        self.assertIn(deleted_prefix + signature_signature_digest, gc_log)


    def checkGarbageCollectionDetails(self, gc, delete_untagged, dry_run,
                                      freed_space_range_start, freed_space_range_end,
                                      purged_blobs, purged_manifests, workers):
        self.assertEqual(gc.job_kind, "MANUAL")
        self.assertEqual(gc.job_name, "GARBAGE_COLLECTION")
        parameters = json.loads(gc.job_parameters)
        self.assertEqual(parameters["delete_untagged"], delete_untagged)
        self.assertEqual(parameters["dry_run"], dry_run)
        self.assertTrue(parameters["freed_space"] >= freed_space_range_start)
        self.assertTrue(parameters["freed_space"] <= freed_space_range_end)
        self.assertEqual(parameters["purged_blobs"], purged_blobs)
        self.assertEqual(parameters["purged_manifests"], purged_manifests)
        self.assertEqual(parameters["workers"], workers)


if __name__ == '__main__':
    unittest.main()
