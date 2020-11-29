from __future__ import absolute_import
import unittest

from testutils import ADMIN_CLIENT, suppress_urllib3_warning
from testutils import TEARDOWN
from library.project import Project
from library.user import User
from library.configurations import Configurations

class TestProjects(unittest.TestCase):
    @suppress_urllib3_warning
    def setUp(self):
        self.conf= Configurations()
        self.project= Project()
        self.user= User()

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def tearDown(self):
        #1. Delete project(PA);
        self.project.delete_project(TestProjects.project_edit_project_creation_id, **TestProjects.USER_edit_project_creation_CLIENT)

        #2. Delete user(UA);
        self.user.delete_user(TestProjects.user_edit_project_creation_id, **ADMIN_CLIENT)

    def testEditProjectCreation(self):
        """
        Test case:
            Edit Project Creation
        Test step and expected result:
            1. Create a new user(UA);
            2. Set project creation to "admin only";
            3. Create a new project(PA) by user(UA), and fail to create a new project;
            4. Set project creation to "everyone";
            5. Create a new project(PA) by user(UA), success to create a project.
        Tear down:
            1. Delete project(PA);
            2. Delete user(UA);
        """
        url = ADMIN_CLIENT["endpoint"]
        user_edit_project_creation_password = "Aa123456"

        #1. Create a new user(UA);
        TestProjects.user_edit_project_creation_id, user_edit_project_creation_name = self.user.create_user(user_password = user_edit_project_creation_password, **ADMIN_CLIENT)

        TestProjects.USER_edit_project_creation_CLIENT=dict(endpoint = url, username = user_edit_project_creation_name, password = user_edit_project_creation_password)

        #2. Set project creation to "admin only";
        self.conf.set_configurations_of_project_creation_restriction("adminonly", **ADMIN_CLIENT)

        #3. Create a new project(PA) by user(UA), and fail to create a new project;
        self.project.create_project(metadata = {"public": "false"}, expect_status_code = 403,
            expect_response_body = "{\"errors\":[{\"code\":\"FORBIDDEN\",\"message\":\"Only system admin can create project\"}]}", **TestProjects.USER_edit_project_creation_CLIENT)

        #4. Set project creation to "everyone";
        self.conf.set_configurations_of_project_creation_restriction("everyone", **ADMIN_CLIENT)

        #5. Create a new project(PA) by user(UA), success to create a project.
        TestProjects.project_edit_project_creation_id, _ = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_edit_project_creation_CLIENT)


if __name__ == '__main__':
    unittest.main()