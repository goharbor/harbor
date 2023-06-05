# -*- coding: utf-8 -*-

from __future__ import absolute_import
import unittest

from testutils import harbor_server, files_directory, ADMIN_CLIENT, suppress_urllib3_warning
from library import cosign, referrers_api
from library.project import Project
from library.user import User
from library.artifact import Artifact
from library.repository import push_self_build_image_to_project
from library import docker_api

class TestReferrersApi(unittest.TestCase):

    @suppress_urllib3_warning
    def setUp(self):
        self.project= Project()
        self.user= User()
        self.artifact = Artifact()
        self.image = "artifact_test"
        self.tag = "dev"
        self.sbom_path = files_directory + "sbom_test.json"
        self.sbom_artifact_type = "application/vnd.dev.cosign.artifact.sbom.v1+json"
        self.signature_artifact_type = "application/vnd.oci.image.config.v1+json"

    def testReferrersApi(self):
        """
        Test case:
            Referrers Api
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Push a new image(IA) in project(PA) by user(UA);
            4. Push image(IA) SBOM to project(PA) by user(UA);
            5. Sign image(IA) with cosign;
            6. Sign image(IA) SBOM with cosign;
            7. Call the referrers api successfully;
            8. Call the referrers api and filter artifact_type;
        """
        url = ADMIN_CLIENT["endpoint"]
        user_password = "Aa123456"

        # 1. Create user(UA)
        _, user_name = self.user.create_user(user_password = user_password, **ADMIN_CLIENT)
        user_client = dict(endpoint = url, username = user_name, password = user_password, with_accessory = True)

        # 2. Create private project(PA) by user(UA)
        _, project_name = self.project.create_project(metadata = {"public": "false"}, **user_client)

        # 3. Push a new image(IA) in project(PA) by user(UA)
        push_self_build_image_to_project(project_name, harbor_server, user_name, user_password, self.image, self.tag)

        # 4. Push image(IA) SBOM to project(PA) by user(UA)
        docker_api.docker_login_cmd(harbor_server, user_name, user_password, enable_manifest = False)
        cosign.push_artifact_sbom("{}/{}/{}:{}".format(harbor_server, project_name, self.image, self.tag), self.sbom_path)
        artifact_info = self.artifact.get_reference_info(project_name, self.image, self.tag, **user_client)
        artifact_digest = artifact_info.digest
        sbom_digest = artifact_info.accessories[0].digest

        # 5. Sign image(IA) with cosign
        cosign.generate_key_pair()
        cosign.sign_artifact("{}/{}/{}:{}".format(harbor_server, project_name, self.image, self.tag))
        artifact_info = self.artifact.get_reference_info(project_name, self.image, self.tag, **user_client)
        self.assertEqual(len(artifact_info.accessories), 2)
        signature_digest = None
        for accessory in artifact_info.accessories:
            if accessory.digest != sbom_digest:
                signature_digest = accessory.digest
                break

        # 6. Sign image(IA) SBOM cosign
        cosign.sign_artifact("{}/{}/{}@{}".format(harbor_server, project_name, self.image, sbom_digest))

        # 7. Call the referrers api successfully
        res_json = referrers_api.call(harbor_server, project_name, self.image, artifact_digest, **user_client).json()
        self.assertEqual(len(res_json["manifests"]), 2)
        for  manifest in res_json["manifests"]:
            self.assertIn(manifest["digest"], [signature_digest, sbom_digest])
            self.assertIn(manifest["artifactType"], [self.signature_artifact_type, self.sbom_artifact_type])
            self.assertIsNotNone(manifest["mediaType"])
            self.assertIsNotNone(manifest["size"])

        res_json = referrers_api.call(harbor_server, project_name, self.image, sbom_digest, **user_client).json()
        self.assertEqual(len(res_json["manifests"]), 1)
        manifest = res_json["manifests"][0]
        self.assertIsNotNone(manifest["digest"])
        self.assertIsNotNone(manifest["artifactType"], [self.signature_artifact_type, self.sbom_artifact_type])
        self.assertIsNotNone(manifest["mediaType"])
        self.assertIsNotNone(manifest["size"])

        # 8. Call the referrers api and filter artifact_type
        res = referrers_api.call(harbor_server, project_name, self.image, artifact_digest, self.sbom_artifact_type, **user_client)
        self.assertEqual(res.headers["Oci-Filters-Applied"], "artifactType")
        res_json = res.json()
        self.assertEqual(len(res_json["manifests"]), 1)
        manifest = res_json["manifests"][0]
        self.assertEqual(manifest["digest"], sbom_digest)
        self.assertIn(manifest["artifactType"], self.sbom_artifact_type)
        self.assertIsNotNone(manifest["mediaType"])
        self.assertIsNotNone(manifest["size"])

        res = referrers_api.call(harbor_server, project_name, self.image, artifact_digest, self.signature_artifact_type, **user_client)
        self.assertEqual(res.headers["Oci-Filters-Applied"], "artifactType")
        res_json = res.json()
        self.assertEqual(len(res_json["manifests"]), 1)
        manifest = res_json["manifests"][0]
        self.assertEqual(manifest["digest"], signature_digest)
        self.assertIn(manifest["artifactType"], self.signature_artifact_type)
        self.assertIsNotNone(manifest["mediaType"])
        self.assertIsNotNone(manifest["size"])

if __name__ == '__main__':
    unittest.main()
