from __future__ import absolute_import

import unittest
from urllib.parse import quote

from testutils import ADMIN_CLIENT, TEARDOWN, harbor_server, suppress_urllib3_warning
from library.artifact import Artifact
from library.base import _random_name
from library.project import Project
from library.registry import Registry
from library.repository import Repository, pull_harbor_image
from library.scan import Scan


class TestProxyCacheSBOMGeneration(unittest.TestCase):
    @suppress_urllib3_warning
    def setUp(self):
        self.artifact = Artifact()
        self.project = Project()
        self.registry = Registry()
        self.repo = Repository()
        self.scan = Scan()

    def _delete_repositories(self, project_name):
        for repository in self.repo.list_repositories(project_name, **ADMIN_CLIENT):
            repo_name = repository.name.split("/", 1)[1]
            self.repo.delete_repository(project_name, quote(repo_name, safe=""), **ADMIN_CLIENT)

    def test_generate_sbom_of_proxy_cache_artifact(self):
        """Regression for #23776: the SBOM job must push into a proxy project.

        Create a proxy cache project, pull an image, verify it is cached,
        generate its SBOM, and verify the generated accessory is stored.
        """
        registry_id, _ = self.registry.create_registry(
            "https://registry.goharbor.io",
            registry_type="harbor",
            name=_random_name("proxy-sbom"),
            access_key="",
            access_secret="",
            **ADMIN_CLIENT
        )
        if TEARDOWN:
            self.addCleanup(self.registry.delete_registry, registry_id, **ADMIN_CLIENT)

        project_id, project_name = self.project.create_project(
            registry_id=registry_id,
            metadata={"public": "false", "auto_sbom_generation": "false"},
            **ADMIN_CLIENT
        )
        if TEARDOWN:
            self.addCleanup(self.project.delete_project, project_id, **ADMIN_CLIENT)
            self.addCleanup(self._delete_repositories, project_name)

        repo_name = "nightly/for_proxy"
        tag = "1.0"
        pull_harbor_image(
            harbor_server,
            ADMIN_CLIENT["username"],
            ADMIN_CLIENT["password"],
            project_name + "/" + repo_name,
            tag,
        )

        api_repo_name = quote(repo_name, safe="")
        cached = self.artifact.waiting_for_reference_exist(
            project_name, api_repo_name, tag, with_sbom_overview=True, **ADMIN_CLIENT
        )
        self.assertTrue(cached.digest)
        self.assertFalse(cached.references, "The SBOM fixture must be a single-platform image")
        self.assertIsNone(cached.sbom_overview)

        self.scan.sbom_generation_of_artifact(
            project_name, api_repo_name, cached.digest, **ADMIN_CLIENT
        )

        self.artifact.check_image_sbom_generation_result(
            project_name, api_repo_name, cached.digest,
            with_sbom_overview=True, **ADMIN_CLIENT
        )
        generated = self.artifact.get_reference_info(
            project_name, api_repo_name, cached.digest,
            with_sbom_overview=True, **ADMIN_CLIENT
        )
        sbom_digest = generated.sbom_overview.sbom_digest
        self.assertTrue(sbom_digest, "Successful SBOM generation must report an accessory digest")
        sbom = self.artifact.waiting_for_reference_exist(
            project_name, api_repo_name, sbom_digest, **ADMIN_CLIENT
        )
        self.assertEqual(sbom.type, "SBOM")
        accessories = self.artifact.list_accessories(
            project_name, api_repo_name, cached.digest, **ADMIN_CLIENT
        )
        self.assertTrue(
            any(
                accessory.type == "sbom.harbor"
                and accessory.digest == sbom_digest
                and accessory.subject_artifact_digest == cached.digest
                and accessory.subject_artifact_repo == project_name + "/" + repo_name
                for accessory in accessories or []
            ),
            "The generated SBOM must be attached to the cached image",
        )


if __name__ == "__main__":
    unittest.main(verbosity=2, failfast=True)
