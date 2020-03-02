from __future__ import absolute_import


import unittest

from testutils import ADMIN_CLIENT
from library.project import Project
from library.user import User

class TestProjects(unittest.TestCase):
    """UserGroup unit test stubs"""
    def setUp(self):
        self.project = Project()
        self.user= User()

    def tearDown(self):
        pass

    def testAddProjectMember(self):
        """
        Test case:
            Add a new user to a certain private project as member
        Test step and Expectation:
            1. Login harbor as admin, then to create a user(UA) with non-admin role;
            2. Login harbor as admin, then to create a private project(PA);
            3. Login harbor as user(UA), then to get all private projects, projects count must be zero;
            4. Login harbor as admin, then to add user(UA) in project(PA);
            5. Login harbor as user(UA), then to get all private project, there must be project(PA) only.
        """
        url = ADMIN_CLIENT["endpoint"]
        user_001_password = "Aa123456"

        #1. Create user-001
        user_001_id, user_001_name = self.user.create_user(user_password = user_001_password, **ADMIN_CLIENT)
        self.assertNotEqual(user_001_id, None, msg="Failed to create user, return user is {}".format(user_001_id))

        USER_001_CLIENT=dict(endpoint = url, username = user_001_name, password = user_001_password)

        #2. Create private project-001
        project_001_id, project_001_name = self.project.create_project(metadata = {"public": "false"}, **ADMIN_CLIENT)
        self.assertNotEqual(project_001_name, None, msg="Project was created failed, return project name is  {} and  id is {}.".format(project_001_name, project_001_id))

        #3.1 Get private projects of user-001
        project_001_data = self.project.get_projects(dict(public=False), **USER_001_CLIENT)

        #3.2 Check user-001 has no any private project
        self.assertEqual(project_001_data, None, msg="user-001 should has no any private project, but we got {}".format(project_001_data))

        #4. Add user-001 as a member of project-001
        result = self.project.add_project_members(project_001_id, user_001_id, **ADMIN_CLIENT)
        self.assertNotEqual(result, False, msg="Failed to add member user_001 to project_001, result is {}".format(result))


        #5 Get private project of uesr-001, uesr-001 can see only one private project which is project-001
        project_data = self.project.get_projects(dict(public=False), **USER_001_CLIENT)
        self.assertEqual(len(project_data), 1, msg="Private project count should be 1.")
        self.assertEqual(str(project_data[0].project_id), str(project_001_id), msg="Project-id check failed, please check this test case.")

if __name__ == '__main__':
    unittest.main()

