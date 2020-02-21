from __future__ import absolute_import

import unittest

from testutils import harbor_server
from testutils import TEARDOWN
from testutils import ADMIN_CLIENT
from library.project import Project
from library.user import User
from library.repository import push_image_to_project
from library.repository import Repository

class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        project = Project()
        self.project= project

        user = User()
        self.user= user

        repo = Repository(api_type='repository')
        self.repo= repo

    @classmethod
    def tearDown(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete repository(RA) by admin;
        self.repo.delete_repoitory(TestProjects.project_alice_name, TestProjects.repo_name.split('/')[1], **ADMIN_CLIENT)

        #2. Delete project(Alice);
        self.project.delete_project(TestProjects.project_alice_id, **ADMIN_CLIENT)

        #3. Delete user Alice, Bob and Carol.
        self.user.delete_user(TestProjects.user_alice_id, **ADMIN_CLIENT)
        self.user.delete_user(TestProjects.user_bob_id, **ADMIN_CLIENT)
        self.user.delete_user(TestProjects.user_carol_id, **ADMIN_CLIENT)

    def testManageProjectMember(self):
        """
        Test case:
            Manage Project members
        Test step and expected result:
            1. Create user Alice, Bob, Carol;
            2. Create private project(Alice) by Alice, Add a repository to project(Alice) by Alice;
            3. Bob is not a member of project(Alice);
            4. Alice Add Bob as a guest member of project(Alice), Check Bob is a guest member of project(Alice);
            5. Update role of Bob to developer of project(Alice), Check Bob is developer member of project(Alice);
            6. Update role of Bob to admin member of project(Alice), Check Bob is admin member of project(Alice);
            7. Bob add Carol to project(Alice) as a guest member, Carol is a member of project(Alice) as a guest;
            8. Alice delete Bob from project(Alice),
               Bob is no longer a member of project(Alice) and Bob can see project(Alice),
               Carol is still a member of project(Alice) as a guest.
        Tear down:
            1. Delete repository(RA) by admin;
            2. Delete project(Alice);
            3. Delete user Alice, Bob and Carol.
        """
        url = ADMIN_CLIENT["endpoint"]
        user_alice_password = "Aa123456"
        user_bob_password = "Test1@34"
        user_carol_password = "Test1@34"

        #1.1 Create user Alice
        TestProjects.user_alice_id, user_alice_name = self.user.create_user(user_password = user_alice_password, **ADMIN_CLIENT)
        USER_ALICE_CLIENT=dict(endpoint = url, username = user_alice_name, password = user_alice_password)

        #1.2 Create user Bob
        TestProjects.user_bob_id, user_bob_name = self.user.create_user(user_password = user_bob_password, **ADMIN_CLIENT)
        USER_BOB_CLIENT=dict(endpoint = url, username = user_bob_name, password = user_bob_password)

        #1.3 Create user Carol
        TestProjects.user_carol_id, user_carol_name = self.user.create_user(user_password = user_carol_password, **ADMIN_CLIENT)

        #2.1 Create private project(PA) by Alice
        TestProjects.project_alice_id, TestProjects.project_alice_name = self.project.create_project(metadata = {"public": "false"}, **USER_ALICE_CLIENT)

        #2.2 Add a repository to project(PA) by Alice
        TestProjects.repo_name, _ = push_image_to_project(TestProjects.project_alice_name, harbor_server, user_alice_name, user_alice_password, "hello-world", "latest")

        #3. Bob is not a member of project(PA);
        self.project.check_project_member_not_exist(TestProjects.project_alice_id, user_bob_name, **USER_ALICE_CLIENT)

        #4.1 Alice Add Bob as a guest member of project(PA)
        member_id_bob = self.project.add_project_members(TestProjects.project_alice_id, TestProjects.user_bob_id, member_role_id = 3, **USER_ALICE_CLIENT)

        #4.2 Check Bob is a guest member of project(PA)
        self.project.check_project_members_exist(TestProjects.project_alice_id, user_bob_name, expected_member_role_id = 3, user_name = user_bob_name, user_password = user_bob_password, **USER_ALICE_CLIENT)

        #5.1 Update role of Bob to developer of project(PA)
        self.project.update_project_member_role(TestProjects.project_alice_id, member_id_bob, 2, **USER_ALICE_CLIENT)

        #5.2 Check Bob is developer member of project(PA)
        self.project.check_project_members_exist(TestProjects.project_alice_id, user_bob_name, expected_member_role_id = 2, user_name = user_bob_name, user_password = user_bob_password, **USER_ALICE_CLIENT)

        #6.1 Update role of Bob to admin member of project(PA)
        self.project.update_project_member_role(TestProjects.project_alice_id, member_id_bob, 1, **USER_ALICE_CLIENT)

        #6.2 Check Bob is admin member of project(PA)
        self.project.check_project_members_exist(TestProjects.project_alice_id, user_bob_name, expected_member_role_id = 1, user_name = user_bob_name, user_password = user_bob_password, **USER_ALICE_CLIENT)

        #7.1 Bob add Carol to project(PA) as a guest member.
        self.project.add_project_members(TestProjects.project_alice_id, TestProjects.user_carol_id, member_role_id = 3, **USER_BOB_CLIENT)

        #7.2 Carol is a member of project(PA) as a guest.
        self.project.check_project_members_exist(TestProjects.project_alice_id, user_carol_name, expected_member_role_id = 3, user_name = user_carol_name, user_password = user_carol_password, **USER_ALICE_CLIENT)

        #8.1 Alice delete Bob from project(PA).
        self.project.delete_project_member(TestProjects.project_alice_id, member_id_bob, **USER_ALICE_CLIENT)

        #8.2 Bob is no longer a member of project(PA) and Bob can see project(PA).
        self.project.check_project_member_not_exist(TestProjects.project_alice_id, user_bob_name, **USER_ALICE_CLIENT)

        #8.3 Carol is still a member of project(PA) as a guest.
        self.project.check_project_members_exist(TestProjects.project_alice_id, user_carol_name, expected_member_role_id = 3, user_name = user_carol_name, user_password = user_carol_password, **USER_ALICE_CLIENT)

if __name__ == '__main__':
    unittest.main()

