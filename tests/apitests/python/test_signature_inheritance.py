# -*- coding: utf-8 -*-

from __future__ import absolute_import
import unittest

from testutils import harbor_server, suppress_urllib3_warning
from testutils import ADMIN_CLIENT
from testutils import TEARDOWN
from library import cosign
from library import notation
from library import docker_api
from library.project import Project
from library.user import User
from library.repository import Repository
from library.repository import push_self_build_image_to_project
from library.artifact import Artifact

class TestSignatureInheritance(unittest.TestCase):

    @suppress_urllib3_warning
    def setUp(self):
        self.project = Project()
        self.user = User()
        self.artifact = Artifact()
        self.repo = Repository()
        self.image_a = "alpine"
        self.image_b = "busybox"
        self.index_name = "ci_test_inherited_signature_index"
        self.index_tag = "test_tag"
        self.expect_cosign_type = "signature.cosign"
        self.expect_notation_type = "signature.notation"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def tearDown(self):
        #1. Delete repositories by user(UA);
        self.repo.delete_repository(TestSignatureInheritance.project_name, self.index_name, **TestSignatureInheritance.user_client)
        self.repo.delete_repository(TestSignatureInheritance.project_name, self.image_a, **TestSignatureInheritance.user_client)
        self.repo.delete_repository(TestSignatureInheritance.project_name, self.image_b, **TestSignatureInheritance.user_client)
        #2. Delete project(PA);
        self.project.delete_project(TestSignatureInheritance.project_id, **TestSignatureInheritance.user_client)
        #3. Delete user(UA).
        self.user.delete_user(TestSignatureInheritance.user_id, **ADMIN_CLIENT)

    def testSignatureInheritance(self):
        """
        Test case:
            Signature Inheritance For OCI Index Children
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Push 2 images(IA,IB) in project(PA) and combine them into an index(IC);
            4. Verify that a child of index(IC) is not signed and inherits nothing;
            5. Sign index(IC) with cosign and with notation;
            6. Verify that index(IC) carries both signatures as its own accessories;
            7. Verify that a child of index(IC) inherits both signatures, and that each
               inherited accessory has the digest of index(IC) as its subject;
            8. Verify that the child still owns no accessory, so the accessory list and
               the default artifact response are unchanged by the inheritance.
        Tear down:
            1. Delete repositories;
            2. Delete project(PA);
            3. Delete user(UA).
        """
        url = ADMIN_CLIENT["endpoint"]
        user_password = "Aa123456"

        # 1. Create user(UA)
        TestSignatureInheritance.user_id, user_name = self.user.create_user(user_password = user_password, **ADMIN_CLIENT)
        TestSignatureInheritance.user_client = dict(endpoint = url, username = user_name, password = user_password)

        # 2. Create private project(PA) by user(UA)
        TestSignatureInheritance.project_id, TestSignatureInheritance.project_name = self.project.create_project(metadata = {"public": "false"}, **TestSignatureInheritance.user_client)

        # 3.1. Push 2 images(IA,IB) in project(PA) by user(UA)
        repo_name_a, tag_a = push_self_build_image_to_project(TestSignatureInheritance.project_name, harbor_server, user_name, user_password, self.image_a, "latest")
        repo_name_b, tag_b = push_self_build_image_to_project(TestSignatureInheritance.project_name, harbor_server, user_name, user_password, self.image_b, "latest")

        # 3.2. Combine them into an index(IC)
        manifests = [harbor_server + "/" + repo_name_a + ":" + tag_a, harbor_server + "/" + repo_name_b + ":" + tag_b]
        index = harbor_server + "/" + TestSignatureInheritance.project_name + "/" + self.index_name + ":" + self.index_tag
        index_digest, child_digests = docker_api.docker_manifest_push_to_harbor(index, manifests, harbor_server, user_name, user_password)
        child_digest = child_digests[0]

        # 4.1. The child owns no accessory;
        child = self.artifact.get_reference_info(TestSignatureInheritance.project_name, self.index_name, child_digest,
                                                 with_accessory = True, **TestSignatureInheritance.user_client)
        self.assertIsNone(child.accessories)
        # 4.2. And it inherits none either, the index is not signed yet;
        child = self.artifact.get_reference_info(TestSignatureInheritance.project_name, self.index_name, child_digest,
                                                 with_inherited_accessory = True, **TestSignatureInheritance.user_client)
        self.assertIsNone(child.inherited_accessories)

        # 5.1. Sign index(IC) with cosign;
        cosign.generate_key_pair()
        docker_api.docker_login_cmd(harbor_server, user_name, user_password, enable_manifest = False)
        cosign.sign_artifact(index)
        # 5.2. Sign index(IC) with notation;
        notation.generate_cert(common_name = "inherited-signature.io")
        notation.sign_artifact(index)

        # 6. Index(IC) carries both signatures of its own;
        index_info = self.artifact.get_reference_info(TestSignatureInheritance.project_name, self.index_name, self.index_tag,
                                                      with_accessory = True, **TestSignatureInheritance.user_client)
        self.assertEqual(index_info.digest, index_digest)
        own_types = [accessory.type for accessory in index_info.accessories]
        self.assertIn(self.expect_cosign_type, own_types)
        self.assertIn(self.expect_notation_type, own_types)

        # 7.1. The child inherits both of them;
        child = self.artifact.get_reference_info(TestSignatureInheritance.project_name, self.index_name, child_digest,
                                                 with_accessory = True, with_inherited_accessory = True, **TestSignatureInheritance.user_client)
        inherited_types = [accessory.type for accessory in child.inherited_accessories]
        self.assertIn(self.expect_cosign_type, inherited_types)
        self.assertIn(self.expect_notation_type, inherited_types)
        # 7.2. Every inherited accessory describes the index, not the child. This is what
        #      keeps the API honest: verifying any of them against the child digest fails;
        for accessory in child.inherited_accessories:
            self.assertEqual(accessory.subject_artifact_digest, index_digest)

        # 8.1. The child still owns no accessory;
        self.assertIsNone(child.accessories)
        # 8.2. So the accessory list of the child stays empty;
        accessory_list = self.artifact.list_accessories(TestSignatureInheritance.project_name, self.index_name, child_digest,
                                                        **TestSignatureInheritance.user_client)
        self.assertTrue(len(accessory_list) == 0)
        # 8.3. And a request that does not ask for inherited accessories gets none;
        child = self.artifact.get_reference_info(TestSignatureInheritance.project_name, self.index_name, child_digest,
                                                 **TestSignatureInheritance.user_client)
        self.assertIsNone(child.inherited_accessories)

if __name__ == '__main__':
    unittest.main()
