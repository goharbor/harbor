from __future__ import absolute_import
import unittest

from testutils import CLIENT
from testutils import harbor_server
from testutils import TEARDOWN
from library.project import Project
from library.user import User
from library.repository import Repository
from library.label import Label

class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        project = Project()
        self.project= project

        user = User()
        self.user= user

        repo = Repository()
        self.repo= repo

        label = Label()
        self.label= label

    @classmethod
    def tearDown(self):

    @unittest.skipIf(TEARDOWN == False, "Test data should be remain in the harbor.")
    def test_ClearData(self):
        #1. Delete repository(RA) by user(UA);
        self.repo.delete_repoitory(TestProjects.repo_name, **TestProjects.USER_001_CLIENT)

        #2. Delete project(PA);
        self.project.delete_project(TestProjects.project_001_id, **TestProjects.USER_001_CLIENT)

        #3. Delete user(UA);
        self.user.delete_user(TestProjects.user_001_id, **TestProjects.ADMIN_CLIENT)

        #4. Delete label(LA).
        self.label.delete_label(TestProjects.label_id, **TestProjects.ADMIN_CLIENT)
        pass

    def testAddSysLabelToRepo(self):
        """
        Test case:
            Delete a repository
        Test step & Expectation:
            1. Create a new user(UA);
            2. Create a new private project(PA) by user(UA);
            3. Add user(UA) as a member of project(PA) with project-admin role;
            4. Get private project of uesr-001, uesr-001 can see only one private project which is project-001;
            5. Create a new repository(RA) and tag(TA) in project(PA) by user(UA);
            6. Create a new label(LA) in project(PA) by admin;;
            7. Add this system global label to repository(RA)/tag(TA);
        Tear down:
            1. Delete repository(RA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA);
            4. Delete label(LA).
        """
        admin_user = "admin"
        admin_pwd = "Harbor12345"
        url = CLIENT["endpoint"]
        user_001_password = "Aa123456"
        TestProjects.ADMIN_CLIENT=dict(endpoint = url, username = admin_user, password =  admin_pwd)

        #1. Create user-001
        TestProjects.user_001_id, user_001_name = self.user.create_user_success(user_password = user_001_password, **TestProjects.ADMIN_CLIENT)

        TestProjects.USER_001_CLIENT=dict(endpoint = url, username = user_001_name, password = user_001_password)

        #2. Create private project-001
        project_001_name, TestProjects.project_001_id = self.project.create_project(metadata = {"public": "false"}, **TestProjects.ADMIN_CLIENT)
 
        #3. Add user-001 as a member of project-001 with project-admin role
        self.project.add_project_members(TestProjects.project_001_id, TestProjects.user_001_id, **TestProjects.ADMIN_CLIENT)

        #4. Get private project of uesr-001, uesr-001 can see only one private project which is project-001
        self.project.projects_should_exist(dict(public=False), expected_count = 1,
            expected_project_id = TestProjects.project_001_id, **TestProjects.USER_001_CLIENT)

        #5. Create a new repository(RA) and tag(TA) in project(PA) by user(UA);
        TestProjects.repo_name, tag = self.repo.create_repository(project_001_name, registry = harbor_server, username = user_001_name, password = user_001_password)

        #6. Create a new label(LA) in project(PA) by admin;
        TestProjects.label_id, _ = self.label.create_label(**TestProjects.ADMIN_CLIENT)

        #7. Add this system global label to repository(RA)/tag(TA).
        self.repo.add_label_to_tag(TestProjects.repo_name, tag, int(TestProjects.label_id), **TestProjects.USER_001_CLIENT)

        pass

if __name__ == '__main__':
    unittest.main()

