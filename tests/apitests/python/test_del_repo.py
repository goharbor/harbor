from __future__ import absolute_import
import os
import sys
import unittest

from library.base import _assert_status_code
from testutils import CLIENT
from testutils import harbor_server
from testutils import USER_ROLE
from testutils import TEARDOWN
from library.project import Project
from library.user import User
from library.repository import Repository

class TestProjects(unittest.TestCase):
    @classmethod
    def setUpClass(self):
        project = Project()
        self.project= project

        user = User()
        self.user= user
          
        repo = Repository()
        self.repo= repo
        pass

    @classmethod
    def tearDownClass(self):
        pass

    @unittest.skipIf(TEARDOWN == False, "Test data should be remain in the harbor.")
    def test02ClearData(self):
        #1. Delete project(PA);
        self.project.delete_project(TestProjects.project_001_id, **TestProjects.USER_001_CLIENT)

        #2. Delete user(UA).
        self.user.delete_user(TestProjects.user_001_id, **TestProjects.ADMIN_CLIENT)
        pass

    def test01DelRepo(self):
        """
        Test case: 
            Delete a repository
        Test step & Expectation:
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
        admin_user = "admin"
        admin_pwd = "Harbor12345"
        url = CLIENT["endpoint"]
        user_001_password = "Aa123456"
        TestProjects.ADMIN_CLIENT=dict(endpoint = url, username = admin_user, password =  admin_pwd)
        
        #1. Create a new user(UA);
        TestProjects.user_001_id, user_001_name = self.user.create_user_success(user_password = user_001_password, **TestProjects.ADMIN_CLIENT)
        
        TestProjects.USER_001_CLIENT=dict(endpoint = url, username = user_001_name, password = user_001_password)

        #2. Create a new project(PA) by user(UA);
        project_001_name, TestProjects.project_001_id = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_001_CLIENT)

        #3. Create a new repository(RA) in project(PA) by user(UA);
        repo_name, tag = self.repo.create_repository(project_001_name, registry = harbor_server)

        #4. Get repository in project(PA), there should be one repository which was created by user(UA);
        repo_data = self.repo.get_repository(TestProjects.project_001_id, **TestProjects.USER_001_CLIENT)
        _assert_status_code(repo_name, repo_data[0].name)

        #5. Delete repository(RA) by user(UA);
        self.repo.delete_repoitory(repo_name, **TestProjects.USER_001_CLIENT)

        #6. Get repository by user(UA), it should get nothing;
        repo_data = self.repo.get_repository(TestProjects.project_001_id, **TestProjects.USER_001_CLIENT)
        _assert_status_code(len(repo_data), 0)

        pass

if __name__ == '__main__':
    unittest.main()

