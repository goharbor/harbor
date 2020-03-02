from __future__ import absolute_import

import unittest

from testutils import harbor_server
from testutils import TEARDOWN
from testutils import ADMIN_CLIENT
from library.artifact import Artifact
from library.project import Project
from library.user import User
from library.repository import Repository
from library.repository import push_image_to_project
from library.label import Label

class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        self.project = Project()
        self.user = User()
        self.artifact = Artifact(api_type='artifact')
        self.repo = Repository(api_type='repository')
        self.label = Label()

    @classmethod
    def tearDown(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete repository(RA) by user(UA);
        self.repo.delete_repoitory(TestProjects.project_add_g_lbl_name, TestProjects.repo_name.split('/')[1], **TestProjects.USER_add_g_lbl_CLIENT)

        #2. Delete project(PA);
        self.project.delete_project(TestProjects.project_add_g_lbl_id, **TestProjects.USER_add_g_lbl_CLIENT)

        #3. Delete user(UA);
        self.user.delete_user(TestProjects.user_add_g_lbl_id, **ADMIN_CLIENT)

        #4. Delete label(LA).
        self.label.delete_label(TestProjects.label_id, **ADMIN_CLIENT)

    def testAddSysLabelToRepo(self):
        """
        Test case:
            Add Global Label To Tag
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new private project(PA) by user(UA);
            3. Add user(UA) as a member of project(PA) with project-admin role;
            4. Get private project of user(UA), user(UA) can see only one private project which is project(PA);
            5. Create a new repository(RA) and tag(TA) in project(PA) by user(UA);
            6. Create a new label(LA) in project(PA) by admin;;
            7. Add this system global label to repository(RA)/tag(TA);
        Tear down:
            1. Delete repository(RA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA);
            4. Delete label(LA).
        """
        url = ADMIN_CLIENT["endpoint"]
        user_001_password = "Aa123456"

        #1. Create user-001
        TestProjects.user_add_g_lbl_id, user_add_g_lbl_name = self.user.create_user(user_password = user_001_password, **ADMIN_CLIENT)

        TestProjects.USER_add_g_lbl_CLIENT=dict(endpoint = url, username = user_add_g_lbl_name, password = user_001_password)

        #2. Create private project-001
        TestProjects.project_add_g_lbl_id, TestProjects.project_add_g_lbl_name = self.project.create_project(metadata = {"public": "false"}, **ADMIN_CLIENT)

        #3. Add user-001 as a member of project-001 with project-admin role
        self.project.add_project_members(TestProjects.project_add_g_lbl_id, TestProjects.user_add_g_lbl_id, **ADMIN_CLIENT)

        #4. Get private project of user(UA), user(UA) can see only one private project which is project(PA);
        self.project.projects_should_exist(dict(public=False), expected_count = 1,
            expected_project_id = TestProjects.project_add_g_lbl_id, **TestProjects.USER_add_g_lbl_CLIENT)

        #5. Create a new repository(RA) and tag(TA) in project(PA) by user(UA);
        TestProjects.repo_name, tag = push_image_to_project(TestProjects.project_add_g_lbl_name, harbor_server, user_add_g_lbl_name, user_001_password, "hello-world", "latest")

        #6. Create a new label(LA) in project(PA) by admin;
        TestProjects.label_id, _ = self.label.create_label(**ADMIN_CLIENT)

        #7. Add this system global label to repository(RA)/tag(TA).
        self.artifact.add_label_to_reference(TestProjects.project_add_g_lbl_name, TestProjects.repo_name.split('/')[1], tag, int(TestProjects.label_id), **TestProjects.USER_add_g_lbl_CLIENT)

if __name__ == '__main__':
    unittest.main()
