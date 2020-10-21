from __future__ import absolute_import


import unittest
import urllib

from testutils import ADMIN_CLIENT
from testutils import harbor_server
from testutils import TEARDOWN
from library.base import _assert_status_code
from library.project import Project
from library.user import User
from library.repository import Repository
from library.repository import push_image_to_project
from library.registry import Registry
from library.repository import pull_harbor_image
from library.artifact import Artifact
import library.containerd

class TestProjects(unittest.TestCase):
    @classmethod
    def setUpClass(self):
        url = ADMIN_CLIENT["endpoint"]
        self.user_password = "Aa123456"
        self.project= Project()
        self.user= User()
        self.repo= Repository()
        self.registry = Registry()
        self.artifact = Artifact()
        self.project_id, self.project_name, self.registry_id, self.user_id, self.user_name = [None] * 5
        self.user_id, self.user_name = self.user.create_user(user_password = self.user_password, **ADMIN_CLIENT)
        self.USER_CLIENT=dict(with_signature = True, endpoint = url, username = self.user_name, password = self.user_password)

    @classmethod
    def tearDownClass(self):
        print("Case completed")

    @unittest.skipIf(TEARDOWN == True, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete project(PA);
        self.project.delete_project(self.project_id , **self.USER_CLIENT)

        #2. Delete user(UA).
        self.user.delete_user(self.user_id, **ADMIN_CLIENT)

    def testDelRepo(self):
        """
        Test case:
            Proxy Cache Image
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
        Tear down:
            1. Delete project(PA);
            2. Delete user(UA).
        """
        user_namespace = "nightly"
        image = "for_proxy"
        tag = "1.0"
        image_for_ctr = "redis"
        tag_for_ctr = "latest"
        index_name = "index081597864867"
        index_tag = "index_tag081597864867"
        #1. Create a new registry;
        #self.registry_id, _ = self.registry.create_registry("https://hub.docker.com", registry_type="docker-hub", access_key = "", access_secret = "", insecure=False, **ADMIN_CLIENT)
        self.registry_id, _ = self.registry.create_registry("https://cicd.harbor.vmwarecna.net", registry_type="harbor", access_key = "", access_secret = "", insecure=False, **ADMIN_CLIENT)

        print("registry_id:", self.registry_id)

        #2. Create a new project;
        project_id, self.project_name = self.project.create_project(registry_id = self.registry_id, metadata = {"public": "false"}, **ADMIN_CLIENT)
        print(project_id)
        print(self.project_name)

        result = self.project.add_project_members(project_id, user_id=self.user_id, **ADMIN_CLIENT)
        self.assertNotEqual(result, False, msg="Failed to add member user_001 to project_001, result is {}".format(result))

        #3. Pull image from this project;
        pull_harbor_image(harbor_server, self.USER_CLIENT["username"], self.USER_CLIENT["password"], self.project_name + "/" +  user_namespace + "/" + image, tag)

        # ctr pull image
        oci_ref = harbor_server + "/" + self.project_name + "/" + user_namespace + "/" + image_for_ctr + ":" + tag_for_ctr
        library.containerd.ctr_images_pull(self.user_name, self.user_password, oci_ref)
        library.containerd.ctr_images_list(oci_ref = oci_ref)

        # Pull index
        index_repo_name =  user_namespace + "/" + index_name
        pull_harbor_image(harbor_server, self.user_name, self.user_password, self.project_name + "/" + index_repo_name, index_tag)

        repo_name = urllib.parse.quote(user_namespace + "/" + image,'utf-8')
        self.artifact.check_reference_exist(self.project_name, repo_name, tag, **self.USER_CLIENT)

        repo_name_for_ctr = urllib.parse.quote(user_namespace + "/" + image_for_ctr,'utf-8')
        self.artifact.check_reference_exist(self.project_name, repo_name_for_ctr, tag_for_ctr, **self.USER_CLIENT)

        self.artifact.check_reference_exist(self.project_name, urllib.parse.quote(index_repo_name,'utf-8'), index_tag, **self.USER_CLIENT)

if __name__ == '__main__':
    unittest.main()

