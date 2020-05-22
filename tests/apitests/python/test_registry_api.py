from __future__ import absolute_import
import unittest

from testutils import harbor_server
from testutils import admin_user
from testutils import admin_pwd
from testutils import TEARDOWN
from testutils import ADMIN_CLIENT
from library.system import System
from library.project import Project
from library.user import User
from library.repository import Repository
from library.repository import push_image_to_project
from library.artifact import Artifact
from library.scanner import Scanner
import os
import library.base
import json

class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        self.system = System()
        self.project= Project()
        self.user= User()
        self.artifact = Artifact()
        self.repo = Repository()
        self.repo_name = "hello-world"

    @classmethod
    def tearDown(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete Alice's repository and Luca's repository;
        self.repo.delete_repoitory(TestProjects.project_Alice_name, TestProjects.repo_a.split('/')[1], **ADMIN_CLIENT)
        self.repo.delete_repoitory(TestProjects.project_Alice_name, TestProjects.repo_b.split('/')[1], **ADMIN_CLIENT)
        self.repo.delete_repoitory(TestProjects.project_Alice_name, TestProjects.repo_c.split('/')[1], **ADMIN_CLIENT)

        #2. Delete Alice's project and Luca's project;
        self.project.delete_project(TestProjects.project_Alice_id, **ADMIN_CLIENT)

        #3. Delete user Alice and Luca.
        self.user.delete_user(TestProjects.user_Alice_id, **ADMIN_CLIENT)

    def testRegistryAPI(self):
        """
        Test case:
            Catalog API to list all repositories by system admin
        Test step and expected result:G
            1. Create user Alice;
            2. Create 1 new private projects project_Alice;
            3. Push 3 images to project_Alice and Add 3 tags to the 3rd image
            4. Call the image_list_tag API
            5. Call the catalog API using admin account without pagination, can get all 3 images
            5.1 Call the catalog API using admin account with pagination n=1, page=2, twice to get all 3 images.
            5.2 Call the catalog API using Alice account, no repos should be found.
        Tear down:
            1. Delete Alice's repository;
            2. Delete Alice's project;
            3. Delete user Alice.
        """
        url = ADMIN_CLIENT["endpoint"]
        user_common_password = "Aa123456"
        #1. Create user Alice and Luca;
        TestProjects.user_Alice_id, user_Alice_name = self.user.create_user(user_password = user_common_password, **ADMIN_CLIENT)

        USER_ALICE_CLIENT=dict(endpoint = url, username = user_Alice_name, password = user_common_password)

        #2. Create 2 new private projects project_Alice and project_Luca;
        TestProjects.project_Alice_id, TestProjects.project_Alice_name = self.project.create_project(metadata = {"public": "false"}, **USER_ALICE_CLIENT)

        #3. Push 3 images to project_Alice and Add 3 tags to the 3rd image.

        src_tag = "latest"
        image_a = "busybox"
        TestProjects.repo_a, tag_a = push_image_to_project(TestProjects.project_Alice_name, harbor_server, user_Alice_name, user_common_password, image_a, src_tag)
        image_b = "alpine"
        TestProjects.repo_b, tag_b = push_image_to_project(TestProjects.project_Alice_name, harbor_server, user_Alice_name, user_common_password, image_b, src_tag)
        image_c = "hello-world"
        TestProjects.repo_c, tag_c = push_image_to_project(TestProjects.project_Alice_name, harbor_server, user_Alice_name, user_common_password, image_c, src_tag)
        create_tags = ["1.0","2.0","3.0"]
        for tag in create_tags:
            self.artifact.create_tag(TestProjects.project_Alice_name, self.repo_name, tag_c, tag, **USER_ALICE_CLIENT)
        #4. Call the image_list_tags API
        tags = library.docker_api.list_image_tags(harbor_server,TestProjects.repo_c,user_Alice_name,user_common_password)
	for tag in create_tags:
            self.assertTrue(tags.count(tag)>0, "Expect tag: %s is not listed"%tag)
        page_tags = library.docker_api.list_image_tags(harbor_server,TestProjects.repo_c,user_Alice_name,user_common_password,len(tags)/2+1)
        page_tags += library.docker_api.list_image_tags(harbor_server,TestProjects.repo_c,user_Alice_name,user_common_password,len(tags)/2+1,tags[len(tags)/2])
	for tag in create_tags:
            self.assertTrue(page_tags.count(tag)>0, "Expect tag: %s is not listed by the pagination query"%tag)
        #5. Call the catalog API;
        repos = library.docker_api.list_repositories(harbor_server,admin_user,admin_pwd)
	self.assertTrue(repos.count(TestProjects.repo_a)>0 and repos.count(TestProjects.repo_b)>0 and repos.count(TestProjects.repo_c)>0, "Expected repo not found")
        for repo in [TestProjects.repo_a,TestProjects.repo_b,TestProjects.repo_c]:
            self.assertTrue(repos.count(repo)>0,"Expected repo: %s is not listed"%repo)
        page_repos = library.docker_api.list_repositories(harbor_server,admin_user,admin_pwd,len(repos)/2+1)
        page_repos += library.docker_api.list_repositories(harbor_server,admin_user,admin_pwd,len(repos)/2+1,repos[len(repos)/2])
        for repo in [TestProjects.repo_a,TestProjects.repo_b,TestProjects.repo_c]:
            self.assertTrue(page_repos.count(repo)>0,"Expected repo: %s is not listed by the pagination query"%repo)

        null_repos = library.docker_api.list_repositories(harbor_server,user_Alice_name,user_common_password)
        self.assertEqual(null_repos, "")

if __name__ == '__main__':
    unittest.main()
