from __future__ import absolute_import


import unittest
import urllib
import sys

from testutils import ADMIN_CLIENT, suppress_urllib3_warning, DOCKER_USER, DOCKER_PWD
from testutils import harbor_server
from testutils import TEARDOWN
from library.base import _random_name
from library.base import _assert_status_code
from library.project import Project
from library.user import User
from library.repository import Repository
from library.registry import Registry
from library.repository import pull_harbor_image
from library.artifact import Artifact
import library.containerd

class TestProxyCache(unittest.TestCase):
    @suppress_urllib3_warning
    def setUp(self):
        self.url = ADMIN_CLIENT["endpoint"]
        self.user_password = "Aa123456"
        self.project= Project()
        self.user= User()
        self.repo= Repository()
        self.registry = Registry()
        self.artifact = Artifact()

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def tearDown(self):
        print("Case completed")

    def do_validate(self, registry_type):
        """
        Test case:
            Proxy Cache Image From Harbor
        Test step and expected result:
            1. Create a new registry;
            2. Create a new project;
            3. Add a new user as a member of project;
            4. Pull image from this project by docker CLI;
            5. Pull image from this project by ctr CLI;
            6. Pull manifest index from this project by docker CLI;
            7. Pull manifest from this project by ctr CLI;
            8. Image pulled by docker CLI should be cached;
            9. Image pulled by ctr CLI should be cached;
            10. Manifest index pulled by docker CLI should be cached;
            11. Manifest index pulled by ctr CLI should be cached;
        Tear down:
            1. Delete project(PA);
            2. Delete user(UA).
        """
        user_id, user_name = self.user.create_user(user_password = self.user_password, **ADMIN_CLIENT)
        USER_CLIENT=dict(with_signature = True, endpoint = self.url, username = user_name, password = self.user_password)

        image_for_docker = dict(image = "for_proxy", tag = "1.0")
        image_for_ctr = dict(image = "redis", tag = "latest")
        index_for_docker = dict(image = "index081597864867", tag = "index_tag081597864867")
        access_key = ""
        access_secret = ""

        #1. Create a new registry;
        if registry_type == "docker-hub":
            user_namespace = DOCKER_USER
            access_key = user_namespace
            access_secret = DOCKER_PWD
            registry = "https://hub.docker.com"
            # Memo: ctr will not send image pull request if manifest list already exist, so we pull different manifest list for different registry;
            index_for_ctr = dict(image = "alpine", tag = "3.12.0")
        else:
            user_namespace = "nightly"
            registry = "https://cicd.harbor.vmwarecna.net"
            index_for_ctr = dict(image = "busybox", tag = "1.32.0")

        registry_id, _ = self.registry.create_registry(registry, name=_random_name(registry_type), registry_type=registry_type, access_key = access_key, access_secret = access_secret, insecure=False, **ADMIN_CLIENT)

        print("registry_id:", registry_id)

        #2. Create a new project;
        project_id, project_name = self.project.create_project(registry_id = registry_id, metadata = {"public": "false"}, **ADMIN_CLIENT)
        print("project_id:",project_id)
        print("project_name:",project_name)

        #3. Add a new user as a member of project;
        self.project.add_project_members(project_id, user_id=user_id, **ADMIN_CLIENT)

        #4. Pull image from this project by docker CLI;
        pull_harbor_image(harbor_server, USER_CLIENT["username"], USER_CLIENT["password"], project_name + "/" +  user_namespace + "/" + image_for_docker["image"], image_for_docker["tag"])

        #5. Pull image from this project by ctr CLI;
        oci_ref = harbor_server + "/" + project_name + "/" + user_namespace + "/" + image_for_ctr["image"] + ":" + image_for_ctr["tag"]
        library.containerd.ctr_images_pull(user_name, self.user_password, oci_ref)
        library.containerd.ctr_images_list(oci_ref = oci_ref)

        #6. Pull manifest index from this project by docker CLI;
        index_repo_name =  user_namespace + "/" + index_for_docker["image"]
        pull_harbor_image(harbor_server, user_name, self.user_password, project_name + "/" + index_repo_name, index_for_docker["tag"])

        #7. Pull manifest from this project by ctr CLI;
        index_repo_name_for_ctr =  user_namespace + "/" + index_for_ctr["image"]
        oci_ref = harbor_server + "/" + project_name + "/" + index_repo_name_for_ctr + ":" + index_for_ctr["tag"]
        library.containerd.ctr_images_pull(user_name, self.user_password, oci_ref)
        library.containerd.ctr_images_list(oci_ref = oci_ref)

        #8. Image pulled by docker CLI should be cached;
        self.artifact.waiting_for_reference_exist(project_name, urllib.parse.quote(user_namespace + "/" + image_for_docker["image"],'utf-8'), image_for_docker["tag"], **USER_CLIENT)

        #9. Image pulled by ctr CLI should be cached;
        self.artifact.waiting_for_reference_exist(project_name, urllib.parse.quote(user_namespace + "/" + image_for_ctr["image"],'utf-8'), image_for_ctr["tag"], **USER_CLIENT)

        #10. Manifest index pulled by docker CLI should be cached;
        ret_index_by_d = self.artifact.waiting_for_reference_exist(project_name, urllib.parse.quote(index_repo_name,'utf-8'), index_for_docker["tag"], **USER_CLIENT)
        print("Index's reference by docker CLI:", ret_index_by_d.references)
        self.assertTrue(len(ret_index_by_d.references) == 1)

        #11. Manifest index pulled by ctr CLI should be cached;
        ret_index_by_c = self.artifact.waiting_for_reference_exist(project_name, urllib.parse.quote(index_repo_name_for_ctr,'utf-8'), index_for_ctr["tag"], **USER_CLIENT)
        print("Index's reference by ctr CLI:", ret_index_by_c.references)
        self.assertTrue(len(ret_index_by_c.references) == 1)

    def test_proxy_cache_from_harbor(self):
        self.do_validate("harbor")

    #def test_proxy_cache_from_dockerhub(self):
    #    self.do_validate("docker-hub")

if __name__ == '__main__':
    suite = unittest.TestSuite(unittest.makeSuite(TestProxyCache))
    result = unittest.TextTestRunner(sys.stdout, verbosity=2, failfast=True).run(suite)
    if not result.wasSuccessful():
        raise Exception(r"Proxy cache test failed: ".format(result))

