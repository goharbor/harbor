from __future__ import absolute_import


import unittest

from library.base import _assert_status_code
from testutils import ADMIN_CLIENT
from testutils import harbor_server

from testutils import TEARDOWN
from library.project import Project
from library.user import User
from library.repository import Repository
from library.repository import push_image_to_project

class TestProjects(unittest.TestCase):
    @classmethod
    def setUpClass(self):
        self.project= Project()
        self.user= User()
        self.repo= Repository(api_type='repository')

    @classmethod
    def tearDownClass(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete project(PA);
        self.project.delete_project(TestProjects.project_del_repo_id, **TestProjects.USER_del_repo_CLIENT)

        #2. Delete user(UA).
        self.user.delete_user(TestProjects.user_del_repo_id, **ADMIN_CLIENT)

    def testDelRepo(self):
        """
        Test case:
            Delete a repository
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Create a new repository(RA) in project(PA) by user(UA);
            4. Get repository in project(PA), there should be one repository which was created by user(UA);
            5. Delete repository(RA) by user(UA);
            6. Get repository by user(UA), it should get nothing;
        Tear down:
            1. Delete project(PA);
            2. Delete user(UA).
        """
        url = ADMIN_CLIENT["endpoint"]
        user_del_repo_password = "Aa123456"

        #1. Create a new user(UA);
        TestProjects.user_del_repo_id, user_del_repo_name = self.user.create_user(user_password = user_del_repo_password, **ADMIN_CLIENT)

        TestProjects.USER_del_repo_CLIENT=dict(endpoint = url, username = user_del_repo_name, password = user_del_repo_password)

        #2. Create a new project(PA) by user(UA);
        TestProjects.project_del_repo_id, TestProjects.project_del_repo_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_del_repo_CLIENT)

        #3. Create a new repository(RA) in project(PA) by user(UA);
        repo_name, _ = push_image_to_project(TestProjects.project_del_repo_name, harbor_server, 'admin', 'Harbor12345', "hello-world", "latest")

        #4. Get repository in project(PA), there should be one repository which was created by user(UA);
        repo_data = self.repo.get_repository(TestProjects.project_del_repo_name, **TestProjects.USER_del_repo_CLIENT)
        _assert_status_code(repo_name, repo_data[0].name)

        #5. Delete repository(RA) by user(UA);
        self.repo.delete_repoitory(TestProjects.project_del_repo_name, repo_name.split('/')[1], **TestProjects.USER_del_repo_CLIENT)

        #6. Get repository by user(UA), it should get nothing;
        repo_data = self.repo.get_repository(TestProjects.project_del_repo_name, **TestProjects.USER_del_repo_CLIENT)
        _assert_status_code(len(repo_data), 0)

if __name__ == '__main__':
    unittest.main()

